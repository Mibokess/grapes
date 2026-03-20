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

Download a binary from [GitHub Releases](https://github.com/Mibokess/grapes/releases), or install with Go:

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
grapes                # launch the TUI (creates .grapes/ if needed)
```

## CLI

| Command              | Description                                              |
| -------------------- | -------------------------------------------------------- |
| `grapes`             | Launch the TUI                                           |
| `grapes issue`       | Allocate next ID, create directory, stamp timestamps     |
| `grapes issue <id>`  | Stamp timestamps on an existing issue                    |
| `grapes validate`    | Validate all issues                                      |
| `grapes validate ID` | Validate specific issues                                 |

`grapes issue` scans the main project and all worktrees, using file locking to prevent ID collisions.

## Features

**Three views** — Kanban board, sortable list, and full detail view with rendered markdown. Switch with `B`, `L`, and `Enter`/`Esc`.

**Filtering and search** — Filter by status, priority, labels, or sub-issue scope (`f`). Full-text search across all views (`/`).

**Inline editing** — Press `e` to open any issue in `$EDITOR`. Press `s`/`p`/`t` to cycle status, priority, or pick labels. Drag cards between columns on the board.

**Sub-issues and dependencies** — Set `parent = 1` to nest under an issue. Set `blocked_by = [3, 5]` to declare dependencies. The TUI renders the full hierarchy with navigable links.

**Live reload** — File changes from agents, editors, or other tools appear instantly via filesystem watching.

**Multi-source** — Aggregates issues from git worktrees. See which worktree an issue came from, click to switch sources.

**Themes** — Ships with 450+ color presets (Catppuccin, Dracula, Gruvbox, Nord, Tokyo Night, and more). Press `C` in the TUI or configure in `.grapes/config.toml`.

**Settings** — Press `C` to configure theme, colors, and keybindings from within the TUI. All keybindings are customizable in `.grapes/config.toml`.

## Agent Integration

Grapes ships with a [Claude Code](https://docs.anthropic.com/en/docs/claude-code) plugin that teaches agents how to create, update, and manage issues.

Install via Claude Code's plugin system:

```sh
/plugin marketplace add Mibokess/grapes
/plugin install grapes@grapes
```

The design principle: agents don't need a special client. Any tool that can read and write files can work with grapes issues.

## Format Reference

See [SPEC.md](SPEC.md) for the full field reference and format specification.

## Inspiration

- [peas](https://github.com/asaaki/peas) — file-based issue tracking
- The Unix philosophy: text files, simple tools, composability
