"""Grapes TUI — File-based issue tracker visualization."""

from __future__ import annotations

from textual.app import App

from grapes_tui.screens.board import BoardScreen
from grapes_tui.screens.list import ListScreen
from grapes_tui.screens.detail import DetailScreen


class GrapesApp(App):
    """Main application."""

    TITLE = "Grapes"
    SUB_TITLE = "Issue Tracker"
    theme = "textual-light"

    CSS = """
    Screen {
        background: $background;
    }
    """

    SCREENS = {
        "board": BoardScreen,
        "list": ListScreen,
    }

    def on_mount(self) -> None:
        self.push_screen("board")

    def push_screen(self, screen, kwargs=None, callback=None):
        """Override to support passing kwargs to screens."""
        if screen == "detail" and isinstance(kwargs, dict):
            detail = DetailScreen(issue_id=kwargs["issue_id"])
            super().push_screen(detail, callback=callback)
        else:
            super().push_screen(screen, callback=callback)


def main():
    app = GrapesApp()
    app.run()


if __name__ == "__main__":
    main()
