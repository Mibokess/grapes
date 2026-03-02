Wire up the filter system into the existing list view.

## Changes to list view

1. Add `FilterSet` to the list model
2. Replace `filteredIssues()` to use `FilterSet.Matches()` instead of title-only matching
3. Move the existing text input filter query into `FilterSet.TextQuery`
4. Add the filter bar to the view layout (between header and table)
5. Handle `f` keypress to open filter menu overlay
6. Handle `FilterChangedMsg` to update the filter state and rebuild the table
7. Update the issue count in header to show filtered count (e.g. "12 of 45 issues")

## Key bindings

- `f` — open filter menu (new)
- `/` — text search (existing, now sets FilterSet.TextQuery)
- Keep existing sort keys (`o`, `O`, column clicks) working alongside filters

## Messages

Add to `internal/tui/common/messages.go`:
- `FilterChangedMsg{FilterSet}` — broadcast when filters change

## Layout

```
Issues (12 of 45)                                  [sort info]
Filters: [Status: In Progress] [Priority: High]    ✕ Clear
──────────────────────────────────────────────────────────────
 ID │ Title          │ Status      │ Priority │ ...
```

The filter bar consumes vertical space only when filters are active.
Adjust table height calculation to account for the bar.
