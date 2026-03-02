// Package testutil provides golden file testing helpers and shared test fixtures
// for the TUI components.
//
// Usage:
//
//	go test ./internal/tui/... -update   # regenerate golden files
//	go test ./internal/tui/...           # compare against golden files
package testutil

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/Mibokess/grapes/internal/data"
)

var update = flag.Bool("update", false, "update golden files")

// ansiRE matches ANSI escape sequences (colors, cursor movement, etc.).
var ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// StripANSI removes ANSI escape sequences from a string.
func StripANSI(s string) string {
	return ansiRE.ReplaceAllString(s, "")
}

// RequireGolden compares the rendered view against a golden file.
// Golden files are stored with ANSI codes stripped for readability and
// robustness (immune to color-code changes across lipgloss versions).
//
// With -update, it writes the current output to the golden file.
// Without -update, it compares and fails on mismatch with a line diff.
func RequireGolden(t *testing.T, got string) {
	t.Helper()

	name := strings.ReplaceAll(t.Name(), "/", "__")
	goldenFile := filepath.Join("testdata", name+".golden")
	gotClean := StripANSI(got)

	if *update {
		if err := os.MkdirAll(filepath.Dir(goldenFile), 0o755); err != nil {
			t.Fatalf("creating testdata dir: %v", err)
		}
		if err := os.WriteFile(goldenFile, []byte(gotClean), 0o644); err != nil {
			t.Fatalf("writing golden file: %v", err)
		}
		t.Logf("updated golden file: %s", goldenFile)
		return
	}

	expected, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("golden file not found: %s\nRun with -update to create it:\n  go test -run %s -update", goldenFile, t.Name())
	}

	if gotClean != string(expected) {
		gotLines := strings.Split(gotClean, "\n")
		expLines := strings.Split(string(expected), "\n")
		maxLines := len(gotLines)
		if len(expLines) > maxLines {
			maxLines = len(expLines)
		}

		var diff strings.Builder
		for i := 0; i < maxLines; i++ {
			gl, el := "", ""
			if i < len(gotLines) {
				gl = gotLines[i]
			}
			if i < len(expLines) {
				el = expLines[i]
			}
			if gl != el {
				diff.WriteString("--- exp line " + itoa(i+1) + ": " + el + "\n")
				diff.WriteString("+++ got line " + itoa(i+1) + ": " + gl + "\n")
			}
		}

		if len(gotLines) != len(expLines) {
			diff.WriteString("--- expected " + itoa(len(expLines)) + " lines, got " + itoa(len(gotLines)) + "\n")
		}

		t.Errorf("golden file mismatch: %s\nRun with -update to accept:\n  go test -run %s -update\n\nDiff:\n%s", goldenFile, t.Name(), diff.String())
	}
}

func itoa(n int) string {
	if n < 0 {
		return "-" + itoa(-n)
	}
	if n < 10 {
		return string(rune('0' + n))
	}
	return itoa(n/10) + string(rune('0'+n%10))
}

// SampleIssues returns a fixed set of test issues covering various statuses,
// priorities, labels, and relationships. Timestamps are fixed for determinism.
func SampleIssues() []data.Issue {
	t1 := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC)
	t3 := time.Date(2025, 2, 10, 9, 0, 0, 0, time.UTC)
	t4 := time.Date(2025, 3, 1, 14, 0, 0, 0, time.UTC)
	t5 := time.Date(2025, 3, 5, 8, 0, 0, 0, time.UTC)
	t6 := time.Date(2025, 3, 10, 16, 0, 0, 0, time.UTC)
	t7 := time.Date(2025, 3, 12, 11, 0, 0, 0, time.UTC)

	parentID := 1
	issues := []data.Issue{
		{
			ID:       1,
			Title:    "Implement user authentication",
			Status:   data.StatusInProgress,
			Priority: data.PriorityHigh,
			Labels:   []string{"backend", "security"},
			Children: []int{5, 6},
			Created:  t1,
			Updated:  t4,
			Content:  "Add JWT-based authentication with refresh tokens.\n\n## Requirements\n\n- Login endpoint\n- Token refresh\n- Logout",
			Comments: []data.Comment{
				{Date: "2025-02-01T12:00", Body: "Investigating OAuth2 vs JWT. JWT is simpler for this use case."},
				{Date: "2025-02-02T09:30", Body: "Started with JWT implementation. Login endpoint scaffolded."},
			},
		},
		{
			ID:       2,
			Title:    "Fix dashboard loading performance",
			Status:   data.StatusTodo,
			Priority: data.PriorityUrgent,
			Labels:   []string{"performance", "frontend"},
			Created:  t2,
			Updated:  t2,
			Content:  "Dashboard takes 5+ seconds to load. Investigate and fix.",
		},
		{
			ID:       3,
			Title:    "Add dark mode support",
			Status:   data.StatusBacklog,
			Priority: data.PriorityMedium,
			Labels:   []string{"frontend", "design"},
			Created:  t3,
		},
		{
			ID:       4,
			Title:    "Write API documentation",
			Status:   data.StatusDone,
			Priority: data.PriorityLow,
			Labels:   []string{"docs"},
			Created:  t4,
			Updated:  t5,
			Content:  "OpenAPI spec for all endpoints.",
		},
		{
			ID:       5,
			Title:    "Implement login endpoint",
			Status:   data.StatusDone,
			Priority: data.PriorityHigh,
			Labels:   []string{"backend"},
			Parent:   &parentID,
			Created:  t5,
			Updated:  t6,
		},
		{
			ID:       6,
			Title:    "Implement token refresh",
			Status:   data.StatusInProgress,
			Priority: data.PriorityHigh,
			Labels:   []string{"backend"},
			Parent:   &parentID,
			Created:  t5,
			Updated:  t7,
		},
		{
			ID:       7,
			Title:    "Remove deprecated v1 API endpoints",
			Status:   data.StatusCancelled,
			Priority: data.PriorityLow,
			Created:  t1,
			Content:  "No longer needed after migration deadline passed.",
		},
	}
	return issues
}
