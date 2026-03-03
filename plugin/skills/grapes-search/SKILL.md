---
name: grapes-search
description: "Use when you need to search across issues — by keyword, content, or to find related issues."
user-invokable: false
---

# Searching Issues

## Search Titles

```bash
grep -l "<query>" .grapes/*/meta.toml
```

## Search Descriptions

```bash
grep -rl "<query>" .grapes/*/content.md
```

## Search Comments

```bash
grep -rl "<query>" .grapes/*/comments.md
```

## Search Everything

```bash
grep -rl "<query>" .grapes/
```

This searches across titles, descriptions, and comments. Results are file paths — extract the issue ID from the path (`.grapes/<id>/...`).

## Find Related Issues

To find issues related to a topic, combine searches:

```bash
grep -rl "auth\|login\|oauth" .grapes/
```

## Find Issues Mentioning Another Issue

Issues may reference each other with `#<id>`:

```bash
grep -rl "#42" .grapes/
```

## After Searching

Once you have matching file paths, read the meta.toml of each matching issue to get the title, status, and other metadata. Don't read the full content unless you need it.
