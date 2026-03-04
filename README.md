# Grapes

A file-based issue tracker. Issues are plain files in a `.grapes/` folder that agents read and write with standard file tools. A terminal UI provides visualization.

![Grapes TUI demo](doc/demo.gif)

## Why

- **Surgical edits** — Change `status = 'todo'` to `status = 'in_progress'` with a single line edit.
- **Standard tools** — Agents use `grep`, `cat`, `sed`, and regular file operations to work with issues.
- **Composable** — Pipe, filter, and transform issues with any Unix tool or scripting language.

## Issue Format

Each issue is a numbered folder under `.grapes/`:

```
.grapes/
  1/
    meta.toml       # structured metadata
    content.md      # description (markdown)
    comments.md     # append-only comment log
  2/
    ...
```

### meta.toml

```toml
title = "Fix login redirect loop"
status = 'todo'
priority = 'high'
labels = ['bug', 'auth']
parent = 1
blocked_by = [3, 5]
created = '2026-02-27T14:00'
updated = '2026-02-28T09:30'
```

### Fields

| Field | Values | Notes |
|-------|--------|-------|
| `status` | `backlog` `todo` `in_progress` `done` `cancelled` | Required |
| `priority` | `urgent` `high` `medium` `low` | Required |
| `labels` | Freeform tags | `[]` for none |
| `parent` | Issue ID | Creates a sub-issue relationship |
| `blocked_by` | List of issue IDs | Inverse `blocks` computed automatically |
| `created` | `YYYY-MM-DD` or `YYYY-MM-DDTHH:MM` | Set once |
| `updated` | `YYYY-MM-DD` or `YYYY-MM-DDTHH:MM` | Updated on every change |

### comments.md

Append-only log with timestamped entries:

```markdown
### 2026-02-28T10:00

Investigated the root cause — session cookie not being cleared on redirect.

### 2026-02-28T14:30

Fix deployed. Monitoring for regressions.
```

## Relationships

**Sub-issues** — Set `parent = 1` on a child issue to nest it under issue 1. Nesting depth is unlimited. The folder structure stays flat; hierarchy is a data relationship.

**Blocking** — Set `blocked_by = [3, 5]` to indicate dependencies. The inverse (`blocks`) is computed at load time and shown in the TUI.

## Querying

```sh
grep -rl "status = " .grapes/*/meta.toml           # find by status
grep -rl "priority = 'urgent'" .grapes/*/meta.toml # find by priority
grep -rl "parent = 1" .grapes/*/meta.toml          # children of issue 1
grep -rl "login" .grapes/*/content.md              # full-text search
ls .grapes/ | sort -n | tail -1                    # latest issue ID
```

## Creating an Issue

```sh
# Reserve the next ID (scans main + all worktrees, uses file locking)
next=$(grapes next-id)

cat > .grapes/$next/meta.toml << 'EOF'
title = "My new issue"
status = 'todo'
priority = 'medium'
labels = []
created = '2026-03-01T10:00'
updated = '2026-03-01T10:00'
EOF

touch .grapes/$next/content.md .grapes/$next/comments.md
```

## Validation

```sh
go run . validate          # validate all issues
go run . validate 42 43    # validate specific issues
```

Checks required fields, valid enum values, date formats, comment header format, and cross-issue integrity (parent/blocked_by references exist, no self-blocking).

## TUI

```sh
go run .
```

Three views — **Board** (kanban by status), **List** (sortable table), and **Detail** (full issue with rendered markdown). Switch between them with `L`/`B` and `Enter`/`Esc`.

### Features

- **Filtering** — Structured filter menu (`f`) for status, priority, labels, and sub-issue scope. Text search (`/`) in list view matches across all fields.
- **Sorting** — Cycle sort mode with `o` (priority, updated, created, ID, title, status). Reverse with `O`.
- **Inline editing** — Press `e` to open the issue in `$EDITOR`. Changes are validated before saving.
- **Comments** — Press `c` in detail view to append a timestamped comment.
- **Status/priority** — Press `s` or `p` to pick a new value from a menu.
- **Drag and drop** — Drag cards between board columns to change status.
- **Live reload** — File changes from agents or other tools appear in real time via fsnotify.
- **Navigation** — Clickable links between parent, child, and blocked issues in detail view.

### Keybindings

| Key | Board | List | Detail |
|-----|-------|------|--------|
| `hjkl` / arrows | Navigate cards | Navigate rows | Scroll |
| `Enter` | Open issue | Open issue | — |
| `Esc` | — | Clear filter | Go back |
| `L` / `B` | To list | To board | Switch view |
| `/` | — | Text search | — |
| `f` | Filter menu | Filter menu | — |
| `s` / `p` | Status / priority filter | Status / priority filter | Cycle status / priority |
| `o` / `O` | Sort / reverse | Sort / reverse | — |
| `e` | Edit in `$EDITOR` | Edit in `$EDITOR` | Edit in `$EDITOR` |
| `c` | — | — | Add comment |
| `r` | Refresh | Refresh | — |
| `?` | Help | Help | Help |
| `q` | Quit | Quit | Quit |

## Agent Integration

Grapes ships with [Claude Code](https://claude.ai/claude-code) skills in `plugin/skills/` for creating, reading, listing, searching, updating, commenting on, and closing issues. Symlink them into `.claude/skills/` to enable them.

Changes made in the TUI write directly to the `.grapes/` files on disk.

## Inspiration

- [peas](https://github.com/asaaki/peas) — file-based issue tracking
- The Unix philosophy: text files, simple tools, composability
