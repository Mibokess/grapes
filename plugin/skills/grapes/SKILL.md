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
  meta.yaml       # status, priority, assignee, labels, dates (~8 lines)
  content.md      # issue description (markdown)
  comments.md     # append-only comment log
```

IDs are numeric folder names. The folder listing is the index.

## meta.yaml Schema

```yaml
title: "Short description of the issue"
status: todo
priority: high
assignee: ""
labels: [bug, auth]
parent: 40
created: 2026-02-27
updated: 2026-02-27
```

### Field Values

- **status**: `backlog`, `todo`, `in_progress`, `done`, `cancelled`
- **priority**: `urgent`, `high`, `medium`, `low`
- **assignee**: username string, or `""` when unassigned
- **labels**: YAML list of freeform tags
- **parent**: numeric ID of parent issue (omit for top-level issues)
- **created** / **updated**: `YYYY-MM-DD`

### Rules

- Always update `updated:` to today's date when modifying meta.yaml.
- Quote titles containing special characters: colons, brackets, etc.

## comments.md Format

```markdown
### alice — 2026-02-27
Comment body here. Can be multiple lines.

### bob — 2026-02-27
Another comment.
```

- Header: `### <author> — <YYYY-MM-DD>` (em-dash `—`, not hyphen)
- Use `agent` as author for AI-authored comments.
- Append-only. Never edit or delete existing comments.

## Principles

- **Read only what you need.** meta.yaml is ~8 lines. Read it first. Only load content.md or comments.md when you need the full description or comment history.
- **Surgical edits.** Change one field in meta.yaml, don't rewrite the file.
- **The filesystem is the database.** Use grep/ls to query, file tools to read/write.
