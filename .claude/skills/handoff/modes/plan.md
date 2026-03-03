# Handoff Mode: Plan

Hand off issue-planning work. The receiving agent will use `/issue plan` to make the issue comprehensive and implementation-ready.

## Guidance

- Include the exact issue ID and current state of its content.
- Include any context you already have (files explored, questions raised, decisions made).
- Note specific areas of concern or things that need attention.
- Point the agent to the issue skill for the full process.

**Don't pre-chew conclusions.** Give context but let the agent research the codebase independently and form their own understanding.

## Template

```markdown
## Issue
#<id>: [title]

## Current State
[What the issue currently contains, what's missing or unclear]

## Context You Have
- [Files explored, patterns noticed, questions raised]
- [Any decisions or constraints discovered]

## Areas of Concern
- [Specific things to focus on or investigate further]

## Instructions
Follow `/issue plan` to make this issue comprehensive and implementation-ready.
Read the issue skill: .claude/skills/issue/SKILL.md
```
