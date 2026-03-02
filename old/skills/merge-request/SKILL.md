---
name: merge-request
description: Create a GitLab merge request for the current branch. Pushes to origin, creates an MR via glab against gitlab.ethz.ch, and returns the MR URL.
argument-hint: [issue-id] [--target <branch>]
---

# Create Merge Request

Parse `$ARGUMENTS` for:
1. **Issue ID** (optional): e.g. `ETH-00042`. If not given, infer from the current branch name or recent commit messages.
2. **`--target <branch>`** (optional): Target branch for the MR. Defaults to the repo default branch (usually `main`).

## Prerequisites

- `glab` CLI installed and authenticated to `gitlab.ethz.ch`
- Remote: `origin` pointing to `gitlab.ethz.ch`
- Changes committed on a feature branch (not `main`)

## Process

### Step 1: Gather Context

1. Get the current branch name: `git branch --show-current`
2. Confirm you are NOT on `main` — refuse to create an MR from `main`
3. Get the issue ID from: `$ARGUMENTS`, branch name (`ETH-XXXXX/...`), or most recent commit message (`ETH-XXXXX: ...`)
4. Find the merge base: `git merge-base HEAD <target>` — this correctly handles stacked branches where the current branch was not created directly from `<target>`
5. Collect all commits on this branch: `git log $(git merge-base HEAD <target>)..HEAD --oneline`
6. Get the diff summary: `git diff $(git merge-base HEAD <target>)..HEAD --stat`

### Step 2: Push

Push the branch to origin if not already pushed:

```bash
git push -u origin <branch-name>
```

### Step 3: Create MR

`glab mr create` has no `--description-file` flag, and inline `$(...)` command substitution
triggers a sandbox permission prompt. Use the reusable `create-mr.sh` helper instead:

**3a.** Use the Write tool to create the description file at `/tmp/claude/ISSUE-ID/mr-description.md`:

```markdown
## Summary
- [1-3 bullet points summarizing changes]

## Issue
ETH-XXXXX

## Changes
[git diff --stat output or file list]

Generated with Claude Code
```

**3b.** Run the helper (no command substitution in the Bash tool call itself):

```bash
bash .claude/skills/merge-request/create-mr.sh \
  "ETH-XXXXX: Short description" \
  /tmp/claude/ISSUE-ID/mr-description.md
```

For a specific target branch, pass extra flags at the end:

```bash
bash .claude/skills/merge-request/create-mr.sh \
  "ETH-XXXXX: Short description" \
  /tmp/claude/ISSUE-ID/mr-description.md \
  --target-branch experiments
```

### Step 4: Report

Output to the user:
- The MR URL (returned by `glab mr create`)
- Title and target branch
- Number of commits included

## MR Title

Format: `ETH-XXXXX: Short description`

- Use the issue ID as prefix
- Keep under 72 characters
- Describe the change, not the issue title (unless they match)

## MR Description

```markdown
## Summary
- [Bullet points describing what changed and why]

## Issue
ETH-XXXXX

## Changes
[File list or diff stat]

Generated with Claude Code
```

Keep it concise. The issue has the full context — the MR description is a quick summary for reviewers.

## Common Operations

```bash
# Create MR (current branch -> default branch)
glab mr create --title "ETH-XXXXX: title" --description "..."

# Create MR with specific target
glab mr create --target-branch experiments --title "..." --description "..."

# List open MRs
glab mr list

# View MR details
glab mr view <MR-number>

# View MR with comments and CI status
glab mr view <MR-number> --comments
glab ci status

# Close/merge MR
glab mr close <MR-number>
glab mr merge <MR-number>
```

## Branch Naming

Use `ETH-XXXXX/short-description`:

```bash
git checkout -b ETH-00042/fix-parser-bug
```

## Notes

- `glab mr create` defaults to the current branch as source and the repo default branch as target
- The command returns the MR URL on success
- If an MR already exists for this branch, `glab` will tell you — do not create a duplicate
- Do not force-push to branches with open MRs without asking the user first
