## Goal

Add interactive label editing to the TUI so users can add/remove labels on issues without opening an external editor or editing `meta.toml` by hand.

## Context

Labels are currently **display-only** in the TUI. They render correctly in all three views (board, list, detail) and can be used for filtering, but there is no way to modify them from within the TUI.

### What exists today

- **Display**: Labels render as colored pills in detail view (`internal/tui/detail/detail.go:252-258`), as colored text in board cards (`internal/tui/board/board.go:525-547`) and list rows (`internal/tui/list/list.go:509-513`).
- **Filtering**: A multi-select label filter picker exists (`internal/tui/filter/picker.go`), toggled via the filter menu. Uses `space`/`x` to toggle, `enter` to apply.
- **Theme**: 10 label color pairs defined in `internal/tui/common/theme.go:301-356`. `RenderLabel()` (fg only) and `RenderLabelPill()` (fg+bg) render functions at lines 548-558.
- **Data model**: `Labels []string` in `internal/data/issue.go:198`. Serialized in `internal/data/writer.go:87-92`.
- **Status/priority pickers**: The existing single-select picker (`internal/tui/picker/picker.go`) handles status and priority changes via `ShowPickerMsg` → `PickerResultMsg` flow. The app handles this in `internal/tui/app.go:503-536`.

### What's missing

1. **No label picker/editor** — No equivalent of the status/priority picker for labels.
2. **No keybinding** — No key mapped to open a label editor. Config (`config.toml`) has `board_status`, `board_priority`, `detail_status`, `detail_priority`, etc., but no `*_label` keys.
3. **No click handling** — Label pills in detail view are not clickable (unlike status/priority which have click zones at `detail.go:24`).
4. **No write path** — `data.UpdateField()` likely doesn't handle array-valued fields like labels.

### Architecture for the solution

The label editor differs from status/priority because labels are **freeform multi-select** (not a fixed enum). The filter picker (`filter/picker.go`) already implements multi-select with checkboxes — this pattern should be reused or adapted.

**Approach**: Create a label editor overlay (similar to filter picker) that:
- Lists all labels currently used across all issues (collected via `collectAllLabels()` in `app.go:1017-1031`)
- Shows checkmarks for labels the current issue has
- Allows toggling labels with `space`/`x`
- Allows typing a new label (input field at bottom)
- Applies changes on `enter`, cancels on `esc`
- Writes the updated label list back to `meta.toml`

**Message flow** (mirroring status/priority):
- `ShowLabelPickerMsg{IssueID int}` — sent by views when `l` is pressed
- `LabelPickerResultMsg{IssueID int, Labels []string}` — sent when user confirms
- `LabelPickerCancelMsg` — sent on escape

**Files to modify**:
- `internal/tui/common/messages.go` — Add new message types
- `internal/tui/picker/` or new `internal/tui/labelpicker/` — Label picker component
- `internal/tui/app.go` — Handle new messages, build label picker, write labels
- `internal/tui/board/board.go` — Send `ShowLabelPickerMsg` on keybinding
- `internal/tui/list/list.go` — Send `ShowLabelPickerMsg` on keybinding
- `internal/tui/detail/detail.go` — Send `ShowLabelPickerMsg` on keybinding + click zone for labels
- `internal/tui/settings/settings.go` — Add `board_label`, `list_label`, `detail_label` key config
- `internal/config/config.go` — Add label key fields
- `internal/data/writer.go` — Support writing label arrays via `UpdateField` or new function

## Acceptance Criteria

- [ ] Pressing a configurable key (default `L`) on an issue in board, list, or detail view opens a label picker overlay
- [ ] The picker lists all labels used across the project, with checkboxes showing which labels the current issue has
- [ ] User can toggle labels on/off with `space` or `x`
- [ ] User can type a new label name and add it
- [ ] Pressing `enter` applies the changes and writes updated labels to `meta.toml`
- [ ] Pressing `esc` cancels without changes
- [ ] Mouse click on label pills in detail view opens the label picker
- [ ] Mouse works in the picker (click to toggle, click outside to cancel)
- [ ] Label key is configurable in settings (`board_label`, `list_label`, `detail_label`)
- [ ] Existing label display, filtering, and search remain unaffected

## Verify

```bash
cd /projects/mboss/dev/grapes && go test ./...
```

## Pass Criteria

All tests pass. Manual verification: open TUI, press `L` on an issue, toggle labels, confirm they persist in `meta.toml`.
