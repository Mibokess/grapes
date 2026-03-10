## Goal
Make j/k (up/down) wrap across columns in the board view so navigation moves between issues rather than being confined to the current column.

## Context
- File: `internal/tui/board/board.go`, lines 170-180 (Up/Down key handling)
- Currently, pressing j at the bottom of a column does nothing; pressing k at the top does nothing
- The user must use h/l to switch columns, making it feel like navigation moves between columns rather than issues

## Acceptance Criteria
- [ ] Pressing j at the last issue in a column moves to the first issue in the next column
- [ ] Pressing k at the first issue in a column moves to the last issue in the previous column
- [ ] j at the last issue of the last column does nothing (no wrap around)
- [ ] k at the first issue of the first column does nothing (no wrap around)
- [ ] Scroll state (scrollRow, scrollCol) updates correctly after wrapping

## Verify
```bash
go test ./internal/tui/board/ -run TestBoard -v
```

## Pass Criteria
All board tests pass, including new wrap-navigation tests.
