Users can override specific rules with explicit instructions (e.g., "skip MR for this", "don't commit"). 
Override only the rule explicitly mentioned — all other rules remain in effect.

# Guidelines

## 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

## 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

## 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

## 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.

---

**These guidelines are working if:** fewer unnecessary changes in diffs, fewer rewrites due to overcomplication, and clarifying questions come before implementation rather than after mistakes.


## 5. Issue Tracking (peas)

This project tracks work with peas (file-based issue tracker). Issues live in `.peas/` as markdown files.

Before making code changes:
1. Check peas for an existing issue: `peas list` or `peas search "..."`
2. If none exists, create one first

What requires a peas issue:
- Code changes, file edits, bug fixes, refactoring, documentation changes

What does not require a peas issue:
- Ad-hoc research, quick exploration, answering questions, reading files

Use `type = "research"` for structured investigation that needs tracking or findings that need to be preserved.

Read and follow `docs/peas/peas-usage.md` for the complete workflow: issue templates, activity log conventions, state management, and handoffs.

## 6. Coding Philosophy

This is a research codebase. Prioritize clarity and correctness over production concerns.

- No backward compatibility! Change interfaces freely. Don't preserve deprecated code paths.
- Fail explicitly! Let errors propagate. No defensive error handling or silent fallbacks.
- No future-proofing! Solve today's problem. Don't add abstractions for hypothetical use cases.
- No over-engineering! Only make changes that are directly requested or clearly necessary.

## 7. Documentation

The `docs/` folder contains research and reference material. Consult it before external sources. When you discover new information not covered in docs, add it.

`CLAUDE.md` files exist throughout `src/` directories and contain module-level documentation: purpose, exports, and usage patterns. Read them when working in those directories.

## 8. Git

### Commit as You Go

Commit format: `ISSUE-ID: message` (e.g., `ETH-00042: Add config loader`)

Do NOT save changes for one big commit at the end. Commit after each meaningful logical unit: one fix, one function, one test. Each commit should be understandable on its own.

### Feature Branches

All implementation work happens on feature branches: `ETH-XXXXX/short-description`. Never commit directly to `main`.

### Merge Requests

When work is done:
1. Append `[DONE]` to peas issue, `peas done ETH-XXXXX`, commit the issue file
2. Push the branch and create an MR via `/merge-request`
3. Report to the user: "MR ready for review: <URL>"
4. **STOP.** Wait for user review.

MRs are squash-merged on GitLab, so individual commits are preserved in the MR but collapse into one clean commit on the target branch.

### Never Self-Merge

**Agents NEVER merge MRs without explicit user approval.**

The issue is marked done and included in the MR. The MR is the review gate:
- If rejected: `peas start ETH-XXXXX` (reopen), revise on the same branch, mark done again, push
- If approved: user merges on GitLab

Ad-hoc/no-issue tasks may finish without an MR unless the user explicitly requests one.

Commit ownership:
- Default: the agent that made the changes creates the commits and MR.
- Exception: if the user explicitly assigns a different integrator/lead, that assignee pushes and creates the MR.
- If commit cannot be created, append `[BLOCKED]` to the peas issue and do not mark Done.

## 9. Surface Results in Responses

**If you mention it, show it. If you saved it, say where.**

- If your response references a table, figure, or result — include it inline. Do not reference results that live only in your internal reasoning.
- If you save any output to a file, state the full file path in your response.

This does not mean dumping all intermediate work into every response. Sub-analysis and internal steps can stay internal. But anything the response points to must be present or have a path.

## 10. Subagents

Provide the peas issue ID to subagents.
Subagents update the peas issue with progress and findings (append log entries to issue body).

Subagent completion rules:
- For tracked implementation work, if a subagent makes file changes, that subagent marks the issue done, commits the issue file, and creates the MR by default.
- If a lead/integrator is explicitly assigned, subagents hand off with `[VERIFY]`/`[PROGRESS]` log entries and the integrator closes out and creates the MR.
- The MR is the review gate. User merges after approval.

## 11. Claude
If you are claude and haven't yet read CLAUDE.md, do so now!