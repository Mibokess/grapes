`yaml.safe_load()` is called without any error handling in both `load_all_issues()` and `load_issue()`. A single malformed `meta.yaml` crashes the entire TUI — you can't view any issues at all.

## Affected code

```python
# data.py:73 — no try/except
meta = yaml.safe_load(meta_path.read_text())

# data.py:110 — same problem
meta = yaml.safe_load(meta_path.read_text())
```

## Reproduction

```bash
echo "invalid: [syntax" > .grapes/99/meta.yaml
python -m grapes_tui
# yaml.scanner.ScannerError — TUI never starts
```

## Expected behavior

Skip the malformed issue and continue loading the rest. Ideally show a warning in the UI (e.g., a red card on the board saying "Issue #99: parse error").

## Fix

Wrap the YAML load in a try/except:

```python
for entry in sorted(issues_dir.iterdir()):
    if entry.is_dir() and entry.name.isdigit():
        meta_path = entry / "meta.yaml"
        if not meta_path.exists():
            continue
        try:
            meta = yaml.safe_load(meta_path.read_text())
        except yaml.YAMLError:
            continue  # skip malformed issues
```

A `None` check on `meta` would also be good — `yaml.safe_load("")` returns `None`, not an empty dict.
