# Quality Criteria

Every issue (parent or sub-issue) must have all of these:

| Section | Requirements |
|---------|-------------|
| **Goal** | What needs to be done (one sentence). Why it matters. Scoped to a single deliverable. |
| **Context** | All file paths (relative from repo root). All symbol names with brief descriptions. Current vs. desired behavior. Constraints and conventions. No assumed knowledge. |
| **Acceptance Criteria** | Each criterion is binary pass/fail. Covers all expected changes. Edge cases noted. |
| **Verify** | Exact shell commands, copy-pasteable without modification. |
| **Pass Criteria** | Exact expected output or behavior. Two readers would agree on pass/fail. |
| **Labels** | `implementation` + `needs-review`. Priority set. Parent linked if sub-issue. |

## Common Problems

| Problem | Fix |
|---------|-----|
| Vague goal ("refactor the pipeline") | Name the exact functions and what changes |
| Missing file paths ("update the config loader") | Find and list the actual path |
| Implicit context ("same pattern as temperature") | Spell out the pattern with code references |
| Untestable criteria ("code should be clean") | Replace with binary pass/fail checks |
| No verify command | Add exact `uv run pytest ...` or equivalent |
| Scope creep (multiple unrelated changes) | Split into sub-issues |
| Sub-issue not standalone ("see parent") | Copy relevant context into the sub-issue |
| Parent duplicates sub-issue details | Remove specifics, keep only the overview |

## Verification Checklist

1. **Every file path** exists and points to the right file
2. **Every symbol name** is spelled correctly and matches current code
3. **Every described behavior** matches what the code actually does
4. **Every claimed relationship** (e.g., "called from X", "depends on Y") is true
5. **No unverified assumptions** remain ("probably", "should be", "similar to")
6. **No open questions** remain ("TBD", "TODO", "need to check", "unclear")
7. **No vague language** hides uncertainty ("as appropriate", "if needed", "etc.")
