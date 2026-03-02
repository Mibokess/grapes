---
name: grapes-comment
description: "Use when you need to add a comment to an issue."
user-invokable: false
---

# Adding a Comment

Append to `.grapes/<id>/comments.md`.

## Format

```markdown
### YYYY-MM-DDTHH:MM
Comment body here.
```

- Header: `### YYYY-MM-DDTHH:MM`
- Set the datetime to now
- Leave a blank line before the header if the file is not empty

## Example

If `.grapes/5/comments.md` currently contains:

```markdown
### 2026-02-27T09:15
Found the root cause in auth/callback.ts.
```

Append:

```markdown

### 2026-02-28T14:30
Fixed the callback to preserve the original URL. See commit abc123.
```

## When to Comment

- When starting work on an issue (what you plan to do)
- When making progress worth noting
- When closing or cancelling an issue (what resolved it)
- When you discover something relevant to an issue you're not currently working on

## Rules

- **Append only.** Never edit or delete existing comments.
- **One comment per action.** Don't append multiple headers at once.
