@AGENTS.md

---

## Grapes

This project uses grapes for issue tracking. Run `/grapes` to learn the format and available skills.

## Parallel Work

Use the `Agent` tool with `isolation: "worktree"` for isolated parallel work on separate issues. 
Each worktree agent gets its own copy of the repo and can work independently.

## Agent Teams

Use the `Agent` tool with subagent types for coordinated work:

- **Explore** — Fast codebase exploration and search
- **general-purpose** — Multi-step tasks, research, implementation
- **Plan** — Architecture and implementation planning

Each agent should have a clear, non-overlapping scope.

## Temporary Files

Save temporary outputs to `.grapes/<id>/tmp/` when working on an issue, not `/tmp/`. Use `.grapes/tmp/` when no issue exists yet. Both paths are gitignored.