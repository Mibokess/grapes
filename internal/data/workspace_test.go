package data

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

// loadWS loads the workspace from a repository's main checkout.
func loadWS(t *testing.T, r *testRepo) Workspace {
	t.Helper()
	ws, err := NewWorkspaceLoader().Load(filepath.Join(r.Root, ".grapes"), WorkspaceOptions{})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return ws
}

func issueByID(t *testing.T, ws Workspace, id int) Issue {
	t.Helper()
	for _, iss := range ws.Issues {
		if iss.ID == id {
			return iss
		}
	}
	t.Fatalf("issue %d not in workspace", id)
	return Issue{}
}

// A worktree that has changed nothing must not appear anywhere: not as a source
// on any issue, not in the worktree list, and not in the watch set.
func TestWorkspace_IdleWorktreesAreInvisible(t *testing.T) {
	r := baseRepo(t)
	r.addWorktree("idle-a")
	r.addWorktree("idle-b")

	ws := loadWS(t, r)

	if len(ws.Worktrees) != 0 {
		t.Errorf("worktrees = %v, want none listed", ws.WorktreeNames())
	}
	for _, iss := range ws.Issues {
		if len(iss.Sources) != 1 {
			t.Errorf("issue %d has %d sources, want only the main copy", iss.ID, len(iss.Sources))
		}
		if iss.Worktree != "" {
			t.Errorf("issue %d attributed to %q, want the main checkout", iss.ID, iss.Worktree)
		}
	}
	if len(ws.WatchDirs) != 1 {
		t.Errorf("watch dirs = %v, want only the canonical store", ws.WatchDirs)
	}
}

// This is the bug that silently lost edits: git stamps a worktree's files at
// checkout time, so a freshly created worktree holds the newest mtimes for every
// issue despite never touching one. Ownership must ignore that, and writes must
// follow ownership.
func TestWorkspace_FreshWorktreeDoesNotStealOwnership(t *testing.T) {
	r := baseRepo(t)
	mainDir := filepath.Join(r.Root, ".grapes")

	// Create the worktree after main's files, so its copies are strictly newer
	// on disk. Guarantee it rather than trusting checkout timing.
	wt := r.addWorktree("fresh")
	future := time.Now().Add(2 * time.Hour)
	for _, id := range []int{1, 2, 3} {
		for _, name := range []string{"meta.toml", "content.md", "comments.md"} {
			p := filepath.Join(wt, ".grapes", strconv.Itoa(id), name)
			if err := os.Chtimes(p, future, future); err != nil {
				t.Fatal(err)
			}
		}
	}

	ws := loadWS(t, r)

	for _, iss := range ws.Issues {
		if iss.Worktree != "" {
			t.Errorf("issue %d owned by %q; a checkout timestamp is not an edit",
				iss.ID, iss.Worktree)
		}
		if iss.SourceDir != mainDir {
			t.Errorf("issue %d would be written to %s, want the canonical store %s",
				iss.ID, iss.SourceDir, mainDir)
		}
	}
}

func TestWorkspace_WorktreeOwnsWhatItChanged(t *testing.T) {
	r := baseRepo(t)
	wt := r.addWorktree("worker")

	r.writeIssue(wt, 2, "Issue 2, worked on")
	r.commit(wt, "work on 2")

	ws := loadWS(t, r)

	if names := ws.WorktreeNames(); len(names) != 1 || names[0] != "worker" {
		t.Fatalf("worktrees = %v, want just worker", names)
	}
	if got := ws.Worktrees[0].Touched; len(got) != 1 || got[0] != 2 {
		t.Errorf("touched = %v, want [2]", got)
	}
	if got := ws.Worktrees[0].Owned; len(got) != 1 || got[0] != 2 {
		t.Errorf("owned = %v, want [2]", got)
	}

	two := issueByID(t, ws, 2)
	if two.Worktree != "worker" {
		t.Errorf("issue 2 owner = %q, want worker", two.Worktree)
	}
	if two.Title != "Issue 2, worked on" {
		t.Errorf("issue 2 title = %q, want the worktree's version", two.Title)
	}
	if two.SourceDir != filepath.Join(wt, ".grapes") {
		t.Errorf("issue 2 writes to %s, want the owning worktree", two.SourceDir)
	}

	// Untouched issues stay with the main checkout.
	if one := issueByID(t, ws, 1); one.Worktree != "" {
		t.Errorf("issue 1 owner = %q, want the main checkout", one.Worktree)
	}
	if len(ws.WatchDirs) != 2 {
		t.Errorf("watch dirs = %v, want the store plus the working worktree", ws.WatchDirs)
	}
}

// When main changes an issue after a branch did, main is the newer version and
// keeps ownership — the branch is behind, not ahead.
func TestWorkspace_MainWinsWhenItChangedLast(t *testing.T) {
	r := baseRepo(t)
	wt := r.addWorktree("worker")

	r.writeIssue(wt, 1, "branch version")
	r.commit(wt, "branch edits 1")

	time.Sleep(1100 * time.Millisecond) // commit dates have one-second resolution
	r.writeIssue(r.Root, 1, "main version")
	r.commit(r.Root, "main edits 1")

	ws := loadWS(t, r)

	one := issueByID(t, ws, 1)
	if one.Worktree != "" {
		t.Errorf("issue 1 owner = %q, want the main checkout", one.Worktree)
	}
	if one.Title != "main version" {
		t.Errorf("issue 1 title = %q, want main's version", one.Title)
	}
	// The divergence is still visible for inspection, even though main won.
	if len(one.Sources) != 2 {
		t.Errorf("issue 1 has %d sources, want main plus the diverging branch", len(one.Sources))
	}
	// The worktree is still listed: it has a conflicting version worth finding.
	if names := ws.WorktreeNames(); len(names) != 1 {
		t.Errorf("worktrees = %v, want the diverging branch listed", names)
	}
}

// An issue created in a worktree and never committed still has to show up.
func TestWorkspace_NewIssueInWorktreeAppears(t *testing.T) {
	r := baseRepo(t)
	wt := r.addWorktree("worker")

	r.writeIssue(wt, 42, "born in a worktree")

	ws := loadWS(t, r)

	iss := issueByID(t, ws, 42)
	if iss.Title != "born in a worktree" {
		t.Errorf("title = %q", iss.Title)
	}
	if iss.Worktree != "worker" {
		t.Errorf("owner = %q, want worker", iss.Worktree)
	}
}

// An uncommitted edit is a real local write, so it beats a committed version.
func TestWorkspace_UncommittedEditWins(t *testing.T) {
	r := baseRepo(t)
	wt := r.addWorktree("worker")

	r.writeIssue(wt, 3, "edited, not committed")

	ws := loadWS(t, r)

	three := issueByID(t, ws, 3)
	if three.Title != "edited, not committed" {
		t.Errorf("title = %q, want the uncommitted version", three.Title)
	}
	if !three.Sources[three.ActiveSource].Dirty {
		t.Error("the winning source should be marked dirty")
	}
}

// Outside a git repository grapes still works, and does not nag about it.
func TestWorkspace_NonGitDirectoryIsQuiet(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".grapes")
	if err := os.MkdirAll(filepath.Join(dir, "1"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "1", "meta.toml"),
		[]byte("title = \"only issue\"\nstatus = \"todo\"\npriority = \"low\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ws, err := NewWorkspaceLoader().Load(dir, WorkspaceOptions{})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(ws.Issues) != 1 {
		t.Fatalf("loaded %d issues, want 1", len(ws.Issues))
	}
	if len(ws.Problems) != 0 {
		t.Errorf("problems = %v; being outside git is normal and must not be reported", ws.Problems)
	}
	if ws.AttributionErr == nil {
		t.Error("the reason worktree attribution is off should still be recorded")
	}
}

// A configured branch that does not exist is reported, because attribution
// against the wrong base would quietly mis-assign every issue.
func TestWorkspace_BadDefaultBranchIsReported(t *testing.T) {
	r := baseRepo(t)

	ws, err := NewWorkspaceLoader().Load(filepath.Join(r.Root, ".grapes"),
		WorkspaceOptions{DefaultBranch: "does-not-exist"})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(ws.Problems) == 0 {
		t.Error("an unresolvable configured branch should be reported")
	}
	if len(ws.Issues) != 3 {
		t.Errorf("loaded %d issues, want the main checkout's 3 despite the bad setting", len(ws.Issues))
	}
}

// Reloading through one loader must give the same answer as a cold load: the
// caches are an optimisation, not a change in behaviour.
func TestWorkspace_CachedReloadMatchesColdLoad(t *testing.T) {
	r := baseRepo(t)
	wt := r.addWorktree("worker")
	r.writeIssue(wt, 2, "worked on")
	r.commit(wt, "work")

	dir := filepath.Join(r.Root, ".grapes")
	loader := NewWorkspaceLoader()
	if _, err := loader.Load(dir, WorkspaceOptions{}); err != nil {
		t.Fatal(err)
	}

	// Change the worktree again; the cache keys on HEAD, which just moved.
	r.writeIssue(wt, 3, "also worked on")
	r.commit(wt, "more work")

	warm, err := loader.Load(dir, WorkspaceOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cold := loadWS(t, r)

	if len(warm.Worktrees) != len(cold.Worktrees) {
		t.Fatalf("warm listed %v, cold listed %v", warm.WorktreeNames(), cold.WorktreeNames())
	}
	for i := range cold.Issues {
		if warm.Issues[i].Worktree != cold.Issues[i].Worktree {
			t.Errorf("issue %d: warm owner %q, cold owner %q",
				cold.Issues[i].ID, warm.Issues[i].Worktree, cold.Issues[i].Worktree)
		}
		if warm.Issues[i].Title != cold.Issues[i].Title {
			t.Errorf("issue %d: warm title %q, cold title %q",
				cold.Issues[i].ID, warm.Issues[i].Title, cold.Issues[i].Title)
		}
	}
}
