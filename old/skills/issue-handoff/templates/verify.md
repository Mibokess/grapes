# Verify Handoff Template

The next agent will independently verify a Linear issue against the current codebase. They trust nothing from the issue author — every claim is verified from first principles. If anything is incomplete, incorrect, or ambiguous, they fix it in the issue before reporting.

**The verification agent is not an assistant — they are an auditor.** They do not confirm what the issue says; they determine independently what is true and compare it to what the issue says.

## Plan Structure

```markdown
## Linear Issue
[Issue ID and title. List all sub-issue IDs if the issue has sub-issues.]

## Verification Scope

### Claims to Verify
[Extract every verifiable claim from the issue. Each claim becomes a verification task.]

- [ ] **File paths**: [List every file path mentioned — verify each exists and contains what the issue says]
- [ ] **Symbol names**: [List every class, function, variable mentioned — verify names, signatures, locations]
- [ ] **Current behavior**: [What the issue says the code currently does — verify by reading the actual code]
- [ ] **Proposed changes**: [What the issue says to change — trace the logic to verify it would achieve the goal]
- [ ] **Dependencies**: [Claimed relationships between files/symbols — verify each one]
- [ ] **Acceptance criteria**: [Each criterion — verify it is binary pass/fail and testable]
- [ ] **Verify commands**: [Each command — verify it would actually test what it claims]

### Logic Trace
[Describe the logical flow the verifier should trace through the codebase.]

1. Start at [entry point / main symbol] — read the code, understand what it actually does
2. Trace to [next symbol/file in the chain] — verify the relationship the issue describes
3. [Continue through the full chain of changes]
4. At each step: does the issue's description match reality? Would the proposed change work?

### Files to Read
[Ordered list of files the verifier must read, with what to look for in each.]

- `path/to/file.py` — verify [specific claims about this file]
- `path/to/other.py` — verify [specific claims about this file]

## Verification Rules

The verification agent MUST follow these rules:

1. **Read the code first, then compare to the issue.** Do not read the issue's description of the code and then look for confirmation. Read the code independently, form your own understanding, then check if the issue matches.
2. **Every claim gets a verdict.** For each claim: CORRECT, INCORRECT (with what's actually true), or INCOMPLETE (with what's missing).
3. **Trace the full logic.** Don't just check that symbols exist — trace through the logic to verify the proposed changes would actually achieve the stated goal.
4. **If something is unclear, it's a defect.** The issue must be complete enough that any agent can implement it without questions. Ambiguity = issue needs updating.
5. **Fix the issue, don't just report.** If you find something wrong or missing, update the issue directly (add missing context, correct wrong paths/names, clarify ambiguous criteria). Then note what you changed.
6. **Check completeness.** Are there files or symbols that SHOULD be mentioned but aren't? Are there edge cases the acceptance criteria miss? Are there side effects the issue doesn't account for?
7. **Verify sub-issues independently.** Each sub-issue must be self-contained. Read each as if you've never seen the parent. Flag any that depend on unstated context.

## Expected Output

The verification agent should post a `[VERIFY]` comment on the Linear issue with:

### Format
```
[VERIFY] Independent verification of ETH-XXX

**Verdict**: PASS / FAIL / PASS WITH CHANGES

### Claims Verified
- [Claim] — CORRECT / INCORRECT: [what's actually true] / INCOMPLETE: [what's missing]
- ...

### Issues Found & Fixed
- [What was wrong] → [What was updated in the issue]
- ...

### Issues Found & Not Fixed
- [What needs human decision — e.g., ambiguous requirements where both interpretations are valid]
- ...

### Completeness Check
- Missing files/symbols: [any that should be mentioned but aren't]
- Missing edge cases: [any the criteria don't cover]
- Missing side effects: [any the issue doesn't account for]

### Logic Trace Summary
[Brief walkthrough of the traced logic and whether the proposed changes would achieve the goal]
```

## Additional Context
[Any specific concerns, areas of doubt, or things the user asked to focus on]
```

## Checklist

1. **Every verifiable claim** in the issue is listed as a verification task
2. **Logic trace** walks through the full chain of changes, not just spot-checks
3. **Files to read** are ordered for the verifier to build understanding incrementally
4. **Verification rules** are clear — the agent knows to read code first, issue second
5. **Expected output** format is specific — the agent knows exactly what to produce
6. **Sub-issues included** — if the issue has sub-issues, each is verified independently
7. **No bias leakage** — the plan does not tell the verifier what the "right" answer is, only what to check
8. **Fix-in-place instructions** — the agent knows to update the issue directly, not just report
9. **A fresh agent can verify this cold**
