# Issue Mode: Plan

Make the issue fully comprehensive, implementation-ready, and self-contained. Any agent must be able to implement it correctly on the first try without asking a single question.

## Step 1: Research

Investigate the codebase to build a complete picture:

- **Find every file** that will be read or modified. Confirm they exist.
- **Find every symbol** (class, function, variable) involved. Confirm names, signatures, and locations match current code.
- **Understand current behavior** by reading the actual code, not trusting the issue description.
- **Understand desired behavior.** If ambiguous, ask the user. Do not guess.
- **Discover constraints** from CLAUDE.md files, existing patterns, and project docs.
- **Check dependencies** on other issues or code that must land first.

Trust nothing from the original issue — verify every claim against the code.

## Step 2: Write

Update the issue's `content.md` so it meets all quality criteria:

| Section | Requirements |
|---------|-------------|
| **Goal** | What needs to be done (one sentence). Why it matters. Single deliverable. |
| **Context** | All file paths (relative from repo root). All symbol names with brief descriptions. Current vs. desired behavior. Constraints and conventions. No assumed knowledge. |
| **Acceptance Criteria** | Each criterion is binary pass/fail. Covers all expected changes. Edge cases noted. |
| **Verify** | Exact shell commands, copy-pasteable without modification. |
| **Pass Criteria** | Exact expected output or behavior. Two readers would agree on pass/fail. |

No vague language: "as appropriate", "if needed", "etc.", "probably", "should be" — replace with specifics.

### When to Split into Sub-Issues

Split when the work has 2+ independently testable changes, touches unrelated modules, or has a natural dependency order. Don't split single cohesive changes.

Sub-issue rules:
- Each sub-issue is **fully standalone** — never "see parent for context."
- The parent contains shared context only: motivation, architecture overview, overall goal.
- The parent does NOT duplicate details covered in sub-issues.
- The union of all sub-issues equals the parent scope.

Create sub-issues using the [grapes-create skill](../../../plugin/skills/grapes-create/SKILL.md) with `parent: <id>`.

## Step 3: Verify

Scan every section of every issue (parent and sub-issues) one final time:

1. Every file path exists and points to the right file.
2. Every symbol name is spelled correctly and matches current code.
3. Every described behavior matches what the code actually does.
4. Every claimed relationship is true.
5. No unverified assumptions remain.
6. No open questions remain.
7. No vague language hides uncertainty.

Fix anything that fails. Only finalize once every statement is confirmed fact.

## Step 4: Report

Add a comment to the issue using the [grapes-comment skill](../../../plugin/skills/grapes-comment/SKILL.md) summarizing what was changed. Tell the user:
- What was changed or added to the issue
- Sub-issues created (if any), with IDs and one-line summaries
- Confirm: "All claims verified against code, no open questions remain"
