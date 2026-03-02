Wire up the filter system into the board (kanban) view.

## Changes to board view

1. Add `FilterSet` to the board model
2. Apply `FilterSet.Matches()` in `groupByStatus()` to filter cards
3. Add the filter bar between header and board columns
4. Handle `f` keypress to open filter menu overlay
5. Handle `FilterChangedMsg` to update filters and re-render

## Board-specific considerations

- **Status filter on board**: When filtering by status, hide empty columns entirely
  (e.g. if filtering to "In Progress" only, don't show empty Backlog/Todo/Done/Cancelled columns)
- **Column counts**: Update card counts per column to reflect filtered count
- **Empty state**: If all cards in all columns are filtered out, show "No issues match filters"
- **Drag-and-drop**: When dragging a card to a new status column, the card should remain
  visible even if the target status isn't in the active status filter (auto-add the status
  to the filter, or temporarily bypass the status filter for the moved card)

## Filter bar placement

```
Board                                               [sort info]
Filters: [Priority: High, Urgent] [Label: bug]     ✕ Clear
┌─ Backlog ──┐ ┌─ Todo ──────┐ ┌─ In Progress ┐ ...
```

## Text search on board

Also add `/` text search to the board view (it currently has none).
This filters cards by title, same as list view.
