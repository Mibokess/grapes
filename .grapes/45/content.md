## Goal
Preserve scroll position in the detail view when a file-change refresh (`RefreshMsg`) occurs, instead of resetting to the top.

## Context
- `internal/tui/app.go:547-553` — on `RefreshMsg`, the detail view is recreated with `detail.New()`, which builds a fresh viewport starting at scroll offset 0.
- `internal/tui/detail/detail.go:54-70` — `New()` creates a new viewport from scratch.
- The viewport's `SetContent()` method (`charm.land/bubbles/v2/viewport`) does NOT reset `yOffset` — it only clamps it if the new content is shorter. This means we can re-render content in-place.
- `SetWorktreeNames()` (detail.go:43-52) already demonstrates the pattern: it calls `renderIssue()` and `viewport.SetContent()` without recreating the viewport.

## Approach
1. Add an `UpdateIssue(issue, allIssues)` method to `detail.Model` that re-renders content and updates the viewport in-place (preserving scroll offset).
2. In `app.go`, call `m.detail.UpdateIssue(...)` instead of `detail.New(...)` when handling `RefreshMsg`.

## Acceptance Criteria
- [x] `detail.Model` has an `UpdateIssue` method that re-renders content without resetting scroll
- [x] `app.go` RefreshMsg handler uses `UpdateIssue` instead of `detail.New`
- [x] Scroll position is preserved when issue content changes (unless content got shorter, in which case it clamps)
- [x] Click zones and click lines are rebuilt correctly after update
- [x] Existing tests pass

## Verify
```bash
cd /projects/mboss/dev/grapes/.claude/worktrees/atomic-hopping-giraffe && go build ./... && go test ./...
```

## Pass Criteria
All tests pass. Build succeeds. The `RefreshMsg` handler no longer calls `detail.New()` for the current detail view.
