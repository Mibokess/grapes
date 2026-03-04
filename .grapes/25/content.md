## Goal
Add mouse click support to the settings screen so users can click on categories and fields to select them.

## Context
- `internal/tui/settings/settings.go` — the settings screen model
- `Update()` (line 163) only handles `tea.KeyPressMsg`, ignoring all mouse events
- Other screens (board, list) handle `tea.MouseClickMsg` and `tea.MouseWheelMsg` for click and scroll
- App router (`internal/tui/app.go` line 653) delegates all messages including mouse to `settings.Update()`
- The settings view has two panes: categories (left, 18 chars wide) and fields (right)
- View renders with 1 line of top padding, then rows of `left + " │ " + right`
- The separator `│` is at roughly x=18

### What needs to happen
1. Handle `tea.MouseClickMsg` in `settings.Update()`:
   - Left click on a category row (x < separator position): set `catIdx`, switch `focus` to `paneCategories`
   - Left click on a field row (x >= separator position): set `fieldIdx`, switch `focus` to `paneFields`
   - Double-click or second click on already-selected field: enter edit mode (like pressing Enter)
2. Handle `tea.MouseWheelMsg` for scrolling through fields when the list is long

## Acceptance Criteria
- [ ] Clicking a category in the left pane selects it and shows its fields
- [ ] Clicking a field in the right pane selects it
- [ ] Clicking an already-selected field enters edit mode (for color/key fields) or cycles value (for enum fields)
- [ ] Mouse wheel scrolls through fields when the list exceeds visible height
- [ ] Existing keyboard navigation still works unchanged

## Verify
```bash
cd /projects/mboss/dev/grapes && go build ./...
```

## Pass Criteria
Project builds without errors. Manual verification: open settings, click categories and fields with mouse.
