# Handoff Mode: Continue

Hand off partial work. The receiving agent will use `/issue continue` to pick up where you stopped.

## Guidance

- Document where you stopped and why.
- List what's done (files changed, commits made).
- Describe the current state of the code (uncommitted changes, half-finished work).
- Provide ordered remaining steps.
- Share gotchas you discovered.

**Document all in-flight state.** Uncommitted changes, branch status, half-finished refactors — if the next agent doesn't know about it, they'll either redo it or break it.

## Template

```markdown
## Issue
#<id>: [title]

## Where I Stopped
[What you were doing when you stopped and why]

## What's Done
- [Completed work with file paths]
- [Commits made: hash + description]
- [Comments posted on issue]

## Current State
- Branch: [name and status]
- Uncommitted changes: [exact files and what's changed, or "none"]
- In-flight work: [anything half-finished]

## What Remains

### Next Step: [Short title]
- **File**: `path/to/file`
- **What**: Exact description
- **Details**: Implementation guidance

### After That: [Short title]
...

## Decisions Made
- [Decision]: [rationale] — don't revisit

## Gotchas
- [Errors you hit and how you solved/avoided them]
- [Things that weren't obvious from the issue]

## Verify
[Exact command(s)]
**Pass criteria**: [Exact expected outcome]

## Instructions
Follow `/issue continue` to pick up this work.
Read the issue skill: .claude/skills/issue/SKILL.md
```
