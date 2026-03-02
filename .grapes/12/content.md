Create the core filter data type and matching logic. This is the foundation everything else builds on.

## Location

New file: `internal/tui/filter/filter.go`

## FilterSet struct

```go
type FilterSet struct {
    Statuses    []data.Status
    Priorities  []data.Priority
    Labels      []string
    HasChildren *bool
    TextQuery   string  // existing title search, also match content
}
```

## Key methods

- `Matches(issue data.Issue) bool` — returns true if issue passes all active filters
  - AND across categories (status AND priority AND labels)
  - OR within a category (status=done OR status=in_progress)
  - TextQuery does case-insensitive substring match on title + content
- `IsEmpty() bool` — returns true if no filters are active
- `Clear()` — resets all filters
- `Summary() string` — human-readable summary for the filter bar
- `Toggle(field, value)` — add/remove a value from a filter category

## Matching rules

- Empty category = no filtering on that field (pass all)
- Status: issue.Status must be in Statuses slice
- Priority: issue.Priority must be in Priorities slice
- Labels: issue must have at least one label in the Labels slice
- HasChildren: if set, issue.Children must be non-empty (or empty if false)
- TextQuery: case-insensitive substring in title OR content

## Tests

Write table-driven tests covering combinations of filters and edge cases.
