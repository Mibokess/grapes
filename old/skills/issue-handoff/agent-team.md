# Agent Team Design Guide

When `--team` is specified, design a team of agents for the receiving agent to spawn. The team design becomes a dedicated section in the handoff plan.

## Thinking Process

Before designing the team, think through:

1. **What are the independent work streams?** Look for tasks that can run in parallel without stepping on each other (different files, different modules, research vs. implementation).
2. **What roles are needed?** Common roles: implementer, researcher, tester, reviewer. Only create roles that have distinct, meaningful work.
3. **What are the dependencies?** Which tasks must finish before others can start? This determines sequencing and team coordination.
4. **What's the minimum team size?** More agents = more coordination overhead. Don't create agents for trivial tasks that the lead can do directly.
5. **What keeps context clean?** Teams let each agent focus on a narrow scope without polluting their context with unrelated work. The lead stays focused on coordination; implementers stay focused on their files.

## Design Rules

- **Each agent must have a clear, non-overlapping scope.** If two agents would touch the same files, merge them into one.
- **Every agent gets a self-contained brief.** The lead will hand each agent its instructions — include everything the agent needs in the plan.
- **Name agents by role, not by number.** `researcher`, `implementer`, `tester` — not `agent-1`, `agent-2`.
- **Specify agent types.** Match agent type to the work: `general-purpose` for implementation, `Explore` for read-only research, `Plan` for architecture.
- **Define the lead's role.** The lead coordinates, assigns tasks, integrates results, and handles Linear updates. Spell out what the lead does vs. delegates.
- **Keep it small.** 2-4 agents covers most cases. If you need more, the issue might need splitting first.

## Team Section Template

Include this section in the handoff plan:

```markdown
## Agent Team

### Team Structure

| Agent | Type | Role | Scope |
|-------|------|------|-------|
| lead | general-purpose | Coordinate, integrate, commit | [what the lead owns] |
| [name] | [type] | [role] | [specific files/tasks] |
| [name] | [type] | [role] | [specific files/tasks] |

### Lead Responsibilities
- Spawn teammates and assign tasks
- [Specific coordination duties]
- Integrate results and handle conflicts
- Commit changes and update Linear

### Agent Briefs

#### [agent-name]
- **Task**: [Exact description of what to do]
- **Files**: [Files to read/modify]
- **Output**: [What to produce — code changes, findings, etc.]
- **Done when**: [Clear completion criteria]
- **Dependencies**: [What must finish before this agent starts, if any]

#### [agent-name]
...

### Execution Order

1. [What runs first — parallel or sequential]
2. [What runs next, and what it depends on]
3. [Integration/final steps by the lead]

### Coordination Notes
- [How agents should communicate — e.g. via task list, messages]
- [Shared constraints — e.g. don't both modify the same file]
- [How the lead should handle merge/integration]
```

## Common Team Patterns

### Research + Implement
```
lead (general-purpose) — coordinates, commits
├── researcher (Explore) — investigates questions, reports findings
└── implementer (general-purpose) — writes code based on findings
```
Use when: implementation depends on answers the agent doesn't have yet.

### Parallel Implementation
```
lead (general-purpose) — coordinates, integrates, commits
├── implementer-a (general-purpose) — implements module A
└── implementer-b (general-purpose) — implements module B
```
Use when: work touches independent modules that don't share files.

### Implement + Verify
```
lead (general-purpose) — coordinates, commits
├── implementer (general-purpose) — writes the code
└── tester (general-purpose) — writes tests, runs verification
```
Use when: substantial test work that can be written in parallel with implementation.

### Research Fan-Out
```
lead (general-purpose) — coordinates, synthesizes findings
├── researcher-a (Explore) — investigates area A
└── researcher-b (Explore) — investigates area B
```
Use when: research has independent areas that benefit from parallel exploration.

## Anti-Patterns

| Don't | Why | Instead |
|-------|-----|---------|
| One agent per file | Too much coordination overhead | Group related files into one agent |
| Agent just for "review" | Adds coordination cost without freeing lead context | Lead reviews directly — it's a quick read, not deep work |
| Agent for a 5-minute task | Not worth the spawn cost | Lead does it |
| Agents sharing files | Merge conflicts, wasted work | One agent per file group |
| More than 4 agents | Coordination dominates | Split the issue first |
