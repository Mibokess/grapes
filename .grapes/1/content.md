`parse_comments()` in `tui/src/grapes_tui/data.py:147` splits on every line starting with `### `, which breaks when a comment body itself contains markdown headings or fenced code blocks with `###`.

## Root cause

```python
# data.py:147
blocks = re.split(r"^### ", raw, flags=re.MULTILINE)
```

This regex matches any line starting with `### `, not specifically the `### YYYY-MM-DD` comment header pattern.

## Reproduction

A `comments.md` like this gets misparsed:

```markdown
### 2026-02-27
Here's how to fix the parser:

### The approach

Use a stricter regex.
```

The parser produces two comments: one dated `2026-02-27` and a malformed one from the `### The approach` heading.

## Fix

Replace the naive split with a regex that matches the full header pattern:

```python
COMMENT_HEADER = re.compile(
    r"^### (\d{4}-\d{2}-\d{2})$",
    re.MULTILINE,
)
```

Then use `COMMENT_HEADER.finditer(raw)` and extract text between matches as the body. This also makes the current `header_match` step on line 154 unnecessary since the date is captured directly.
