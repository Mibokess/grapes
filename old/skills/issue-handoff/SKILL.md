---
name: issue-handoff
description: Enter plan mode and write a detailed handoff plan from your current working context so another Claude Code agent can pick up the work. Modes - implement (full implementation), continue (finish partial work), research (investigate something), verify (independent issue verification), comprehensive (make issue implementation-ready).
argument-hint: <implement|continue|research|verify|comprehensive> [issue-id] [additional context...]
disable-model-invocation: true
---

# Write Handoff Plan

You are working on a Linear issue and need to hand off to another Claude Code agent. Parse `$ARGUMENTS` for:
1. **Mode** (first word): `implement`, `continue`, `research`, `verify`, or `comprehensive`. If no mode is given, ask the user.
2. **Issue ID** (optional): e.g. `ETH-123`. If not given, use the issue from your current context.
3. **Flags** (optional):
   - `--team` — Design an agent team for the receiving agent to use. See [agent-team.md](agent-team.md) for guidance. Include the team design as a dedicated section in the plan.
4. **Additional context** (optional): Everything else is free-form guidance from the user — specific areas to focus on, questions to answer, constraints, priorities, etc. Incorporate this into the handoff plan prominently.

## Modes

### `implement` — Hand off for implementation

The next agent will implement the issue from scratch. Your job is to give them a complete implementation roadmap so they can write the code without doing any research.

Focus on:
- The full picture: goal, all files involved, all symbols, all constraints
- Concrete implementation steps in execution order
- Every decision already made (with rationale)
- Patterns and conventions to follow
- Test strategy and verify commands

Use the [implement template](templates/implement.md).

### `continue` — Hand off to continue partial work

The next agent will pick up where you left off. Your job is to give them a clear picture of current state so they don't redo or break your work.

Focus on:
- What's done — files changed, commits made, current state of the code
- What remains — ordered list of remaining steps
- Where you stopped and why
- Anything in-flight (uncommitted changes, half-finished refactors)
- Gotchas you hit that they'll hit too

Use the [continue template](templates/continue.md).

### `research` — Hand off for investigation

The next agent will research a question or explore a problem. Your job is to give them a focused scope and everything you already know so they don't retrace your steps.

Focus on:
- The specific questions to answer
- What you've already looked at (files, symbols, docs) and what you found
- What you haven't looked at yet and why it matters
- Where to start and suggested approach
- What the findings should look like

Use the [research template](templates/research.md).

### `verify` — Hand off for independent issue verification

The next agent will independently verify a Linear issue against the current codebase. Their job is to treat the issue as untrusted — trace every claim, check every file path and symbol, verify the proposed changes would achieve the goal, and surface anything incomplete or incorrect. **The verification agent must not be biased by the original author's assumptions.** They verify everything themselves from first principles.

Focus on:
- The exact issue to verify (ID, all sub-issues if any)
- What claims the issue makes that need independent verification
- Which files, symbols, and behaviors the verifier should trace through
- What "complete and correct" looks like for this issue
- Instructions to update the issue directly if anything is missing, wrong, or unclear

Use the [verify template](templates/verify.md).

### `comprehensive` — Hand off to make issue implementation-ready

The next agent will take a Linear issue and make it fully comprehensive, self-contained, and implementable by any agent on the first try. Your job is to give them the issue and any context you have. **The receiving agent should read and follow the [issue-comprehensive skill](../issue-comprehensive/SKILL.md) and its [quality criteria](../issue-comprehensive/quality-criteria.md) for the full process.**

Focus on:
- The exact issue to make comprehensive (ID, all sub-issues if any)
- Any context you already have about the issue (files explored, questions raised, decisions made)
- Specific areas of concern or things you noticed that need attention
- The agent must research the codebase independently — don't pre-chew conclusions

The receiving agent's output is a fully updated Linear issue (and sub-issues if needed), not a plan.

---

## Process

1. **Enter plan mode** using `EnterPlanMode`
2. **Write the handoff** using the template for your mode
3. **Verify** — read through it as a fresh agent with zero context (see template checklist)
4. **Exit plan mode** with `ExitPlanMode`

## Core Principle

**The receiving agent starts with zero context.** Everything you know — files you've read, decisions you've made, gotchas you've hit, patterns you've noticed — must be in the plan. If it's not written down, it doesn't transfer.

## Report

Tell the user:
- Mode used and what the plan covers
- Summary of key content (what's handed off)
- Any open questions or risks for the next agent
- Confirm: "Plan is ready for handoff."
