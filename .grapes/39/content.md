## Goal

Add mouse support to all overlay popups that currently only support keyboard navigation: filter menu, filter multi-picker, and add a clickable apply/cancel footer to the label picker and filter multi-picker.

## Context

### Current state
- **Picker** (`internal/tui/picker/picker.go`) — Has full mouse support: click to select, click outside to cancel, mouse motion to move cursor.
- **Label Picker** (`internal/tui/labelpicker/labelpicker.go`) — Has mouse support for toggling labels and clicking outside to cancel, but no way to apply (confirm) via mouse.
- **Filter Menu** (`internal/tui/filter/menu.go`) — **No mouse support at all.** Only handles `tea.KeyPressMsg`.
- **Filter MultiPicker** (`internal/tui/filter/picker.go`) — **No mouse support at all.** Only handles `tea.KeyPressMsg`.
- **App routing** (`internal/tui/app.go`, lines 325-343) — Mouse events are routed to `picker` and `labelPicker` overlays, but **not** to `filterMenu` or `filterPicker` overlays. Mouse events fall through to the background screen.

### Gaps
1. Filter Menu: no `tea.MouseClickMsg` or `tea.MouseMotionMsg` handlers, no `ScreenX/ScreenY` fields.
2. Filter MultiPicker: no mouse handlers, no `ScreenX/ScreenY` fields.
3. App.go: mouse routing skips `filterMenu` and `filterPicker`.
4. Label Picker: no clickable "Apply" area — mouse users can toggle but can't confirm.
5. Filter MultiPicker: same as #4 once mouse support is added.

## Acceptance Criteria

- [ ] Filter Menu handles `tea.MouseClickMsg`: clicking a menu item selects it (same as Enter); clicking outside cancels.
- [ ] Filter Menu handles `tea.MouseMotionMsg`: hovering moves the cursor.
- [ ] Filter MultiPicker handles `tea.MouseClickMsg`: clicking an option toggles it (same as Space); clicking outside cancels.
- [ ] Filter MultiPicker handles `tea.MouseMotionMsg`: hovering moves the cursor.
- [ ] App.go routes mouse events to `filterMenu` and `filterPicker` when active (before falling through to background screen).
- [ ] App.go computes centered screen positions for filter overlays (like `updatePickerPosition`/`updateLabelPickerPosition`).
- [ ] Label Picker: clicking the hint/footer area applies the selection.
- [ ] Filter MultiPicker: clicking the hint/footer area applies the selection.
- [ ] All existing tests pass.
- [ ] New tests cover mouse click select, mouse click outside cancel, and mouse motion cursor tracking for filter menu and filter multi-picker.

## Verify

```bash
cd /projects/mboss/dev/grapes && go test ./internal/tui/...
```

## Pass Criteria

All tests pass including new mouse interaction tests for filter menu, filter multi-picker, and label picker apply-by-click.
