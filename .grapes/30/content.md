## Goal
Add integration tests verifying that the TUI correctly reflects changes after file writes and RefreshMsg processing. Currently, changes made within the TUI (status changes, priority changes via picker, drag-drop) rely on fsnotify to trigger a refresh, but there are no tests verifying this roundtrip.

## Context
- `internal/tui/app.go`: Main app model, handles MoveIssueMsg, PickerResultMsg, RefreshMsg
- `internal/data/writer.go`: UpdateField uses sed to modify meta.toml on disk
- `internal/data/loader.go`: LoadAllIssues/LoadAllSources reloads from disk
- Write operations return `nil` (not RefreshMsg), relying on fsnotify to trigger refresh
- No existing app-level tests exist (only board/list/detail component tests)

## Acceptance Criteria
- [ ] Tests verify MoveIssueMsg writes status to disk and RefreshMsg picks it up
- [ ] Tests verify PickerResultMsg writes status/priority to disk and RefreshMsg picks it up
- [ ] Tests verify subissue status change is visible in parent detail view after refresh
- [ ] Tests verify multiple rapid changes are all reflected after a single RefreshMsg
- [ ] All tests use real temp directories with actual file I/O (not mocks)
- [ ] Tests pass: `go test ./internal/tui/ -run TestApp_Refresh`

## Verify
```bash
go test ./internal/tui/ -run TestApp_Refresh -v
```

## Pass Criteria
All TestApp_Refresh* tests pass with no failures.
