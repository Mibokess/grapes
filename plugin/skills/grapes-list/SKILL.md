---
name: grapes-list
description: "Use when you need to find, list, browse, or filter issues by status, priority, or labels."
user-invokable: false
---

# Listing & Filtering Issues

## By Status

```bash
grep -l "status: todo" .grapes/*/meta.yaml
```

Replace `todo` with: `backlog`, `in_progress`, `done`, `cancelled`.

## All Open Issues

```bash
grep -l "status: todo\|status: in_progress\|status: backlog" .grapes/*/meta.yaml
```

## By Priority

```bash
grep -l "priority: urgent" .grapes/*/meta.yaml
```

## By Label

```bash
grep -l "labels:.*bug" .grapes/*/meta.yaml
```

## All Issue IDs

```bash
ls .grapes/
```

## Get Summaries

To get a quick overview, read just the meta.yaml files. Each is ~8 lines:

```bash
head -2 .grapes/*/meta.yaml
```

This prints the title and status of every issue — enough for a summary without reading descriptions.

## Sub-Issues

Direct children of issue 40:

```bash
grep -l "parent: 40" .grapes/*/meta.yaml
```

Top-level issues (no parent):

```bash
grep -rL "parent:" .grapes/*/meta.yaml
```
