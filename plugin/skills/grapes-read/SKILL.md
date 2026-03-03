---
name: grapes-read
description: "Use when you need to read full details of a specific issue — its metadata, description, or comments."
user-invokable: false
---

# Reading an Issue

## Metadata Only (Most Common)

```
Read .grapes/<id>/meta.toml
```

This is ~7 lines. Enough for status, priority, labels, and title.

## Full Description

```
Read .grapes/<id>/content.md
```

Only read this when you need to understand what the issue is about, not just its status.

## Comments

```
Read .grapes/<id>/comments.md
```

Only read when you need the discussion history.

## Reading Order

1. **meta.toml first** — always. It's tiny and tells you if you even need the rest.
2. **content.md second** — only if you need the full description.
3. **comments.md last** — only if you need the discussion.

Don't read all three by default. Read what you need.

## Sub-Issues

To find children of an issue:

```bash
grep -l "parent = <id>" .grapes/*/meta.toml
```

Then read their meta.toml files to understand the breakdown.

## Dependencies

If meta.toml contains `blocked_by = [3, 5]`, this issue depends on issues 3 and 5.

To find what issues a given issue blocks (the inverse), search for it in other issues' `blocked_by` lists:

```bash
grep -l "blocked_by.*<id>" .grapes/*/meta.toml
```
