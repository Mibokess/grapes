"""Kanban board screen."""

from __future__ import annotations

from textual.app import ComposeResult
from textual.binding import Binding
from textual.containers import HorizontalScroll, Vertical, VerticalScroll
from textual.events import MouseDown, MouseMove, MouseUp, MouseScrollDown, MouseScrollLeft, MouseScrollRight, MouseScrollUp
from textual.message import Message
from textual.screen import Screen
from textual.widgets import Header, Footer, Static, Label

from grapes_tui.data import (
    Issue,
    STATUSES,
    STATUS_LABELS,
    PRIORITY_LABELS,
    load_all_issues,
)

COLUMN_WIDTH = 40
COLUMN_STEP = COLUMN_WIDTH + 2  # column + margin


class SnapHorizontalScroll(HorizontalScroll, can_focus=False):
    """HorizontalScroll that snaps by column on mouse wheel and drag."""

    def _snap(self, direction: int) -> None:
        target = self.scroll_x + (COLUMN_STEP * direction)
        self.scroll_to(x=max(0, target), animate=True, duration=0.3)

    def on_mouse_scroll_up(self, event: MouseScrollUp) -> None:
        event.stop()
        self._snap(-1)

    def on_mouse_scroll_down(self, event: MouseScrollDown) -> None:
        event.stop()
        self._snap(1)

    def on_mouse_scroll_left(self, event: MouseScrollLeft) -> None:
        event.stop()
        self._snap(-1)

    def on_mouse_scroll_right(self, event: MouseScrollRight) -> None:
        event.stop()
        self._snap(1)

    # --- Click-drag panning (raw 1:1 movement, no animation) ---

    def on_mouse_down(self, event: MouseDown) -> None:
        self._dragging = False
        self._drag_start_x = event.screen_x
        self._mouse_pressed = True

    def on_mouse_move(self, event: MouseMove) -> None:
        if not getattr(self, "_mouse_pressed", False):
            return
        delta = self._drag_start_x - event.screen_x
        if delta != 0:
            if not self._dragging:
                self._dragging = True
                self.capture_mouse()
            self.scroll_to(x=self.scroll_x + delta, animate=False)
            self._drag_start_x = event.screen_x

    def on_mouse_up(self, event: MouseUp) -> None:
        if self._dragging:
            self.release_mouse()
        self._dragging = False
        self._mouse_pressed = False


class IssueCard(Static):
    """A clickable issue card."""

    class Selected(Message):
        """Posted when a card is selected (click or keyboard)."""

        def __init__(self, card: IssueCard) -> None:
            super().__init__()
            self.card = card
            self.issue = card.issue

    DEFAULT_CSS = """
    IssueCard {
        background: $surface;
        border: solid $primary-muted;
        padding: 0 1;
        margin-bottom: 1;
        height: auto;
        min-height: 3;
    }
    IssueCard:hover {
        border: solid $primary;
        background: $surface-lighten-1;
    }
    IssueCard.-highlight {
        border: tall $accent;
        background: $accent 15%;
    }
    IssueCard .card-id {
        color: $text-muted;
    }
    IssueCard .card-title {
        text-style: bold;
    }
    IssueCard .card-meta {
        color: $text-muted;
    }
    IssueCard .priority-urgent {
        color: $error;
        text-style: bold;
    }
    IssueCard .priority-high {
        color: #fd7e14;
    }
    IssueCard .priority-medium {
        color: $warning;
    }
    IssueCard .priority-low {
        color: $text-muted;
    }
    """

    def __init__(self, issue: Issue) -> None:
        super().__init__()
        self.issue = issue

    def compose(self) -> ComposeResult:
        i = self.issue
        yield Label(f"[b]#{i.id}[/b]  {i.title}")
        meta_parts = []
        if i.assignee:
            meta_parts.append(i.assignee)
        meta_parts.append(PRIORITY_LABELS[i.priority])
        if i.labels:
            meta_parts.append(" ".join(f"[{l}]" for l in i.labels))
        yield Label(" · ".join(meta_parts), classes="card-meta")

    def on_click(self) -> None:
        self.post_message(self.Selected(self))


class StatusColumn(Vertical):
    """A single status column in the board."""

    DEFAULT_CSS = """
    StatusColumn {
        width: 40;
        background: $panel;
        border: solid $primary-muted;
        padding: 0;
        margin: 0 1 0 0;
    }
    StatusColumn .column-header {
        background: $primary-muted;
        text-align: center;
        text-style: bold;
        padding: 0 1;
        height: 1;
    }
    """

    def __init__(self, status: str, issues: list[Issue]) -> None:
        super().__init__()
        self.status = status
        self.issues = issues

    def compose(self) -> ComposeResult:
        yield Label(
            f"{STATUS_LABELS[self.status]} ({len(self.issues)})",
            classes="column-header",
        )
        with VerticalScroll():
            for issue in self.issues:
                yield IssueCard(issue)


class BoardScreen(Screen):
    """Kanban board view."""

    BINDINGS = [
        Binding("up,k", "cursor_up", "Up", show=False),
        Binding("down,j", "cursor_down", "Down", show=False),
        Binding("left,h", "cursor_left", "Left", show=False),
        Binding("right,l", "cursor_right", "Right", show=False),
        Binding("enter", "select_card", "Open"),
        ("L", "switch_list", "List View"),
        ("r", "refresh", "Refresh"),
        ("q", "quit", "Quit"),
    ]

    DEFAULT_CSS = """
    BoardScreen {
        layout: vertical;
    }
    BoardScreen #board-container {
        height: 1fr;
        scrollbar-size-horizontal: 0;
    }
    """

    def __init__(self, **kwargs) -> None:
        super().__init__(**kwargs)
        self.cursor_col = 0
        self.cursor_row = 0

    def compose(self) -> ComposeResult:
        yield Header()
        with SnapHorizontalScroll(id="board-container"):
            issues = load_all_issues()
            for status in STATUSES:
                col_issues = [i for i in issues if i.status == status]
                yield StatusColumn(status, col_issues)
        yield Footer()

    # --- Helpers ---

    def _get_columns(self) -> list[StatusColumn]:
        return list(self.query(StatusColumn))

    def _get_cards(self, col: int) -> list[IssueCard]:
        columns = self._get_columns()
        if 0 <= col < len(columns):
            return list(columns[col].query(IssueCard))
        return []

    def _nonempty_cols(self) -> list[int]:
        return [i for i, c in enumerate(self._get_columns()) if c.query(IssueCard)]

    # --- Highlight ---

    def _update_highlight(self) -> None:
        for card in self.query(IssueCard):
            card.remove_class("-highlight")
        cards = self._get_cards(self.cursor_col)
        if cards and 0 <= self.cursor_row < len(cards):
            card = cards[self.cursor_row]
            card.add_class("-highlight")
            # Scroll card into view vertically
            for ancestor in card.ancestors:
                if isinstance(ancestor, VerticalScroll):
                    ancestor.scroll_to_widget(card, animate=False)
                    break
            # Snap board horizontally to the active column
            container = self.query_one("#board-container", SnapHorizontalScroll)
            container.scroll_to(x=self.cursor_col * COLUMN_STEP, animate=True, duration=0.3)

    def _move_cursor(self, col: int, row: int) -> None:
        cards = self._get_cards(col)
        if not cards:
            return
        self.cursor_col = col
        self.cursor_row = max(0, min(row, len(cards) - 1))
        self._update_highlight()

    # --- Lifecycle ---

    def on_mount(self) -> None:
        nonempty = self._nonempty_cols()
        if nonempty:
            self._move_cursor(nonempty[0], 0)

    def on_screen_resume(self) -> None:
        self._update_highlight()

    # --- Navigation actions ---

    def action_cursor_up(self) -> None:
        if self.cursor_row > 0:
            self._move_cursor(self.cursor_col, self.cursor_row - 1)

    def action_cursor_down(self) -> None:
        cards = self._get_cards(self.cursor_col)
        if self.cursor_row < len(cards) - 1:
            self._move_cursor(self.cursor_col, self.cursor_row + 1)

    def action_cursor_left(self) -> None:
        candidates = [i for i in self._nonempty_cols() if i < self.cursor_col]
        if candidates:
            self._move_cursor(candidates[-1], self.cursor_row)

    def action_cursor_right(self) -> None:
        candidates = [i for i in self._nonempty_cols() if i > self.cursor_col]
        if candidates:
            self._move_cursor(candidates[0], self.cursor_row)

    def action_select_card(self) -> None:
        cards = self._get_cards(self.cursor_col)
        if cards and 0 <= self.cursor_row < len(cards):
            self.app.push_screen("detail", {"issue_id": cards[self.cursor_row].issue.id})

    def on_issue_card_selected(self, event: IssueCard.Selected) -> None:
        # Update cursor to match clicked card, then open detail
        for col_idx, col in enumerate(self._get_columns()):
            cards = list(col.query(IssueCard))
            for row_idx, card in enumerate(cards):
                if card is event.card:
                    self.cursor_col = col_idx
                    self.cursor_row = row_idx
                    self._update_highlight()
                    self.app.push_screen("detail", {"issue_id": event.issue.id})
                    return

    def action_switch_list(self) -> None:
        self.app.switch_screen("list")

    def action_refresh(self) -> None:
        self.app.switch_screen("board")

    def action_quit(self) -> None:
        self.app.exit()
