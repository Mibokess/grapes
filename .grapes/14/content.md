Build the filter menu that opens when the user presses `f`. This is the entry point for adding filters.

## Design

A popup menu listing available filter categories:

```
┌─ Add Filter ──────────────┐
│  ▸ Status                  │
│    Priority                │
│    Label                   │
│    Has sub-issues          │
└────────────────────────────┘
```

## Location

`internal/tui/filter/menu.go`

## Behavior

- Opens as an overlay on `f` keypress
- `j/k` or `↑/↓` to navigate categories
- `Enter` to select a category → opens the multi-select picker (#13) for that field
- `Esc` to close without doing anything
- Show indicator if a category already has active filters (e.g. dot or count)
- For "Has sub-issues" — toggle directly (yes/no/any) instead of opening a picker

## Flow

1. User presses `f` → filter menu opens
2. User selects "Status" → multi-select picker opens with all statuses
3. User toggles desired statuses, presses Enter
4. Filter menu closes, filter bar updates, issues are filtered
5. User presses `f` again → menu shows dot next to "Status" indicating active filter

## Labels special case

The label picker should dynamically collect all labels from the current issue set
rather than having a hardcoded list. Use a helper to extract unique labels from all issues.
