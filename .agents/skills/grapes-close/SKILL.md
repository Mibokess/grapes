---
name: grapes-close
description: "Use when an issue is resolved, completed, or should be cancelled."
user-invokable: false
---

# Closing an Issue

## Completing an Issue

Three steps:

1. **Update status** in `.grapes/<id>/meta.toml`:
   ```
   status = 'todo' → status = 'done'
   ```

2. **Update datetime**:
   ```
   updated = '<current datetime YYYY-MM-DDTHH:MM>'
   ```

3. **Add a closing comment** to `.grapes/<id>/comments.md` explaining what was done:
   ```markdown
   ### 2026-02-27T14:30
   Fixed in commit abc123. The issue was caused by X, resolved by Y.
   ```

Always add a closing comment. Future readers need to know what resolved the issue.

## Cancelling an Issue

Same steps but use `status = 'cancelled'` and explain why in the closing comment:

```markdown
### 2026-02-27T14:30
Cancelled: duplicate of #12.
```

## Don't Forget

- Update `updated` datetime.
- Add a closing comment — never close silently.
