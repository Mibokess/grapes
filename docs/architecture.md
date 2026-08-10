# Architecture

This document describes stable responsibilities and data flow. It is a navigation
aid, not a substitute for checking the implementation named in each section.

## Process Entry Points

`main.go` is the only executable entry point.

1. Help, version, and unknown commands are handled before filesystem discovery.
2. `data.FindIssuesDir` searches from the current directory into descendants, up to
   ten levels deep. It does not walk toward parent directories.
3. If no `.grapes` directory exists, interactive startup offers to create one.
4. `grapes issue` delegates ID allocation and timestamps to `internal/data`.
5. `grapes validate` delegates schema and relationship checks to
   `internal/data/validate.go`.
6. Bare `grapes` loads config, loads all issue sources, constructs `tui.Model`, and
   runs the Bubble Tea program.

## On-Disk Model

`internal/data/issue.go` defines the public in-memory model. `meta`, the private TOML
shape in `internal/data/loader.go`, defines the persisted metadata.

Persisted fields are title, status, priority, labels, parent, blocked-by IDs, and
timestamps. Description and comments live in separate Markdown files. `Children`
and `Blocks` are reverse relationships rebuilt in memory by `RewireRelationships`.

Valid statuses, in board order, are `backlog`, `todo`, `in_progress`, `done`, and
`cancelled`. Valid priorities are `urgent`, `high`, `medium`, and `low`.

## Loading and Source Merging

The primary load path is `data.LoadAllSources`:

1. `LoadAllIssues` reads numeric directories from the selected main `.grapes`.
2. `FindGitWorktreeGrapesDirs` calls `git worktree list --porcelain` and finds a
   `.grapes` directory in each worktree.
3. `FindWorktreeIssuesDirs` expands optional `[sources].dirs` globs relative to the
   project root, unless a pattern is absolute.
4. Copies with the same numeric ID become `IssueSource` entries on one `Issue`.
5. Sources are ordered with main first and named sources alphabetically.
6. The copy with the newest filesystem modification time becomes active. Calling
   `Issue.SwitchSource` copies that source's fields onto the top-level issue.
7. Relationships are rebuilt from the active versions, and results are sorted by ID.

Loading deliberately tolerates some source failures: malformed issue directories
and unreadable secondary sources can be skipped. Validation is the strict path.

## ID Allocation

`data.NextID` finds the main repository through Git's common directory, locks
`.grapes/.lock`, scans the main source plus all known worktree/configured sources,
and creates `max ID + 1` in the local source. Platform-specific locking lives in
`flock_unix.go` and `flock_windows.go`.

Keep the scan and directory creation inside the lock. Moving either operation out
of the critical section reintroduces collisions between parallel agents.

## Persistence

All production write helpers live in `internal/data/writer.go`:

- `UpdateField` changes status or priority and stamps timestamps.
- `UpdateLabels` rewrites parsed metadata and updates `updated`.
- `StampTimestamps` is the canonical timestamp operation used by the CLI.
- `AppendComment` appends a timestamped Markdown section.
- `SerializeIssue` creates the combined document passed to an external editor.
- `SaveIssueFromText` validates the edited frontmatter, then separates it back into
  metadata, description, and comments.

The root TUI resolves an issue's active source before calling these helpers. Preserve
that behavior: edits made while viewing a worktree version belong to that worktree.

## TUI Ownership and Message Flow

`internal/tui/app.go` contains the root `tui.Model`. It owns:

- the complete issue slice and active screen;
- board, list, detail, and settings child models;
- navigation history, global sorting, and filters;
- picker and filter overlays;
- configuration and theme state;
- filesystem watching, external-editor sessions, writes, and status messages.

Screens live in separate packages and implement local navigation and rendering.
They communicate upward with the types in `internal/tui/common/messages.go`.

```text
keyboard/mouse/fs event
  -> active child model
  -> common message (open, move, edit, filter, switch, save)
  -> root tui.Model.Update
  -> data write or app-state transition
  -> common.RefreshMsg / filesystem event
  -> LoadAllSources
  -> refreshed child models
```

Cross-view behavior belongs in the root model. Screen-specific selection, layout,
and rendering belong in the screen package. Shared message types, key maps, and theme
primitives belong in `internal/tui/common`.

The filesystem watcher watches each source directory and its numeric issue
directories. A failed main watch is surfaced in the status bar rather than silently
disabling live reload.

## Configuration

`internal/config/config.go` defines the TOML schema and defaults. Configuration is
loaded from `.grapes/config.toml` and covers:

- startup screen, sort mode, child auto-close, and empty board columns;
- extra source directory globs;
- dark/light theme mode, preset, and color overrides;
- global and screen-specific keybindings.

A missing file returns defaults. A malformed file returns clean defaults plus an
error; the TUI displays that error in its status bar. Settings writes use
`config.Save`, then send `common.ConfigSavedMsg` so the root can apply the result.

## Package Dependency Direction

```text
main
  -> config, data, tui
tui/app
  -> config, data, tui child packages, tui/common
tui child packages
  -> data, tui/common
data and config
  -> standard library and format/system dependencies
```

Avoid importing the root `tui` package from child views. Use a type or message in
`tui/common` to keep coordination flat and prevent import cycles.
