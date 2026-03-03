---
name: grapes
description: "Foundational reference for the Grapes file-based issue tracker. Use when working in a project with a .grapes/ directory."
user-invokable: false
---

# Grapes — File-Based Issue Tracker

Issues are plain files in `.grapes/`. No database, no CLI. You manipulate them directly with file tools.

## Structure

```
.grapes/<id>/
  meta.yaml       # status, priority, labels, dates (~7 lines)
  content.md      # issue description (markdown)
  comments.md     # append-only comment log
```

IDs are numeric folder names. The folder listing is the index.

## meta.yaml Schema

```yaml
title: "Short description of the issue"
status: todo
priority: high
labels: [bug, auth]
parent: 40
blocked_by: [3, 5]
created: 2026-02-27T09:15
updated: 2026-02-27T14:30
```

### Field Values

- **status**: `backlog`, `todo`, `in_progress`, `done`, `cancelled`
- **priority**: `urgent`, `high`, `medium`, `low`
- **labels**: YAML list of freeform tags
- **parent**: numeric ID of parent issue (omit for top-level issues)
- **blocked_by**: YAML list of issue IDs this issue depends on (omit if none). The inverse (`blocks`) is computed at load time — only `blocked_by` is stored on disk.
- **created** / **updated**: `YYYY-MM-DDTHH:MM` (24-hour time)

### Rules

- Always update `updated:` to the current datetime when modifying meta.yaml.
- Quote titles containing special characters: colons, brackets, etc.

## comments.md Format

```markdown
### 2026-02-27T09:15
Comment body here. Can be multiple lines.

### 2026-02-28T14:30
Another comment.
```

- Header: `### YYYY-MM-DDTHH:MM`
- Append-only. Never edit or delete existing comments.

## Principles

- **Read only what you need.** meta.yaml is ~7 lines. Read it first. Only load content.md or comments.md when you need the full description or comment history.
- **Surgical edits.** Change one field in meta.yaml, don't rewrite the file.
- **The filesystem is the database.** Use grep/ls to query, file tools to read/write.
