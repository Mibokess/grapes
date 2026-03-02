---
name: grapes-create
description: "Use when you need to create a new issue or sub-issue in the tracker."
user-invokable: false
---

# Creating an Issue

## Step 1: Generate the Next ID

```bash
ls .grapes/ | sort -n | tail -1
```

Add 1 to the result. If the directory is empty, start at 1.

## Step 2: Create the Folder

```bash
mkdir -p .grapes/<id>
```

## Step 3: Write meta.yaml

```yaml
title: "Short description of the issue"
status: backlog
priority: medium
labels: []
created: YYYY-MM-DDTHH:MM
updated: YYYY-MM-DDTHH:MM
```

- Set `created` and `updated` to the current datetime.
- Set `status` to `backlog` for new issues unless there's reason to start higher.
- Add `parent: <id>` if this is a sub-issue.
- Quote the title if it contains colons, brackets, or other YAML-special characters.

## Step 4: Write content.md

Write the issue description in plain markdown. Include:
- What the problem or feature is
- Context and reproduction steps (for bugs)
- Any relevant code references

## Step 5: Create Empty comments.md

Create the file but leave it empty. Comments are added later via `grapes-comment`.

## Sub-Issues

To create a sub-issue, add `parent: <id>` to meta.yaml:

```yaml
title: "Implement auth callback fix"
status: todo
priority: high
labels: [auth]
parent: 40
created: 2026-02-27T09:15
updated: 2026-02-27T09:15
```

The folder structure stays flat. Nesting is a data relationship only.
