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
	return loadAllIssues(dir, true)
}

// loadAllIssues reads issue directories without optionally deriving reverse
// relationships. Workspace loading merges several sources and rewires once
// after that merge, so skipping the per-source pass avoids duplicate work.
func loadAllIssues(dir string, wireRelationships bool) ([]Issue, []LoadProblem, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("reading %s: %w", dir, err)
	}

	var issues []Issue
	var problems []LoadProblem

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
	}

	if wireRelationships {
		RewireRelationships(issues)
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
	if id <= 0 {
		return Issue{}, fmt.Errorf("issue ID must be positive: %d", id)
	}
	issue, err := loadIssueMeta(dir, id)
	if err != nil {
		return Issue{}, err
	}
	name := strconv.Itoa(id)
	content, err := readOptional(filepath.Join(dir, name, "content.md"))
	if err != nil {
		return Issue{}, fmt.Errorf("reading %s/%s/content.md: %w", dir, name, err)
	}
	comments, err := readOptional(filepath.Join(dir, name, "comments.md"))
	if err != nil {
		return Issue{}, fmt.Errorf("reading %s/%s/comments.md: %w", dir, name, err)
	}
	issue.Content = content
	issue.Comments = ParseComments(comments)
	issue.SourceDir = dir
	return issue, nil
}

func loadIssueMeta(dir string, id int) (Issue, error) {
	if id <= 0 {
		return Issue{}, fmt.Errorf("issue ID must be positive: %d", id)
	}
	path := filepath.Join(dir, strconv.Itoa(id), "meta.toml")
	raw, err := os.ReadFile(path)
	if err != nil {
		return Issue{}, err
	}
	var m meta
	if err := toml.Unmarshal(raw, &m); err != nil {
		return Issue{}, fmt.Errorf("parsing %s: %w", path, err)
	}
	if validation := validateMeta(id, m); len(validation) != 0 {
		return Issue{}, validation[0]
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

func readOptional(path string) (string, error) {
	if _, err := os.Lstat(path); os.IsNotExist(err) {
		return "", nil
	} else if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ProjectRoot returns the parent directory of a .grapes/ path.
func ProjectRoot(issuesDir string) string {
	return filepath.Dir(issuesDir)
}

// maxIDInDirErr returns the highest positive numeric folder name in dir.
func maxIDInDirErr(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}
	max := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id, err := strconv.Atoi(e.Name())
		if err != nil || id <= 0 {
			continue
		}
		if id > max {
			max = id
		}
	}
	return max, nil
}

// maxIDInDir returns the highest numeric folder name in dir, or 0 if none.
// Kept for callers that only need a best-effort probe.
func maxIDInDir(dir string) int {
	max, _ := maxIDInDirErr(dir)
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
// project and all worktrees. Extra directories are glob patterns scanned in
// addition to Git-discovered checkouts.
func NextID(issuesDir string, extraDirs ...string) (int, error) {
	var mainRoot, grapesRel string
	if _, resolvedRoot, resolvedRel, err := resolveLayout(issuesDir); err == nil {
		// resolveLayout already found the repository root and relative store
		// path. Reuse them instead of resolving the Git layout a second time.
		mainRoot = resolvedRoot
		grapesRel = resolvedRel
	} else {
		mainRoot = FindMainProjectRoot(issuesDir)
		grapesRel = ".grapes"
	}
	mainGrapes := filepath.Join(mainRoot, filepath.FromSlash(grapesRel))
	lockRoot := mainGrapes
	if info, err := os.Stat(lockRoot); err != nil || !info.IsDir() {
		// A configured/nested store may be the only local store. Keep its
		// reservation lock beside the store rather than scanning/creating an
		// unrelated <project>/.grapes.
		lockRoot = issuesDir
	}

	// Acquire exclusive lock. The lock file is deliberately left on disk:
	// unlinking it while holding the lock lets a waiting process keep a lock on
	// the orphaned inode while the next process locks a freshly created file.
	lockPath := filepath.Join(lockRoot, ".lock")
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return 0, fmt.Errorf("opening lock file: %w", err)
	}
	defer lockFile.Close()
	if err := flockExclusive(lockFile.Fd()); err != nil {
		return 0, fmt.Errorf("acquiring lock: %w", err)
	}
	defer flockUnlock(lockFile.Fd())

	// Every copy has to be scanned, unlike a normal load: an ID already used
	// on some branch must not be handed out again. Include the local store
	// explicitly so nested/non-git stores work too, and deduplicate the main
	// checkout rather than scanning it twice.
	scanDirs := make(map[string]bool)
	addDir := func(dir string) {
		if dir != "" {
			scanDirs[filepath.Clean(dir)] = true
		}
	}
	addDir(issuesDir)
	if info, err := os.Stat(mainGrapes); err == nil && info.IsDir() {
		addDir(mainGrapes)
	} else if err != nil && !os.IsNotExist(err) {
		return 0, fmt.Errorf("scanning %s: %w", mainGrapes, err)
	}
	if checkouts, err := Checkouts(mainRoot, grapesRel); err == nil {
		for _, co := range checkouts {
			addDir(co.Dir)
		}
	} else if _, statErr := os.Stat(filepath.Join(mainRoot, ".git")); statErr == nil {
		return 0, fmt.Errorf("discovering worktrees: %w", err)
	}
	for _, dir := range FindExternalIssuesDirs(mainRoot, extraDirs...) {
		addDir(dir)
	}

	max := 0
	for dir := range scanDirs {
		m, err := maxIDInDirErr(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue // a removed worktree may race discovery
			}
			return 0, fmt.Errorf("scanning %s: %w", dir, err)
		}
		if m > max {
			max = m
		}
	}

	next := max + 1
	if err := os.MkdirAll(filepath.Join(issuesDir, strconv.Itoa(next)), 0o755); err != nil {
		return 0, fmt.Errorf("creating issue directory: %w", err)
	}
	return next, nil
}

// FindExternalIssuesDirs resolves glob patterns to issue directories and returns
// a map of unique display name → directory path. Relative patterns are resolved
// against projectRoot. The display name is the parent directory of each match.
// When names collide, a stable "#N" suffix preserves every configured source.
//
// These are stores outside this repository's history. Worktrees of this
// repository are discovered through git instead, by Checkouts.
func FindExternalIssuesDirs(projectRoot string, patterns ...string) map[string]string {
	result := make(map[string]string)
	seen := make(map[string]bool)
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
			match = filepath.Clean(match)
			if seen[match] {
				continue
			}
			seen[match] = true
			name := filepath.Base(filepath.Dir(match))
			base := name
			for n := 2; ; n++ {
				if _, exists := result[name]; !exists {
					break
				}
				name = fmt.Sprintf("%s#%d", base, n)
			}
			result[name] = match
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
			issues[i].Children = sortUniqueInts(kids)
		}
		if blocked, ok := blocksMap[issues[i].ID]; ok {
			issues[i].Blocks = sortUniqueInts(blocked)
		}
	}
}

func sortUniqueInts(values []int) []int {
	if len(values) < 2 {
		return values
	}
	sort.Ints(values)
	n := 1
	for _, value := range values[1:] {
		if value != values[n-1] {
			values[n] = value
			n++
		}
	}
	return values[:n]
}

// ParseComments parses comments.md using strict "### YYYY-MM-DD" headers.
//
// Text before the first header is kept as a leading comment with an empty Date
// rather than discarded, so hand-written preambles survive the edit round-trip.
func ParseComments(raw string) []Comment {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	// The final newline terminates the Markdown file; remove only that
	// delimiter so blank lines intentionally included in a body survive.
	raw = strings.TrimSuffix(raw, "\n")
	lines := strings.Split(raw, "\n")
	var comments []Comment
	current := &Comment{} // the dateless preamble, dropped below if empty
	finish := func(separator bool) {
		trim := 1 // the newline terminating the final body line
		if separator {
			trim++ // the blank line separating this comment from the next
		}
		for trim > 0 && strings.HasSuffix(current.Body, "\n") {
			current.Body = current.Body[:len(current.Body)-1]
			trim--
		}
		if current.Date != "" || current.Body != "" {
			comments = append(comments, *current)
		}
	}

	for _, line := range lines {
		slashes := 0
		for slashes < len(line) && line[slashes] == '\\' {
			slashes++
		}
		if slashes > 0 && commentHeader.MatchString(line[slashes:]) {
			// Remove one escape marker and keep any literal leading
			// backslashes in the comment body.
			current.Body += line[1:] + "\n"
			continue
		}
		if m := commentHeader.FindStringSubmatch(line); m != nil {
			finish(true)
			current = &Comment{Date: m[1]}
		} else {
			current.Body += line + "\n"
		}
	}
	finish(false)
	return comments
}
