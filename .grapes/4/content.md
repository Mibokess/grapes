Opening a single issue in the detail view triggers two full scans of every issue on disk.

## Call chain

1. `detail.py:77` calls `load_issue(self.issue_id)`
2. Inside `load_issue()`, `data.py:123` calls `load_all_issues(issues_dir)` to compute children
3. Back in `detail.py:113`, if the issue has children, it calls `load_all_issues()` **again** to get child titles

```python
# detail.py:77 — first full load (hidden inside load_issue)
issue = load_issue(self.issue_id)

# data.py:123 — load_issue internally loads ALL issues
all_issues = load_all_issues(issues_dir)
children = [i.id for i in all_issues if i.parent == issue_id]

# detail.py:113 — second full load for sub-issue display
all_issues = load_all_issues()
issue_map = {i.id: i for i in all_issues}
```

With 100 issues, that's 200+ YAML files parsed just to view one issue.

## Fix options

1. **Have `load_issue()` return child Issue objects directly** instead of just IDs, so the detail screen doesn't need to reload
2. **Cache `load_all_issues()` results** with a short TTL or filesystem mtime check
3. **Pass the already-loaded issue map** from the calling screen into the detail screen, avoiding reloads entirely
