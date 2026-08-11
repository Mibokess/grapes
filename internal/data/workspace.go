package data

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// WorkspaceOptions configures a workspace load.
type WorkspaceOptions struct {
	// DefaultBranch overrides the ref worktree branches are compared against.
	// Empty means "work it out from the repository".
	DefaultBranch string
	// ExtraDirs holds glob patterns for issue directories outside this
	// repository. They are loaded whole, because git attribution says nothing
	// about a store that is not part of this history.
	ExtraDirs []string
}

// WorktreeInfo describes a worktree that is actually doing work: it has changed
// at least one issue relative to where it branched off.
type WorktreeInfo struct {
	Name    string
	Branch  string
	Path    string
	Dir     string
	Touched []int // issues this worktree changed
	Owned   []int // issues where its version is the most recent
}

// Workspace is the result of a load: the issues, the worktrees worth showing,
// the directories worth watching, and anything that went wrong on the way.
type Workspace struct {
	Issues    []Issue
	Worktrees []WorktreeInfo
	// WatchDirs contains the canonical store, active worktree stores, and
	// configured external stores. These roots are expanded to numeric issue
	// directories by the TUI watcher.
	WatchDirs []string
	// WatchRoots contains one checkout root per repository worktree. They are
	// intentionally not expanded: idle worktrees are discovered by the
	// periodic workspace poll without an issue-directory watcher per copy.
	WatchRoots []string
	Problems   []LoadProblem

	// AttributionErr says why worktree attribution is off, when it is. Being
	// outside a git repository is the common, unremarkable case, so this is
	// shown where a user would look for worktrees rather than reported on every
	// load.
	AttributionErr error
}

// WorktreeNames returns the worktree names in display order.
func (w Workspace) WorktreeNames() []string {
	names := make([]string, 0, len(w.Worktrees))
	for _, wt := range w.Worktrees {
		names = append(names, wt.Name)
	}
	return names
}

// WorkspaceLoader loads issues across a repository's checkouts. It holds the
// per-worktree claim cache, so reloads that find unchanged worktrees skip most
// of the git work. Reuse one loader for the lifetime of the process.
type WorkspaceLoader struct {
	cache *claimCache

	// Where the repository is and what its default branch is called cannot
	// change while grapes is running, so they are resolved once. Repeating them
	// costs several git processes on every reload, which is most of the fixed
	// cost of a load in a project with no worktrees at all.
	mu       sync.Mutex
	layout   *layout
	branch   string
	branchOf string // the override the cached branch was resolved for

	activityMu    sync.Mutex
	activityReady bool
	activitySig   string
}

type layout struct {
	mainDir   string
	mainRoot  string
	grapesRel string
	err       error
}

func NewWorkspaceLoader() *WorkspaceLoader {
	return &WorkspaceLoader{cache: newClaimCache()}
}

// layoutFor resolves, and then remembers, where the canonical store lives.
func (wl *WorkspaceLoader) layoutFor(issuesDir string) layout {
	wl.mu.Lock()
	defer wl.mu.Unlock()
	if wl.layout == nil {
		mainDir, mainRoot, grapesRel, err := resolveLayout(issuesDir)
		wl.layout = &layout{mainDir: mainDir, mainRoot: mainRoot, grapesRel: grapesRel, err: err}
	}
	return *wl.layout
}

// defaultBranchFor resolves, and then remembers, the comparison branch.
func (wl *WorkspaceLoader) defaultBranchFor(mainRoot, override string) (string, error) {
	wl.mu.Lock()
	defer wl.mu.Unlock()
	if wl.branch != "" && wl.branchOf == override {
		return wl.branch, nil
	}
	branch, err := DefaultBranch(mainRoot, override)
	if err != nil {
		return "", err
	}
	wl.branch, wl.branchOf = branch, override
	return branch, nil
}

// ActivityChanged performs the cheap part of a workspace reload: it checks
// worktree claims and checkout identities without reading every issue file.
// A successful Load seeds the baseline; a fresh loader's first call establishes
// one without requesting a redundant reload.
func (wl *WorkspaceLoader) ActivityChanged(issuesDir string, opts WorkspaceOptions) (bool, error) {
	lay := wl.layoutFor(issuesDir)
	if lay.err != nil {
		return false, nil
	}
	checkouts, err := Checkouts(lay.mainRoot, lay.grapesRel)
	if err != nil {
		return false, err
	}
	if len(checkouts) == 1 {
		dirty, err := dirtyIssues(checkouts[0].Path, lay.grapesRel)
		if err != nil {
			return false, err
		}
		claim := Claim{Checkout: checkouts[0], Touched: make(map[int]Touch, len(dirty))}
		for id, when := range dirty {
			claim.Touched[id] = Touch{Changed: when, Dirty: true}
		}
		return wl.updateActivitySignature(activitySignature([]Claim{claim}, opts)), nil
	}
	branch, err := wl.defaultBranchFor(lay.mainRoot, opts.DefaultBranch)
	if err != nil {
		return false, err
	}

	wl.cache.prune(checkouts)
	claims := GatherClaims(checkouts, lay.grapesRel, branch, wl.cache)
	return wl.updateActivitySignature(activitySignature(claims, opts)), nil
}

func (wl *WorkspaceLoader) updateActivitySignature(sig string) bool {
	wl.activityMu.Lock()
	defer wl.activityMu.Unlock()
	if !wl.activityReady {
		wl.activityReady = true
		wl.activitySig = sig
		return false
	}
	if sig == wl.activitySig {
		return false
	}
	wl.activitySig = sig
	return true
}

func (wl *WorkspaceLoader) seedActivitySignature(sig string) {
	wl.activityMu.Lock()
	wl.activityReady = true
	wl.activitySig = sig
	wl.activityMu.Unlock()
}

// InvalidateActivity forces the next poll to establish a fresh baseline. It is
// used after a failed full load so a transient Git/filesystem error is retried.
func (wl *WorkspaceLoader) InvalidateActivity() {
	wl.activityMu.Lock()
	wl.activityReady = false
	wl.activitySig = ""
	wl.activityMu.Unlock()
}

func activitySignature(claims []Claim, opts WorkspaceOptions) string {
	var b strings.Builder
	b.WriteString(opts.DefaultBranch)
	b.WriteByte(0)
	for _, dir := range opts.ExtraDirs {
		b.WriteString(dir)
		b.WriteByte(0)
	}
	for _, cl := range claims {
		b.WriteString(cl.Checkout.Path)
		b.WriteByte(0)
		b.WriteString(cl.Checkout.Head)
		b.WriteByte(0)
		b.WriteString(cl.Checkout.Branch)
		b.WriteByte(0)
		b.WriteString(cl.Base)
		b.WriteByte(0)
		if cl.Err != nil {
			b.WriteString(cl.Err.Error())
		}
		b.WriteByte(0)
		for _, id := range sortedIDs(cl.Touched) {
			t := cl.Touched[id]
			b.WriteString(strconv.Itoa(id))
			b.WriteByte(':')
			b.WriteString(strconv.FormatInt(t.Changed.UnixNano(), 10))
			if t.Dirty {
				b.WriteByte('d')
			}
			b.WriteByte(';')
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// Load reads every issue that matters and decides which checkout owns each one.
//
// Only the main checkout is read in full. Worktrees contribute just the issues
// git reports them as having changed, which is what keeps the cost independent
// of how many worktrees exist. Losing git, or the repository, degrades to the
// main directory alone and is reported rather than hidden.
func (wl *WorkspaceLoader) Load(issuesDir string, opts WorkspaceOptions) (Workspace, error) {
	var ws Workspace

	lay := wl.layoutFor(issuesDir)
	mainDir, mainRoot, grapesRel, setupErr := lay.mainDir, lay.mainRoot, lay.grapesRel, lay.err
	if setupErr != nil {
		// Outside a git repository there are no worktrees to attribute, and
		// that is an ordinary way to use grapes. Saying so on every reload
		// would be nagging, so it is recorded for anyone who asks rather than
		// pushed into the status bar.
		ws.AttributionErr = setupErr
		return wl.loadFlat(issuesDir, opts, ws)
	}

	// From here the directory *is* in a repository, so worktree support is
	// expected to work. A failure now is worth reporting.
	checkouts, err := Checkouts(mainRoot, grapesRel)
	if err != nil {
		ws.AttributionErr = err
		ws.Problems = append(ws.Problems, LoadProblem{Dir: mainRoot, Err: err})
		return wl.loadFlat(mainDir, opts, ws)
	}
	branch, err := wl.defaultBranchFor(mainRoot, opts.DefaultBranch)
	if err != nil {
		ws.AttributionErr = err
		ws.Problems = append(ws.Problems, LoadProblem{Dir: mainRoot, Err: err})
		return wl.loadFlat(mainDir, opts, ws)
	}

	wl.cache.prune(checkouts)
	claims := GatherClaims(checkouts, grapesRel, branch, wl.cache)
	for _, cl := range claims {
		if !cl.Checkout.IsMain() || cl.Err == nil {
			continue
		}
		// Without the main claim, ownership timestamps are incomplete and
		// successful worktree claims can produce a misleading merged view.
		// Keep the error visible and load the un-attributed baseline instead.
		ws.AttributionErr = cl.Err
		ws.Problems = append(ws.Problems, LoadProblem{Dir: cl.Checkout.Path, Err: cl.Err})
		return wl.loadFlat(mainDir, opts, ws)
	}

	// The main checkout is the baseline: every issue is read from it once.
	mainIssues, problems, err := loadAllIssues(mainDir, false)
	if err != nil {
		return ws, err
	}
	ws.Problems = append(ws.Problems, problems...)

	byID := make(map[int]*Issue, len(mainIssues))
	var mainTouched map[int]Touch
	for _, cl := range claims {
		if cl.Checkout.IsMain() {
			if cl.Err != nil {
				ws.Problems = append(ws.Problems, LoadProblem{Dir: cl.Checkout.Path, Err: cl.Err})
			} else {
				mainTouched = cl.Touched
			}
		}
	}
	for _, iss := range mainIssues {
		c := iss
		c.SourceDir = mainDir
		c.Sources = []IssueSource{sourceOf(iss, "", mainDir, mainTouched[iss.ID])}
		byID[iss.ID] = &c
	}

	// Worktrees contribute only what they changed.
	for _, cl := range claims {
		if cl.Checkout.IsMain() {
			continue
		}
		if cl.Err != nil {
			ws.Problems = append(ws.Problems, LoadProblem{Dir: cl.Checkout.Path, Err: cl.Err})
			continue
		}
		if len(cl.Touched) == 0 {
			continue // an untouched worktree is a stale copy and says nothing
		}
		info := WorktreeInfo{
			Name:   cl.Checkout.Name,
			Branch: cl.Checkout.Branch,
			Path:   cl.Checkout.Path,
			Dir:    cl.Checkout.Dir,
		}
		for _, id := range sortedIDs(cl.Touched) {
			iss, err := LoadIssue(cl.Checkout.Dir, id)
			if err != nil {
				// A branch that deleted an issue still shows as having touched
				// it, and has no files to read. That is not a failure.
				if os.IsNotExist(err) {
					continue
				}
				ws.Problems = append(ws.Problems, LoadProblem{Dir: cl.Checkout.Dir, ID: id, Err: err})
				continue
			}
			info.Touched = append(info.Touched, id)
			src := sourceOf(iss, cl.Checkout.Name, cl.Checkout.Dir, cl.Touched[id])
			if existing, ok := byID[id]; ok {
				existing.Sources = append(existing.Sources, src)
			} else {
				// An issue that exists only on this branch.
				c := iss
				c.SourceDir = cl.Checkout.Dir
				c.Worktree = cl.Checkout.Name
				c.Sources = []IssueSource{src}
				byID[id] = &c
			}
		}
		if len(info.Touched) > 0 {
			ws.Worktrees = append(ws.Worktrees, info)
		}
	}

	// External stores configured by glob are outside this history, so they are
	// loaded whole and always count as candidates.
	knownDirs := map[string]bool{filepath.Clean(mainDir): true}
	for _, co := range checkouts {
		knownDirs[filepath.Clean(co.Dir)] = true
	}
	externalDirs := externalDirs(mainRoot, opts.ExtraDirs, knownDirs)
	ws.Problems = append(ws.Problems, wl.mergeExternal(byID, mainRoot, opts.ExtraDirs, knownDirs)...)

	ws.Issues = resolveOwners(byID)
	attributeOwnership(ws.Issues, &ws)
	ws.WatchDirs = watchDirs(mainDir, ws.Worktrees, externalDirs...)
	ws.WatchRoots = worktreeWatchRoots(checkouts)
	RewireRelationships(ws.Issues)
	activityClaims := claims
	if len(checkouts) == 1 {
		claim := Claim{Checkout: checkouts[0], Touched: map[int]Touch{}}
		if len(claims) > 0 {
			for id, touch := range claims[0].Touched {
				if touch.Dirty {
					claim.Touched[id] = touch
				}
			}
		}
		activityClaims = []Claim{claim}
	}
	wl.seedActivitySignature(activitySignature(activityClaims, opts))
	return ws, nil
}

// resolveLayout works out where the canonical issue directory lives and where
// the issue directory sits inside any checkout of this repository.
func resolveLayout(issuesDir string) (mainDir, mainRoot, grapesRel string, err error) {
	issuesDir, err = filepath.Abs(issuesDir)
	if err != nil {
		return "", "", "", fmt.Errorf("resolving issue directory %s: %w", issuesDir, err)
	}
	local := ProjectRoot(issuesDir)
	top, err := git(local, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", "", "", fmt.Errorf("worktree attribution unavailable: %w", err)
	}
	rel, err := filepath.Rel(strings.TrimSpace(top), issuesDir)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", "", "", fmt.Errorf("issue directory %s is outside its repository", issuesDir)
	}
	grapesRel = filepath.ToSlash(rel)

	mainRoot = FindMainProjectRoot(issuesDir)
	mainDir = filepath.Join(mainRoot, filepath.FromSlash(grapesRel))
	if info, err := os.Stat(mainDir); err != nil || !info.IsDir() {
		// The main checkout has no issue directory; the one we were given is
		// the only store there is.
		return issuesDir, ProjectRoot(issuesDir), grapesRel, nil
	}
	return mainDir, mainRoot, grapesRel, nil
}

// loadFlat is the degraded path: one directory, no attribution.
func (wl *WorkspaceLoader) loadFlat(dir string, opts WorkspaceOptions, ws Workspace) (Workspace, error) {
	issues, problems, err := loadAllIssues(dir, false)
	if err != nil {
		return ws, err
	}
	ws.Problems = append(ws.Problems, problems...)

	byID := make(map[int]*Issue, len(issues))
	for _, iss := range issues {
		c := iss
		c.SourceDir = dir
		c.Sources = []IssueSource{sourceOf(iss, "", dir, Touch{})}
		byID[iss.ID] = &c
	}
	externalDirs := externalDirs(ProjectRoot(dir), opts.ExtraDirs, map[string]bool{filepath.Clean(dir): true})
	ws.Problems = append(ws.Problems, wl.mergeExternal(byID, ProjectRoot(dir), opts.ExtraDirs, map[string]bool{filepath.Clean(dir): true})...)

	ws.Issues = resolveOwners(byID)
	attributeOwnership(ws.Issues, &ws)
	ws.WatchDirs = watchDirs(dir, ws.Worktrees, externalDirs...)
	RewireRelationships(ws.Issues)
	return ws, nil
}

// mergeExternal folds in issue directories matched by configured glob patterns.
func (wl *WorkspaceLoader) mergeExternal(byID map[int]*Issue, root string, patterns []string, known map[string]bool) []LoadProblem {
	var problems []LoadProblem
	for name, dir := range FindExternalIssuesDirs(root, patterns...) {
		if known[filepath.Clean(dir)] {
			continue // already covered as a checkout of this repository
		}
		issues, probs, err := loadAllIssues(dir, false)
		problems = append(problems, probs...)
		if err != nil {
			problems = append(problems, LoadProblem{Dir: dir, Err: err})
			continue
		}
		for _, iss := range issues {
			src := sourceOf(iss, name, dir, Touch{Changed: newestFileIn(filepath.Join(dir, strconv.Itoa(iss.ID)))})
			if existing, ok := byID[iss.ID]; ok {
				existing.Sources = append(existing.Sources, src)
			} else {
				c := iss
				c.SourceDir = dir
				c.Worktree = name
				c.Sources = []IssueSource{src}
				byID[iss.ID] = &c
			}
		}
	}
	return problems
}

// resolveOwners picks the winning source for each issue and returns the issues
// sorted by ID.
//
// The rule: the most recent real change wins, and the main checkout wins ties.
// "Real change" is a commit date or the mtime of an uncommitted edit, never the
// mtime of a checked-out file — git stamps those at checkout time, so they say
// when a worktree was created rather than when anyone touched the issue.
func resolveOwners(byID map[int]*Issue) []Issue {
	result := make([]Issue, 0, len(byID))
	for _, iss := range byID {
		sort.SliceStable(iss.Sources, func(i, j int) bool {
			if (iss.Sources[i].Name == "") != (iss.Sources[j].Name == "") {
				return iss.Sources[i].Name == "" // main leads
			}
			return iss.Sources[i].Name < iss.Sources[j].Name
		})
		best := 0
		for i := 1; i < len(iss.Sources); i++ {
			current, winner := iss.Sources[i], iss.Sources[best]
			if current.Dirty != winner.Dirty {
				if current.Dirty {
					best = i
				}
				continue
			}
			if current.Changed.After(winner.Changed) {
				best = i
			}
		}
		iss.SwitchSource(best)
		result = append(result, *iss)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

// attributeOwnership records which worktree ended up owning each issue.
func attributeOwnership(issues []Issue, ws *Workspace) {
	owned := make(map[string][]int)
	for _, iss := range issues {
		if iss.Worktree != "" {
			owned[iss.Worktree] = append(owned[iss.Worktree], iss.ID)
		}
	}
	for i := range ws.Worktrees {
		ws.Worktrees[i].Owned = owned[ws.Worktrees[i].Name]
	}
	sort.Slice(ws.Worktrees, func(i, j int) bool { return ws.Worktrees[i].Name < ws.Worktrees[j].Name })
}

// watchDirs lists roots whose numeric issue directories should be watched:
// the canonical store, active worktrees, and configured external stores.
func watchDirs(mainDir string, worktrees []WorktreeInfo, extra ...string) []string {
	dirs := []string{mainDir}
	for _, wt := range worktrees {
		dirs = append(dirs, wt.Dir)
	}
	dirs = append(dirs, extra...)
	return uniquePaths(dirs)
}

// worktreeWatchRoots keeps one cheap root watch per checkout. Idle worktrees
// are not expanded into issue-directory watches; periodic claims detect edits.
func worktreeWatchRoots(checkouts []Checkout) []string {
	roots := make([]string, 0, len(checkouts))
	for _, co := range checkouts {
		if !co.IsMain() {
			roots = append(roots, co.Path)
		}
	}
	return uniquePaths(roots)
}

func externalDirs(root string, patterns []string, known map[string]bool) []string {
	var dirs []string
	for _, dir := range FindExternalIssuesDirs(root, patterns...) {
		if !known[filepath.Clean(dir)] {
			dirs = append(dirs, dir)
		}
	}
	return uniquePaths(dirs)
}

func uniquePaths(paths []string) []string {
	seen := make(map[string]bool, len(paths))
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		path = filepath.Clean(path)
		if path == "." || seen[path] {
			continue
		}
		seen[path] = true
		out = append(out, path)
	}
	return out
}

func sourceOf(iss Issue, name, dir string, t Touch) IssueSource {
	return IssueSource{
		Name:      name,
		Dir:       dir,
		Changed:   t.Changed,
		Dirty:     t.Dirty,
		Title:     iss.Title,
		Status:    iss.Status,
		Priority:  iss.Priority,
		Labels:    iss.Labels,
		Parent:    iss.Parent,
		BlockedBy: iss.BlockedBy,
		Created:   iss.Created,
		Updated:   iss.Updated,
		Content:   iss.Content,
		Comments:  iss.Comments,
	}
}

func sortedIDs(m map[int]Touch) []int {
	ids := make([]int, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	return ids
}
