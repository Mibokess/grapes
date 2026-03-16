## Goal
Re-render issue content when the terminal is resized so text wrapping and layout update correctly.

## Context
- `internal/tui/detail/detail.go`: `SetSize()` only updates viewport dimensions but does not call `renderIssue()`, so markdown text stays wrapped at the original width.
- The list view (`list.go`) and board view (`board.go`) already rebuild their layouts in `SetSize()`.

## Acceptance Criteria
- [x] `SetSize()` calls `renderIssue()` to re-wrap content at the new width
- [x] Click lines and click zones are recalculated on resize
- [x] All existing tests pass

## Verify
```bash
go test ./internal/tui/detail/ -v -count=1
```

## Pass Criteria
All tests PASS.
