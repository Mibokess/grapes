Users can override specific rules with explicit instructions (e.g., "skip PR for this", "don't commit"). Override only the rule explicitly mentioned — all other rules remain in effect.

## Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them — don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

## Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

## Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it — don't delete it.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: every changed line should trace directly to the request.

## Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

## Issue-First Policy

**Before editing code, have a grapes issue.** No exceptions.

1. Search for an existing issue: `grep -rl "keyword" .grapes/*/content.md` or scan `meta.toml` titles.
2. If none exists, create one (see `plugin/skills/grapes-create/SKILL.md`).
3. New issues start in `backlog`. **Stop and tell the user.** Do not start work until the issue is moved to `todo`.
4. If the issue is `todo` or `in_progress` — proceed.

If you are about to edit a file and there is no issue — stop. Create the issue first. Issue first, code second.

What requires an issue:
- Code changes, file edits, bug fixes, refactoring, documentation changes.

What does not require an issue:
- Ad-hoc research, quick exploration, answering questions, reading files.

## Issue Lifecycle

```
backlog ──[human approves]──> todo ──> in_progress ──> done
                                                    ──> cancelled
```

- **backlog**: Created but not approved. Work does not start.
- **todo**: Approved and ready for implementation.
- **in_progress**: Being worked on.
- **done**: Complete and verified.
- **cancelled**: Abandoned.

**Research issues can skip the approval gate.** Move directly to `todo` or `in_progress` since research informs decisions rather than changing code.

## Git Conventions

### Branch Naming

```
<id>/short-description
```

Example: `42/fix-parser-crash`

### Commit Messages

```
#<id>: Imperative description of what this commit does
```

Example: `#42: Fix off-by-one in token boundary check`

Keep under 72 characters. The issue ID prefix is required for all tracked work.

### Commit as You Go

Commit after each meaningful unit of work. Do not save everything for one big commit at the end.

```
git checkout -b 42/fix-parser-crash

# fix the bug
git add src/parser.go
git commit -m "#42: Fix off-by-one in token boundary check"

# add test
git add src/parser_test.go
git commit -m "#42: Add test for malformed input edge case"
```

Each commit should be a single logical change: one fix, one function, one test.

## Activity Log

Use `comments.md` to record structured progress with tags:

| Tag | When | Content |
|-----|------|---------|
| `[STARTED]` | Beginning work | Brief description of approach |
| `[PROGRESS]` | Intermediate update | What's done, what's next |
| `[FINDINGS]` | Research results | Files found, patterns, data |
| `[DECISION]` | Choice made | Decision + rationale |
| `[VERIFY]` | Test results | Command, output, PASS/FAIL |
| `[BLOCKED]` | Can't proceed | Blocker, what was tried, what's needed |
| `[DONE]` | Task complete | Files changed, summary |

If it's not in the issue, it doesn't exist. The next agent will waste time rediscovering it.

## PR as Review Gate

When implementation is complete:

1. Push the branch.
2. Create a PR with title `#<id>: description`.
3. **Stop.** Report the PR URL to the user and wait for review.

**Agents never merge without explicit human approval.** The PR is the gate, not the issue status.

If the reviewer requests changes: revise on the same branch, push new commits, notify the user.

Always use / read the `/pr` skill to create merge requests.

## Issue Quality Criteria

Every issue must be **self-contained and implementation-ready**. Any agent must be able to implement it correctly on the first try without asking questions.

| Section | Requirements |
|---------|-------------|
| **Goal** | What needs to be done (one sentence). Why it matters. Single deliverable. |
| **Context** | All file paths (relative from repo root). All symbol names. Current vs. desired behavior. Constraints. No assumed knowledge. |
| **Acceptance Criteria** | Each criterion is binary pass/fail. Covers all expected changes. Edge cases noted. |
| **Verify** | Exact shell commands, copy-pasteable without modification. |
| **Pass Criteria** | Exact expected output or behavior. Two readers would agree on pass/fail. |

### Common Problems

| Problem | Fix |
|---------|-----|
| Vague goal ("refactor the pipeline") | Name the exact functions and what changes |
| Missing file paths ("update the config loader") | Find and list the actual path |
| Implicit context ("same pattern as X") | Spell out the pattern with code references |
| Untestable criteria ("code should be clean") | Replace with binary pass/fail checks |
| No verify command | Add exact test commands |
| Vague language ("as appropriate", "if needed", "etc.") | Replace with specifics |

### Verification Checklist

1. Every file path exists and points to the right file.
2. Every symbol name is spelled correctly and matches current code.
3. Every described behavior matches what the code actually does.
4. Every claimed relationship (e.g., "called from X") is true.
5. No unverified assumptions remain ("probably", "should be").
6. No open questions remain ("TBD", "TODO", "need to check").
7. No vague language hides uncertainty.

## Issue Template

```markdown
## Goal
[What needs to be achieved — clear and actionable]

## Context
[ALL relevant information]
- File paths discovered
- Code patterns identified
- Constraints and dependencies

## Acceptance Criteria
- [ ] [Binary pass/fail criterion]
- [ ] [Binary pass/fail criterion]

## Verify
```bash
[Exact command(s)]
```

## Pass Criteria
[Exact expected output or behavior]
```

## Sub-Issue Rules

Split when the work has 2+ independently testable changes, touches unrelated modules, or has a natural dependency order. Don't split single cohesive changes.

- Each sub-issue is **fully standalone** — an agent reading only that sub-issue has everything it needs. Never write "see parent for context."
- The parent contains **shared context only**: motivation, architecture overview, overall goal, shared constraints.
- The parent does NOT duplicate details covered in sub-issues.
- The union of all sub-issues equals the parent scope — nothing lost, nothing added.
- Each sub-issue must be independently verifiable.

## Dependencies

Use `blocked_by` in `meta.yaml` to declare dependencies between issues.

```yaml
blocked_by: [19, 20]
```

Independent tasks run in parallel. Dependent tasks wait for their blockers to complete.

## Temporary Files

**Save temporary outputs to `.grapes/<id>/tmp/`, not `/tmp/`.**

When working on an issue, all temporary files — images, scratch files, debug logs, generated artifacts — go in `.grapes/<id>/tmp/`. Create the directory with `mkdir -p` on first use.

When no issue exists yet (e.g., early exploration before creating an issue), use `.grapes/tmp/` as a project-level scratch space. Once an issue is created for the work, move any relevant files from `.grapes/tmp/` into `.grapes/<id>/tmp/`.

Both paths are gitignored.

## Surface Results

**If you mention it, show it. If you saved it, say where.**

- If your response references a table, figure, or result — include it inline.
- If you save output to a file, state the full file path in your response.

Sub-analysis and internal steps can stay internal. But anything the response points to must be present or have a path.

## Subagents

When delegating to subagents:
- Provide the grapes issue ID.
- Subagents update the issue with progress and findings (append comments via `grapes-comment`).
- For tracked implementation work, the agent that makes file changes creates the commits and PR by default.
- If a lead/integrator is explicitly assigned, subagents hand off with `[PROGRESS]`/`[VERIFY]` comments and the integrator closes out.

## Grapes Skills Reference

Detailed mechanics for issue manipulation are the grapes skills.

- `.claude/skills/grapes/SKILL.md` — Format reference
- `.claude/skills/grapes-create/SKILL.md` — Creating issues
- `.claude/skills/grapes-read/SKILL.md` — Reading issues
- `.claude/skills/grapes-update/SKILL.md` — Updating metadata
- `.claude/skills/grapes-list/SKILL.md` — Listing and filtering
- `.claude/skills/grapes-search/SKILL.md` — Searching
- `.claude/skills/grapes-comment/SKILL.md` — Adding comments
- `.claude/skills/grapes-close/SKILL.md` — Closing issues
