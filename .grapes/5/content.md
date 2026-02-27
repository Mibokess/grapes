The detail screen renders issue descriptions using `Markdown()` but renders comment bodies using plain `Label()`. This means code blocks, bold text, links, and lists in comments show up as raw markdown syntax.

## Current code

```python
# detail.py:104-105 — description uses Markdown widget (correct)
if issue.content.strip():
    yield Markdown(issue.content)

# detail.py:133 — comments use Label widget (wrong)
yield Label(f"    {c.body}", classes="comment-body")
```

Comments frequently contain code blocks and formatting (agents especially tend to write markdown in comments). These should render the same way descriptions do.

## Fix

Replace the `Label` with `Markdown` for comment bodies:

```python
for c in comments:
    yield Label(f"  {c.author} — {c.date}", classes="comment-header")
    yield Markdown(c.body)
```

## Depends on

Should fix #1 first — the comment parser needs to correctly extract bodies before we can render them as markdown. If the parser is splitting mid-code-block, rendering as markdown will look even worse.
