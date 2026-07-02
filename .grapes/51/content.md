## Goal

Make the `grapes` CLI discoverable and safe for non-interactive/agent use: add a self-contained `grapes --help`/`grapes version`, and make any unrecognized argument (e.g. `grapes next`, `grapes --help` today) print an error/help to stderr and exit non-zero instead of silently launching the full-screen TUI (which fails in agent environments).

## Context

Entry point is `main.go` (package `main`, binary `grapes`). Current dispatch in `main()` (`main.go:16-64`):

1. `os.Getwd()` then `data.FindIssuesDir(cwd)` — if no `.grapes/` is found it **interactively prompts** to create one (`main.go:23-38`). This runs *before* any subcommand dispatch.
2. Subcommand switch (`main.go:41-48`) handles only `validate` and `issue`.
3. Anything else — including `--help`, `-h`, `version`, and typos like `grapes next` — falls through to launching the Bubble Tea TUI (`main.go:50-63`).

### Current vs. desired behavior

**Current:**
- `grapes --help` / `grapes next` / any unknown arg → tries to launch the TUI. In an agent/non-tty environment this errors out (no interactive terminal), giving no hint about valid usage.
- `grapes --help` also first hits the `.grapes/` creation prompt (`main.go:23-38`) before falling through.
- There is no way to print the version (`version` var at `main.go:14`) or usage help.

**Desired:**
- `grapes help`, `grapes --help`, `grapes -h` → print a self-contained usage/help text to **stdout**, exit 0. Works with **no `.grapes/` directory present** (must run before `FindIssuesDir`, so it never triggers the creation prompt).
- `grapes version`, `grapes --version`, `grapes -v` → print `version` (`main.go:14`) to stdout, exit 0. Also works with no `.grapes/`.
- Any other unrecognized argument → print `Unknown command: <arg>` plus the help text to **stderr**, exit code 2. Do **not** launch the TUI.
- Bare `grapes` (no args) → unchanged: launch the TUI.
- `grapes issue [...]` and `grapes validate [...]` → unchanged.

### Existing subcommand surface (must be reflected in help text)

- `grapes` — launch TUI (interactive terminal only).
- `grapes issue` — allocate next ID, create `.grapes/<id>/`, print ID (`runIssue`, `main.go:66`).
- `grapes issue <id>` — bump `updated` timestamp (`main.go:83-101`).
- `grapes validate [<id>...]` — validate all issues, or specific IDs (`runValidate`, `main.go:104`).

Issue file format (from `plugin/skills/grapes/SKILL.md`) — briefly summarize in help: `.grapes/<id>/` holds `meta.toml`, `content.md`, `comments.md`; statuses `backlog|todo|in_progress|done|cancelled`; priorities `urgent|high|medium|low`.

### Constraints

- Keep changes surgical and confined to `main.go` (plus README/help text). No new packages or flag libraries — hand-rolled `switch` on `os.Args[1]` matches the existing style.
- Help/version must not depend on `FindIssuesDir` or the cwd having a `.grapes/`.
- Do not change the behavior of bare `grapes`, `grapes issue`, or `grapes validate`.
- Update the README CLI table (`README.md:52-58`) to list `help`/`version`.

## Approach

1. In `main()`, **before** `os.Getwd()`/`FindIssuesDir`, add a switch on `os.Args[1]` (guarded by `len(os.Args) > 1`) handling `help|--help|-h` → `writeHelp(os.Stdout)`, exit 0; and `version|--version|-v` → print `version`, exit 0.
2. Extend the existing subcommand switch (`main.go:41-48`) with a `default:` case for a non-empty first arg: write `Unknown command: <arg>\n\n` + `writeHelp` to `os.Stderr`, `os.Exit(2)`.
3. Add `func writeHelp(w io.Writer)` printing the usage block (commands, issue-format summary, a short agent workflow example, and a note that bare `grapes` needs an interactive terminal).
4. Update `README.md` CLI table with `grapes help` and `grapes version` rows.

## Acceptance Criteria

- [x] `grapes --help`, `grapes -h`, and `grapes help` all print usage to stdout and exit 0, listing `issue`, `issue <id>`, `validate`, `help`, `version` and the bare-TUI behavior.
- [x] `grapes --version`, `grapes -v`, and `grapes version` print the `version` string to stdout and exit 0.
- [x] Help and version work when run in a directory with **no** `.grapes/` dir and produce **no** interactive "Create one?" prompt.
- [x] An unknown argument (e.g. `grapes next`) prints `Unknown command: next` + help to stderr and exits with code 2 — it does **not** attempt to launch the TUI.
- [x] Bare `grapes` still launches the TUI; `grapes issue`, `grapes issue <id>`, and `grapes validate` behave exactly as before.
- [x] Help text summarizes the issue file layout (`meta.toml`/`content.md`/`comments.md`), valid statuses and priorities, and shows a `id=$(grapes issue)` workflow example.
- [x] `README.md` CLI table includes `grapes help` and `grapes version`.

## Verify

```bash
cd /home/mboss/dev/grapes
go build -o /tmp/grapes-test .
cd $(mktemp -d)   # a dir with no .grapes/
/tmp/grapes-test --help;      echo "help exit: $?"
/tmp/grapes-test version;     echo "version exit: $?"
/tmp/grapes-test next 2>&1;   echo "unknown exit: $?"
```

## Pass Criteria

- `--help` prints usage to stdout, exit 0, no prompt.
- `version` prints the version string (e.g. `0.1.9`), exit 0.
- `next` prints `Unknown command: next` and usage to stderr, exit code 2, and does not hang waiting on a TUI/terminal.
- `go build -o /tmp/grapes-test .` exits 0.
