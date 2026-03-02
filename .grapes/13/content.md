Build a reusable multi-select picker overlay component for choosing filter values.

## Context

The existing status/priority pickers (`ShowPickerMsg` / `PickerResultMsg`) are single-select.
Filters need multi-select (e.g. select both "In Progress" AND "Done" statuses).

## Design

A popup overlay that shows a list of options with checkboxes:

```
┌─ Status ──────────────┐
│  [✓] Backlog           │
│  [✓] In Progress       │
│  [ ] Todo              │
│  [ ] Done              │
│  [ ] Cancelled         │
├────────────────────────┤
│  Enter: apply  Esc: cancel │
└────────────────────────┘
```

## Location

`internal/tui/filter/picker.go`

## Behavior

- `j/k` or `↑/↓` to navigate options
- `Space` to toggle selection
- `Enter` to confirm and apply selections
- `Esc` to cancel without changes
- Pre-select currently active filter values when opening
- Show status/priority icons and colors matching the rest of the TUI

## Interface

```go
type MultiPickerModel struct { ... }

func NewMultiPicker(title string, options []PickerOption, selected []string) MultiPickerModel
```

Where `PickerOption` has a key, display label, and optional style.

## Messages

- `MultiPickerResultMsg{Field string, Selected []string}` — emitted on confirm
- Cancelled = no message emitted, just close overlay
