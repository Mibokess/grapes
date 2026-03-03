# Handoff Mode: Implement

Hand off implementation. The receiving agent will write code. Every decision must already be made — no choices left for the implementer.

## Guidance

- Include the full picture: goal, all files, all symbols, all constraints.
- Provide concrete implementation steps in execution order.
- Include every decision already made (with rationale).
- Show patterns and conventions to follow (with examples from the codebase).
- Specify test strategy and verify commands.
- Define exact expected outcomes.

## Template

```markdown
## Issue
#<id>: [title]

## Goal
[One-sentence summary of what to build/change and why]

## Key Files
- `path/to/file` — role, relevant symbols, current state
- `path/to/other` — role, relevant symbols, current state

## Implementation Steps

### Step 1: [Short title]
- **File**: `path/to/file`
- **What**: Exact description of the change
- **Symbols**: `ClassName.method_name` (line ~N) — current signature, what to change
- **Details**: Types, patterns to follow, code examples
- **Watch out**: Gotchas or constraints

### Step 2: [Short title]
...

## Decisions Made
- [Decision]: [rationale]

## Patterns & Conventions
- [Rules from CLAUDE.md or project conventions]
- [Code style patterns, with examples from existing code]

## Gotchas
- [Non-obvious things discovered during research]

## Test Strategy
- **Existing tests**: `path/to/test` — what they cover
- **New tests needed**: What to add and where
- **Run**: [exact test command]

## Verify
[Exact command(s)]
**Pass criteria**: [Exact expected outcome]
```

## Checklist

1. Every file path exists and is correct.
2. Every symbol matches current code (name, signature, location).
3. Every step is actionable without additional research.
4. All decisions made — no choices left for the implementer.
5. No vague language ("as needed", "appropriately", "etc.").
6. Step ordering respects dependencies.
7. Patterns section has concrete examples, not abstract rules.
8. Test strategy covers all acceptance criteria.
9. Verify command is copy-pasteable with unambiguous pass criteria.
10. A fresh agent can implement this cold.
