# Issue Mode: Continue

Pick up partial work on an issue. Another agent (or you in a previous session) started but didn't finish.

## Process

1. Read the issue's `content.md` and `comments.md` to understand the full history.
2. Identify the last activity log entry — understand where work stopped and why.
3. Check the git state: current branch, uncommitted changes, recent commits.
4. Resume from where work was left off. Do not redo completed work.
5. Add a `[STARTED]` comment noting you are continuing and from what state.
6. Continue the work, adding `[PROGRESS]` comments as you go.

## Report

Follow the normal completion flow for the work being done (implementation -> PR, research -> findings summary).
