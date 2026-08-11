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

- `id=$(grapes issue)` — scans all known repository worktrees and configured
  external sources under a lock, creates the next globally unused ID in the
  current store, and prints it.
- `grapes issue <id>` — updates the issue's timestamps in the current store.
- IDs must be positive integers; `issue` accepts at most one ID, and `validate`
  accepts zero or more positive IDs. Extra or malformed arguments are errors.

Run `grapes issue <id>` after modifying any issue files. Never write
`created` or `updated` timestamps manually. Grapes preserves existing metadata,
sets `updated` to the current UTC time, and fills `created` only when it is
missing. The TUI writes to the issue's active source directory, which may be a
working worktree rather than the main checkout.

The new `meta.toml` only has timestamps — you still need to populate title,
status, priority, etc. and write `content.md`. Read the newly created files
before editing them.

## Worktrees and External Sources

`.grapes` is tracked by git, so every worktree contains a copy of every issue.
Grapes attributes a worktree only when Git reports that its branch changed an
issue relative to the branch base; an idle or merely stale worktree is not a
second source version. When sources disagree, the TUI selects the most recent
real change and lets you switch source versions. Edits follow the active
`SourceDir`, not a path selected by file mtime.

Additional stores can be configured with `[sources].dirs` glob patterns.
External stores are loaded in full and do not receive Git worktree attribution.
Set `[sources].default_branch` when automatic default-branch detection chooses
the wrong comparison ref; leave it empty to use repository detection.

## Validation and Writes

Normal loading is intentionally tolerant: malformed issue directories and
unreadable secondary files may be skipped and reported in the TUI. Run
`grapes validate` for strict checks of all issues, or `grapes validate <id>...`
for selected positive IDs, before relying on issue data. Validation checks
metadata, relationships, and comment structure.

The TUI and editor write paths validate metadata before replacing files.
Metadata writes are atomic; comments are append-only. Do not edit existing
comments in place or persist derived `children`/`blocks` relationship fields.

## Git Workflow

Issue files are tracked by git.
Commit them after creation or modification using the format:
`#<id>: Create issue` or `#<id>: Update issue description`.

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
