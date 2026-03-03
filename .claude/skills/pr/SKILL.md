---
name: pr
description: "Push the current branch and create a GitHub PR. Gathers context, pushes, creates the PR, and reports the URL."
argument-hint: "[issue-id] [--target <branch>]"
user-invokable: true
---

# Create Pull Request

Parse `$ARGUMENTS` for:
1. **Issue ID** (optional): numeric ID. If not given, infer from branch name (`<id>/...`) or recent commit messages (`#<id>: ...`).
2. **`--target <branch>`** (optional): Target branch for the PR. Defaults to the repo default branch (usually `main`).

## Prerequisites

- `gh` CLI installed and authenticated
- Changes committed on a feature branch (not `main`)

## Process

### Step 1: Gather Context

1. Get the current branch: `git branch --show-current`
2. Confirm you are NOT on `main` — refuse to create a PR from `main`.
3. Get the issue ID from: arguments, branch name (`<id>/...`), or most recent commit (`#<id>: ...`).
4. Find the merge base: `git merge-base HEAD <target>` — handles stacked branches correctly.
5. Collect commits: `git log $(git merge-base HEAD <target>)..HEAD --oneline`
6. Get the diff summary: `git diff $(git merge-base HEAD <target>)..HEAD --stat`

### Step 2: Push

Push the branch to origin:

```bash
git push -u origin <branch-name>
```

### Step 3: Create PR

Read the issue's `content.md` for context. Create the PR:

```bash
gh pr create --title "#<id>: Short description" --body "$(cat <<'EOF'
## Summary
- [1-3 bullet points summarizing changes]

## Issue
#<id>

## Changes
[git diff --stat output]

Generated with Claude Code
EOF
)"
```

For a specific target branch:

```bash
gh pr create --title "#<id>: Short description" --base <target> --body "..."
```

### Step 4: Report

Tell the user:
- The PR URL (returned by `gh pr create`)
- Title and target branch
- Number of commits included

## PR Title

Format: `#<id>: Short description`

- Use the issue ID as prefix
- Keep under 72 characters
- Describe the change, not the issue title (unless they match)

## Notes

- If a PR already exists for this branch, `gh` will tell you — do not create a duplicate.
- Do not force-push to branches with open PRs without asking the user first.
- Agents never merge PRs without explicit human approval.
