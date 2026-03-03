# Grapes Format Specification

Version 1.0

## Directory Structure

```
.grapes/
  <id>/
    meta.toml
    content.md
    comments.md
```

Each issue is a directory named with its numeric ID (positive integer, assigned sequentially). All three files are required but `content.md` and `comments.md` may be empty.

## meta.toml

```toml
title = "Short description of the issue"
status = 'todo'
priority = 'high'
labels = ['bug', 'auth']
parent = 40
created = '2026-02-27T09:15'
updated = '2026-02-27T14:30'
```

| Field | Type | Required | Values |
|-------|------|----------|--------|
| `title` | string | yes | Non-empty. |
| `status` | string | yes | `backlog` · `todo` · `in_progress` · `done` · `cancelled` |
| `priority` | string | yes | `urgent` · `high` · `medium` · `low` |
| `labels` | string list | yes | Freeform tags. `[]` for none. |
| `parent` | integer | no | ID of parent issue. Omit for top-level issues. Must reference an existing issue. |
| `created` | datetime | yes | Set once at creation. |
| `updated` | datetime | yes | Updated on every modification to any file in the issue. |

### Datetimes

Canonical format: `YYYY-MM-DDTHH:MM` (24-hour, minute precision). `YYYY-MM-DD` is accepted on read.

## content.md

Issue description in Markdown. May be empty.

## comments.md

Append-only log of timestamped comments. May be empty.

```markdown
### 2026-02-27T09:15
Comment body. Can span multiple lines.

### 2026-02-28T14:30
Another comment.
```

- Header: `### YYYY-MM-DDTHH:MM` (one per comment, `### YYYY-MM-DD` also accepted on read)
- Body: everything between one header and the next (or end of file)
- Comments are separated by a blank line
- Existing comments must never be edited or deleted

## Sub-Issues

Parent-child relationships are expressed via the `parent` field. The directory structure is flat — all issues live directly under `.grapes/` regardless of nesting depth. An issue's children are those issues whose `parent` equals its ID.
