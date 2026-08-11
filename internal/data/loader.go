package data

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	toml "github.com/pelletier/go-toml/v2"
)

// commentHeader matches "### YYYY-MM-DD" or "### YYYY-MM-DDTHH:MM" headers,
// as well as legacy "### author — YYYY-MM-DD" headers (em-dash only).
var commentHeader = regexp.MustCompile(`^### (?:\S+ \x{2014} )?(\d{4}-\d{2}-\d{2}(?:T\d{2}:\d{2})?)$`)

// meta is the on-disk TOML structure.
type meta struct {
	Title     string    `toml:"title"`
	Status    string    `toml:"status"`
	Priority  string    `toml:"priority"`
	Labels    []string  `toml:"labels"`
	Parent    *int      `toml:"parent,omitempty"`
	BlockedBy []int     `toml:"blocked_by,omitempty"`
	Comments  []Comment `toml:"comments,omitempty"`
	Created   time.Time `toml:"created"`
	Updated   time.Time `toml:"updated"`
}

// maxSearchDepth is how many directory levels deep to search for .grapes/.
const maxSearchDepth = 10

// LoadProblem records an issue directory that could not be loaded. Skipping a
// malformed issue keeps the rest of the board usable, but the skip is reported
// rather than swallowed, so the issue does not simply vanish from the TUI.
type LoadProblem struct {
	Dir string
	ID  int // 0 when the whole directory failed to load
	Err error
}

func (p LoadProblem) Error() string {
	if p.ID == 0 {
		return fmt.Sprintf("%s: %v", p.Dir, p.Err)
	}
	return fmt.Sprintf("#%d: %v", p.ID, p.Err)
}

// FindIssuesDir locates the .grapes/ directory to use for startDir.
//
// It walks up from startDir first, the way git finds .git, so grapes works from
// any subdirectory of a project. Only when no ancestor holds one does it scan
// downward (up to maxSearchDepth), which still allows running grapes from a
// directory that merely contains the project.
func FindIssuesDir(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	if found, ok := findIssuesDirUp(dir); ok {
		return found, nil
	}
	found, err := findIssuesDirDown(dir)
	if err != nil {
		return "", fmt.Errorf("searching %s for .grapes/: %w", startDir, err)
	}
	if found != "" {
		return found, nil
	}
	return "", fmt.Errorf(".grapes/ directory not found in %s or its parents", startDir)
}

// findIssuesDirUp walks from dir to the filesystem root looking for .grapes/.
func findIssuesDirUp(dir string) (string, bool) {
	for {
		candidate := filepath.Join(dir, ".grapes")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// findIssuesDirDown scans subdirectories of dir for a .grapes/ directory.
// Returns "" when the scan completes without finding one.
func findIssuesDirDown(dir string) (string, error) {
	var found string
	baseDepth := strings.Count(dir, string(filepath.Separator))
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if path == dir {
				return err // the starting directory itself is unreadable
			}
			return filepath.SkipDir // an unreadable subtree is not fatal
		}
		if found != "" {
			return filepath.SkipAll
		}
		if d.IsDir() && d.Name() == ".grapes" {
			found = path
			return filepath.SkipAll
		}
		if d.IsDir() && d.Name() != "." {
			depth := strings.Count(path, string(filepath.Separator)) - baseDepth
			if depth >= maxSearchDepth {
				return filepath.SkipDir
			}
			if strings.HasPrefix(d.Name(), ".") || d.Name() == "node_modules" {
				return filepath.SkipDir
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return found, nil
}

// LoadAllIssues scans the .grapes/ directory and returns all issues with
// parent→children relationships built. Content and comments are loaded too.
// Issue directories that fail to load are reported as problems and left out of
// the returned slice.
func LoadAllIssues(dir string) ([]Issue, []LoadProblem, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("reading %s: %w", dir, err)
	}

	var issues []Issue
	var problems []LoadProblem
	childrenMap := make(map[int][]int) // parent ID → child IDs
	blocksMap := make(map[int][]int)   // blocked ID → IDs it blocks

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id, err := strconv.Atoi(e.Name())
		if err != nil {
			continue // skip non-numeric directories
		}
		issue, err := LoadIssue(dir, id)
		if err != nil {
			problems = append(problems, LoadProblem{Dir: dir, ID: id, Err: err})
			continue
		}
		issues = append(issues, issue)
		if issue.Parent != nil {
			childrenMap[*issue.Parent] = append(childrenMap[*issue.Parent], id)
		}
		for _, blockerID := range issue.BlockedBy {
			blocksMap[blockerID] = append(blocksMap[blockerID], id)
		}
	}

	// Wire up children and blocks
	for i := range issues {
		if kids, ok := childrenMap[issues[i].ID]; ok {
			sort.Ints(kids)
			issues[i].Children = kids
		}
		if blocked, ok := blocksMap[issues[i].ID]; ok {
			sort.Ints(blocked)
			issues[i].Blocks = blocked
		}
	}

	sort.Slice(issues, func(i, j int) bool {
		return issues[i].ID < issues[j].ID
	})

	return issues, problems, nil
}

// LoadIssue reads one issue — metadata, content, and comments — from an issue
// directory. Loading a named issue is what lets a worktree contribute only the
// issues git says it changed, instead of its whole copy of the store.
func LoadIssue(dir string, id int) (Issue, error) {
	issue, err := loadIssueMeta(dir, id)
	if err != nil {
		return Issue{}, err
	}
	name := strconv.Itoa(id)
	issue.Content = readFileOr(filepath.Join(dir, name, "content.md"), "")
	issue.Comments = ParseComments(readFileOr(filepath.Join(dir, name, "comments.md"), ""))
	issue.SourceDir = dir
	return issue, nil
}

func loadIssueMeta(dir string, id int) (Issue, error) {
	path := filepath.Join(dir, strconv.Itoa(id), "meta.toml")
	raw, err := os.ReadFile(path)
	if err != nil {
		return Issue{}, err
	}
	var m meta
	if err := toml.Unmarshal(raw, &m); err != nil {
		return Issue{}, fmt.Errorf("parsing %s: %w", path, err)
	}

	return Issue{
		ID:        id,
		Title:     m.Title,
		Status:    Status(m.Status),
		Priority:  Priority(m.Priority),
		Labels:    m.Labels,
		Parent:    m.Parent,
		BlockedBy: m.BlockedBy,
		Created:   m.Created,
		Updated:   m.Updated,
	}, nil
}

func readFileOr(path, fallback string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return fallback
	}
	return string(data)
}

// ProjectRoot returns the parent directory of a .grapes/ path.
func ProjectRoot(issuesDir string) string {
	return filepath.Dir(issuesDir)
}

// maxIDInDir returns the highest numeric folder name in dir, or 0 if none.
func maxIDInDir(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	max := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		if id > max {
			max = id
		}
	}
	return max
}

// FindMainProjectRoot returns the main project root, even when issuesDir is
// inside a worktree. It uses git rev-parse --git-common-dir to find the shared
// .git directory, then takes its parent. Falls back to ProjectRoot if git fails.
func FindMainProjectRoot(issuesDir string) string {
	projectRoot := ProjectRoot(issuesDir)
	cmd := exec.Command("git", "rev-parse", "--git-common-dir")
	cmd.Dir = projectRoot
	out, err := cmd.Output()
	if err != nil {
		return projectRoot
	}
	gitCommon := strings.TrimSpace(string(out))
	if !filepath.IsAbs(gitCommon) {
		gitCommon = filepath.Join(projectRoot, gitCommon)
	}
	return filepath.Dir(filepath.Clean(gitCommon))
}

// NextID atomically reserves the next available issue ID across the main
// project and all worktrees. It acquires an exclusive lock, scans all .grapes/
// directories, creates the new issue directory locally, then releases the lock.
// Extra worktree directories can be passed to scan beyond .claude/worktrees.
func NextID(issuesDir string, extraDirs ...string) (int, error) {
	mainRoot := FindMainProjectRoot(issuesDir)
	mainGrapes := filepath.Join(mainRoot, ".grapes")

	// Acquire exclusive lock. The lock file is deliberately left on disk:
	// unlinking it while holding the lock lets a waiting process keep a lock on
	// the orphaned inode while the next process locks a freshly created file,
	// so two callers would scan for the max ID at the same time.
	lockPath := filepath.Join(mainGrapes, ".lock")
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return 0, fmt.Errorf("opening lock file: %w", err)
	}
	defer lockFile.Close()

	if err := flockExclusive(lockFile.Fd()); err != nil {
		return 0, fmt.Errorf("acquiring lock: %w", err)
	}
	defer flockUnlock(lockFile.Fd())

	// Every copy has to be scanned, unlike a normal load: an ID already used on
	// some branch must not be handed out again, even when that branch has done
	// nothing this workspace considers interesting. This runs under the lock and
	// off the render path, so the full scan is affordable here.
	scanDirs := make(map[string]bool)
	if checkouts, err := Checkouts(mainRoot, ".grapes"); err == nil {
		for _, co := range checkouts {
			scanDirs[filepath.Clean(co.Dir)] = true
		}
	}
	for _, dir := range FindExternalIssuesDirs(mainRoot, extraDirs...) {
		scanDirs[filepath.Clean(dir)] = true
	}

	// Find max ID across all sources
	max := maxIDInDir(mainGrapes)
	for dir := range scanDirs {
		if m := maxIDInDir(dir); m > max {
			max = m
		}
	}

	next := max + 1

	// Create the directory in the local .grapes/
	if err := os.MkdirAll(filepath.Join(issuesDir, strconv.Itoa(next)), 0o755); err != nil {
		return 0, fmt.Errorf("creating issue directory: %w", err)
	}

	return next, nil
}

// FindExternalIssuesDirs resolves glob patterns to issue directories and returns
// a map of display name → directory path. Relative patterns are resolved against
// projectRoot. The display name is the parent directory of each matched path.
//
// These are stores outside this repository's history. Worktrees of this
// repository are discovered through git instead, by Checkouts.
func FindExternalIssuesDirs(projectRoot string, patterns ...string) map[string]string {
	result := make(map[string]string)
	for _, pattern := range patterns {
		if !filepath.IsAbs(pattern) {
			pattern = filepath.Join(projectRoot, pattern)
		}
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil || !info.IsDir() {
				continue
			}
			name := filepath.Base(filepath.Dir(match))
			if _, exists := result[name]; !exists {
				result[name] = match
			}
		}
	}
	return result
}

// RewireRelationships rebuilds Children and Blocks slices from all issues'
// Parent and BlockedBy fields. Use after merging issues from multiple sources.
func RewireRelationships(issues []Issue) {
	childrenMap := make(map[int][]int)
	blocksMap := make(map[int][]int)
	for _, iss := range issues {
		if iss.Parent != nil {
			childrenMap[*iss.Parent] = append(childrenMap[*iss.Parent], iss.ID)
		}
		for _, blockerID := range iss.BlockedBy {
			blocksMap[blockerID] = append(blocksMap[blockerID], iss.ID)
		}
	}
	for i := range issues {
		issues[i].Children = nil
		issues[i].Blocks = nil
		if kids, ok := childrenMap[issues[i].ID]; ok {
			sort.Ints(kids)
			issues[i].Children = kids
		}
		if blocked, ok := blocksMap[issues[i].ID]; ok {
			sort.Ints(blocked)
			issues[i].Blocks = blocked
		}
	}
}

// ParseComments parses comments.md using strict "### YYYY-MM-DD" headers.
//
// Text before the first header is kept as a leading comment with an empty Date
// rather than discarded, so hand-written preambles survive the edit round-trip.
func ParseComments(raw string) []Comment {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	lines := strings.Split(raw, "\n")
	var comments []Comment
	current := &Comment{} // the dateless preamble, dropped below if empty

	for _, line := range lines {
		if m := commentHeader.FindStringSubmatch(line); m != nil {
			// Save previous comment
			current.Body = strings.TrimSpace(current.Body)
			if current.Date != "" || current.Body != "" {
				comments = append(comments, *current)
			}
			current = &Comment{Date: m[1]}
		} else {
			current.Body += line + "\n"
		}
	}
	// Don't forget the last comment
	current.Body = strings.TrimSpace(current.Body)
	if current.Date != "" || current.Body != "" {
		comments = append(comments, *current)
	}

	return comments
}
