---
name: grapes-update
description: "Use when you need to change issue metadata — status, priority, labels, title, parent, or blocked_by."
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

## Changing Title

```
Edit .grapes/<id>/meta.yaml
  old: "title: \"Old title\""
  new: "title: \"New title\""
```

Quote the title if it contains colons, brackets, or other YAML-special characters.

## Changing Parent

```
Edit .grapes/<id>/meta.yaml
  old: "parent: 5"
  new: "parent: 10"
```

Add `parent: <id>` to make an issue a sub-issue, or remove the line to make it top-level.

## Changing Blocked By

```
Edit .grapes/<id>/meta.yaml
  old: "blocked_by: [3]"
  new: "blocked_by: [3, 7]"
```

Add `blocked_by: [id1, id2]` to mark dependencies, or remove the line to clear them. The inverse (`blocks`) is computed at load time — only `blocked_by` is stored on disk.

## Always Update the Datetime

Every edit to meta.yaml must also update `updated:` to the current datetime:

```
Edit .grapes/<id>/meta.yaml
  old: "updated: 2026-02-20T10:00"
  new: "updated: 2026-02-27T14:30"
```

Do both edits (the field change + the datetime update) when modifying an issue.
