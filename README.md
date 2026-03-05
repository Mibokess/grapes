# Grapes

An issue tracker built for AI agents. Issues are plain files that agents edit with standard tools — no APIs, no databases. A terminal UI gives humans a live view of what's happening.

![Grapes TUI demo](doc/demo.gif)

## How It Works

Issues live in a `.grapes/` folder as numbered directories. Each issue is three files:

```
.grapes/42/
  meta.toml       # title, status, priority, labels
  content.md      # description
  comments.md     # append-only log
```

An agent changes an issue's status by editing one line in a TOML file. No client libraries, no authentication, no learning curve. `grep`, `sed`, and `cat` are the entire API.

The TUI watches the filesystem and updates in real time — so when an agent moves an issue to `in_progress`, you see the card slide across the board immediately.

## Install

```sh
go install github.com/Mibokess/grapes@latest
```

Or build from source:

```sh
git clone https://github.com/Mibokess/grapes.git
cd grapes && go build -o grapes .
```

## Quick Start

```sh
mkdir -p .grapes
grapes                # launch the TUI
```

Create an issue:

```sh
next=$(grapes next-id)

cat > .grapes/$next/meta.toml << 'EOF'
title = "Fix login redirect loop"
status = 'todo'
priority = 'high'
labels = ['bug']
created = '2026-03-05T10:00'
updated = '2026-03-05T10:00'
EOF

touch .grapes/$next/content.md .grapes/$next/comments.md
```

Validate issues:

```sh
grapes validate          # all issues
grapes validate 42 43    # specific issues
```

## Features

**Three views** — Kanban board, sortable list, and full detail view with rendered markdown. Switch with `B`, `L`, and `Enter`/`Esc`.

**Filtering and search** — Filter by status, priority, labels, or sub-issue scope (`f`). Full-text search in list view (`/`).

**Inline editing** — Press `e` to open any issue in `$EDITOR`. Press `s`/`p` to cycle status or priority. Drag cards between columns.

**Sub-issues and dependencies** — Set `parent = 1` to nest under an issue. Set `blocked_by = [3, 5]` to declare dependencies. The TUI renders the full hierarchy with navigable links.

**Live reload** — File changes from agents, editors, or other tools appear instantly via filesystem watching.

**Themes** — Ships with 450+ color presets (Catppuccin, Dracula, Gruvbox, Nord, Tokyo Night, and more). Configure in `~/.grapes/config.toml` or press `C` in the TUI.

**Concurrent-safe IDs** — `grapes next-id` scans the main project and all worktrees, using file locking to prevent collisions.

## Agent Integration

Grapes ships with [Claude Code](https://docs.anthropic.com/en/docs/claude-code) skills for creating, updating, searching, and commenting on issues. Copy the `plugin/` directory to use them in your project.

The design principle: agents don't need a special client. Any tool that can read and write files can work with grapes issues.

## Format Reference

See [SPEC.md](SPEC.md) for the full field reference and format specification.

## Inspiration

- [peas](https://github.com/asaaki/peas) — file-based issue tracking
- The Unix philosophy: text files, simple tools, composability
