"""Issue detail screen."""

from __future__ import annotations

from textual.app import ComposeResult
from textual.containers import VerticalScroll
from textual.screen import Screen
from textual.widgets import Header, Footer, Static, Label, Markdown, Rule

from grapes_tui.data import (
    STATUS_LABELS,
    PRIORITY_LABELS,
    load_issue,
    load_all_issues,
    parse_comments,
)


class DetailScreen(Screen):
    """Full issue detail view."""

    BINDINGS = [
        ("escape", "go_back", "Back"),
        ("j", "scroll_down", "Scroll Down"),
        ("k", "scroll_up", "Scroll Up"),
        ("b", "switch_board", "Board"),
        ("l", "switch_list", "List"),
        ("q", "quit", "Quit"),
    ]

    DEFAULT_CSS = """
    DetailScreen {
        layout: vertical;
    }
    DetailScreen #detail-scroll {
        height: 1fr;
        padding: 1 2;
    }
    DetailScreen .detail-title {
        text-style: bold;
        color: $primary;
        margin-bottom: 1;
    }
    DetailScreen .meta-section {
        margin-bottom: 1;
    }
    DetailScreen .meta-line {
        margin: 0;
    }
    DetailScreen .section-header {
        text-style: bold;
        color: $secondary;
        margin-top: 1;
        margin-bottom: 0;
    }
    DetailScreen .comment-header {
        text-style: bold;
        margin-top: 1;
    }
    DetailScreen .comment-body {
        margin-left: 2;
        margin-bottom: 1;
    }
    DetailScreen .sub-issue {
        margin-left: 2;
    }
    DetailScreen .sub-issue:hover {
        text-style: underline;
    }
    """

    def __init__(self, issue_id: int, **kwargs) -> None:
        super().__init__(**kwargs)
        self.issue_id = issue_id

    def compose(self) -> ComposeResult:
        yield Header()
        with VerticalScroll(id="detail-scroll"):
            issue = load_issue(self.issue_id)
            if issue is None:
                yield Label(f"Issue #{self.issue_id} not found.")
                return

            yield Label(f"#{issue.id}: {issue.title}", classes="detail-title")
            yield Rule()

            # Metadata
            yield Label(
                f"  Status: {STATUS_LABELS.get(issue.status, issue.status)}   "
                f"Priority: {PRIORITY_LABELS.get(issue.priority, issue.priority)}   "
                f"Assignee: {issue.assignee or 'Unassigned'}",
                classes="meta-line",
            )
            yield Label(
                f"  Labels: {', '.join(issue.labels) or 'None'}   "
                f"Created: {issue.created}   "
                f"Updated: {issue.updated}",
                classes="meta-line",
            )
            if issue.parent is not None:
                yield Label(f"  Parent: #{issue.parent}", classes="meta-line")

            # Description
            yield Label("Description", classes="section-header")
            yield Rule()
            if issue.content.strip():
                yield Markdown(issue.content)
            else:
                yield Label("  No description.")

            # Sub-issues
            if issue.children:
                yield Label("Sub-issues", classes="section-header")
                yield Rule()
                all_issues = load_all_issues()
                issue_map = {i.id: i for i in all_issues}
                for child_id in issue.children:
                    child = issue_map.get(child_id)
                    if child:
                        status = STATUS_LABELS.get(child.status, child.status)
                        yield Label(
                            f"  #{child.id}: {child.title} — {status}",
                            classes="sub-issue",
                        )

            # Comments
            comments = parse_comments(issue.comments_raw)
            yield Label(f"Comments ({len(comments)})", classes="section-header")
            yield Rule()
            if not comments:
                yield Label("  No comments yet.")
            else:
                for c in comments:
                    yield Label(f"  {c.author} — {c.date}", classes="comment-header")
                    yield Label(f"    {c.body}", classes="comment-body")

        yield Footer()

    def action_scroll_down(self) -> None:
        self.query_one("#detail-scroll", VerticalScroll).scroll_down(animate=False)

    def action_scroll_up(self) -> None:
        self.query_one("#detail-scroll", VerticalScroll).scroll_up(animate=False)

    def action_go_back(self) -> None:
        self.app.pop_screen()

    def action_switch_board(self) -> None:
        self.app.switch_screen("board")

    def action_switch_list(self) -> None:
        self.app.switch_screen("list")

    def action_quit(self) -> None:
        self.app.exit()
