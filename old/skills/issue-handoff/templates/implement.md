# Implement Handoff Template

The next agent will implement the issue from scratch. Give them everything they need to write the code without doing any research.

## Plan Structure

```markdown
## Linear Issue
[Issue ID and title]

## Goal
[One-sentence summary of what needs to be built/changed and why]

## Key Files

- `path/to/file.py` — role, relevant symbols, current state
- `path/to/other.py` — role, relevant symbols, current state

## Implementation Steps

### Step 1: [Short title]
- **File**: `path/to/file.py`
- **What**: Exact description of the change
- **Symbols**: `ClassName.method_name` (line ~N) — current signature and what to change
- **Details**: Types, patterns to follow, code examples from the codebase
- **Watch out**: Gotchas or constraints

### Step 2: [Short title]
...

## Decisions Made
- [Decision]: [rationale]
- [Decision]: [rationale]

## Patterns & Conventions
- [Relevant CLAUDE.md rules]
- [Code style patterns to follow, with examples from existing code]

## Gotchas
- [Non-obvious things discovered during research]
- [Common mistakes to avoid]

## Test Strategy
- **Existing tests**: `path/to/test_file.py` — what they cover
- **New tests needed**: What to add and where
- **Run**: `uv run pytest path/to/tests/ -x`

## Verify
```bash
[Exact command(s)]
```
**Pass criteria**: [Exact expected outcome]
```

## Checklist

1. **Every file path** exists and is correct
2. **Every symbol** matches current code (name, signature, location)
3. **Every step** is actionable without additional research
4. **All decisions made** — no choices left for the implementing agent
5. **No vague language** ("as needed", "appropriately", "etc.")
6. **Step ordering** respects dependencies
7. **Patterns section** has concrete examples from the codebase, not abstract rules
8. **Test strategy** covers all acceptance criteria
9. **Verify command** is copy-pasteable with unambiguous pass criteria
10. **A fresh agent can implement this cold**
