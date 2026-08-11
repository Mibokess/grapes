package data

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// benchRepo builds a repository with `worktrees` worktrees and `issues` issues,
// of which `active` worktrees have actually changed an issue. This is the shape
// that matters: a real project accumulates many worktrees, but only a couple are
// working on issues at any moment, and every other one holds a stale full copy
// of the store purely because .grapes/ is tracked.
func benchRepo(tb testing.TB, worktrees, issues, active int) string {
	tb.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		tb.Skip("git not available")
	}
	root := tb.TempDir()
	run := func(dir string, args ...string) {
		tb.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			tb.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	writeIssue := func(dir string, id int, title string) {
		tb.Helper()
		d := filepath.Join(dir, ".grapes", fmt.Sprint(id))
		mustMkdirAll(tb, d)
		mustWrite(tb, filepath.Join(d, "meta.toml"), fmt.Sprintf(
			"id = %d\ntitle = %q\nstatus = \"todo\"\npriority = \"medium\"\n"+
				"created = 2024-01-01T00:00:00Z\nupdated = 2024-01-01T00:00:00Z\n", id, title))
		mustWrite(tb, filepath.Join(d, "content.md"), "body text for issue\n")
		mustWrite(tb, filepath.Join(d, "comments.md"), "### 2024-01-01\n\nhello\n")
	}

	run(root, "init", "-q", "-b", "main")
	run(root, "config", "user.email", "bench@example.com")
	run(root, "config", "user.name", "Bench")
	for i := 1; i <= issues; i++ {
		writeIssue(root, i, fmt.Sprintf("Issue %d", i))
	}
	run(root, "add", "-A")
	run(root, "commit", "-q", "-m", "issues")

	for w := 0; w < worktrees; w++ {
		name := fmt.Sprintf("wt-%02d", w)
		path := filepath.Join(root, "wt", name)
		run(root, "worktree", "add", "-q", "-b", name, path)
		if w < active {
			writeIssue(path, w+1, fmt.Sprintf("Issue %d, reworked", w+1))
			run(path, "add", "-A")
			run(path, "commit", "-q", "-m", "work")
		}
	}
	return filepath.Join(root, ".grapes")
}

// BenchmarkLoadWorkspace measures the cost of a reload as worktrees accumulate.
// The old loader read every issue from every worktree, so it grew linearly with
// worktree count; attribution should make idle worktrees nearly free.
func BenchmarkLoadWorkspace(b *testing.B) {
	for _, n := range []int{0, 5, 20, 40} {
		b.Run(fmt.Sprintf("worktrees=%d", n), func(b *testing.B) {
			dir := benchRepo(b, n, 40, min(n, 2))
			loader := NewWorkspaceLoader()
			if _, err := loader.Load(dir, WorkspaceOptions{}); err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := loader.Load(dir, WorkspaceOptions{}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func mustMkdirAll(tb testing.TB, dir string) {
	tb.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		tb.Fatal(err)
	}
}

func mustWrite(tb testing.TB, path, body string) {
	tb.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		tb.Fatal(err)
	}
}
