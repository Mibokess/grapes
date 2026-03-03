---
name: issue
description: "Work on a grapes issue in a specific mode: plan (make comprehensive), verify (check against code), research (investigate), or continue (resume partial work)."
argument-hint: <plan|verify|research|continue> [issue-id]
user-invokable: true
---

# Work on Issue

Parse `$ARGUMENTS` for:
1. **Mode** (first word): `plan`, `verify`, `research`, or `continue`. If missing, ask the user.
2. **Issue ID** (optional): numeric ID. If not given, infer from current branch name or ask.

Read the issue using the [grapes-read skill](../../plugin/skills/grapes-read/SKILL.md) before starting.

Then read and follow the mode-specific instructions:

| Mode | File | Purpose |
|------|------|---------|
| `plan` | [modes/plan.md](modes/plan.md) | Make issue comprehensive and implementation-ready |
| `verify` | [modes/verify.md](modes/verify.md) | Independently audit issue against codebase |
| `research` | [modes/research.md](modes/research.md) | Investigate a question, capture findings |
| `continue` | [modes/continue.md](modes/continue.md) | Resume partial work from where it stopped |

Read **only** the mode file you need. Do not read the others.
