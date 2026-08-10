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
		issue, err := loadIssueMeta(dir, id)
		if err != nil {
			problems = append(problems, LoadProblem{Dir: dir, ID: id, Err: err})
			continue
		}
		// Load content and comments
		issue.Content = readFileOr(filepath.Join(dir, e.Name(), "content.md"), "")
		issue.Comments = ParseComments(readFileOr(filepath.Join(dir, e.Name(), "comments.md"), ""))

		issue.SourceDir = dir
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

	// Collect every .grapes/ to scan: this repo's git worktrees (auto-discovered),
	// plus any configured glob patterns. Keyed by directory path to de-duplicate a
	// worktree that both git and a glob report.
	scanDirs := make(map[string]bool)
	for _, dir := range FindGitWorktreeGrapesDirs(mainRoot) {
		scanDirs[filepath.Clean(dir)] = true
	}
	for _, dir := range FindWorktreeIssuesDirs(mainRoot, extraDirs...) {
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

// FindWorktreeIssuesDirs resolves glob patterns to issue directories and returns
// a map of display name → directory path. Relative patterns are resolved against
// projectRoot. The display name is the parent directory of each matched path.
func FindWorktreeIssuesDirs(projectRoot string, patterns ...string) map[string]string {
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

// FindGitWorktreeGrapesDirs enumerates this repository's worktrees via
// git worktree list --porcelain and returns a map of display name → .grapes/
// directory for each worktree that has one. The display name is the base name of
// the worktree path. Returns an empty map when git is unavailable or mainRoot is
// not a git repository, so callers degrade to glob-based discovery.
func FindGitWorktreeGrapesDirs(mainRoot string) map[string]string {
	result := make(map[string]string)
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = mainRoot
	out, err := cmd.Output()
	if err != nil {
		return result
	}
	for _, line := range strings.Split(string(out), "\n") {
		path, ok := strings.CutPrefix(line, "worktree ")
		if !ok {
			continue
		}
		grapesDir := filepath.Join(path, ".grapes")
		if info, err := os.Stat(grapesDir); err != nil || !info.IsDir() {
			continue
		}
		result[filepath.Base(path)] = grapesDir
	}
	return result
}

// computeIssueMtime returns the most recent mtime across meta.toml, content.md,
// and comments.md for the given issue.
func computeIssueMtime(dir string, id int) time.Time {
	idStr := strconv.Itoa(id)
	files := []string{"meta.toml", "content.md", "comments.md"}
	var latest time.Time
	for _, f := range files {
		info, err := os.Stat(filepath.Join(dir, idStr, f))
		if err == nil && info.ModTime().After(latest) {
			latest = info.ModTime()
		}
	}
	return latest
}

// issueToSource creates an IssueSource from an Issue and its source metadata.
func issueToSource(iss Issue, name string, dir string, mtime time.Time) IssueSource {
	return IssueSource{
		Name:      name,
		Dir:       dir,
		Mtime:     mtime,
		Title:     iss.Title,
		Status:    iss.Status,
		Priority:  iss.Priority,
		Labels:    iss.Labels,
		Parent:    iss.Parent,
		BlockedBy: iss.BlockedBy,
		Created:   iss.Created,
		Updated:   iss.Updated,
		Content:   iss.Content,
		Comments:  iss.Comments,
	}
}

// LoadAllSources loads issues from main and all worktree .grapes/ directories,
// merging copies of the same issue ID into Sources. The active source is set to
// the one with the most recent file mtime. Extra worktree directories can be
// passed to scan beyond .claude/worktrees.
//
// Issues that fail to load are reported as problems; only a main directory that
// cannot be read at all is returned as an error.
func LoadAllSources(mainDir string, projectRoot string, extraDirs ...string) ([]Issue, []LoadProblem, error) {
	mainIssues, problems, err := LoadAllIssues(mainDir)
	if err != nil {
		return nil, nil, err
	}

	// Build map: issueID → *Issue with Sources populated
	issueMap := make(map[int]*Issue)
	for _, iss := range mainIssues {
		mtime := computeIssueMtime(mainDir, iss.ID)
		src := issueToSource(iss, "", mainDir, mtime)
		issCopy := iss
		issCopy.Sources = []IssueSource{src}
		issCopy.SourceDir = mainDir
		issueMap[iss.ID] = &issCopy
	}

	// Load all worktree issues: this repo's git worktrees (auto-discovered), plus
	// any configured glob patterns.
	worktrees := FindWorktreeIssuesDirs(projectRoot, extraDirs...)
	seenDirs := make(map[string]bool)
	for _, dir := range worktrees {
		seenDirs[filepath.Clean(dir)] = true
	}
	currentDir := filepath.Clean(mainDir)
	for name, dir := range FindGitWorktreeGrapesDirs(FindMainProjectRoot(mainDir)) {
		clean := filepath.Clean(dir)
		if clean == currentDir || seenDirs[clean] {
			continue // already loaded as main, or already found via glob
		}
		if _, ok := worktrees[name]; ok {
			continue // name already used (mirrors FindWorktreeIssuesDirs de-dup)
		}
		seenDirs[clean] = true
		worktrees[name] = dir
	}
	var wtNames []string
	for name := range worktrees {
		wtNames = append(wtNames, name)
	}
	sort.Strings(wtNames)

	for _, name := range wtNames {
		dir := worktrees[name]
		wtIssues, wtProblems, err := LoadAllIssues(dir)
		problems = append(problems, wtProblems...)
		if err != nil {
			problems = append(problems, LoadProblem{Dir: dir, Err: err})
			continue
		}
		for _, iss := range wtIssues {
			mtime := computeIssueMtime(dir, iss.ID)
			src := issueToSource(iss, name, dir, mtime)

			if existing, ok := issueMap[iss.ID]; ok {
				existing.Sources = append(existing.Sources, src)
			} else {
				issCopy := iss
				issCopy.Worktree = name
				issCopy.SourceDir = dir
				issCopy.Sources = []IssueSource{src}
				issueMap[iss.ID] = &issCopy
			}
		}
	}

	// For each issue, sort sources and pick the most recent as active
	var result []Issue
	for _, iss := range issueMap {
		// Sort sources: main first, then alphabetical by worktree name
		sort.SliceStable(iss.Sources, func(i, j int) bool {
			if iss.Sources[i].Name == "" {
				return true
			}
			if iss.Sources[j].Name == "" {
				return false
			}
			return iss.Sources[i].Name < iss.Sources[j].Name
		})

		// Find most recent mtime and switch to it
		bestIdx := 0
		for i, s := range iss.Sources {
			if s.Mtime.After(iss.Sources[bestIdx].Mtime) {
				bestIdx = i
			}
		}
		iss.SwitchSource(bestIdx)
		result = append(result, *iss)
	}

	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	RewireRelationships(result)
	return result, problems, nil
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
