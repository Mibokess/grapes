@AGENTS.md

---

## Grapes

The `grapes` skill (`.agents/skills/grapes/`) is the reference for the file-based issue tracker. It covers format, rules, and issue creation.

Workflow skills:

- `/issue <mode> [issue-id]` — Work on an issue (plan, verify, research, continue)
- `/handoff <mode> [issue-id] [context...]` — Write a handoff for another agent
- `/pr [issue-id]` — Push and create a PR

## Handoffs

Use `EnterPlanMode` to write handoff plans. The plan document is the handoff — it contains everything the receiving agent needs.

## Parallel Work

Use the `Agent` tool with `isolation: "worktree"` for isolated parallel work on separate issues. Each worktree agent gets its own copy of the repo and can work independently.

## Agent Teams

Use the `Agent` tool with subagent types for coordinated work:

- **Explore** — Fast codebase exploration and search
- **general-purpose** — Multi-step tasks, research, implementation
- **Plan** — Architecture and implementation planning

Each agent should have a clear, non-overlapping scope.

## Temporary Files

Save temporary outputs to `.grapes/<id>/tmp/` when working on an issue, not `/tmp/`. Use `.grapes/tmp/` when no issue exists yet. Both paths are gitignored.


