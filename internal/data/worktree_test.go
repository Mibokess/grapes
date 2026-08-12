package data

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

// testRepo drives a real git repository. Worktree attribution is defined by what
// git reports, so faking git would only test the fake.
type testRepo struct {
	t    *testing.T
	Root string
}

func newTestRepo(t *testing.T) *testRepo {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	r := &testRepo{t: t, Root: t.TempDir()}
	r.git(r.Root, "init", "-q", "-b", "main")
	r.git(r.Root, "config", "user.email", "test@example.com")
	r.git(r.Root, "config", "user.name", "Test")
	return r
}

func (r *testRepo) git(dir string, args ...string) string {
	r.t.Helper()
	return r.gitEnv(dir, nil, args...)
}

func (r *testRepo) gitEnv(dir string, env []string, args ...string) string {
	r.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return string(out)
}

// writeIssue creates or overwrites an issue in the given checkout.
func (r *testRepo) writeIssue(dir string, id int, title string) {
	r.t.Helper()
	d := filepath.Join(dir, ".grapes", strconv.Itoa(id))
	if err := os.MkdirAll(d, 0o755); err != nil {
		r.t.Fatal(err)
	}
	meta := fmt.Sprintf("id = %d\ntitle = %q\nstatus = \"todo\"\npriority = \"medium\"\n"+
		"created = 2024-01-01T00:00:00Z\nupdated = 2024-01-01T00:00:00Z\n", id, title)
	for name, body := range map[string]string{
		"meta.toml":   meta,
		"content.md":  "body of " + title + "\n",
		"comments.md": "",
	} {
		if err := os.WriteFile(filepath.Join(d, name), []byte(body), 0o644); err != nil {
			r.t.Fatal(err)
		}
	}
}

func (r *testRepo) commit(dir, msg string) {
	r.t.Helper()
	r.git(dir, "add", "-A")
	r.git(dir, "commit", "-q", "-m", msg)
}

func (r *testRepo) commitAt(dir, msg string, when time.Time) {
	r.t.Helper()
	r.git(dir, "add", "-A")
	date := when.UTC().Format(time.RFC3339)
	r.gitEnv(dir, []string{
		"GIT_AUTHOR_DATE=" + date,
		"GIT_COMMITTER_DATE=" + date,
	}, "commit", "-q", "-m", msg)
}

// addWorktree creates a worktree on a new branch and returns its path.
func (r *testRepo) addWorktree(name string) string {
	r.t.Helper()
	path := filepath.Join(r.Root, "wt", name)
	r.git(r.Root, "worktree", "add", "-q", "-b", name, path)
	return path
}

// claims runs the full attribution pass and returns claims keyed by checkout name.
func (r *testRepo) claims() map[string]Claim {
	r.t.Helper()
	checkouts, err := Checkouts(r.Root, ".grapes")
	if err != nil {
		r.t.Fatalf("Checkouts: %v", err)
	}
	branch, err := DefaultBranch(r.Root, "")
	if err != nil {
		r.t.Fatalf("DefaultBranch: %v", err)
	}
	result := make(map[string]Claim)
	for _, cl := range GatherClaims(checkouts, ".grapes", branch, newClaimCache()) {
		if cl.Err != nil {
			r.t.Fatalf("claim for %q: %v", cl.Checkout.Name, cl.Err)
		}
		result[cl.Checkout.Name] = cl
	}
	return result
}

func touchedIDs(cl Claim) []int {
	var ids []int
	for id := range cl.Touched {
		ids = append(ids, id)
	}
	for i := 1; i < len(ids); i++ {
		for j := i; j > 0 && ids[j] < ids[j-1]; j-- {
			ids[j], ids[j-1] = ids[j-1], ids[j]
		}
	}
	return ids
}

// baseRepo: main with three issues committed, ready for worktrees.
func baseRepo(t *testing.T) *testRepo {
	r := newTestRepo(t)
	for _, id := range []int{1, 2, 3} {
		r.writeIssue(r.Root, id, fmt.Sprintf("Issue %d", id))
	}
	r.commit(r.Root, "add issues")
	return r
}

func TestCheckouts_ListsMainAndWorktrees(t *testing.T) {
	r := baseRepo(t)
	r.addWorktree("feature-a")
	r.addWorktree("feature-b")

	got, err := Checkouts(r.Root, ".grapes")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d checkouts, want 3", len(got))
	}
	if !got[0].IsMain() {
		t.Errorf("first checkout is %q, want the main checkout", got[0].Name)
	}
	if got[1].Name != "feature-a" || got[2].Name != "feature-b" {
		t.Errorf("worktrees = %q, %q; want feature-a, feature-b sorted", got[1].Name, got[2].Name)
	}
	if got[1].Branch != "feature-a" {
		t.Errorf("branch = %q, want feature-a", got[1].Branch)
	}
	if got[1].Dir != filepath.Join(got[1].Path, ".grapes") {
		t.Errorf("Dir = %q, want the checkout's .grapes/", got[1].Dir)
	}
}

// The headline case: a worktree that has changed nothing claims nothing, even
// though its .grapes/ is a full copy of every issue.
func TestClaims_UntouchedWorktreeClaimsNothing(t *testing.T) {
	r := baseRepo(t)
	r.addWorktree("idle")

	cl := r.claims()["idle"]
	if len(cl.Touched) != 0 {
		t.Errorf("idle worktree claims %v, want nothing", touchedIDs(cl))
	}
}

func TestClaims_WorktreeClaimsWhatItCommitted(t *testing.T) {
	r := baseRepo(t)
	wt := r.addWorktree("worker")

	r.writeIssue(wt, 2, "Issue 2 reworked")
	r.commit(wt, "work on 2")

	cl := r.claims()["worker"]
	if got := touchedIDs(cl); len(got) != 1 || got[0] != 2 {
		t.Errorf("claims %v, want just issue 2", got)
	}
	if cl.Touched[2].Changed.IsZero() {
		t.Error("a committed change should carry the commit date")
	}
	if cl.Touched[2].Dirty {
		t.Error("a committed change is not dirty")
	}
}

func TestClaims_UncommittedChangeIsClaimed(t *testing.T) {
	r := baseRepo(t)
	wt := r.addWorktree("worker")

	r.writeIssue(wt, 3, "edited but not committed")

	cl := r.claims()["worker"]
	if got := touchedIDs(cl); len(got) != 1 || got[0] != 3 {
		t.Fatalf("claims %v, want just issue 3", got)
	}
	if !cl.Touched[3].Dirty {
		t.Error("an uncommitted change should be marked dirty")
	}
}

// An issue created only inside a worktree is untracked there, and must still be
// claimed — otherwise new issues written from a worktree would be invisible.
func TestClaims_NewIssueOnlyInWorktree(t *testing.T) {
	r := baseRepo(t)
	wt := r.addWorktree("worker")

	r.writeIssue(wt, 99, "born in a worktree")

	cl := r.claims()["worker"]
	if _, ok := cl.Touched[99]; !ok {
		t.Errorf("claims %v, want issue 99", touchedIDs(cl))
	}
}

// The case the whole design turns on: main moves ahead, so the worktree's copy
// differs from main's — but it differs because the worktree is behind, which
// says nothing about what the worktree is working on.
func TestClaims_WorktreeBehindMainClaimsNothing(t *testing.T) {
	r := baseRepo(t)
	r.addWorktree("stale")

	r.writeIssue(r.Root, 1, "Issue 1 moved on")
	r.writeIssue(r.Root, 2, "Issue 2 moved on")
	r.commit(r.Root, "main moves ahead")

	cl := r.claims()["stale"]
	if len(cl.Touched) != 0 {
		t.Errorf("a worktree that is merely behind claims %v, want nothing", touchedIDs(cl))
	}
}

// Two checkouts changed the same issue; the later commit date must win, and it
// must be main's when main committed last.
func TestClaims_MainAndWorktreeBothChange(t *testing.T) {
	r := baseRepo(t)
	wt := r.addWorktree("worker")

	r.writeIssue(wt, 1, "worktree version")
	r.commit(wt, "worktree edits 1")

	// Commit dates have one-second resolution, so separate the two commits.
	time.Sleep(1100 * time.Millisecond)
	r.writeIssue(r.Root, 1, "main version")
	r.commit(r.Root, "main edits 1")

	claims := r.claims()
	wtWhen := claims["worker"].Touched[1].Changed
	mainWhen := claims[""].Touched[1].Changed
	if wtWhen.IsZero() || mainWhen.IsZero() {
		t.Fatalf("both checkouts should claim issue 1: main=%v worktree=%v", mainWhen, wtWhen)
	}
	if !mainWhen.After(wtWhen) {
		t.Errorf("main committed later but its change time %v is not after %v", mainWhen, wtWhen)
	}
}

func TestDefaultBranch_PrefersOverrideThenConvention(t *testing.T) {
	r := baseRepo(t)

	got, err := DefaultBranch(r.Root, "")
	if err != nil {
		t.Fatal(err)
	}
	if got != "main" {
		t.Errorf("default branch = %q, want main", got)
	}

	r.git(r.Root, "branch", "trunk")
	if got, err := DefaultBranch(r.Root, "trunk"); err != nil || got != "trunk" {
		t.Errorf("override: got %q, %v; want trunk", got, err)
	}
	if _, err := DefaultBranch(r.Root, "nope"); err == nil {
		t.Error("an override that does not resolve should be an error, not a silent fallback")
	}
}

func TestCheckouts_NotAGitRepoIsAnError(t *testing.T) {
	if _, err := Checkouts(t.TempDir(), ".grapes"); err == nil {
		t.Error("a non-git directory should report an error so the caller can fall back to main only")
	}
}

// The committed half of a claim is cached on HEAD, but an uncommitted edit does
// not move HEAD — so caching must not hide it.
func TestClaims_CacheStillSeesUncommittedEdits(t *testing.T) {
	r := baseRepo(t)
	wt := r.addWorktree("worker")

	checkouts, err := Checkouts(r.Root, ".grapes")
	if err != nil {
		t.Fatal(err)
	}
	branch, err := DefaultBranch(r.Root, "")
	if err != nil {
		t.Fatal(err)
	}
	cache := newClaimCache()

	GatherClaims(checkouts, ".grapes", branch, cache) // warm the cache

	r.writeIssue(wt, 2, "edited after the cache was warm")

	for _, cl := range GatherClaims(checkouts, ".grapes", branch, cache) {
		if cl.Checkout.Name != "worker" {
			continue
		}
		if _, ok := cl.Touched[2]; !ok {
			t.Errorf("cached claim hid an uncommitted edit; claims %v", touchedIDs(cl))
		}
	}
}

func TestIssueIDInPath(t *testing.T) {
	tests := []struct {
		line string
		want int
		ok   bool
	}{
		{".grapes/11/meta.toml", 11, true},
		{" M .grapes/7/content.md", 7, true},
		{"?? .grapes/99/meta.toml", 99, true},
		{"R  .grapes/1/meta.toml -> .grapes/2/meta.toml", 1, true},
		{".grapes/notanumber/meta.toml", 0, false},
		{"README.md", 0, false},
		{".grapes/0/meta.toml", 0, false},
	}
	for _, tt := range tests {
		got, ok := issueIDInPath(".grapes", tt.line)
		if ok != tt.ok || got != tt.want {
			t.Errorf("issueIDInPath(%q) = %d, %v; want %d, %v", tt.line, got, ok, tt.want, tt.ok)
		}
	}
}

func TestIssueIDsInPath_IncludesBothRenameSides(t *testing.T) {
	got := issueIDsInPath(".grapes", "R  .grapes/1/meta.toml -> .grapes/2/meta.toml")
	if len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("issueIDsInPath rename = %v, want [1 2]", got)
	}
}

func TestGatherClaims_PropagatesMainError(t *testing.T) {
	claims := GatherClaims([]Checkout{{Path: t.TempDir()}}, ".grapes", "main", newClaimCache())
	if len(claims) != 1 || claims[0].Err == nil {
		t.Fatalf("main claim error was not propagated: %+v", claims)
	}
}
