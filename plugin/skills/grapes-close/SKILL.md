---
name: grapes-close
description: "Use when an issue is resolved, completed, or should be cancelled."
user-invokable: false
---

# Closing an Issue

## Completing an Issue

Three steps:

1. **Update status** in `.grapes/<id>/meta.yaml`:
   ```
   status: todo → status: done
   ```

2. **Update date**:
   ```
   updated: <today's date>
   ```

3. **Add a closing comment** to `.grapes/<id>/comments.md` explaining what was done:
   ```markdown
   ### agent — 2026-02-27
   Fixed in commit abc123. The issue was caused by X, resolved by Y.
   ```

Always add a closing comment. Future readers need to know what resolved the issue.

## Cancelling an Issue

Same steps but use `status: cancelled` and explain why in the closing comment:

```markdown
### agent — 2026-02-27
Cancelled: duplicate of #12.
```

## Don't Forget

- Update `updated:` date.
- Add a closing comment — never close silently.
