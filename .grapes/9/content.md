`idea.md` describes a web UI as a core visualization layer, but only the TUI exists today.

## What the spec says

From `idea.md:103-120`:

> A lightweight web app that reads `.grapes/` and renders a board.
>
> **Stack**: Single-page app (vanilla JS or lightweight framework). Minimal server that reads the `.grapes/` directory, parses frontmatter, serves JSON, optionally watches for file changes (live reload).
>
> **Views**: Board view (Kanban), List view (sortable/filterable table), Detail view (full issue with comments).

## Current state

The project only has `tui/` — a Textual-based terminal UI. No web server, no frontend, no HTML/JS/CSS.

## Approach options

1. **Simple file server** — Python/Node script that serves `.grapes/` as JSON + a static SPA
2. **Vite + vanilla JS** — Lightweight, no framework overhead, matches the "minimal" philosophy
3. **Textual-web** — Textual has a `textual-web` driver that serves the TUI as a web app (https://github.com/Textualize/textual-web). Could reuse existing screens with zero new code
4. **Skip it** — The TUI is sufficient, remove the web UI section from the spec

## Key principle from spec

> The web UI is **read-heavy, write-light**. The primary write path is the agent editing files directly. The UI mostly visualizes.

This means the web UI could be as simple as a static file that fetches JSON and renders it. No auth, no write endpoints needed for v1.
