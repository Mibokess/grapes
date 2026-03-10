package data

import (
	"os"
	"path/filepath"
	"testing"
)

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
	id, err := NextID(grapes, ".claude/worktrees")
	if err != nil {
		t.Fatalf("NextID: %v", err)
	}
	if id != 9 {
		t.Errorf("got %d, want 9 (should see worktree ID 8)", id)
	}
}

func TestFindWorktreeIssuesDirsExtraDirs(t *testing.T) {
	root := t.TempDir()

	// Create default worktree location
	defaultWT := filepath.Join(root, ".claude", "worktrees", "default-wt", ".grapes")
	os.MkdirAll(defaultWT, 0o755)

	// Create extra worktree location (absolute path)
	extraDir := t.TempDir()
	extraWT := filepath.Join(extraDir, "custom-wt", ".grapes")
	os.MkdirAll(extraWT, 0o755)

	// Without any dirs, nothing is found
	result := FindWorktreeIssuesDirs(root)
	if len(result) != 0 {
		t.Errorf("without dirs: got %d worktrees, want 0", len(result))
	}

	// With default dir, only default is found
	result = FindWorktreeIssuesDirs(root, ".claude/worktrees")
	if len(result) != 1 {
		t.Errorf("with default dir: got %d worktrees, want 1", len(result))
	}
	if _, ok := result["default-wt"]; !ok {
		t.Error("with default dir: missing default-wt")
	}

	// With both dirs, both are found
	result = FindWorktreeIssuesDirs(root, ".claude/worktrees", extraDir)
	if len(result) != 2 {
		t.Errorf("with both dirs: got %d worktrees, want 2", len(result))
	}
	if _, ok := result["default-wt"]; !ok {
		t.Error("with both dirs: missing default-wt")
	}
	if _, ok := result["custom-wt"]; !ok {
		t.Error("with both dirs: missing custom-wt")
	}
}

func TestFindWorktreeIssuesDirsRelativePath(t *testing.T) {
	root := t.TempDir()

	// Create a relative extra dir inside the project
	relDir := filepath.Join(root, "my-worktrees", "wt1", ".grapes")
	os.MkdirAll(relDir, 0o755)

	// Pass relative path
	result := FindWorktreeIssuesDirs(root, "my-worktrees")
	if len(result) != 1 {
		t.Errorf("relative path: got %d worktrees, want 1", len(result))
	}
	if _, ok := result["wt1"]; !ok {
		t.Error("relative path: missing wt1")
	}
}

func TestNextIDWithExtraDirs(t *testing.T) {
	root := t.TempDir()
	grapes := filepath.Join(root, ".grapes")
	os.Mkdir(grapes, 0o755)
	os.Mkdir(filepath.Join(grapes, "1"), 0o755)

	// Create extra worktree location with a higher ID
	extraDir := t.TempDir()
	extraWT := filepath.Join(extraDir, "ext-wt", ".grapes")
	os.MkdirAll(extraWT, 0o755)
	os.Mkdir(filepath.Join(extraWT, "10"), 0o755)

	// NextID should see the extra dir's ID 10
	id, err := NextID(grapes, extraDir)
	if err != nil {
		t.Fatalf("NextID: %v", err)
	}
	if id != 11 {
		t.Errorf("got %d, want 11 (should see extra dir ID 10)", id)
	}
}
