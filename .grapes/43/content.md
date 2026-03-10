## Goal

Make the directories grapes scans for worktree `.grapes/` folders configurable via `config.toml`, so users with worktrees at non-default locations can have their issues discovered.

## Context

- `internal/data/loader.go:247` — `FindWorktreeIssuesDirs()` hardcodes `.claude/worktrees` as the only scan path
- `internal/config/config.go` — `Config` struct, loaded from `.grapes/config.toml`
- Callers of `FindWorktreeIssuesDirs`: `NextID()`, `LoadAllSources()`, `NewModel()`, reload handler in `app.go`
- Default `.claude/worktrees` must remain as a built-in path for backward compat

## Acceptance Criteria

- [ ] New `[sources]` section in config with `worktree_dirs` string list
- [ ] Each entry is scanned for `*/.grapes/` subdirectories (same as `.claude/worktrees`)
- [ ] Paths can be absolute or relative to project root
- [ ] Default `.claude/worktrees` is always scanned (no config needed for current behavior)
- [ ] `NextID` sees issues from extra dirs for correct ID allocation
- [ ] `LoadAllSources` merges issues from extra dirs
- [ ] File watcher watches extra dirs
- [ ] Existing tests still pass
- [ ] New test for `FindWorktreeIssuesDirs` with extra dirs

## Verify

```bash
go test ./...
```

## Pass Criteria

All tests pass. A `config.toml` with `[sources] worktree_dirs = ["/some/path"]` causes that path to be scanned for `.grapes/` subdirectories.
