# Handoff Mode: Verify

Hand off independent verification. The receiving agent will use `/issue verify` to audit the issue against the codebase.

## Guidance

- Include the exact issue ID and all sub-issue IDs.
- Extract the claims that need verification from the issue.
- List files, symbols, and behaviors to trace.

**Don't bias the agent.** Don't tell them what the "right" answer is. Don't share your opinions about whether the issue is correct. Let them read code first and form independent conclusions.

## Template

```markdown
## Issue
#<id>: [title]
Sub-issues: [list IDs if any]

## Verification Scope
[What the issue claims — file paths, symbols, behaviors, proposed changes]

## Files to Read
- `path/to/file` — verify [specific claims about this file]

## Instructions
Follow `/issue verify` to independently verify this issue against the codebase.
Read the issue skill: .claude/skills/issue/SKILL.md

Do not trust the issue. Read the code first, form your own understanding, then compare.
```
