"""List/table screen."""

from __future__ import annotations

from textual.app import ComposeResult
from textual.screen import Screen
from textual.widgets import Header, Footer, DataTable

from grapes_tui.data import (
    STATUS_LABELS,
    PRIORITY_LABELS,
    load_all_issues,
)


class ListScreen(Screen):
    """Sortable list view of all issues."""

    BINDINGS = [
        ("b", "switch_board", "Board View"),
        ("r", "refresh", "Refresh"),
        ("q", "quit", "Quit"),
    ]

    DEFAULT_CSS = """
    ListScreen {
        layout: vertical;
    }
    ListScreen DataTable {
        height: 1fr;
    }
    """

    def compose(self) -> ComposeResult:
        yield Header()
        table = DataTable(cursor_type="row")
        table.add_columns("ID", "Title", "Status", "Priority", "Assignee", "Labels")
        yield table
        yield Footer()

    def on_mount(self) -> None:
        self._load_data()

    def _load_data(self) -> None:
        table = self.query_one(DataTable)
        table.clear()
        issues = load_all_issues()
        for issue in issues:
            table.add_row(
                f"#{issue.id}",
                issue.title,
                STATUS_LABELS.get(issue.status, issue.status),
                PRIORITY_LABELS.get(issue.priority, issue.priority),
                issue.assignee or "—",
                ", ".join(issue.labels),
                key=str(issue.id),
            )

    def on_data_table_row_selected(self, event: DataTable.RowSelected) -> None:
        if event.row_key and event.row_key.value:
            issue_id = int(event.row_key.value)
            self.app.push_screen("detail", {"issue_id": issue_id})

    def action_switch_board(self) -> None:
        self.app.switch_screen("board")

    def action_refresh(self) -> None:
        self._load_data()

    def action_quit(self) -> None:
        self.app.exit()
