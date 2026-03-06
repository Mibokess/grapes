## Goal
Fix the board view so that clicking the "+N more" indicator at the bottom of a column scrolls to reveal hidden issues instead of opening the first hidden issue's detail view.

## Context
- File: `internal/tui/board/board.go`
- `cardAt()` (line 612) maps screen coordinates to a column/row index. It checks `ri < len(col.issues)` but does not check whether `ri` falls within the **visible** range (`scrollOff` to `scrollOff + maxCards`).
- The "+N more" indicator is rendered at line 460-463 in `renderColumn()`, occupying the y-position immediately after the last visible card.
- When you click on "+N more", `cardAt()` computes `ri = yOffset/cardH + scrollOff`, which maps to the first hidden card (index `endIdx`). Since `endIdx < len(col.issues)`, `cardAt` returns `ok=true`.
- The click handler at line 206 then selects that card and sets up a mouse-down, and the release handler at line 265-268 opens its detail view.
- The same issue exists for the "↑ N more" indicator at the top when scrolled — line 643-648 returns `ok=false` for `yOffset == 0`, which correctly rejects clicks on it, but only for the active/scrolled column.

## Acceptance Criteria
- [ ] Clicking the "+N more" indicator at the bottom of a column scrolls down to reveal more issues (instead of opening a hidden issue).
- [ ] Clicking the "↑ N more" indicator at the top of a column scrolls up to reveal earlier issues.
- [ ] `cardAt()` returns `ok=false` for coordinates that land on either "more" indicator.
- [ ] Existing card click, drag-and-drop, and scroll behavior is unchanged.
- [ ] A test verifies that clicking the "+N more" area does not produce an `OpenDetailMsg`.

## Verify
```bash
cd /home/mboss/dev/grapes/.claude/worktrees/mutable-dreaming-creek && go test ./internal/tui/board/... -v -run TestBoard
```

## Pass Criteria
All board tests pass, including a new test for clicking the "+N more" indicator area.
