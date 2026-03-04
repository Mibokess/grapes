## Goal
Fix keyboard navigation in the settings screen so arrow keys and Esc behave intuitively.

## Context
- File: `internal/tui/settings/settings.go` — `updateNavigating()` (line 312)
- File: `internal/tui/common/keys.go` — `SettingsKeys` struct (line 227)
- Currently only Tab/Enter switch between categories and fields panes
- Left/Right arrow keys (and h/l) do nothing
- Esc always exits the settings screen, even when focused on the fields pane

## Acceptance Criteria
- [ ] `→` and `l` move focus from categories pane to fields pane
- [ ] `←` and `h` move focus from fields pane to categories pane
- [ ] Esc when focused on fields pane moves focus back to categories pane
- [ ] Esc when focused on categories pane exits settings screen (existing behavior)
- [ ] Esc while editing a field still cancels the edit (existing behavior unchanged)
