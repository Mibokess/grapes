---
name: grapes-create
description: "Use when you need to create a new issue or sub-issue in the tracker."
user-invokable: false
---

# Creating an Issue

## Step 1: Generate the Next ID

Scan the main `.grapes/` directory **and** any worktree `.grapes/` directories to find the highest existing ID across all of them:

```bash
{ ls .grapes/; ls .claude/worktrees/*/.grapes/ 2>/dev/null; } | grep -E '^[0-9]+$' | sort -n | tail -1
```

Add 1 to the result. If no numeric directories exist, start at 1.

This prevents ID collisions with issues created in worktrees.

## Step 2: Create the Folder

```bash
mkdir -p .grapes/<id>
```

## Step 3: Write meta.yaml

Get the current timestamp first:

```bash
date +%Y-%m-%dT%H:%M
```

Use this value for both `created` and `updated`.

```yaml
title: "Short description of the issue"
status: backlog
priority: medium
labels: []
created: YYYY-MM-DDTHH:MM
updated: YYYY-MM-DDTHH:MM
```

- Set `status` to `backlog` for new issues unless there's reason to start higher.
- Add `parent: <id>` if this is a sub-issue.
- Add `blocked_by: [id1, id2]` if the issue depends on other issues being completed first.
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

## Dependencies

To mark an issue as blocked by other issues, add `blocked_by` to meta.yaml:

```yaml
title: "Build preferences UI"
status: backlog
priority: medium
labels: [frontend]
blocked_by: [19, 20]
created: 2026-03-02T14:10
updated: 2026-03-02T14:10
```

The inverse (`blocks`) is computed at load time — only `blocked_by` is stored on disk.
