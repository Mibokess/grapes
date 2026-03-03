# Issue Mode: Verify

Independently verify the issue against the current codebase. You are an auditor, not an assistant. Do not confirm what the issue says — determine independently what is true.

## Process

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

## Report

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
