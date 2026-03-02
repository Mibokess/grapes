package data

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// meta is the on-disk YAML structure.
type meta struct {
	Title     string    `yaml:"title"`
	Status    string    `yaml:"status"`
	Priority  string    `yaml:"priority"`
	Labels    []string  `yaml:"labels"`
	Parent    *int      `yaml:"parent,omitempty"`
	BlockedBy []int     `yaml:"blocked_by,omitempty"`
	Comments  []Comment `yaml:"comments,omitempty"`
	Created   string    `yaml:"created"`
	Updated   string    `yaml:"updated"`
}

// FindIssuesDir walks up from startDir looking for a .grapes/ directory.
func FindIssuesDir(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, ".grapes")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf(".grapes/ directory not found (searched up from %s)", startDir)
		}
		dir = parent
	}
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
	path := filepath.Join(dir, strconv.Itoa(id), "meta.yaml")
	raw, err := os.ReadFile(path)
	if err != nil {
		return Issue{}, err
	}
	var m meta
	if err := yaml.Unmarshal(raw, &m); err != nil {
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
