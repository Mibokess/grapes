---
name: grapes
description: "Foundational reference for the Grapes file-based issue tracker. Use when working in a project with a .grapes/ directory."
user-invokable: false
---

# Grapes — File-Based Issue Tracker

Issues are plain files in `.grapes/`. IDs are numeric folder names.

```
.grapes/<id>/
  meta.toml       # status, priority, labels, dates
  content.md      # issue description (markdown)
  comments.md     # append-only comment log
```

## Creating and Updating Issues

- `id=$(grapes issue)` — creates the directory and prints the next ID
- `grapes issue <id>` — bumps `updated`

Run `grapes issue <id>` after modifying any issue files. Never write timestamps manually.

The new `meta.toml` only has timestamps — you still need to populate title, status, priority, etc. and write `content.md`.
Read the newly created files before editing them.

Issue files are tracked by git. 
Commit them after creation or modification using the format: `#<id>: Create issue` or `#<id>: Update issue description`.

## meta.toml

```toml
title = "Short description of the issue"
status = 'todo'
priority = 'high'
labels = ['bug', 'auth']
parent = 40
blocked_by = [3, 5]
created = 2026-02-27T09:15:00Z
updated = 2026-02-27T14:30:00Z
```

- **status**: `backlog`, `todo`, `in_progress`, `done`, `cancelled`
- **priority**: `urgent`, `high`, `medium`, `low`
- **labels**: freeform tags
- **parent**: numeric ID of parent issue (omit for top-level)
- **blocked_by**: issue IDs this depends on (omit if none)
- **created** / **updated**: managed by `grapes issue`, never write manually

## comments.md

Append-only. Never edit or delete existing comments.

```markdown
### 2026-02-27T09:15
Comment body here.

### 2026-02-28T14:30
Another comment.
```

`comments.md` contains progress updates, decisions, and context that may not be in `content.md`.
When building a full picture of an issue, read both.
