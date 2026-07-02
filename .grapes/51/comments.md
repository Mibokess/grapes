### 2026-07-02T18:00
[STARTED] `grapes` launches the TUI for any unrecognized arg (`--help`, `next`, ...), which fails in agent/non-tty environments. Adding self-contained `help`/`version` handling (before the `.grapes/` lookup) and a `default` case that rejects unknown commands with exit 2 instead of launching the TUI. Scope: `main.go` + README CLI table.

### 2026-07-02T18:05
[DONE] Implemented in `main.go`:
- Early switch (before `FindIssuesDir`) handles `help|--help|-h` → usage to stdout exit 0; `version|--version|-v` → version to stdout exit 0; and rejects any unknown command with `Unknown command: <arg>` + help to **stderr**, exit 2 — so it never hits the `.grapes/` creation prompt or launches the TUI.
- `issue`/`validate` fall through to dispatch after `issuesDir` is resolved (unchanged).
- Added `writeHelp(io.Writer)` with a self-contained usage block: commands, issue file layout, valid statuses/priorities, and an `id=$(grapes issue)` agent workflow example.
- README CLI table now lists `grapes help` and `grapes version`.

[VERIFY] `go build ./...` → 0. In an empty (no `.grapes/`) dir: `--help`/`-h`/`help` exit 0 (stdout); `version`/`-v`/`--version` print `0.1.8` exit 0; `next` prints error+help to stderr, exit 2, stdout empty, no prompt/hang. `validate 51` → "All issues valid." `issue 51` bump → exit 0.
