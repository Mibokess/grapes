Replace the current title-only text filter with a structured, composable filter system.

## Goals

- Filter issues by **Status**, **Priority**, **Labels**, and **has sub-issues**
- Keep the existing `/` text search for title/content matching
- Add a chip-based filter bar showing active filters
- Add a filter menu overlay with multi-select pickers per field
- Works in both **list view** and **board view**
- Filters are AND'd across categories, OR'd within a category

## UX Design

```
Filters: [Status: In Progress, Done] [Priority: High] [Label: bug]  ✕ Clear
─────────────────────────────────────────────────────────────────────────────
 ID  Title              Status       Priority  Created   Updated   Labels
```

- `f` opens filter menu overlay
- `/` keeps existing text search
- Filter menu shows categories → selecting one opens multi-select picker
- Active filters shown as chips, removable with backspace/delete
- `Esc` closes menus, chips navigable with ←/→

## Architecture

- `FilterSet` type holding selected values per field + text query
- `FilterSet.Matches(issue)` method for applying filters
- Filter menu + chip bar as Bubble Tea models
- `FilterChangedMsg` for views to react to filter updates
