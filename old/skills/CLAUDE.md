@AGENTS.md
@docs/peas/peas-usage.md

---

## Issue Tracking — No Exceptions

**BEFORE editing any file, you MUST have a peas issue.**

1. Check peas for an existing issue covering this work: `peas list` or `peas search "..."`
2. If none exists, create one (see AGENTS_peas.md for workflow)
3. If the issue is in **todo** or **in-progress** — proceed with code changes
4. If you just created the issue or it is in **draft** — **STOP.** Report the issue to the user ("Created ETH-XXXXX, ready for review") and **wait for approval.** Do NOT start implementing.

If you are about to edit a file and there is no peas issue — STOP. Create the issue first. No rationalizing, no "I'll create it after." Issue first, code second.

## No Assumptions

**NEVER work on assumptions. Always verify.**
Run code, check sources, ask questions.

## Git — Commit as You Go, Never Self-Close

**Commit meaningful units of work throughout implementation, not one big commit at the end.**

1. Create a feature branch: `git checkout -b ETH-XXXXX/short-description`
2. Commit after each logical unit: `git commit -m "ETH-XXXXX: what this unit does"`
3. When done: append `[DONE]` to peas issue, `peas done ETH-XXXXX`, commit the issue file
4. Push branch and create MR via `/merge-request` skill
5. **STOP.** Report the MR to the user and wait for review.

**Agents NEVER merge an MR without explicit user approval.**

The issue is marked done and included in the MR. If the MR is rejected, the issue is reopened (`peas start ETH-XXXXX`), revisions are made, and the cycle repeats. The MR is the review gate, not the issue status.

Commit ownership:
- Default: the agent that made the file changes commits and creates the MR.
- If a lead/integrator is explicitly assigned, that lead/integrator pushes and creates the MR.
- If a commit cannot be created, append `[BLOCKED]` to the peas issue and do not mark Done.

## Teams

Use teams for complex work needing coordination between agents. Use a solo subagent for isolated, self-contained tasks.

**Team tasks ≠ peas sub-issues.** Peas sub-issues are the logical decomposition for humans. Team tasks are the operational decomposition—how agents actually divide labor. These don't need to map 1:1. Agents may slice work by role (one writes Python, one verifies, one researches) even if that cuts across multiple peas sub-issues.

Rules:
- **Lead reconciles with peas.** The lead maps team progress onto peas sub-issues—regardless of which team tasks contributed.
- **Commit owner must be explicit for tracked implementation issues.** If a lead/integrator is assigned, they commit. Otherwise, each implementing agent commits their own changes.
- **Durable knowledge goes to peas.** `[FINDINGS]`, `[DECISION]`, and other tagged log entries go to the peas issue body. Team messages (`SendMessage`) are for ephemeral coordination only.
- **Implementation requires approval.** Teams can be used for research or issue creation freely, but implementation work still requires an approved issue first.

---

## Running Scripts

**NEVER use inline heredocs.** They fail silently in the sandbox.

```bash
# WRONG - will fail:
python << 'EOF'
print("hello")
EOF

# WRONG - will fail:
python -c "print('hello')"

# CORRECT:
# When working on peas issue ETH-00042:
# 1. Write script to /tmp/claude/ETH-00042/
# 2. Run with: uv run /tmp/claude/ETH-00042/script.py

# When not working on any issue:
# 1. Write script to /tmp/claude/
# 2. Run with: uv run /tmp/claude/script.py
```

Always write Python to a file first, then execute it. Use issue-specific directories (`/tmp/claude/ISSUE-ID/`) to organize temporary files when working on peas issues.

## Reading Files in src/

When reading any file in `src/`, also read all CLAUDE.md files from that file's directory up to `src/npnf/`.

Example: Reading `src/npnf/data/synthetic/temperature.py` → also read:
- `src/npnf/data/synthetic/CLAUDE.md`
- `src/npnf/data/CLAUDE.md`
- `src/npnf/CLAUDE.md`
