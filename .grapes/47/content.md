## Goal

Fix the detail view so that the Children and Blocks lists are computed from the viewing source's perspective, not from globally active sources.

## Context

Issue #46 fixed two things:
1. `RewireRelationships` is now called after `SwitchSourceMsg` to recompute `Children`/`Blocks`
2. `findRelatedIssue` shows title/status/labels from the viewing source

But `RewireRelationships` computes `Children`/`Blocks` from each issue's *currently active* `Parent`/`BlockedBy` fields — a global property. When viewing issue 10 from worktree "feature", the children shown depend on what source is active for ALL other issues, not what those issues say in the "feature" source.

### Example

- Issue 10 (parent) — viewing from worktree "feature"
- Issue 20: main has `parent=10`, "feature" has `parent=nil`
- Issue 30: main has `parent=nil`, "feature" has `parent=10`

If issue 20's active source is main and issue 30's active source is main:
- `RewireRelationships` sets issue 10's `Children = [20]`
- But from the "feature" perspective, children should be `[30]`

### Fix

In `renderIssue` (`internal/tui/detail/detail.go`), compute children and blocks dynamically by inspecting each issue's source matching the viewing worktree, instead of using the pre-computed `issue.Children` and `issue.Blocks`.

- `childrenForSource(allIssues, parentID, worktree)` — find issues whose Parent in the matching source equals parentID
- `blocksForSource(allIssues, blockerID, worktree)` — find issues whose BlockedBy in the matching source contains blockerID
- Fall back to the issue's active source when the viewing source doesn't exist for that issue (consistent with `findRelatedIssue`)

Files:
- `internal/tui/detail/detail.go` — `renderIssue` lines 394-408 (Children), lines 331-339 (Blocks)
- `internal/tui/detail/detail_test.go` — add test for correct child set per source

## Acceptance Criteria

- [x] Detail view Children section shows only issues whose Parent (in the viewing source, falling back to active) matches the current issue's ID
- [x] Detail view Blocks section shows only issues whose BlockedBy (in the viewing source, falling back to active) contains the current issue's ID
- [x] Existing tests pass (`go test ./...`)

## Verify

```bash
go test ./...
```

## Pass Criteria

All tests pass. `TestDetailView_MultiSource_ChildrenComputedPerSource` verifies that viewing from a worktree shows the correct set of children based on that worktree's parent relationships.
