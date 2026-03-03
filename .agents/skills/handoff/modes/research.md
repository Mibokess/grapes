# Handoff Mode: Research

Hand off investigation. The receiving agent will use `/issue research` to explore a question.

## Guidance

- List the specific research questions to answer.
- Document what you've already investigated so they don't retrace steps.
- List what hasn't been investigated yet and why it matters.
- Provide suggested starting points and approach.
- Define what "done" looks like.

## Template

```markdown
## Issue
#<id>: [title]

## Research Questions
1. [Specific question to answer]
2. [Specific question to answer]

## What's Already Known

### Investigated
- `path/to/file` — what was looked at, what was found
- [Source] — key findings

### Findings So Far
- [Verified fact]
- [Hypothesis — clearly marked as unverified]

## What Hasn't Been Investigated
- `path/to/unexplored` — why it matters, what to look for
- [Area/topic] — why it's relevant

## Suggested Approach
1. [Start here — why]
2. [Then look at this — why]

## Scope
- **In scope**: [what to investigate]
- **Out of scope**: [what to skip]
- **Stop when**: [what "done" looks like]

## Expected Output
- [What the findings should look like]
- [Post as `[FINDINGS]` comments on the issue]

## Instructions
Follow `/issue research` to investigate.
Read the issue skill: .claude/skills/issue/SKILL.md
```
