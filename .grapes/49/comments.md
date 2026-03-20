### 2026-03-20T00:00
[STARTED] Build fails on Windows with `undefined: syscall.Flock`. Will extract file locking into platform-specific files using Go build tags.

### 2026-03-20T00:01
[DONE] Fixed by extracting file locking into platform-specific files:
- `internal/data/flock_unix.go` — `syscall.Flock` (build tag `!windows`)
- `internal/data/flock_windows.go` — `windows.LockFileEx`/`UnlockFileEx`
- `internal/data/loader.go` — removed `syscall` import, calls `flockExclusive()`/`flockUnlock()`
- `go.sum` / `go.mod` — upgraded `golang.org/x/sys` (already a transitive dep)
- Verified: `go build ./...` passes on Windows and `GOOS=linux go build ./...` passes too.
