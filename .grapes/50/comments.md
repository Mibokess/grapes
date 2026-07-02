### 2026-07-02T00:00
[FINDINGS] Root cause is glob-only worktree discovery in `FindWorktreeIssuesDirs` (`internal/data/loader.go:248`), consumed by both `NextID` (:209) and `LoadAllSources` (:335). Git worktrees created outside the configured glob (default `.claude/worktrees/*/.grapes`) — e.g. `Agent isolation: "worktree"` or `git worktree add ../foo` — are never scanned, causing duplicate IDs and missing issues in the TUI.

The `flock` on `<mainRoot>/.grapes/.lock` already resolves to the same main across all worktrees (via `git rev-parse --git-common-dir`), so cross-worktree atomicity is already correct. Only discovery is broken.

### 2026-07-02T00:01
[DECISION] Scope widened (per user): replace glob-based worktree discovery with automatic `git worktree list --porcelain` discovery for both `NextID` and `LoadAllSources`, and demote `Sources.Dirs` to an optional additive setting (default empty) for non-worktree external sources only (e.g. a separate repo like `../other-project/.potatoes`, which git worktree list cannot know about). Config globs are kept as a superset-reach fallback, not removed.

### 2026-07-02T00:02
[VERIFY] `go build ./...` OK. `go test ./...` all packages pass, including new tests (real git worktrees, PASS not SKIP):
- `TestFindGitWorktreeGrapesDirs` — parses porcelain output; non-git dir → empty map.
- `TestNextIDGitWorktree` — worktree outside `.claude/worktrees/*`, empty config; NextID from main and from worktree both skip the worktree's ID.
- `TestLoadAllSourcesGitWorktree` — TUI load picks up the git-discovered worktree's issue.

[VERIFY] End-to-end with the built CLI: temp repo (main IDs 1,2) + two worktrees at sibling paths (not under `.claude/worktrees/*`), no `sources.dirs` config. `grapes issue`: wa→3, wb→4 (no collision), wa→5, main→6. Each got a distinct global ID; before the fix wa and wb would both have returned 3.

[DONE] Files changed:
- `internal/data/loader.go` — added `FindGitWorktreeGrapesDirs`; folded git discovery into `NextID` and `LoadAllSources` (de-dup by path; main/current excluded).
- `internal/config/config.go` — default `Sources.Dirs` now empty; doc comment describes auto-discovery.
- `internal/data/nextid_test.go` — git-worktree helper + 3 new tests.
