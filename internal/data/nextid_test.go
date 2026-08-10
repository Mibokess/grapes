package data

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
)

// setupGitRepoWithWorktree creates a git repo at <base>/main and adds a linked
// worktree at <base>/wt (deliberately NOT under .claude/worktrees/*). It returns
// the main root and the worktree path. The test is skipped if git is missing.
func setupGitRepoWithWorktree(t *testing.T) (mainRoot, wtPath string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	base := t.TempDir()
	mainRoot = filepath.Join(base, "main")
	wtPath = filepath.Join(base, "wt")

	git := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	os.MkdirAll(mainRoot, 0o755)
	git(mainRoot, "init", "-q")
	git(mainRoot, "config", "user.email", "test@example.com")
	git(mainRoot, "config", "user.name", "Test")
	os.WriteFile(filepath.Join(mainRoot, "README"), []byte("x"), 0o644)
	git(mainRoot, "add", "README")
	git(mainRoot, "commit", "-q", "-m", "init")
	git(mainRoot, "worktree", "add", "-q", "-b", "feature", wtPath)
	return mainRoot, wtPath
}

func TestMaxIDInDir(t *testing.T) {
	dir := t.TempDir()

	// Empty dir → 0
	if got := maxIDInDir(dir); got != 0 {
		t.Errorf("empty dir: got %d, want 0", got)
	}

	// Create some numeric dirs
	os.Mkdir(filepath.Join(dir, "5"), 0o755)
	os.Mkdir(filepath.Join(dir, "10"), 0o755)
	os.Mkdir(filepath.Join(dir, "3"), 0o755)
	os.Mkdir(filepath.Join(dir, "tmp"), 0o755) // non-numeric, should be skipped

	if got := maxIDInDir(dir); got != 10 {
		t.Errorf("got %d, want 10", got)
	}
}

func TestNextID(t *testing.T) {
	// Set up a fake project with .grapes/
	root := t.TempDir()
	grapes := filepath.Join(root, ".grapes")
	os.Mkdir(grapes, 0o755)
	os.Mkdir(filepath.Join(grapes, "1"), 0o755)
	os.Mkdir(filepath.Join(grapes, "5"), 0o755)

	id, err := NextID(grapes)
	if err != nil {
		t.Fatalf("NextID: %v", err)
	}
	if id != 6 {
		t.Errorf("got %d, want 6", id)
	}

	// Directory should exist
	if _, err := os.Stat(filepath.Join(grapes, "6")); os.IsNotExist(err) {
		t.Error("directory .grapes/6 was not created")
	}

	// Calling again should increment
	id2, err := NextID(grapes)
	if err != nil {
		t.Fatalf("NextID second call: %v", err)
	}
	if id2 != 7 {
		t.Errorf("second call: got %d, want 7", id2)
	}
}

func TestNextIDWithWorktree(t *testing.T) {
	// Set up main project with .grapes/ and a fake worktree
	root := t.TempDir()
	grapes := filepath.Join(root, ".grapes")
	os.Mkdir(grapes, 0o755)
	os.Mkdir(filepath.Join(grapes, "1"), 0o755)
	os.Mkdir(filepath.Join(grapes, "5"), 0o755)

	// Create a worktree with a higher ID
	wtGrapes := filepath.Join(root, ".claude", "worktrees", "test-wt", ".grapes")
	os.MkdirAll(wtGrapes, 0o755)
	os.Mkdir(filepath.Join(wtGrapes, "8"), 0o755)

	// NextID from main should see the worktree's ID 8
	id, err := NextID(grapes, ".claude/worktrees/*/.grapes")
	if err != nil {
		t.Fatalf("NextID: %v", err)
	}
	if id != 9 {
		t.Errorf("got %d, want 9 (should see worktree ID 8)", id)
	}
}

func TestNextIDGitWorktree(t *testing.T) {
	mainRoot, wtPath := setupGitRepoWithWorktree(t)

	mainGrapes := filepath.Join(mainRoot, ".grapes")
	os.MkdirAll(filepath.Join(mainGrapes, "1"), 0o755)
	os.MkdirAll(filepath.Join(mainGrapes, "5"), 0o755)

	// Worktree lives outside .claude/worktrees/* and holds a higher ID.
	wtGrapes := filepath.Join(wtPath, ".grapes")
	os.MkdirAll(filepath.Join(wtGrapes, "8"), 0o755)

	// From main, with NO glob patterns, git discovery must still find ID 8.
	id, err := NextID(mainGrapes)
	if err != nil {
		t.Fatalf("NextID from main: %v", err)
	}
	if id != 9 {
		t.Errorf("from main: got %d, want 9 (should see git worktree ID 8)", id)
	}
	os.RemoveAll(filepath.Join(mainGrapes, "9")) // undo reservation for next check

	// From inside the worktree, it must see both main and worktree IDs.
	id, err = NextID(wtGrapes)
	if err != nil {
		t.Fatalf("NextID from worktree: %v", err)
	}
	if id != 9 {
		t.Errorf("from worktree: got %d, want 9", id)
	}
	if _, err := os.Stat(filepath.Join(wtGrapes, "9")); os.IsNotExist(err) {
		t.Error("new issue dir should be created in the local (worktree) .grapes/")
	}
}

func TestFindGitWorktreeGrapesDirs(t *testing.T) {
	mainRoot, wtPath := setupGitRepoWithWorktree(t)
	os.MkdirAll(filepath.Join(wtPath, ".grapes", "3"), 0o755)

	dirs := FindGitWorktreeGrapesDirs(mainRoot)
	if _, ok := dirs[filepath.Base(wtPath)]; !ok {
		t.Errorf("expected worktree %q in %v", filepath.Base(wtPath), dirs)
	}

	// Non-git directory degrades to an empty map, no error.
	if got := FindGitWorktreeGrapesDirs(t.TempDir()); len(got) != 0 {
		t.Errorf("non-git dir: got %v, want empty", got)
	}
}

func TestLoadAllSourcesGitWorktree(t *testing.T) {
	mainRoot, wtPath := setupGitRepoWithWorktree(t)

	writeIssue := func(grapes string, id int, title string) {
		t.Helper()
		dir := filepath.Join(grapes, strconv.Itoa(id))
		os.MkdirAll(dir, 0o755)
		os.WriteFile(filepath.Join(dir, "meta.toml"), []byte("title = \""+title+"\"\n"), 0o644)
	}

	mainGrapes := filepath.Join(mainRoot, ".grapes")
	writeIssue(mainGrapes, 1, "main issue")

	// Worktree issue outside .claude/worktrees/* with a distinct ID.
	wtGrapes := filepath.Join(wtPath, ".grapes")
	writeIssue(wtGrapes, 7, "worktree issue")

	// From main, with NO glob patterns, the worktree issue must be discovered.
	issues, problems, err := LoadAllSources(mainGrapes, mainRoot)
	if err != nil {
		t.Fatalf("LoadAllSources: %v", err)
	}
	if len(problems) > 0 {
		t.Fatalf("unexpected load problems: %v", problems)
	}
	byID := make(map[int]Issue)
	for _, iss := range issues {
		byID[iss.ID] = iss
	}
	if _, ok := byID[1]; !ok {
		t.Error("missing main issue 1")
	}
	wt, ok := byID[7]
	if !ok {
		t.Fatalf("missing git-worktree issue 7; got IDs %v", byID)
	}
	if wt.Title != "worktree issue" {
		t.Errorf("issue 7 title: got %q, want %q", wt.Title, "worktree issue")
	}
}

func TestFindWorktreeIssuesDirsGlobPatterns(t *testing.T) {
	root := t.TempDir()

	// Create default worktree location
	defaultWT := filepath.Join(root, ".claude", "worktrees", "default-wt", ".grapes")
	os.MkdirAll(defaultWT, 0o755)

	// Create extra worktree location with absolute path
	extraDir := t.TempDir()
	extraWT := filepath.Join(extraDir, "custom-wt", ".grapes")
	os.MkdirAll(extraWT, 0o755)

	// Without any patterns, nothing is found
	result := FindWorktreeIssuesDirs(root)
	if len(result) != 0 {
		t.Errorf("without patterns: got %d, want 0", len(result))
	}

	// Relative glob pattern
	result = FindWorktreeIssuesDirs(root, ".claude/worktrees/*/.grapes")
	if len(result) != 1 {
		t.Errorf("relative glob: got %d, want 1", len(result))
	}
	if _, ok := result["default-wt"]; !ok {
		t.Error("relative glob: missing default-wt")
	}

	// Absolute glob pattern
	absPattern := filepath.Join(extraDir, "*", ".grapes")
	result = FindWorktreeIssuesDirs(root, absPattern)
	if len(result) != 1 {
		t.Errorf("absolute glob: got %d, want 1", len(result))
	}
	if _, ok := result["custom-wt"]; !ok {
		t.Error("absolute glob: missing custom-wt")
	}

	// Multiple patterns
	result = FindWorktreeIssuesDirs(root, ".claude/worktrees/*/.grapes", absPattern)
	if len(result) != 2 {
		t.Errorf("multiple patterns: got %d, want 2", len(result))
	}
}

func TestFindWorktreeIssuesDirsCustomName(t *testing.T) {
	root := t.TempDir()

	// Create a non-.grapes issue dir
	os.MkdirAll(filepath.Join(root, "worktrees", "proj1", ".potatoes"), 0o755)

	result := FindWorktreeIssuesDirs(root, "worktrees/*/.potatoes")
	if len(result) != 1 {
		t.Errorf("custom name: got %d, want 1", len(result))
	}
	if _, ok := result["proj1"]; !ok {
		t.Error("custom name: missing proj1")
	}
}

func TestFindWorktreeIssuesDirsExactPath(t *testing.T) {
	root := t.TempDir()

	// Create an exact path (no glob)
	os.MkdirAll(filepath.Join(root, "other", ".grapes"), 0o755)

	result := FindWorktreeIssuesDirs(root, "other/.grapes")
	if len(result) != 1 {
		t.Errorf("exact path: got %d, want 1", len(result))
	}
	if _, ok := result["other"]; !ok {
		t.Error("exact path: missing 'other'")
	}
}

func TestNextIDWithGlobPattern(t *testing.T) {
	root := t.TempDir()
	grapes := filepath.Join(root, ".grapes")
	os.Mkdir(grapes, 0o755)
	os.Mkdir(filepath.Join(grapes, "1"), 0o755)

	// Create extra worktree location with a higher ID
	extraDir := t.TempDir()
	extraWT := filepath.Join(extraDir, "ext-wt", ".grapes")
	os.MkdirAll(extraWT, 0o755)
	os.Mkdir(filepath.Join(extraWT, "10"), 0o755)

	// NextID should see the extra dir's ID 10 via glob
	absPattern := filepath.Join(extraDir, "*", ".grapes")
	id, err := NextID(grapes, absPattern)
	if err != nil {
		t.Fatalf("NextID: %v", err)
	}
	if id != 11 {
		t.Errorf("got %d, want 11 (should see extra dir ID 10)", id)
	}
}
