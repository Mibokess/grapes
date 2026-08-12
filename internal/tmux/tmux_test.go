package tmux

import (
	"errors"
	"os/exec"
	"strings"
	"testing"
)

func TestSessionNameStableAndSanitized(t *testing.T) {
	first := sessionName("/tmp/Project Name", 42, "feature/foo bar")
	second := sessionName("/tmp/Project Name", 42, "feature/foo bar")
	if first != second {
		t.Fatalf("session name changed between calls: %q != %q", first, second)
	}
	if !strings.Contains(first, "Project-Name") || !strings.Contains(first, "feature-foo-bar") {
		t.Fatalf("session name does not retain sanitized identity: %q", first)
	}
	for _, r := range first {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.') {
			t.Errorf("session name contains unsafe rune %q in %q", r, first)
		}
	}
}

func TestParseSessionsFiltersMetadataAndSelectsPane(t *testing.T) {
	root := "/repo/project"
	output := strings.Join([]string{
		"renamed\t1\t/repo/project\t7\tworker\tagent\t1\t2\t/repo/project\tvim",
		"renamed\t1\t/repo/project\t7\tworker\tagent\t0\t0\t/repo/project\tbash",
		"unmanaged\t0\t/repo/project\t8\tworker\t\t0\t0\t/repo\tsh",
		"other-project\t0\t/repo/other\t7\tworker\t\t0\t0\t/repo/other\tsh",
		"bad-issue\t0\t/repo/project\tnope\tworker\t\t0\t0\t/repo\tsh",
	}, "\n")

	got := parseSessions(output, root)
	if len(got) != 2 {
		t.Fatalf("parsed %d sessions, want 2: %#v", len(got), got)
	}
	var renamed Session
	for _, session := range got {
		if session.Name == "renamed" {
			renamed = session
		}
	}
	if renamed.Target != "renamed:0.0" || renamed.Path != root || renamed.Pane != "bash" {
		t.Errorf("representative pane = %#v, want target renamed:0.0, path %q, pane bash", renamed, root)
	}
	if !renamed.Attached || renamed.IssueID != 7 || renamed.Worktree != "worker" || renamed.Agent != "agent" {
		t.Errorf("metadata = %#v", renamed)
	}
}

func TestListNoServerIsEmpty(t *testing.T) {
	oldLookPath := tmuxLookPath
	oldRun := tmuxRunCommand
	t.Cleanup(func() {
		tmuxLookPath = oldLookPath
		tmuxRunCommand = oldRun
	})
	tmuxLookPath = func(string) (string, error) { return "/usr/bin/tmux", nil }
	tmuxRunCommand = func(string, ...string) ([]byte, []byte, error) {
		return nil, []byte("no server running on /tmp/tmux.sock\n"), errors.New("exit status 1")
	}

	got, err := List("/repo/project")
	if err != nil {
		t.Fatalf("List returned error for absent server: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("List = %#v, want empty non-nil list", got)
	}
}

func TestListMissingTmuxIsEmpty(t *testing.T) {
	oldLookPath := tmuxLookPath
	t.Cleanup(func() { tmuxLookPath = oldLookPath })
	tmuxLookPath = func(string) (string, error) { return "", exec.ErrNotFound }

	got, err := List("/repo/project")
	if err != nil || len(got) != 0 {
		t.Fatalf("List = %#v, %v; want empty list and nil error", got, err)
	}
}
func TestEnsureCreatesAndReusesSession(t *testing.T) {
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		t.Skip("tmux is not installed")
	}

	root := t.TempDir()
	const issueID = 73
	const worktree = "feature/test"
	name := sessionName(root, issueID, worktree)
	cleanupNames := map[string]bool{name: true}
	t.Cleanup(func() {
		for target := range cleanupNames {
			_, _, _ = tmuxRunCommand(tmuxPath, "kill-session", "-t", target)
		}
	})

	first, err := Ensure(root, issueID, worktree, root)
	if err != nil {
		t.Fatalf("Ensure(create) failed: %v", err)
	}
	if first.Name == "" || first.Target == "" {
		t.Fatalf("Ensure(create) returned incomplete session: %#v", first)
	}

	listed, err := List(root)
	if err != nil {
		t.Fatalf("List after create failed: %v", err)
	}
	if len(listed) != 1 || listed[0].IssueID != issueID || listed[0].Worktree != worktree {
		t.Fatalf("List after create = %#v, want issue %d/worktree %q", listed, issueID, worktree)
	}

	second, err := Ensure(root, issueID, worktree, root)
	if err != nil {
		t.Fatalf("Ensure(reuse) failed: %v", err)
	}
	if second.Name != first.Name {
		t.Fatalf("Ensure(reuse) changed session name: %q -> %q", first.Name, second.Name)
	}

	renamed := first.Name + "-renamed"
	if _, stderr, err := tmuxRunCommand(tmuxPath, "rename-session", "-t", first.Name, renamed); err != nil {
		t.Fatalf("rename-session failed: %v (%s)", err, stderr)
	}
	delete(cleanupNames, first.Name)
	cleanupNames[renamed] = true

	third, err := Ensure(root, issueID, worktree, root)
	if err != nil {
		t.Fatalf("Ensure(renamed reuse) failed: %v", err)
	}
	if third.Name != renamed {
		t.Fatalf("Ensure(renamed reuse) returned %q, want %q", third.Name, renamed)
	}
}
