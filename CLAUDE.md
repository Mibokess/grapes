@AGENTS.md

---

## Grapes Skills

Available `/grapes-*` skills for issue tracking:

- `/grapes-create` — Create a new issue or sub-issue
- `/grapes-read` — Read issue metadata, content, or comments
- `/grapes-update` — Update issue metadata (status, priority, labels, title, parent, blocked_by)
- `/grapes-list` — List, filter, or browse issues
- `/grapes-search` — Search across issues by keyword or content
- `/grapes-comment` — Add a comment to an issue
- `/grapes-close` — Complete or cancel an issue

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

## PR Creation

Use `gh pr create` for GitHub PRs. The `/pr` skill handles the full workflow.
