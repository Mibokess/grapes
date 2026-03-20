## Goal
Fix `go build` on Windows. Currently fails because `internal/data/loader.go` uses `syscall.Flock` and `syscall.LOCK_EX`/`LOCK_UN`, which only exist on Linux/Unix.

## Context
- File: `internal/data/loader.go`, lines 222–226
- `syscall.Flock` is used to acquire an exclusive file lock when generating the next issue ID (`NextID` function)
- Windows equivalent is `LockFileEx`/`UnlockFileEx` from the Windows API via `golang.org/x/sys/windows` or `syscall.LockFileEx`

## Approach
Extract file locking into platform-specific files using Go build tags:
- `internal/data/flock_unix.go` — existing `syscall.Flock` logic (build tag `//go:build !windows`)
- `internal/data/flock_windows.go` — Windows `LockFileEx`/`UnlockFileEx` implementation
- `loader.go` calls a shared `lockFile(fd)` / `unlockFile(fd)` function

## Acceptance Criteria
- [x] `go build ./...` succeeds on Windows
- [x] `go build ./...` still succeeds on Linux (no regressions)
- [x] File locking in `NextID` works correctly on both platforms

## Verify
```bash
go build ./...
```

## Pass Criteria
Build completes with exit code 0 on Windows.
