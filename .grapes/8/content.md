The board and list screens handle `r` (refresh) differently. The list screen reloads data in-place, but the board screen replaces itself with a new instance, losing scroll position.

## Comparison

```python
# list.py:67-68 — reloads data in-place (good)
def action_refresh(self) -> None:
    self._load_data()

# board.py:210-211 — replaces entire screen (bad)
def action_refresh(self) -> None:
    self.app.switch_screen("board")
```

The board does this because its columns are built in `compose()`, which only runs once. There's no `_load_data()` equivalent that can rebuild the column widgets.

## Fix

Add a `_rebuild()` method to `BoardScreen` that:
1. Removes existing `StatusColumn` widgets
2. Calls `load_all_issues()`
3. Creates and mounts new `StatusColumn` widgets

```python
def action_refresh(self) -> None:
    container = self.query_one("#board-container")
    container.remove_children()
    issues = load_all_issues()
    for status in STATUSES:
        col_issues = [i for i in issues if i.status == status]
        container.mount(StatusColumn(status, col_issues))
```

This preserves the screen instance and avoids the scroll-to-top reset.
