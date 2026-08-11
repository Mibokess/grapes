package data

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
)

func TestParseComments(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []Comment
	}{
		{"empty", "   \n", nil},
		{
			"single",
			"### 2026-01-01\nhello\n",
			[]Comment{{Date: "2026-01-01", Body: "hello"}},
		},
		{
			"with time",
			"### 2026-01-01T09:30\nhello\n\n### 2026-01-02T10:00\nagain\n",
			[]Comment{{Date: "2026-01-01T09:30", Body: "hello"}, {Date: "2026-01-02T10:00", Body: "again"}},
		},
		{
			"legacy author header",
			"### alice — 2026-01-01\nhello\n",
			[]Comment{{Date: "2026-01-01", Body: "hello"}},
		},
		{
			// Text before the first header used to be dropped on the floor.
			"preamble kept as dateless comment",
			"free text\n\n### 2026-01-01\nhello\n",
			[]Comment{{Date: "", Body: "free text"}, {Date: "2026-01-01", Body: "hello"}},
		},
		{
			"no headers at all",
			"just notes\n",
			[]Comment{{Date: "", Body: "just notes"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseComments(tt.raw)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d comments, want %d: %+v", len(got), len(tt.want), got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("comment %d = %+v, want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// A malformed issue is skipped so the rest of the board still loads, but the
// skip has to be reported — otherwise the issue silently disappears.
func TestLoadAllIssues_ReportsMalformedIssues(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".grapes")
	writeIssueMeta(t, dir, 1, "title = \"good\"\nstatus = \"todo\"\npriority = \"low\"\n")
	writeIssueMeta(t, dir, 2, "this is not valid toml = = =\n")

	issues, problems, err := LoadAllIssues(dir)
	if err != nil {
		t.Fatalf("LoadAllIssues: %v", err)
	}
	if len(issues) != 1 || issues[0].ID != 1 {
		t.Fatalf("expected only issue #1 to load, got %+v", issues)
	}
	if len(problems) != 1 {
		t.Fatalf("expected 1 problem, got %d: %v", len(problems), problems)
	}
	if problems[0].ID != 2 {
		t.Errorf("problem is for #%d, want #2", problems[0].ID)
	}
	if problems[0].Error() == "" {
		t.Error("problem has no message")
	}
}

func TestFindIssuesDir_WalksUpFromSubdirectory(t *testing.T) {
	root := t.TempDir()
	grapes := filepath.Join(root, ".grapes")
	if err := os.MkdirAll(grapes, 0o755); err != nil {
		t.Fatal(err)
	}
	deep := filepath.Join(root, "internal", "data")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := FindIssuesDir(deep)
	if err != nil {
		t.Fatalf("FindIssuesDir: %v", err)
	}
	if got != grapes {
		t.Errorf("FindIssuesDir(%s) = %s, want %s", deep, got, grapes)
	}
}

func TestFindIssuesDir_FallsBackToSubdirectories(t *testing.T) {
	root := t.TempDir()
	grapes := filepath.Join(root, "project", ".grapes")
	if err := os.MkdirAll(grapes, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := FindIssuesDir(root)
	if err != nil {
		t.Fatalf("FindIssuesDir: %v", err)
	}
	if got != grapes {
		t.Errorf("FindIssuesDir(%s) = %s, want %s", root, got, grapes)
	}
}

func TestFindIssuesDir_NotFound(t *testing.T) {
	if _, err := FindIssuesDir(t.TempDir()); err == nil {
		t.Error("expected an error when no .grapes/ exists anywhere")
	}
}

// The lock file must survive NextID. Deleting it while holding the lock lets a
// waiting caller keep a lock on the unlinked inode while the next caller locks
// a freshly created file — two "exclusive" holders, one duplicated ID. The race
// needs an unlucky interleaving to show up, so the invariant is asserted
// directly rather than fished for with concurrency.
func TestNextID_LeavesLockFileInPlace(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".grapes")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	if _, err := NextID(dir); err != nil {
		t.Fatalf("NextID: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".lock")); err != nil {
		t.Errorf("lock file missing after NextID: %v", err)
	}
}

// NextID's whole job is handing out distinct IDs under concurrency.
func TestNextID_ConcurrentCallersGetDistinctIDs(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".grapes")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	const callers = 8
	ids := make([]int, callers)
	errs := make([]error, callers)
	var wg sync.WaitGroup
	for i := 0; i < callers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ids[i], errs[i] = NextID(dir)
		}(i)
	}
	wg.Wait()

	seen := make(map[int]bool, callers)
	for i, id := range ids {
		if errs[i] != nil {
			t.Fatalf("NextID: %v", errs[i])
		}
		if seen[id] {
			t.Errorf("ID %d handed out more than once (all: %v)", id, ids)
		}
		seen[id] = true
	}
	for i := 1; i <= callers; i++ {
		if !seen[i] {
			t.Errorf("expected IDs 1..%d, missing %d (all: %v)", callers, i, ids)
		}
	}
}

func writeIssueMeta(t *testing.T, grapesDir string, id int, metaTOML string) {
	t.Helper()
	issueDir := filepath.Join(grapesDir, strconv.Itoa(id))
	if err := os.MkdirAll(issueDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(issueDir, "meta.toml"), []byte(metaTOML), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadIssue_OptionalFilesAndMetadataValidation(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".grapes")
	writeIssueMeta(t, dir, 1, "title = \"valid\"\nstatus = \"todo\"\npriority = \"low\"\n")
	issue, err := LoadIssue(dir, 1)
	if err != nil {
		t.Fatalf("missing optional files should load: %v", err)
	}
	if issue.Content != "" || issue.Comments != nil {
		t.Fatalf("missing optional files should be empty: %+v", issue)
	}

	writeIssueMeta(t, dir, 2, "title = \"bad\"\nstatus = \"invalid\"\npriority = \"low\"\n")
	if _, err := LoadIssue(dir, 2); err == nil {
		t.Fatal("invalid metadata should be rejected")
	}
	if _, err := LoadIssue(dir, 0); err == nil {
		t.Fatal("nonpositive issue ID should be rejected")
	}
}

func TestLoadIssue_SecondaryReadFailureIsReported(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".grapes")
	writeIssueMeta(t, dir, 1, "title = \"valid\"\nstatus = \"todo\"\npriority = \"low\"\n")
	if err := os.Mkdir(filepath.Join(dir, "1", "content.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadIssue(dir, 1); err == nil {
		t.Fatal("unreadable secondary file should be reported")
	}
}

func TestFindExternalIssuesDirs_PreservesCollidingNames(t *testing.T) {
	root := t.TempDir()
	first := filepath.Join(root, "one", ".grapes")
	second := filepath.Join(root, "two", ".grapes")
	if err := os.MkdirAll(first, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(second, 0o755); err != nil {
		t.Fatal(err)
	}
	// Both stores have the same parent display name.
	dupA := filepath.Join(root, "a", "shared", ".grapes")
	dupB := filepath.Join(root, "b", "shared", ".grapes")
	if err := os.MkdirAll(dupA, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dupB, 0o755); err != nil {
		t.Fatal(err)
	}
	got := FindExternalIssuesDirs(root, filepath.Join(root, "*", "*", ".grapes"))
	if len(got) != 2 {
		t.Fatalf("got %d sources, want 2: %#v", len(got), got)
	}
	if got["shared"] == "" || got["shared#2"] == "" {
		t.Fatalf("colliding sources were not uniquely named: %#v", got)
	}
}
