package data

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Checkout is one working copy of the repository: the main checkout or a git
// worktree. Every checkout holds a full copy of the issue directory, because
// .grapes/ is tracked, so a checkout is only interesting when git says its
// branch actually changed an issue.
type Checkout struct {
	Path   string // working directory root
	Name   string // display name; "" for the main checkout
	Branch string // short branch name, empty when detached
	Head   string // HEAD commit sha
	Dir    string // .grapes/ directory inside this checkout
}

// IsMain reports whether c is the main checkout rather than a worktree.
func (c Checkout) IsMain() bool { return c.Name == "" }

// Touch is one checkout's claim on one issue: when the issue last really
// changed there, and whether that change is still uncommitted.
type Touch struct {
	Changed time.Time
	Dirty   bool
}

// issueIDInPath returns the issue ID a git path refers to. Scanning for the
// issue directory anywhere in the line means the same helper handles
// `diff --name-only`, `log --name-only`, and the status-porcelain forms that
// carry a two-letter code, a quoted path, or a rename arrow.
func issueIDInPath(grapesRel, line string) (int, bool) {
	prefix := grapesRel + "/"
	idx := strings.Index(line, prefix)
	if idx < 0 {
		return 0, false
	}
	rest := line[idx+len(prefix):]
	slash := strings.Index(rest, "/")
	if slash <= 0 {
		return 0, false
	}
	id, err := strconv.Atoi(rest[:slash])
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

// git runs a git command in dir and returns its stdout. Stderr is folded into
// the error, because a git failure that reaches the UI should say what git said.
func git(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return string(out), nil
}

// DefaultBranch resolves the ref that worktree branches are compared against.
// An explicit override wins; otherwise the remote's published HEAD is the best
// available answer, with the conventional names as fallbacks. Returns an error
// only when none of them resolve, which is the signal to skip worktree
// attribution entirely rather than to guess.
func DefaultBranch(root, override string) (string, error) {
	resolves := func(ref string) bool {
		if ref == "" {
			return false
		}
		_, err := git(root, "rev-parse", "--verify", "--quiet", ref+"^{commit}")
		return err == nil
	}

	// A configured branch is a statement of intent, so it is used or reported.
	// Falling back to a guess would silently attribute work against the wrong
	// base, which is worse than saying the setting is wrong.
	if override != "" {
		if resolves(override) {
			return override, nil
		}
		return "", fmt.Errorf("configured default branch %q does not resolve", override)
	}

	var candidates []string
	if out, err := git(root, "symbolic-ref", "--short", "refs/remotes/origin/HEAD"); err == nil {
		candidates = append(candidates, strings.TrimSpace(out))
	}
	candidates = append(candidates, "origin/main", "origin/master", "main", "master")
	for _, ref := range candidates {
		if resolves(ref) {
			return ref, nil
		}
	}
	return "", fmt.Errorf("no default branch found (tried origin/HEAD, main, master)")
}

// Checkouts lists the main checkout and every git worktree of the repository
// containing mainRoot. grapesRel is the issue directory's path relative to a
// checkout root, so it can be joined onto each worktree.
//
// A repository with no worktrees yields just the main checkout. Not being a git
// repository is an error, which callers turn into "main only" plus a reported
// problem.
func Checkouts(mainRoot, grapesRel string) ([]Checkout, error) {
	out, err := git(mainRoot, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}

	var result []Checkout
	var cur Checkout
	flush := func() {
		if cur.Path == "" {
			return
		}
		cur.Dir = filepath.Join(cur.Path, grapesRel)
		if filepath.Clean(cur.Path) == filepath.Clean(mainRoot) {
			cur.Name = "" // the main checkout is never named
		} else {
			cur.Name = filepath.Base(cur.Path)
		}
		result = append(result, cur)
		cur = Checkout{}
	}

	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")
		switch {
		case strings.HasPrefix(line, "worktree "):
			flush()
			cur.Path = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "HEAD "):
			cur.Head = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			cur.Branch = strings.TrimPrefix(strings.TrimPrefix(line, "branch "), "refs/heads/")
		case line == "bare":
			cur.Path = "" // a bare checkout has no working files to read
		}
	}
	flush()

	// Sort so the main checkout leads and worktree order is stable between
	// reloads, which keeps assigned display colors from shuffling.
	mainFirst := make([]Checkout, 0, len(result))
	for _, c := range result {
		if c.IsMain() {
			mainFirst = append(mainFirst, c)
		}
	}
	rest := make([]Checkout, 0, len(result))
	for _, c := range result {
		if !c.IsMain() {
			rest = append(rest, c)
		}
	}
	sortCheckouts(rest)
	return append(mainFirst, rest...), nil
}

func sortCheckouts(cs []Checkout) {
	for i := 1; i < len(cs); i++ {
		for j := i; j > 0 && cs[j].Name < cs[j-1].Name; j-- {
			cs[j], cs[j-1] = cs[j-1], cs[j]
		}
	}
}

// dirtyIssues returns the issues with uncommitted changes in a checkout, mapped
// to the mtime of the newest changed file. An uncommitted change is a real local
// write, so unlike a checked-out file its mtime is meaningful.
func dirtyIssues(dir, grapesRel string) (map[int]time.Time, error) {
	// --no-optional-locks keeps this a pure read: without it `git status` may
	// refresh and rewrite the index, so merely displaying issues would mutate
	// git state, and forty concurrent calls would contend on the index lock.
	out, err := git(dir, "--no-optional-locks", "status", "--porcelain", "--untracked-files=all", "--", grapesRel)
	if err != nil {
		return nil, err
	}
	result := make(map[int]time.Time)
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		id, ok := issueIDInPath(grapesRel, line)
		if !ok {
			continue
		}
		if t := newestFileIn(filepath.Join(dir, grapesRel, strconv.Itoa(id))); t.After(result[id]) {
			result[id] = t
		}
	}
	return result, nil
}

// newestFileIn returns the newest mtime among an issue's files.
func newestFileIn(issueDir string) time.Time {
	var latest time.Time
	for _, name := range []string{"meta.toml", "content.md", "comments.md"} {
		if info, err := os.Stat(filepath.Join(issueDir, name)); err == nil && info.ModTime().After(latest) {
			latest = info.ModTime()
		}
	}
	return latest
}

// changedSince returns, for every issue changed in the commit range, the date of
// the most recent commit that changed it. One `git log --name-only` walk covers
// every issue in the range, which is what keeps attribution off a per-issue path.
func changedSince(dir, grapesRel, rang string) (map[int]time.Time, error) {
	out, err := git(dir, "log", "--format=C%cI", "--name-only", rang, "--", grapesRel)
	if err != nil {
		return nil, err
	}
	result := make(map[int]time.Time)
	var when time.Time
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "C") {
			if t, err := time.Parse(time.RFC3339, line[1:]); err == nil {
				when = t
			}
			continue
		}
		id, ok := issueIDInPath(grapesRel, line)
		if !ok {
			continue
		}
		// Commits arrive newest first, so the first date seen for an issue is
		// its most recent change; later, older commits must not overwrite it.
		if _, seen := result[id]; !seen && !when.IsZero() {
			result[id] = when
		}
	}
	return result, nil
}

// Claim is what one checkout has to say about the issues it changed.
type Claim struct {
	Checkout Checkout
	Base     string        // merge-base with the default branch
	Touched  map[int]Touch // issue ID → when it really changed here
	Err      error         // attribution failed; the checkout is skipped
}

// claimCache remembers a worktree's committed claims. Those change only when
// HEAD moves, so a reload that finds the same sha can skip two git calls per
// worktree. Uncommitted changes are never cached: writing a file does not move
// HEAD, so a cached dirty set would go stale immediately.
type claimCache struct {
	mu      sync.Mutex
	entries map[string]cachedClaim // keyed by checkout path
}

type cachedClaim struct {
	head      string
	branchSha string // the default branch tip the merge-base was computed against
	base      string
	committed map[int]time.Time
}

func newClaimCache() *claimCache {
	return &claimCache{entries: make(map[string]cachedClaim)}
}

func (c *claimCache) get(path, head, branchSha string) (cachedClaim, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[path]
	if !ok || e.head != head || e.branchSha != branchSha {
		return cachedClaim{}, false
	}
	return e, true
}

func (c *claimCache) put(path string, e cachedClaim) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[path] = e
}

// prune drops cache entries for checkouts that no longer exist, so a long
// session does not hold state for deleted worktrees.
func (c *claimCache) prune(live []Checkout) {
	alive := make(map[string]bool, len(live))
	for _, co := range live {
		alive[co.Path] = true
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for path := range c.entries {
		if !alive[path] {
			delete(c.entries, path)
		}
	}
}

// claimFor computes what a worktree changed relative to where it branched off.
//
// Comparing against the merge-base rather than against current main is the whole
// point: a worktree that has fallen behind differs from main in ways that say
// nothing about what it is working on, while its diff against its own base is
// exactly the work it has done.
func claimFor(co Checkout, grapesRel, defaultBranch, branchSha string, cache *claimCache) Claim {
	cl := Claim{Checkout: co}

	// The merge-base and the commits since it depend only on this checkout's
	// HEAD and the default branch tip. While neither moves, two git processes
	// per worktree per reload can be skipped entirely.
	var committed map[int]time.Time
	var err error
	if e, ok := cache.get(co.Path, co.Head, branchSha); ok {
		cl.Base, committed = e.base, e.committed
	} else {
		base, err := git(co.Path, "merge-base", defaultBranch, "HEAD")
		if err != nil {
			cl.Err = err
			return cl
		}
		cl.Base = strings.TrimSpace(base)
		committed, err = changedSince(co.Path, grapesRel, cl.Base+"..HEAD")
		if err != nil {
			cl.Err = err
			return cl
		}
		cache.put(co.Path, cachedClaim{head: co.Head, branchSha: branchSha, base: cl.Base, committed: committed})
	}

	// Uncommitted changes are never cached: writing a file moves neither HEAD
	// nor the branch tip, so a cached answer would go stale immediately.
	dirty, err := dirtyIssues(co.Path, grapesRel)
	if err != nil {
		cl.Err = err
		return cl
	}

	cl.Touched = make(map[int]Touch, len(committed)+len(dirty))
	for id, when := range committed {
		cl.Touched[id] = Touch{Changed: when}
	}
	// An uncommitted change is newer than anything committed here by definition,
	// so it replaces rather than competes with the committed timestamp.
	for id, when := range dirty {
		cl.Touched[id] = Touch{Changed: when, Dirty: true}
	}
	return cl
}

// mainClaim describes what the main checkout changed since floor, the oldest
// point any worktree branched from. Changes older than that are already shared
// with every worktree, so they cannot decide ownership and need not be read.
func mainClaim(co Checkout, grapesRel, floor string) Claim {
	cl := Claim{Checkout: co, Touched: map[int]Touch{}}

	rang := "HEAD"
	if floor != "" {
		rang = floor + "..HEAD"
	}
	committed, err := changedSince(co.Path, grapesRel, rang)
	if err != nil {
		cl.Err = err
		return cl
	}
	for id, when := range committed {
		cl.Touched[id] = Touch{Changed: when}
	}
	if dirty, err := dirtyIssues(co.Path, grapesRel); err == nil {
		for id, when := range dirty {
			cl.Touched[id] = Touch{Changed: when, Dirty: true}
		}
	}
	return cl
}

// GatherClaims asks every checkout what it has changed, running the worktrees
// concurrently because each one is an independent set of git calls.
//
// The main checkout is queried last and only back to the oldest merge-base, so
// its history walk stays bounded by how far the worktrees have diverged rather
// than by the age of the repository.
func GatherClaims(checkouts []Checkout, grapesRel, defaultBranch string, cache *claimCache) []Claim {
	var worktrees []Checkout
	var main *Checkout
	for i, co := range checkouts {
		if co.IsMain() {
			main = &checkouts[i]
			continue
		}
		worktrees = append(worktrees, co)
	}

	// Resolved once and passed down: it is part of every worktree's cache key.
	branchSha := ""
	if main != nil {
		if out, err := git(main.Path, "rev-parse", defaultBranch); err == nil {
			branchSha = strings.TrimSpace(out)
		}
	}

	claims := make([]Claim, len(worktrees))
	limit := runtime.NumCPU()
	if limit < 1 {
		limit = 1
	}
	sem := make(chan struct{}, limit)
	var wg sync.WaitGroup
	for i, co := range worktrees {
		wg.Add(1)
		go func(i int, co Checkout) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			claims[i] = claimFor(co, grapesRel, defaultBranch, branchSha, cache)
		}(i, co)
	}
	wg.Wait()

	if main == nil {
		return claims
	}

	// The floor is the oldest base among worktrees that actually claimed
	// something; main's changes before that are common ancestry.
	floor := ""
	oldest := time.Time{}
	for _, cl := range claims {
		if cl.Err != nil || len(cl.Touched) == 0 || cl.Base == "" {
			continue
		}
		out, err := git(main.Path, "show", "-s", "--format=%cI", cl.Base)
		if err != nil {
			continue
		}
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(out))
		if err != nil {
			continue
		}
		if oldest.IsZero() || t.Before(oldest) {
			oldest = t
			floor = cl.Base
		}
	}
	return append([]Claim{mainClaim(*main, grapesRel, floor)}, claims...)
}
