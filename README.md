# Grapes

An issue tracker built for AI agents. Issues are plain files — no APIs, no databases. A terminal UI gives humans a live view of what's happening.

![Grapes TUI demo](doc/demo.gif)

## How It Works

Issues live in a `.grapes/` folder as numbered directories. Each issue is three files:

```
.grapes/42/
  meta.toml       # title, status, priority, labels
  content.md      # description
  comments.md     # append-only log
```

Agents change issue status by editing a line in a TOML file. No client libraries, no authentication, no learning curve.

The TUI watches the filesystem and updates in real time — when an agent moves an issue to `in_progress`, you see the card slide across the board immediately.

## Install

Download a binary from [GitHub Releases](https://github.com/Mibokess/grapes/releases), or install with Go:

```sh
go install github.com/Mibokess/grapes@latest
```

## Claude Code

Grapes ships with a [Claude Code](https://docs.anthropic.com/en/docs/claude-code) plugin that teaches agents the issue format so they can create, update, and manage issues out of the box.

```sh
/plugin marketplace add Mibokess/grapes
/plugin install grapes@grapes
```

**Worktree support** — The TUI also picks up issues from git worktrees, so you can see agent progress across parallel work streams.

## Skills

The `.agents/` directory contains workflow skills that can be copied into any project:

- `/issue` — plan, verify, research, or continue work on an issue
- `/handoff` — write a handoff plan so another agent can pick up work
- `/pr` — push and create a pull request


## CLI

| Command              | Description                                          |
| -------------------- | ---------------------------------------------------- |
| `grapes`             | Launch the TUI                                       |
| `grapes issue`       | Allocate next ID, create directory, stamp timestamps |
| `grapes issue <id>`  | Stamp timestamps on an existing issue                |
| `grapes validate`    | Validate all issues                                  |
| `grapes validate ID` | Validate specific issues                             |

`grapes issue` scans the main project and all worktrees, using file locking to prevent ID collisions.