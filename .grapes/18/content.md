Add golden file tests to cover the new filter UI states in both list and board views.

## Test cases for list view

- List with active status filter (e.g. only "in_progress" issues shown)
- List with active priority filter
- List with multiple filters active (status + priority + label)
- List with filter bar visible + text search active simultaneously
- List with filters that match no issues (empty state)
- List with filter bar truncation (many active filters in narrow terminal)

## Test cases for board view

- Board with priority filter active
- Board with status filter hiding empty columns
- Board with label filter
- Board with filters matching no issues

## Test cases for filter components

- Filter menu overlay rendering
- Multi-select picker with some items checked
- Filter bar with various chip combinations

## Approach

Follow existing golden file test patterns in:
- `internal/tui/board/board_test.go`
- `internal/tui/list/list_test.go`
- Use `testutil.go` shared helpers and sample fixtures

Regenerate with `go test ./internal/tui/... -update` after implementation.
