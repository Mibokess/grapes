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
2. **Read the mode-specific file** (see table below) and write the handoff following its template.
3. **Verify** — read through it as a fresh agent with zero context. Can they start immediately? Is anything missing?
4. **Exit plan mode** with `ExitPlanMode`.

| Mode | File | Purpose |
|------|------|---------|
| `plan` | [modes/plan.md](modes/plan.md) | Hand off issue-planning work |
| `verify` | [modes/verify.md](modes/verify.md) | Hand off independent verification |
| `implement` | [modes/implement.md](modes/implement.md) | Hand off implementation (all decisions pre-made) |
| `research` | [modes/research.md](modes/research.md) | Hand off investigation |
| `continue` | [modes/continue.md](modes/continue.md) | Hand off partial work |

Read **only** the mode file you need. Do not read the others.

## Report

Tell the user:
- Mode used and what the plan covers
- Summary of key content handed off
- Any open questions or risks for the next agent
- Confirm: "Plan is ready for handoff."
