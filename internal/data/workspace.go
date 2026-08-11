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
	WatchDirs []string
	Problems  []LoadProblem

	// AttributionErr says why worktree attribution is off, when it is. Being
	// outside a git repository is the common, unremarkable case, so this is
	// shown where a user would look for worktrees rather than reported as a
	// problem on every load.
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

	// The main checkout is the baseline: every issue is read from it once.
	mainIssues, problems, err := LoadAllIssues(mainDir)
	if err != nil {
		return ws, err
	}
	ws.Problems = append(ws.Problems, problems...)

	byID := make(map[int]*Issue, len(mainIssues))
	var mainTouched map[int]Touch
	for _, cl := range claims {
		if cl.Checkout.IsMain() {
			mainTouched = cl.Touched
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
	ws.Problems = append(ws.Problems, wl.mergeExternal(byID, mainRoot, opts.ExtraDirs, knownDirs)...)

	ws.Issues = resolveOwners(byID)
	attributeOwnership(ws.Issues, &ws)
	ws.WatchDirs = watchDirs(mainDir, ws.Worktrees)
	RewireRelationships(ws.Issues)
	return ws, nil
}

// resolveLayout works out where the canonical issue directory lives and where
// the issue directory sits inside any checkout of this repository.
func resolveLayout(issuesDir string) (mainDir, mainRoot, grapesRel string, err error) {
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
	issues, problems, err := LoadAllIssues(dir)
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
	ws.Problems = append(ws.Problems, wl.mergeExternal(byID, ProjectRoot(dir), opts.ExtraDirs, map[string]bool{filepath.Clean(dir): true})...)

	ws.Issues = resolveOwners(byID)
	attributeOwnership(ws.Issues, &ws)
	ws.WatchDirs = watchDirs(dir, ws.Worktrees)
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
		issues, probs, err := LoadAllIssues(dir)
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
		for i := range iss.Sources {
			if iss.Sources[i].Changed.After(iss.Sources[best].Changed) {
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

// watchDirs lists the directories live reload needs to follow: the canonical
// store, plus the worktrees that are actually working on something. Watching
// every worktree would mean thousands of descriptors for copies that never
// change.
func watchDirs(mainDir string, worktrees []WorktreeInfo) []string {
	dirs := []string{mainDir}
	for _, wt := range worktrees {
		dirs = append(dirs, wt.Dir)
	}
	return dirs
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
