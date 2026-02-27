The spec in `idea.md:116` describes the list view as a "sortable/filterable table", but `screens/list.py` currently just dumps all issues into a `DataTable` with no way to filter or sort.

## Current state

`ListScreen` has three keybindings: `b` (board), `r` (refresh), `q` (quit). No interaction with the data itself.

```python
# list.py:19-23 — no filter/sort bindings
BINDINGS = [
    ("b", "switch_board", "Board View"),
    ("r", "refresh", "Refresh"),
    ("q", "quit", "Quit"),
]
```

Data is loaded in ID order and that's it:

```python
# list.py:44-57
def _load_data(self) -> None:
    table = self.query_one(DataTable)
    table.clear()
    issues = load_all_issues()
    for issue in issues:
        table.add_row(...)
```

## Proposed features

1. **`/` — Quick filter**: Open a text input at the top, filter rows by title in real time
2. **`s` — Status filter**: Cycle through `all → backlog → todo → in_progress → done → cancelled`
3. **`p` — Priority sort**: Toggle between ID order and priority order (urgent first)
4. **Column click sorting**: Textual's `DataTable` supports `sort_key` on columns — wire it up

## Implementation notes

Textual's `DataTable` doesn't support removing individual rows efficiently. The pattern used in the Textual docs is to `table.clear()` and re-add filtered rows. With `PRIORITY_ORDER` already defined in `data.py:25`, sorting is straightforward:

```python
issues.sort(key=lambda i: PRIORITY_ORDER.get(i.priority, 99))
```
