package data

import (
	"fmt"
	"os"
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
	Created   string    `toml:"created"`
	Updated   string    `toml:"updated"`
}

// maxSearchDepth is how many directory levels deep to search for .grapes/.
const maxSearchDepth = 10

// FindIssuesDir searches startDir and subdirectories (up to maxSearchDepth) for a .grapes/ directory.
func FindIssuesDir(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	var found string
	baseDepth := strings.Count(dir, string(filepath.Separator))
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || found != "" {
			return filepath.SkipDir
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
	if found != "" {
		return found, nil
	}
	return "", fmt.Errorf(".grapes/ directory not found in %s", startDir)
}

// LoadAllIssues scans the .grapes/ directory and returns all issues with
// parent→children relationships built. Content and comments are loaded too.
func LoadAllIssues(dir string) ([]Issue, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", dir, err)
	}

	var issues []Issue
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
			continue // skip malformed issues gracefully
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

	return issues, nil
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

	created := parseDate(m.Created)
	updated := parseDate(m.Updated)

	return Issue{
		ID:        id,
		Title:     m.Title,
		Status:    Status(m.Status),
		Priority:  Priority(m.Priority),
		Labels:    m.Labels,
		Parent:    m.Parent,
		BlockedBy: m.BlockedBy,
		Created:   created,
		Updated:   updated,
	}, nil
}

func parseDate(s string) time.Time {
	if t, err := time.Parse("2006-01-02T15:04", s); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t
	}
	return time.Time{}
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

// FindWorktreeIssuesDirs scans .claude/worktrees/*/.grapes/ relative to
// projectRoot and returns a map of worktree name → .grapes/ directory path.
func FindWorktreeIssuesDirs(projectRoot string) map[string]string {
	worktreesDir := filepath.Join(projectRoot, ".claude", "worktrees")
	entries, err := os.ReadDir(worktreesDir)
	if err != nil {
		return nil
	}
	result := make(map[string]string)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		grapesDir := filepath.Join(worktreesDir, e.Name(), ".grapes")
		if info, err := os.Stat(grapesDir); err == nil && info.IsDir() {
			result[e.Name()] = grapesDir
		}
	}
	return result
}

// LoadWorktreeIssues loads issues from all worktree .grapes/ directories,
// returning only issues whose IDs don't exist in mainIDs.
// Deprecated: Use LoadAllSources instead for multi-source tracking.
func LoadWorktreeIssues(projectRoot string, mainIDs map[int]bool) ([]Issue, error) {
	worktrees := FindWorktreeIssuesDirs(projectRoot)
	var all []Issue
	seen := make(map[int]bool) // dedup across worktrees
	for name, dir := range worktrees {
		issues, err := LoadAllIssues(dir)
		if err != nil {
			continue
		}
		for i := range issues {
			if mainIDs[issues[i].ID] || seen[issues[i].ID] {
				continue
			}
			issues[i].Worktree = name
			seen[issues[i].ID] = true
			all = append(all, issues[i])
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })
	return all, nil
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
// the one with the most recent file mtime.
func LoadAllSources(mainDir string, projectRoot string) ([]Issue, error) {
	mainIssues, err := LoadAllIssues(mainDir)
	if err != nil {
		return nil, err
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

	// Load all worktree issues
	worktrees := FindWorktreeIssuesDirs(projectRoot)
	var wtNames []string
	for name := range worktrees {
		wtNames = append(wtNames, name)
	}
	sort.Strings(wtNames)

	for _, name := range wtNames {
		dir := worktrees[name]
		wtIssues, err := LoadAllIssues(dir)
		if err != nil {
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
	return result, nil
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
func ParseComments(raw string) []Comment {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	lines := strings.Split(raw, "\n")
	var comments []Comment
	var current *Comment

	for _, line := range lines {
		if m := commentHeader.FindStringSubmatch(line); m != nil {
			// Save previous comment
			if current != nil {
				current.Body = strings.TrimSpace(current.Body)
				comments = append(comments, *current)
			}
			current = &Comment{
				Date: m[1],
			}
		} else if current != nil {
			current.Body += line + "\n"
		}
	}
	// Don't forget the last comment
	if current != nil {
		current.Body = strings.TrimSpace(current.Body)
		comments = append(comments, *current)
	}

	return comments
}
