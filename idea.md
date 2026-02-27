# Grapes — File-Based Issue Tracker

## The Idea

An issue tracker where issues are plain files in a folder. No database, no CLI tool, no API. AI agents manipulate issues directly using standard file tools (grep, edit, read, write). A web UI provides board visualization.

## Why

**Context efficiency** — Tools like Linear require fetching entire issue objects to read or update them. With local files, an agent can surgically edit a single line (e.g. change `status: todo` to `status: in_progress`) without loading anything else.

**Zero tooling overhead** — No custom tool definitions, no SDK, no authentication. The agent already has file tools built in. The filesystem *is* the database, the agent *is* the CLI.

**Performance** — Standard Linux commands handle hundreds/thousands of issues effortlessly:
- `grep -rl "status: todo" .issues/*/meta.yaml` → query by status
- `grep -rl "assignee: alice" .issues/*/meta.yaml` → query by assignee
- `ls .issues/ | sort -n | tail -1` → get next ID
- `grep -rl "login bug" .issues/*/content.md` → full-text search
- `cat .issues/42/meta.yaml` → read only metadata (7 lines)

## Folder Structure

Each issue is a folder containing separate files for metadata, content, and comments:

```
.issues/
  42/
    meta.yaml       # status, priority, assignee, labels, dates
    content.md      # issue description
    comments.md     # append-only comment log
  43/
    meta.yaml
    content.md
    comments.md
```

### meta.yaml — Metadata (tiny, most-touched file)

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

### content.md — Description

```markdown
Users are getting stuck in a redirect loop after login when they have
2FA enabled. The OAuth callback doesn't preserve the original URL.
```

### comments.md — Append-only log

```markdown
### alice — 2026-02-27
Reproduced locally. The issue is in `auth/callback.ts`.

### bob — 2026-02-27
Confirmed. Working on a fix now.
```

### Fields
- **title** — short description
- **status** — `backlog`, `todo`, `in_progress`, `done`, `cancelled`
- **priority** — `urgent`, `high`, `medium`, `low`
- **assignee** — who's working on it
- **labels** — freeform tags
- **parent** — ID of parent issue (omit for top-level issues)
- **created** / **updated** — dates

### Sub-Issues (Nesting)

Issues support unlimited nesting via the `parent` field in `meta.yaml`. The folder structure stays **flat** — nesting is a data relationship, not a filesystem relationship.

```
.issues/
  40/meta.yaml   # top-level issue (no parent)
  42/meta.yaml   # parent: 40 (child of 40)
  44/meta.yaml   # parent: 42 (grandchild of 40)
```

**Queries:**
- `grep -rl "parent: 40" .issues/*/meta.yaml` → direct children of 40
- `grep -rL "parent:" .issues/*/meta.yaml` → all top-level issues (no parent field)

**Why flat + `parent` field instead of nested folders:**
- All IDs stay globally unique — no composite IDs like `40/1/1`
- Moving a sub-issue = edit one line in `meta.yaml`
- Grep/ls work identically regardless of nesting depth
- No deep folder traversals needed

### Why Folders Per Issue
- **Context efficiency** — Changing status only touches `meta.yaml` (~7 lines). The agent never loads the description or comments.
- **Comments scale independently** — 50 comments don't bloat metadata reads.
- **Appending comments is trivial** — Just append to `comments.md`, no need to parse or find insertion points.
- Numeric folder names (`42/`) give natural ordering and easy ID generation.
- The folder listing *is* the index — no manifest needed.

## Visualization — Web UI

A lightweight web app that reads `.issues/` and renders a board.

### Stack
- **Frontend**: Single-page app (vanilla JS or lightweight framework)
- **Backend**: Minimal server that reads the `.issues/` directory
  - Parses frontmatter from each file
  - Serves JSON to the frontend
  - Optionally watches for file changes (live reload)

### Views
- **Board view** — Kanban columns by status
- **List view** — Sortable/filterable table
- **Detail view** — Full issue with comments

### Key Principle
The web UI is **read-heavy, write-light**. The primary write path is the agent editing files directly. The UI mostly visualizes. If the UI supports editing, it just writes back to the files.

## ID Generation

The agent handles this:
1. List folders in `.issues/`
2. Find the highest number
3. Create `.issues/<highest + 1>/` with `meta.yaml`, `content.md`, `comments.md`

No counter file needed. The folder names *are* the counter.

## Inspiration

- [peas](https://github.com/asaaki/peas) — file-based issue tracking
- The Unix philosophy: text files, simple tools, composability
