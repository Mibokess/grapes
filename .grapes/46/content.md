## Goal

Fix the detail view so that cross-issue relationships (Children, Blocks, Parent, BlockedBy) are consistent with the active source when viewing a multi-source issue.

Currently, switching sources on an issue updates its own fields but related issues display data from their own independently-chosen active source, and computed relationship lists (`Children`, `Blocks`) are never recomputed.

## Context

### How it works today

- `LoadAllSources` (`internal/data/loader.go:332-410`) loads issues from main + all worktrees, picks the most recently modified source as active, then calls `RewireRelationships` once.
- `SwitchSource` (`internal/data/issue.go:216-234`) copies per-source fields (`Title`, `Status`, `Parent`, `BlockedBy`, etc.) but does **not** update `Children` or `Blocks`.
- `RewireRelationships` (`internal/data/loader.go:415-438`) computes `Children` and `Blocks` from the current active `Parent`/`BlockedBy` of all issues. It runs once at load time and is never called again after a source switch.
- `SwitchSourceMsg` handler (`internal/tui/app.go:426-436`) calls `SwitchSource` on the issue and recreates the detail view, but does **not** call `RewireRelationships`.
- `renderIssue` (`internal/tui/detail/detail.go:296-428`) looks up related issues from `allIssues` by ID, using each related issue's own active source for title/status/labels.

### Bug 1: `Children` and `Blocks` are stale after source switch

`Children` and `Blocks` are top-level fields on `Issue` (not part of `IssueSource`). They are computed once by `RewireRelationships` at load time. When you switch an issue's source, its `Parent` and `BlockedBy` IDs change (correctly), but the inverse relationships (`Children`, `Blocks`) on other issues are never recomputed.

**Example:** Issue #5 has `parent_id = 10` in main but no parent in the worktree. At load time, main is the active source, so `RewireRelationships` adds #5 to #10's `Children`. If you then switch #5 to its worktree source (no parent), #10 still lists #5 as a child.

### Bug 2: Related issue data comes from the wrong source

When rendering relationships in the detail view, related issues are looked up from `allIssues`. Each related issue displays its own active source's title/status — not the source matching the viewing issue.

**Example:** You're viewing issue #5 from main. Its parent is #3. The detail view shows #3's title, but that title comes from #3's own active source (maybe a worktree where #3 was edited more recently), not from main.

This affects all four relationship types: Parent (line 298-304), BlockedBy (line 316-321), Blocks (line 333-338), Children (line 404-409).

## Acceptance Criteria

- [x] `SwitchSourceMsg` handler calls `RewireRelationships` (or equivalent) after switching a source so that `Children` and `Blocks` reflect the new active `Parent`/`BlockedBy` values
- [x] Related issues rendered in the detail view (Parent, BlockedBy, Blocks, Children) display title/status/labels from the **same source directory** as the viewing issue's active source, falling back to the related issue's own active source if it doesn't exist in that source
- [x] Switching sources on an issue updates the sub-issues section to reflect the new relationship state
- [x] Existing tests pass (`go test ./...`)
- [x] No regressions in single-source behavior (issues with only one source render identically to before)

## Verify

```bash
go test ./...
```

## Pass Criteria

All tests pass. Manual verification: open an issue that exists in both main and a worktree with different parent/child relationships, switch sources, and confirm the sub-issues section updates accordingly.
