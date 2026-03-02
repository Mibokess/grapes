---
name: issue-comprehensive
description: Make a Linear issue fully comprehensive, self-contained, and independently implementable by any agent on the first try. Researches the codebase, fills in missing context, adds sub-issues if needed, and ensures the parent correctly encompasses all sub-issue work.
argument-hint: [--handoff] [issue-id]
disable-model-invocation: true
---

# Make Issue Implementation-Ready

Parse `$ARGUMENTS` for:
1. **`--handoff`** (optional flag): If present, do NOT execute the steps yourself. Instead, read the [issue-handoff skill](../.claude/skills/issue-handoff/SKILL.md) and use it in `comprehensive` mode to write a handoff plan for a fresh agent to do this work.
2. **Issue ID** (optional): e.g. `ETH-123`. If not given, use the issue from your current context.

Without `--handoff`: fetch the Linear issue and transform it into a fully standalone, implementation-ready issue.

## Core Principle

**Any agent must be able to implement this issue correctly on the first try without asking a single question.** Every statement in the issue must be verified fact. No assumptions, no "TBD", no "probably", no vague language.

---

## Step 1: Research

Fetch the issue and all its comments. Then investigate the codebase to build a complete picture:

- **Find every file** that will be read or modified. Confirm they exist.
- **Find every symbol** (class, function, variable) involved. Confirm names, signatures, and locations match current code.
- **Understand current behavior** by reading the actual code, not trusting the issue description.
- **Understand desired behavior.** If ambiguous, ask the user. Do not guess.
- **Discover constraints** from CLAUDE.md files, existing patterns, and docs/.
- **Check dependencies** on other issues or code that must land first.

Trust nothing from the original issue -- verify every claim against the code.

## Step 2: Write

Update the issue so it meets all quality criteria (see [quality-criteria.md](quality-criteria.md)). If the work should be split, create sub-issues.

### When to Split into Sub-Issues

Split when the work has 2+ independently testable changes, touches unrelated modules, or has a natural dependency order. Don't split single cohesive changes or when sub-issues can't be verified on their own.

### Sub-Issue Rules

- Each sub-issue is **fully standalone** -- an agent reading only that sub-issue has everything it needs. Never write "see parent for context."
- The parent contains **shared context only**: motivation, architecture overview, overall goal, shared constraints.
- The parent does NOT duplicate details covered in sub-issues.
- The parent's acceptance criteria describe the **overall outcome**, not per-sub-issue work.
- The parent lists all sub-issues with one-line summaries and dependency order.
- The union of all sub-issues equals the parent scope -- nothing lost, nothing added.

## Step 3: Verify

Scan every section of every issue (parent and sub-issues) one final time. See the [verification checklist](quality-criteria.md#verification-checklist) for the full list.

For anything that fails: fix it from the codebase, or ask the user. Only finalize once every statement is confirmed fact.

## Step 4: Report

Tell the user:
- What was changed or added to the issue
- Sub-issues created (if any), with IDs and one-line summaries
- Confirm: "All claims verified against code, no open questions remain"
