package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFile_CreatesAndReplaces(t *testing.T) {
	path := filepath.Join(t.TempDir(), "meta.toml")

	if err := WriteFile(path, []byte("first"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := WriteFile(path, []byte("second"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "second" {
		t.Errorf("content = %q, want %q", got, "second")
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o644 {
		t.Errorf("mode = %o, want 644", perm)
	}
}

// The temp file must not survive a successful write, and must not survive a
// failed one either — a directory littered with .meta.toml.tmp-* files would be
// picked up by nothing but would confuse anyone reading .grapes/.
func TestWriteFile_LeavesNoTempFiles(t *testing.T) {
	dir := t.TempDir()
	if err := WriteFile(filepath.Join(dir, "meta.toml"), []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "meta.toml" {
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}
		t.Errorf("directory contains %v, want only meta.toml", names)
	}
}

func TestWriteFile_MissingDirectoryIsAnError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nope", "meta.toml")
	if err := WriteFile(path, []byte("x"), 0o644); err == nil {
		t.Error("writing into a nonexistent directory should fail")
	}
}

func TestWriteFiles_PreflightsAllDestinations(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "first")
	second := filepath.Join(dir, "second")
	if err := os.WriteFile(first, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(second, 0o755); err != nil {
		t.Fatal(err)
	}
	err := WriteFiles([]AtomicFile{
		{Path: first, Data: []byte("new"), Perm: 0o644},
		{Path: second, Data: []byte("new"), Perm: 0o644},
	})
	if err == nil {
		t.Fatal("expected non-regular destination error")
	}
	got, readErr := os.ReadFile(first)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(got) != "old" {
		t.Fatalf("first destination changed despite preflight failure: %q", got)
	}
}
