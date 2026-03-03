---
name: handoff
description: "Write a handoff plan so another agent can pick up work. Modes: plan, verify, implement, research, continue."
argument-hint: <plan|verify|implement|research|continue> [issue-id] [context...]
user-invokable: true
---

# Write Handoff Plan

You need to hand off work to another agent. Parse `$ARGUMENTS` for:
1. **Mode** (first word): `plan`, `verify`, `implement`, `research`, or `continue`. If missing, ask the user.
2. **Issue ID** (optional): numeric ID. If not given, infer from current context or ask.
3. **Additional context** (optional): Everything else is free-form guidance — areas to focus on, constraints, priorities. Incorporate prominently into the handoff.

## Core Principle

**The receiving agent starts with zero context.** Everything you know — files you've read, decisions you've made, gotchas you've hit, patterns you've noticed — must be in the plan. If it's not written down, it doesn't transfer.

## Process

1. **Enter plan mode** using `EnterPlanMode`.
2. **Write the handoff** using the mode-specific guidance below.
3. **Verify** — read through it as a fresh agent with zero context. Can they start immediately? Is anything missing?
4. **Exit plan mode** with `ExitPlanMode`.

---

## Mode: `plan`

Hand off issue-planning work. The receiving agent will use `/issue plan` to make the issue comprehensive and implementation-ready.

**Include:**
- The exact issue ID and current state of its content
- Any context you already have (files explored, questions raised, decisions made)
- Specific areas of concern or things you noticed that need attention
- Point the agent to the [issue skill](../issue/SKILL.md) for the full process

**Don't pre-chew conclusions.** Give context but let the agent research the codebase independently and form their own understanding.

### Template

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
Read the issue skill: plugin/skills/issue/SKILL.md
```

---

## Mode: `verify`

Hand off independent verification. The receiving agent will use `/issue verify` to audit the issue against the codebase.

**Include:**
- The exact issue ID and all sub-issue IDs
- What claims need verification (extract from the issue)
- Which files, symbols, and behaviors to trace

**Don't bias the agent.** Don't tell them what the "right" answer is. Don't share your opinions about whether the issue is correct. Let them read code first and form independent conclusions.

### Template

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
Read the issue skill: plugin/skills/issue/SKILL.md

Do not trust the issue. Read the code first, form your own understanding, then compare.
```

---

## Mode: `implement`

Hand off implementation. The receiving agent will write code. Every decision must already be made — no choices left for the implementer.

**Include:**
- The full picture: goal, all files, all symbols, all constraints
- Concrete implementation steps in execution order
- Every decision already made (with rationale)
- Patterns and conventions to follow (with examples from the codebase)
- Test strategy and verify commands
- Exact expected outcomes

### Template

```markdown
## Issue
#<id>: [title]

## Goal
[One-sentence summary of what to build/change and why]

## Key Files
- `path/to/file` — role, relevant symbols, current state
- `path/to/other` — role, relevant symbols, current state

## Implementation Steps

### Step 1: [Short title]
- **File**: `path/to/file`
- **What**: Exact description of the change
- **Symbols**: `ClassName.method_name` (line ~N) — current signature, what to change
- **Details**: Types, patterns to follow, code examples
- **Watch out**: Gotchas or constraints

### Step 2: [Short title]
...

## Decisions Made
- [Decision]: [rationale]

## Patterns & Conventions
- [Rules from CLAUDE.md or project conventions]
- [Code style patterns, with examples from existing code]

## Gotchas
- [Non-obvious things discovered during research]

## Test Strategy
- **Existing tests**: `path/to/test` — what they cover
- **New tests needed**: What to add and where
- **Run**: [exact test command]

## Verify
[Exact command(s)]
**Pass criteria**: [Exact expected outcome]
```

### Checklist

1. Every file path exists and is correct.
2. Every symbol matches current code (name, signature, location).
3. Every step is actionable without additional research.
4. All decisions made — no choices left for the implementer.
5. No vague language ("as needed", "appropriately", "etc.").
6. Step ordering respects dependencies.
7. Patterns section has concrete examples, not abstract rules.
8. Test strategy covers all acceptance criteria.
9. Verify command is copy-pasteable with unambiguous pass criteria.
10. A fresh agent can implement this cold.

---

## Mode: `research`

Hand off investigation. The receiving agent will use `/issue research` to explore a question.

**Include:**
- The specific research questions to answer
- What you've already investigated (so they don't retrace steps)
- What hasn't been investigated yet and why it matters
- Suggested starting points and approach
- What "done" looks like

### Template

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
Read the issue skill: plugin/skills/issue/SKILL.md
```

---

## Mode: `continue`

Hand off partial work. The receiving agent will use `/issue continue` to pick up where you stopped.

**Include:**
- Where you stopped and why
- What's done (files changed, commits made)
- Current state of the code (uncommitted changes, half-finished work)
- What remains (ordered steps)
- Gotchas you discovered

**Document all in-flight state.** Uncommitted changes, branch status, half-finished refactors — if the next agent doesn't know about it, they'll either redo it or break it.

### Template

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
Read the issue skill: plugin/skills/issue/SKILL.md
```

---

## Report

Tell the user:
- Mode used and what the plan covers
- Summary of key content handed off
- Any open questions or risks for the next agent
- Confirm: "Plan is ready for handoff."
