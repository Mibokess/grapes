Build the persistent filter bar that displays active filters as chips below the view header.

## Design

```
Filters: [Status: In Progress, Done] [Priority: High] [Label: bug]  ✕ Clear all
```

When no filters are active, the bar is hidden (no wasted vertical space).

## Location

`internal/tui/filter/bar.go`

## Behavior

- Renders horizontally, one chip per active filter category
- Each chip shows the category name and selected values
- Truncate long value lists (e.g. "Status: In Progress +2 more")
- "✕ Clear all" at the end to reset all filters
- Chips should use the same colors/styles as the rest of the TUI (status colors, priority colors, label colors)

## Interaction (stretch goal)

- `←/→` to navigate between chips when bar is focused
- `Backspace/Delete` to remove the focused chip
- `Enter` on a chip to re-open that category's picker for editing
- This navigation could be triggered by a key like `F` (shift-f) to "edit filters"

For v1, the bar can be display-only (no navigation), with `f` to modify and a clear-all shortcut.

## Sizing

- Should respect terminal width
- If filters don't fit in one line, truncate with "..." or wrap to second line
- Calculate height dynamically so the list/board view adjusts its available space
