---
name: grapes-update
description: "Use when you need to change issue metadata — status, priority, or labels."
user-invokable: false
---

# Updating Issue Metadata

Edit `.grapes/<id>/meta.yaml` to change fields. Use surgical single-line edits, not full rewrites.

## Changing Status

```
Edit .grapes/<id>/meta.yaml
  old: "status: todo"
  new: "status: in_progress"
```

Valid values: `backlog`, `todo`, `in_progress`, `done`, `cancelled`

## Changing Priority

```
Edit .grapes/<id>/meta.yaml
  old: "priority: medium"
  new: "priority: urgent"
```

Valid values: `urgent`, `high`, `medium`, `low`

## Changing Labels

```
Edit .grapes/<id>/meta.yaml
  old: "labels: [bug]"
  new: "labels: [bug, auth]"
```

## Always Update the Datetime

Every edit to meta.yaml must also update `updated:` to the current datetime:

```
Edit .grapes/<id>/meta.yaml
  old: "updated: 2026-02-20T10:00"
  new: "updated: 2026-02-27T14:30"
```

Do both edits (the field change + the datetime update) when modifying an issue.
