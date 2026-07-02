## Goal

Discover this repo's worktrees automatically via `git worktree list --porcelain` instead of relying on the configured `Sources.Dirs` glob, so that (a) creating an issue from any worktree never reuses an ID that exists on `main` or another worktree, and (b) the TUI shows issues from all worktrees regardless of where they live. Demote `Sources.Dirs` to an optional, additive setting for non-worktree external sources.

## Context

Issue IDs are numeric directory names under `.grapes/`. Two consumers need to know about all worktrees:

- **ID assignment** — `internal/data/loader.go:209` `NextID(issuesDir, extraDirs...)` reserves `max(existing IDs) + 1`, scanning `<mainRoot>/.grapes` plus `FindWorktreeIssuesDirs(mainRoot, extraDirs...)`.
- **Loading/display** — `internal/data/loader.go:335` `LoadAllSources(mainDir, projectRoot, extraDirs...)` aggregates issues from worktrees for the TUI.

Both currently discover worktrees **only via `filepath.Glob`**: `internal/data/loader.go:248` `FindWorktreeIssuesDirs(projectRoot, patterns...)` never calls git. The patterns come from `Sources.Dirs` (`internal/config/config.go:169`), default `.claude/worktrees/*/.grapes` (`internal/config/config.go:188`).

Call sites that pass `cfg.Sources.Dirs`:
- `main.go:52` — `LoadAllSources` (TUI startup)
- `main.go:70` — `NextID` (CLI `grapes issue`)
- `internal/tui/app.go:81`, `internal/tui/app.go:522`, `internal/tui/app.go:559` — TUI load + refresh

`FindMainProjectRoot` (`internal/data/loader.go:190`) already resolves the shared main root from inside any worktree via `git rev-parse --git-common-dir`, falling back to the local project root when git fails.

### Current vs. desired behavior

**Current:** worktrees are only found if they sit under a configured glob (default `.claude/worktrees/*`). A worktree created elsewhere — by `Agent isolation: "worktree"` or `git worktree add ../foo` — is invisible. Result: `NextID` from two such worktrees both return `N+1` (duplicate issue), and the TUI won't display their issues. Users must manually configure `Sources.Dirs` to match their worktree layout.

**Desired:** worktrees of this repo are discovered automatically from `git worktree list --porcelain`, wherever they are — zero config. `Sources.Dirs` remains only for sources git can't know about (a *separate* repo with a custom `.grapes` name, e.g. the documented `../other-project/.potatoes`). Its default becomes empty.

### Constraints

- `git worktree list` failing or not-a-git-repo must degrade gracefully (return no git-discovered dirs; existing glob + main behavior still works), mirroring the fallback in `FindMainProjectRoot` (`loader.go:195-196`).
- `maxIDInDir` already returns 0 for missing/unreadable dirs (`loader.go:167-169`) — skip worktree paths that have no `.grapes/`.
- Git-worktree discovery and glob discovery overlap (e.g. a worktree under `.claude/worktrees/*` that a user still lists in config); results must be de-duplicated by resolved directory path.
- The main checkout appears in `git worktree list` too; scanning it as a "worktree source" must not duplicate the main issues that `LoadAllSources` already loads from `mainDir`.
- Keep the change surgical: add discovery, don't restructure the `flock` in `NextID` (it already correctly serializes across worktrees via the shared main `.grapes/.lock`).

## Approach

1. Add `FindGitWorktreeGrapesDirs(mainRoot string) map[string]string` in `internal/data/loader.go`: run `git worktree list --porcelain` (with `cmd.Dir = mainRoot`), parse the `worktree <path>` lines, and for each `<path>` whose `<path>/.grapes` exists as a directory, add `base(path) -> <path>/.grapes`. Return an empty map on any git error.
2. Fold git-discovered dirs into both scan paths so every consumer sees them:
   - `NextID`: include `FindGitWorktreeGrapesDirs(mainRoot)` alongside the existing `FindWorktreeIssuesDirs(mainRoot, extraDirs...)` in the max-ID loop, de-duplicated by directory path.
   - `LoadAllSources`: include git-discovered worktree dirs alongside the glob results, excluding the entry that resolves to `<mainRoot>/.grapes` (already loaded as the main source), de-duplicated by path.
3. Change the default `Sources.Dirs` to `[]string{}` in `config.Defaults()` (`internal/config/config.go:187-189`) and update the doc comment on `SourcesConfig.Dirs` (`internal/config/config.go:169-173`) to say worktrees are auto-discovered and `Dirs` is only for additional non-worktree sources.

## Acceptance Criteria

- [x] `FindGitWorktreeGrapesDirs(mainRoot)` exists, parses `git worktree list --porcelain`, returns each worktree's existing `.grapes/` dir, and returns an empty map (no error propagated) when the dir is not a git repo.
- [x] `NextID` skips past IDs held in **any** git worktree's `.grapes/`, even one located outside `.claude/worktrees/*` and absent from `Sources.Dirs`.
- [x] `LoadAllSources` returns issues from git-discovered worktrees, without duplicating main's own issues and without duplicating a worktree that also matches a `Sources.Dirs` glob.
- [x] Default `Sources.Dirs` is empty; the `SourcesConfig.Dirs` doc comment reflects auto-discovery.
- [x] New test: create a real git repo, `git worktree add` a worktree at a path **not** matching `.claude/worktrees/*`, seed an issue ID in its `.grapes/`, and assert `NextID` (from both main and the worktree) skips past it with default (empty) config.
- [x] New test: `LoadAllSources` picks up issues from that git-discovered worktree.
- [x] Existing tests in `internal/data/nextid_test.go` and `internal/data/*_test.go` still pass; a still-configured `Sources.Dirs` glob keeps working (additive).

## Verify

```bash
cd /home/mboss/dev/grapes
go build ./...
go test ./internal/data/... ./internal/config/...
```

## Pass Criteria

`go build ./...` exits 0. `go test ./internal/data/... ./internal/config/...` passes, including the new git-worktree collision and loading tests. With empty `Sources.Dirs`, two worktrees created off the same `main` and each running `grapes issue` receive distinct IDs, and the TUI lists issues from both.
