## Goal
Add `grapes issue` subcommand that replaces `next-id` and manual timestamp maintenance.

- `grapes issue` (no args) → allocate next ID, create dir, set `created` + `updated`, print ID
- `grapes issue <id>` → create dir if needed, set `created` if missing, bump `updated`

Remove the `next-id` subcommand.

## Context

### Current state
- `main.go:45-46` — `next-id` subcommand calls `data.NextID`
- `internal/data/loader.go:209-243` — `NextID` allocates IDs with file locking
- `internal/data/writer.go:20-36` — `UpdateField` runs sed to bump `updated`
- `internal/data/writer.go:39-60` — `UpdateLabels` sets `m.Updated`
- `internal/data/writer.go:174-265` — `SaveIssueFromText` preserves `created`, sets `updated`
- `plugin/skills/grapes/SKILL.md` — skill instructions tell agents to manually maintain timestamps

### Design
`grapes issue [id]`:
- No args: run `NextID` to allocate, then set both `created` and `updated` to now, print ID
- With ID: if dir doesn't exist create it. Read meta.toml — if `created` is zero/missing set it to now. Always set `updated` to now. Write back.
- Timestamps: UTC, truncated to minute

Agent workflow becomes:
1. `id=$(grapes issue)` — get new issue
2. Write meta.toml fields, content.md, etc.
3. `grapes issue $id` — stamp timestamps

### Files to change
- `main.go` — replace `next-id` with `issue` subcommand
- `internal/data/writer.go` — add `StampTimestamps(issuesDir string, issueID int) error`; remove `updated` bumping from `UpdateField`, `UpdateLabels`
- `plugin/skills/grapes/SKILL.md` — replace `next-id` and manual timestamp rules with `grapes issue`

## Acceptance Criteria
- [ ] `grapes issue` allocates next ID, creates dir, sets `created` + `updated`, prints ID
- [ ] `grapes issue 35` creates dir if needed, sets `created` if missing, bumps `updated`
- [ ] `grapes issue 35` preserves existing `created`
- [ ] `next-id` subcommand removed
- [ ] `UpdateField` no longer bumps `updated`
- [ ] `UpdateLabels` no longer sets `m.Updated`
- [ ] Skill instructions updated
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes

## Verify
```bash
cd /projects/mboss/dev/grapes && go build ./... && go test ./...
```

## Pass Criteria
- Build and tests pass.
- `grapes issue` prints a numeric ID and creates `.grapes/<id>/`.
- `grapes issue <id>` on existing issue updates `updated` without changing `created`.
