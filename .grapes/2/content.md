The board screen uses direct dictionary access for `PRIORITY_LABELS` and `STATUS_LABELS`, which crashes with `KeyError` if an issue has a value not in the map. The list and detail screens already handle this correctly with `.get()`.

## Inconsistency

```python
# board.py:132 — CRASHES on unknown priority
meta_parts.append(PRIORITY_LABELS[i.priority])

# board.py:168 — CRASHES on unknown status
f"{STATUS_LABELS[self.status]} ({len(self.issues)})"

# list.py:52-53 — SAFE, uses .get()
STATUS_LABELS.get(issue.status, issue.status),
PRIORITY_LABELS.get(issue.priority, issue.priority),

# detail.py:87-88 — SAFE, uses .get()
STATUS_LABELS.get(issue.status, issue.status)
PRIORITY_LABELS.get(issue.priority, issue.priority)
```

## Reproduction

Create an issue with a typo or custom status:

```yaml
title: Test issue
status: wip
priority: critical
```

Open the board view — `KeyError: 'critical'` crash.

## Fix

Use `.get()` with the raw value as fallback, matching what list and detail screens already do:

```python
# board.py:132
meta_parts.append(PRIORITY_LABELS.get(i.priority, i.priority))

# board.py:168
f"{STATUS_LABELS.get(self.status, self.status)} ({len(self.issues)})"
```
