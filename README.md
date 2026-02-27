# Grapes

A file-based issue tracker designed for AI agents. Issues are plain files in a `.grapes/` folder — no database, no CLI tool, no API. Agents manipulate issues directly using standard file tools (grep, edit, read, write). A web UI provides board visualization.

## Why

- **Context efficiency** — Agents can surgically edit a single line (e.g. change `status: todo` to `status: in_progress`) without loading entire issue objects.
- **Zero tooling overhead** — No custom tool definitions, no SDK, no authentication. The filesystem *is* the database, the agent *is* the CLI.
- **Performance** — Standard Linux commands handle hundreds/thousands of issues effortlessly.

## How It Works

Each issue is a numbered folder with three files:

```
.grapes/
  42/
    meta.yaml       # status, priority, assignee, labels, dates
    content.md      # issue description
    comments.md     # append-only comment log
```

### meta.yaml

```yaml
title: Fix login redirect loop
status: todo
priority: high
assignee: alice
labels: [bug, auth]
parent: 40
created: 2026-02-27
updated: 2026-02-27
```

### Fields

| Field | Values |
|-------|--------|
| `status` | `backlog`, `todo`, `in_progress`, `done`, `cancelled` |
| `priority` | `urgent`, `high`, `medium`, `low` |
| `assignee` | freeform |
| `labels` | freeform tags |
| `parent` | ID of parent issue (omit for top-level) |

### Querying

```sh
grep -rl "status: todo" .grapes/*/meta.yaml       # issues by status
grep -rl "assignee: alice" .grapes/*/meta.yaml     # issues by assignee
grep -rl "login bug" .grapes/*/content.md          # full-text search
grep -rl "parent: 40" .grapes/*/meta.yaml          # children of issue 40
```

### Creating an Issue

1. Find the next ID: `ls .grapes/ | sort -n | tail -1`
2. Create `.grapes/<next>/` with `meta.yaml`, `content.md`, `comments.md`

No counter file needed — the folder names *are* the counter.

## Sub-Issues

Issues support unlimited nesting via the `parent` field. The folder structure stays flat — nesting is a data relationship, not a filesystem relationship. Moving a sub-issue means editing one line.

## Web UI

A lightweight web app reads `.grapes/` and renders:

- **Board view** — Kanban columns by status
- **List view** — Sortable/filterable table
- **Detail view** — Full issue with comments

The UI is read-heavy, write-light. The primary write path is the agent editing files directly.

## Inspiration

- [peas](https://github.com/asaaki/peas) — file-based issue tracking
- The Unix philosophy: text files, simple tools, composability
