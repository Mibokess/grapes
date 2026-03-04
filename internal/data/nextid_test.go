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
	id, err := NextID(grapes)
	if err != nil {
		t.Fatalf("NextID: %v", err)
	}
	if id != 9 {
		t.Errorf("got %d, want 9 (should see worktree ID 8)", id)
	}
}
