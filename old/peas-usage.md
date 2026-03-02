# Peas Workflow for AI Agents

Work happens in two phases: first create a detailed issue, then implement it after human approval.

Scope: this document governs tracked peas work. Ad-hoc/no-issue tasks are outside this workflow and do not require a commit unless explicitly requested.

peas is a file-based issue tracker. Issues are markdown files with TOML frontmatter stored in `.peas/`, fully git-trackable. No server, no internet required.

## Two Phases

| Phase          | Focus                                 | Output         | End State                    |
| -------------- | ------------------------------------- | -------------- | ---------------------------- |
| Issue Creation | Research, gather context, write issue | Detailed issue | `draft` + `needs-review` tag |
| Implementation | Execute the approved issue            | Code + commits | `completed`                  |

### Issue Creation

1. Research codebase, gather context
2. Create issue, then write the body directly into the `.peas/` file:
   ```bash
   peas create "Title" --type task --status draft --tag needs-review
   # Then edit .peas/ETH-XXXXX--*.md to add the body
   ```
   **Always pass `--status draft` explicitly.** The config default is not reliably applied.

   > **Note:** `--body-file` exists but only accepts paths relative to the project root, not absolute paths like `/tmp/...`. Prefer direct file edit instead.
3. Report to user: "Created ETH-XXXXX, ready for review"
4. **Stop.** Wait for approval.

### Approval Gate

- User reviews issue (reads `.peas/` file or uses `peas show`)
- If changes needed: user comments or asks agent to refine
- If approved: user removes `needs-review` tag and sets status to `todo`:
  ```bash
  peas update ETH-XXXXX --remove-tag needs-review --status todo
  ```

### Implementation

Starts only after issue is in `todo`:

1. Read issue: `peas show ETH-XXXXX` and read the `.peas/ETH-XXXXX--*.md` file for full body
2. Create a feature branch: `git checkout -b ETH-XXXXX/short-description`
3. `peas start ETH-XXXXX`, append [STARTED]
4. Work in meaningful commits (see [Git Workflow](#git-workflow) below)
5. When done: append `[DONE]` log entry with summary, `peas done ETH-XXXXX`, commit the issue file
6. Push branch and create MR via `/merge-request` skill
7. **Stop.** Report MR to user and wait for review.

The issue is marked `completed` and included in the MR. If the MR is rejected, the issue is reopened.

### Review Gate

After MR creation:

- User reviews MR (code, tests, CI)
- If changes needed: user comments, agent reopens issue (`peas start ETH-XXXXX`), revises on the same branch, marks done again, pushes
- If approved: user merges MR on GitLab (squash merge for clean history)

**Agents NEVER merge an MR without explicit user approval.** Marking the issue done is fine — the MR is the gate, not the issue status.

### Commit Ownership (Implementation Issues)

- Default: the agent that made file changes creates the commits and MR.
- Exception: if a lead/integrator is explicitly assigned for an issue, that lead/integrator pushes and creates the MR.
- If a commit cannot be created, append `[BLOCKED]` log entry with details and keep the issue `in-progress`.
- If there is no tracked implementation issue, this commit/Done gate does not apply.

## Research Issues

Use `type = "research"` for investigation/exploration where no code changes are expected. Research issues capture knowledge before deciding whether (or what) to implement.

### When to Use

- Exploratory research before knowing what to implement
- Investigating a problem/bug before proposing a solution
- Technology/library evaluation
- Understanding existing code patterns

### Research vs Implementation Issues

| Aspect       | Implementation Issue                | Research Issue                            |
| ------------ | ----------------------------------- | ----------------------------------------- |
| Type         | `task`, `feature`, `bug`, etc.      | `research`                                |
| Approval tag | `needs-review`                      | (none -- can start freely)                |
| Purpose      | Define what to build                | Answer questions, gather knowledge        |
| Output       | Code + commits                      | `[FINDINGS]` log entries                  |
| Approval     | Required before implementation      | Can start freely, Done needs confirmation |
| Template     | Acceptance Criteria, Verify command | Research Questions, Scope                 |
| Done when    | Tests pass, code committed          | Questions answered, human confirms        |

### Research Issue Template

```markdown
## Research Questions
- [Question 1 to answer]
- [Question 2 to answer]

## Scope
[What to investigate, boundaries]
- Files/areas to explore
- Out of scope

## Context
[Why this research is needed]

## Expected Output
[What findings should be captured]
```

### Research Flow

```
draft -> in-progress -> completed (requires human confirmation)

1. Create issue: peas create "Title" --type research
2. Start work: peas start ETH-XXXXX, append [STARTED]
3. Investigate, append [FINDINGS] log entries as discoveries are made
4. When questions answered, append final [FINDINGS] summary
5. Request human confirmation to mark completed
6. Human confirms -> peas done ETH-XXXXX
```

### Sub-tasks

Research issues support sub-tasks for parallel investigation:

```
Research Parent: [Overall research goal]
|-- Sub-task 1: Investigate X           [research, independent]
|-- Sub-task 2: Investigate Y           [research, independent]
+-- Sub-task 3: Synthesize findings     [research, blocked by 1, 2]
```

### Research -> Implementation

If research concludes that implementation is needed:

1. Research issue stays `completed` (knowledge preserved)
2. Create NEW implementation issue
3. Reference research issue ID in Context section
4. Add `needs-review` tag
5. Follow standard implementation flow

## Writing Good Issues

### Quality Checklist

- [ ] Task is specific and actionable
- [ ] Context includes ALL file paths, code references, constraints
- [ ] No assumed knowledge -- completely self-contained
- [ ] Verify section has exact commands
- [ ] Pass Criteria is unambiguous

### Template

```markdown
## Goal / Task
[What needs to be achieved - clear and actionable]

## Context
[ALL relevant information]
- File paths discovered
- Code patterns identified
- Constraints and dependencies
- Tech stack details

## Acceptance Criteria
- [Criterion 1]
- [Criterion 2]

## Verify
[Exact command to test]

## Pass Criteria
[Expected output or behavior that indicates success]
```

### Example

**Bad** (assumes context):
```markdown
## Task
Find where the config is loaded
```

**Good** (self-contained):
```markdown
## Task
Find where YAML configuration files are loaded in this project

## Context
- Project: /home/user/dev/project-name
- Looking for: Config/settings loading from .yaml or .yml files
- Tech stack: Python, uses pydantic

## Output Expected
- File paths where config loading occurs
- Function names responsible for loading
```

## Implementation Loop

### States

| Status        | Meaning                             |
| ------------- | ----------------------------------- |
| `draft`       | Not yet approved for implementation |
| `todo`        | Approved, ready for implementation  |
| `in-progress` | Being worked on (one per agent)     |
| `completed`   | Complete and verified               |
| `scrapped`    | Cancelled                           |

### Execution

```
WHILE parent not completed:
    1. SELECT next sub-task in todo (no unresolved blockers)
    2. START: create branch, peas start ETH-XXXXX, append [STARTED]
    3. EXECUTE: Do work, commit meaningful units, append [PROGRESS]/[DECISION] log entries
    4. VERIFY: Run verify command, append [VERIFY] log entry
    5. EVALUATE:
       - Pass -> Append [DONE], peas done, commit issue file, push, create MR, STOP
       - Fail -> Fix and retry, or append [BLOCKED]
    6. WAIT for user review
       - Changes requested -> peas start (reopen), revise, repeat from step 4
       - Approved -> user merges MR

WHEN all sub-tasks completed:
    - Verify parent Acceptance Criteria
    - Append [DONE] to parent, peas done ETH-XXXXX
```

### Handoff (Subagents)

**Commit ownership must be explicit.**

1. Main agent creates issue with full context and defines commit owner for each sub-issue
2. Subagent reads issue file, creates branch, runs `peas start ETH-XXXXX`, appends [STARTED]
3. Subagent works, commits meaningful units, appends [PROGRESS]/[FINDINGS]/[DECISION] log entries
4. Subagent appends [VERIFY] with results
5. If subagent is commit owner: subagent appends [DONE], `peas done`, commits issue file, pushes, creates MR, stops
6. If lead/main is commit owner: subagent appends handoff [PROGRESS] with files + verify result, then lead/main closes out and creates MR
7. User reviews MR. If changes needed: `peas start` (reopen), revise, repeat. If approved: user merges.

### Sub-Issue Completion

**Mark sub-issues `completed` before creating the MR.**

- Agent finishes work -> appends [DONE], `peas done`, commits issue file, pushes, creates MR, stops
- User reviews MR. If rejected: `peas start` (reopen), revise, repeat
- If commit is blocked, append [BLOCKED] and keep the issue `in-progress`

The issue status travels with the MR. The MR is the review gate, not the issue status.

## Activity Log

peas has no comment system. Instead, append structured log entries to the issue body, separated from the issue description by a `---` divider.

### Format

```markdown
+++
id = "ETH-00042"
title = "Fix parser bug"
type = "bug"
status = "in-progress"
priority = "high"
tags = []
parent = "ETH-00040"
created = "2026-02-26T14:00:00Z"
updated = "2026-02-26T14:30:00Z"
+++

## Goal
Fix the parser crash on malformed input.

## Acceptance Criteria
- Parser handles malformed input gracefully
- Test added for edge case

---

### [STARTED] 2026-02-26T14:30:00Z
Beginning work on parser fix.

### [PROGRESS] 2026-02-26T14:45:00Z
Found root cause in `src/parser/lexer.py:142`. Off-by-one in token boundary check.

### [VERIFY] 2026-02-26T14:55:00Z
`uv run pytest tests/test_parser.py` -- PASS (12 passed)

### [DONE] 2026-02-26T15:00:00Z
Files changed: `src/parser/lexer.py`, `tests/test_parser.py`.
```

### How to Append Log Entries

**Option A -- Direct file edit (preferred for agents with file tools):**

Read the `.peas/ETH-XXXXX--*.md` file, append the log entry to the body, save.

**Option B -- CLI with body-file:**

1. Read current body: `peas query '{ pea(id: "ETH-XXXXX") { body } }'`
2. Write updated body to a temp file with the new log entry appended
3. Update: `peas update ETH-XXXXX --body-file path/relative/to/project/root.md`

> **Note:** `--body-file` only accepts paths relative to the project root. Absolute paths (e.g. `/tmp/...`) are rejected.

**Option C -- Direct file write:**

Since peas files are plain markdown, agents can directly edit `.peas/ETH-XXXXX--*.md` to append log entries. The file is the source of truth.

### Log Entry Tags

| Tag        | When                | Content                                 |
| ---------- | ------------------- | --------------------------------------- |
| [STARTED]  | Beginning work      | Brief description                       |
| [PROGRESS] | Intermediate update | What's done, what's next                |
| [FINDINGS] | Research results    | Files found, patterns, questions        |
| [DECISION] | Choice made         | Decision + rationale                    |
| [VERIFY]   | Test results        | Command, output, PASS/FAIL              |
| [BLOCKED]  | Can't proceed       | Blocker, what was tried, what's needed  |
| [DONE]     | Task complete       | Files changed, summary of what was done |

## Knowledge Capture

Capture significant findings, decisions, and blockers in the issue's activity log as they happen.

```
DISCOVERED something? -> Append [FINDINGS] log entry
MADE a decision?     -> Append [DECISION] log entry with rationale
HIT a blocker?       -> Append [BLOCKED] log entry with details
FOUND a solution?    -> Append [PROGRESS] log entry
```

Capture: file paths, code patterns, design decisions, errors/solutions, dependencies, questions.

**Why:** If it's not in the issue, it doesn't exist. The next agent will waste time rediscovering it.

### Persistent Knowledge (Memory)

For knowledge that outlives a single issue, use the peas memory system:

```bash
peas memory save "architecture-decisions" --body "Key decisions about..."
peas memory query "architecture-decisions"
peas memory list
```

Memory entries are stored in `.peas/memory/` as markdown files, also git-trackable.

## Git Workflow

### Commit as You Go

Do NOT save all changes for one big commit at the end. Instead, commit after each meaningful unit of work:

```
git checkout -b ETH-00042/fix-parser-bug

# ... fix the lexer ...
git add src/parser/lexer.py
git commit -m "ETH-00042: Fix off-by-one in token boundary check"

# ... add tests ...
git add tests/test_parser.py
git commit -m "ETH-00042: Add test for malformed input edge case"

# ... update docs ...
git add docs/parser.md
git commit -m "ETH-00042: Document input validation behavior"
```

### What Makes a Good Commit

Each commit should be a **single logical change** that could be understood on its own:

- One bug fix = one commit
- One new function/class = one commit
- One test addition = one commit
- Formatting/linting fixes = one commit (separate from logic changes)

**Bad:** one commit with "fix parser, add tests, update docs, refactor utils"
**Good:** four commits, one for each of those changes

### Commit Message Format

```
ETH-XXXXX: Imperative description of what this commit does
```

Keep under 72 characters. The issue ID prefix is required.

### Squash Merge

MRs are merged with GitLab's **squash commits** option. This means:

- Individual commits on the feature branch are preserved in the MR for review
- On merge, they collapse into a single commit on the target branch
- The MR title becomes the squash commit message

This gives the best of both worlds: granular commits during development, clean history on `main`.

### Revisions After Review

If the user requests changes after MR creation:

1. Stay on the same branch
2. Make fixes as new commits (do NOT amend or force-push unless asked)
3. Push: `git push`
4. Notify user that revisions are ready

New commits appear in the existing MR automatically.

## Merge Requests

After committing, create a merge request using the `/merge-request` skill (see `.claude/skills/merge-request/SKILL.md`). The skill handles push and MR creation via `glab` against `gitlab.ethz.ch`.

## Dependencies

peas uses `blocking` (forward direction): "this ticket blocks those tickets."

To express "A is blocked by B": set `blocking = ["A"]` on ticket B.

```bash
# B blocks A (meaning A is blocked by B)
peas create "Task B" --blocking ETH-00042
```

The `peas suggest` command respects dependencies -- it won't suggest tickets that have unmet blockers.

```
|-- Sub-task 1: Explore patterns        [independent]
|-- Sub-task 2: Research libraries      [independent]
|-- Sub-task 3: Design schema           [blocked by 1, 2 -- set blocking on 1 and 2]
+-- Sub-task 4: Implement              [blocked by 3 -- set blocking on 3]
```

Independent tasks run in parallel. Dependent tasks wait.

### Querying What Blocks a Ticket

To find what blocks a specific ticket, search for it in blocking fields:

```bash
peas search "ETH-XXXXX"  # finds tickets that reference this ID in their blocking field
```

Or use GraphQL:

```bash
peas query '{ peas(filter: { isOpen: true }) { id title blocking } }'
```

## Issue Structure

```
Parent Issue: [High-level goal]
|-- 1. [Atomic unit]
|-- 2. [Atomic unit]
+-- 3. [Atomic unit] (blocked by 1, 2)
```

**Number sub-issue titles** with `1.`, `2.`, `3.` prefixes so execution order is visible at a glance.

Create sub-issues with `--parent`:

```bash
peas create "1. Explore patterns" --type task --parent ETH-XXXXX
peas create "2. Research libraries" --type task --parent ETH-XXXXX
peas create "3. Design schema" --type task --parent ETH-XXXXX
```

Aim for 2-7 sub-tasks per parent. Each must be independently verifiable.

Query sub-issues:

```bash
peas list --parent ETH-XXXXX
peas list --parent ETH-XXXXX --json
```

## Issue Types

| Type        | Purpose                    |
| ----------- | -------------------------- |
| `task`      | Generic tasks (default)    |
| `feature`   | Feature requests           |
| `bug`       | Bug reports                |
| `research`  | Investigation/exploration  |
| `chore`     | Maintenance tasks          |
| `story`     | User stories               |
| `epic`      | Groups of related features |
| `milestone` | Major releases/goals       |

Use `--template` for pre-populated body templates:

```bash
peas create "Fix parser crash" --template bug
peas create "Investigate caching options" --template research
peas create "Add export feature" --template feature
```

## Tags

Tags are free-form strings. The workflow uses one reserved tag:

| Tag            | Meaning                                          |
| -------------- | ------------------------------------------------ |
| `needs-review` | Issue awaiting human approval before work begins |

Add/remove tags:

```bash
peas update ETH-XXXXX --add-tag needs-review
peas update ETH-XXXXX --remove-tag needs-review
```

Filter by tag:

```bash
peas list --tag needs-review
```

## Smart Suggestions

`peas suggest` recommends the next ticket to work on, considering:
- Priority (critical > high > normal > low)
- Blocking count (tickets that unblock more work ranked higher)
- Type (bugs > features > stories > tasks)
- Dependency status (won't suggest blocked tickets)

```bash
peas suggest              # Single suggestion
peas suggest --limit 3    # Top 3 suggestions
peas suggest --json       # Machine-readable output
```

## Quick Reference

### Common Operations

```bash
# Create (always pass --status draft explicitly)
peas create "Title" --type task --status draft --tag needs-review
peas create "Title" --type research --status draft
peas create "Sub-task" --type task --status draft --parent ETH-XXXXX
# Then write the body by editing .peas/ETH-XXXXX--*.md directly

# Read
peas show ETH-XXXXX                                    # Summary view
peas query '{ pea(id: "ETH-XXXXX") { id title body status tags } }'  # Full body via GraphQL
peas list                                              # All open issues
peas list --status in-progress                         # Filter by status
peas list --parent ETH-XXXXX                           # Sub-issues
peas list --tag needs-review                           # Awaiting review
peas search "keyword"                                  # Text search

# Update
peas update ETH-XXXXX --status todo                    # Change status
peas update ETH-XXXXX --add-tag needs-review           # Add tag
peas update ETH-XXXXX --remove-tag needs-review        # Remove tag
peas update ETH-XXXXX --body-file relative/path.md     # Replace body (relative path only!)
peas start ETH-XXXXX                                   # Shortcut: set in-progress
peas done ETH-XXXXX                                    # Shortcut: set completed

# Bulk
peas bulk done ETH-00001 ETH-00002 ETH-00003           # Complete multiple
peas bulk tag "needs-review" ETH-00001 ETH-00002       # Tag multiple

# Context (for agent bootstrapping)
peas suggest --json                                    # Next ticket recommendation
peas context                                           # Project stats as JSON

# Memory (persistent knowledge)
peas memory save "key" --body "content"
peas memory query "key"
peas memory list
```

### Issue File Structure

Issues are stored as `.peas/ETH-XXXXX--slugified-title.md`:

```
.peas/
  config.toml                          # Project configuration
  ETH-00001--fix-parser-bug.md         # Issue file
  ETH-00002--add-export-feature.md     # Issue file
  archive/                             # Archived issues
  memory/                              # Knowledge base
```

Agents can read/edit these files directly -- they are the source of truth.

### Configuration

`.peas/config.toml`:

```toml
[peas]
prefix = "ETH-"
id_length = 5
id_mode = "sequential"
default_status = "draft"
default_type = "task"
frontmatter = "toml"
```

**Note:** `default_status = "draft"` is set in config but not reliably applied by peas v0.2.0. Always pass `--status draft` explicitly when creating issues.

### Concurrency

- Do not update the same issue file concurrently (peas detects concurrent modification and rejects)
- Appending log entries via file edit is safe if only one agent writes at a time
- Multiple agents can read issues concurrently without issue

### Health Check

```bash
peas doctor       # Validate config, tickets, references
peas doctor --fix # Auto-fix common issues
```
