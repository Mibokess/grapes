---
name: grapes-list
description: "Use when you need to find, list, browse, or filter issues by status, priority, or labels."
user-invokable: false
---

# Listing & Filtering Issues

**Always** start by globbing for `.grapes/*/meta.toml`. If no matches, there are no issues yet. The commands below assume at least one issue exists.

## By Status

```bash
grep -l "status = 'todo'" .grapes/*/meta.toml
```

Replace `todo` with: `backlog`, `in_progress`, `done`, `cancelled`.

## All Open Issues

```bash
grep -l "status = 'todo'\|status = 'in_progress'\|status = 'backlog'" .grapes/*/meta.toml
```

## By Priority

```bash
grep -l "priority = 'urgent'" .grapes/*/meta.toml
```

## By Label

```bash
grep -l "labels.*bug" .grapes/*/meta.toml
```

## All Issue IDs

```bash
ls .grapes/
```

## Get Summaries

To get a quick overview, read just the meta.toml files. Each is ~8 lines:

```bash
head -2 .grapes/*/meta.toml
```

This prints the title and status of every issue — enough for a summary without reading descriptions.

## Sub-Issues

Direct children of issue 40:

```bash
grep -l "parent = 40" .grapes/*/meta.toml
```

Top-level issues (no parent):

```bash
grep -rL "parent = " .grapes/*/meta.toml
```

## Blocked Issues

Issues that are blocked by something:

```bash
grep -l "blocked_by = " .grapes/*/meta.toml
```

Issues blocked by a specific issue (e.g. issue 5):

```bash
grep -l "blocked_by.*5" .grapes/*/meta.toml
```
