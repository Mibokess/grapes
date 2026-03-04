## Goal
When a parent issue is set to "done" in the TUI, automatically set all non-cancelled sub-issues to "done". This behavior should be configurable via the Config page.

## Context
- Status changes happen in `internal/tui/app.go` via `PickerResultMsg` and `MoveIssueMsg` handlers
- Config is defined in `internal/config/config.go` (`ViewConfig` struct)
- Settings UI is in `internal/tui/settings/settings.go` (View category)
- Writer is in `internal/data/writer.go` (`UpdateField` function)
- Parent/child relationships: `Issue.Children` (computed), `Issue.Parent` (stored)

## Acceptance Criteria
- [ ] `ViewConfig` has `AutoCloseSubs bool` field (default false)
- [ ] Settings View category shows "Auto-close sub-issues" toggle (on/off)
- [ ] When enabled and a parent is set to "done", all children not in "cancelled" status are also set to "done"
- [ ] Works via both PickerResultMsg and MoveIssueMsg code paths
- [ ] Disabled by default

## Verify
```bash
cd /projects/mboss/dev/grapes/.claude/worktrees/abstract-petting-donut && go build ./...
```

## Pass Criteria
Build succeeds. Config field exists. Settings toggle appears. Status cascade logic is in both handlers.
