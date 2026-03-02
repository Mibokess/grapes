# Continue Handoff Template

The next agent will pick up where you left off. Give them a clear picture of current state so they don't redo or break your work.

## Plan Structure

```markdown
## Linear Issue
[Issue ID and title]

## Where I Stopped
[What you were doing when you stopped and why — be specific]

## What's Done
- [Completed work with file paths]
- [Commits made: hash + description]
- [Linear comments posted]

## Current State of the Code
- [Uncommitted changes, if any — exact files and what's changed]
- [Anything half-finished or in-flight]
- [Branch name and status]

## What Remains

### Next Step: [Short title]
- **File**: `path/to/file.py`
- **What**: Exact description of the change
- **Symbols**: `ClassName.method_name` — current state and what to change
- **Details**: Implementation guidance
- **Watch out**: Gotchas or constraints

### After That: [Short title]
...

## Decisions Made
- [Decision]: [rationale] — don't revisit this
- [Decision]: [rationale]

## Problems & Gotchas
- [Errors you hit and how you solved or avoided them]
- [Things that weren't obvious from the issue]
- [Constraints you discovered]

## Verify
```bash
[Exact command(s)]
```
**Pass criteria**: [Exact expected outcome]
```

## Checklist

1. **Where you stopped** is specific enough to resume immediately
2. **What's done** lists every file touched and commit made
3. **Current state** covers uncommitted changes and in-flight work
4. **Remaining steps** are ordered and actionable
5. **All decisions made** — the next agent continues, not re-decides
6. **Problems & gotchas** captures hard-won knowledge
7. **No gaps** between done and remaining — nothing falls through
8. **A fresh agent can pick this up cold**
