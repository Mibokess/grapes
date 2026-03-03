---
name: issue
description: "Work on a grapes issue in a specific mode: plan (make comprehensive), verify (check against code), research (investigate), or continue (resume partial work)."
argument-hint: <plan|verify|research|continue> [issue-id]
user-invokable: true
---

# Work on Issue

Parse `$ARGUMENTS` for:
1. **Mode** (first word): `plan`, `verify`, `research`, or `continue`. If missing, ask the user.
2. **Issue ID** (optional): numeric ID. If not given, infer from current branch name or ask.

Read the issue using the [grapes-read skill](../grapes-read/SKILL.md) before starting any mode.

---

## Mode: `plan`

Make the issue fully comprehensive, implementation-ready, and self-contained. Any agent must be able to implement it correctly on the first try without asking a single question.

### Step 1: Research

Investigate the codebase to build a complete picture:

- **Find every file** that will be read or modified. Confirm they exist.
- **Find every symbol** (class, function, variable) involved. Confirm names, signatures, and locations match current code.
- **Understand current behavior** by reading the actual code, not trusting the issue description.
- **Understand desired behavior.** If ambiguous, ask the user. Do not guess.
- **Discover constraints** from CLAUDE.md files, existing patterns, and project docs.
- **Check dependencies** on other issues or code that must land first.

Trust nothing from the original issue — verify every claim against the code.

### Step 2: Write

Update the issue's `content.md` so it meets all quality criteria:

| Section | Requirements |
|---------|-------------|
| **Goal** | What needs to be done (one sentence). Why it matters. Single deliverable. |
| **Context** | All file paths (relative from repo root). All symbol names with brief descriptions. Current vs. desired behavior. Constraints and conventions. No assumed knowledge. |
| **Acceptance Criteria** | Each criterion is binary pass/fail. Covers all expected changes. Edge cases noted. |
| **Verify** | Exact shell commands, copy-pasteable without modification. |
| **Pass Criteria** | Exact expected output or behavior. Two readers would agree on pass/fail. |

No vague language: "as appropriate", "if needed", "etc.", "probably", "should be" — replace with specifics.

#### When to Split into Sub-Issues

Split when the work has 2+ independently testable changes, touches unrelated modules, or has a natural dependency order. Don't split single cohesive changes.

Sub-issue rules:
- Each sub-issue is **fully standalone** — never "see parent for context."
- The parent contains shared context only: motivation, architecture overview, overall goal.
- The parent does NOT duplicate details covered in sub-issues.
- The union of all sub-issues equals the parent scope.

Create sub-issues using the [grapes-create skill](../grapes-create/SKILL.md) with `parent: <id>`.

### Step 3: Verify

Scan every section of every issue (parent and sub-issues) one final time:

1. Every file path exists and points to the right file.
2. Every symbol name is spelled correctly and matches current code.
3. Every described behavior matches what the code actually does.
4. Every claimed relationship is true.
5. No unverified assumptions remain.
6. No open questions remain.
7. No vague language hides uncertainty.

Fix anything that fails. Only finalize once every statement is confirmed fact.

### Step 4: Report

Add a comment to the issue using the [grapes-comment skill](../grapes-comment/SKILL.md) summarizing what was changed. Tell the user:
- What was changed or added to the issue
- Sub-issues created (if any), with IDs and one-line summaries
- Confirm: "All claims verified against code, no open questions remain"

---

## Mode: `verify`

Independently verify the issue against the current codebase. You are an auditor, not an assistant. Do not confirm what the issue says — determine independently what is true.

### Process

1. **Read the code first, then compare to the issue.** Do not read the issue's description and then look for confirmation. Read the code independently, form your own understanding, then check if the issue matches.

2. **Extract every verifiable claim** from the issue:
   - File paths — do they exist? Do they contain what the issue says?
   - Symbol names — correct names, signatures, locations?
   - Current behavior — does the code actually do what the issue says?
   - Proposed changes — would they achieve the stated goal?
   - Dependencies — are the claimed relationships true?
   - Acceptance criteria — are they binary pass/fail and testable?
   - Verify commands — would they actually test what they claim?

3. **Trace the full logic.** Don't just spot-check. Trace through the chain of changes to verify the proposed approach would work.

4. **Assign a verdict to each claim**: `CORRECT`, `INCORRECT` (with what's actually true), or `INCOMPLETE` (with what's missing).

5. **Fix issues in-place.** If you find something wrong or missing, update `content.md` directly. Add missing context, correct wrong paths/names, clarify ambiguous criteria.

6. **Check completeness.** Are there files or symbols that should be mentioned but aren't? Edge cases the criteria miss? Side effects not accounted for?

7. **Verify sub-issues independently.** Each must be self-contained. Read each as if you've never seen the parent.

### Report

Add a `[VERIFY]` comment to the issue:

```
[VERIFY] Independent verification

Verdict: PASS / FAIL / PASS WITH CHANGES

Claims Verified:
- [Claim] — CORRECT / INCORRECT: [actual truth] / INCOMPLETE: [missing info]

Issues Found & Fixed:
- [What was wrong] → [What was updated]

Issues Found & Not Fixed:
- [What needs human decision]

Completeness:
- Missing files/symbols: [any not mentioned]
- Missing edge cases: [any not covered]
```

---

## Mode: `research`

Investigate a question or explore a problem. Research captures knowledge — it doesn't change code.

### Process

1. Read the issue to understand what questions need answering.
2. Investigate the codebase systematically. Record what you look at and what you find.
3. Add `[FINDINGS]` comments as discoveries are made — don't wait until the end.
4. When research questions are answered, add a final `[FINDINGS]` summary.
5. If research reveals that implementation is needed, note this but do not start implementing. Create a new issue if appropriate.

### Report

Tell the user:
- Questions answered and key findings
- Any new issues created as a result of research
- Recommendations for next steps

---

## Mode: `continue`

Pick up partial work on an issue. Another agent (or you in a previous session) started but didn't finish.

### Process

1. Read the issue's `content.md` and `comments.md` to understand the full history.
2. Identify the last activity log entry — understand where work stopped and why.
3. Check the git state: current branch, uncommitted changes, recent commits.
4. Resume from where work was left off. Do not redo completed work.
5. Add a `[STARTED]` comment noting you are continuing and from what state.
6. Continue the work, adding `[PROGRESS]` comments as you go.

### Report

Follow the normal completion flow for the work being done (implementation → PR, research → findings summary).
