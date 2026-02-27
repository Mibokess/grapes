`find_issues_dir()` walks up the filesystem starting from the Python package's install location (`Path(__file__).resolve().parent`), not the user's working directory. This means the TUI only works when the package happens to be installed inside the project tree.

## Current code

```python
# data.py:51-59
def find_issues_dir() -> Path:
    """Find .grapes/ directory by walking up from this file's location."""
    d = Path(__file__).resolve().parent
    for _ in range(10):
        candidate = d / ".grapes"
        if candidate.is_dir():
            return candidate
        d = d.parent
    raise FileNotFoundError("Could not find .grapes/ directory")
```

## Problem

If the TUI is installed as a package (e.g., via `pip install -e .` or into a venv), `__file__` resolves to the venv or site-packages path — nowhere near the user's project. The function will never find `.grapes/`.

This only works today because the dev venv is inside the project tree:
```
grapes/
  .grapes/         ← found because venv is a child of this
  tui/
    .venv/         ← __file__ starts here, walks up, finds .grapes/
    src/grapes_tui/
```

## Fix

Start from `Path.cwd()` instead:

```python
def find_issues_dir() -> Path:
    d = Path.cwd()
    for _ in range(10):
        candidate = d / ".grapes"
        if candidate.is_dir():
            return candidate
        d = d.parent
    raise FileNotFoundError("Could not find .grapes/ directory")
```

Or better, accept a CLI argument: `grapes-tui --issues-dir /path/to/.issues`.
