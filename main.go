package main

import (
	"fmt"
	"io"
	"os"
	"strconv"

	tea "charm.land/bubbletea/v2"
	"github.com/Mibokess/grapes/internal/config"
	"github.com/Mibokess/grapes/internal/data"
	"github.com/Mibokess/grapes/internal/tui"
)

var version = "0.1.9"

func main() {
	// Handle help/version and reject unknown commands before touching the
	// filesystem, so they work anywhere without triggering the .grapes/
	// creation prompt below. issue/validate need issuesDir and are
	// dispatched further down.
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "help", "--help", "-h":
			writeHelp(os.Stdout)
			os.Exit(0)
		case "version", "--version", "-v":
			fmt.Println(version)
			os.Exit(0)
		case "issue", "validate":
			// handled after issuesDir is resolved below
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
			writeHelp(os.Stderr)
			os.Exit(2)
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	issuesDir, err := data.FindIssuesDir(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "No .grapes/ directory found.\n")
		fmt.Fprintf(os.Stderr, "Create one in %s? [y/N] ", cwd)
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			os.Exit(1)
		}
		issuesDir = cwd + "/.grapes"
		if err := os.MkdirAll(issuesDir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating .grapes/: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Created %s\n", issuesDir)
	}

	// Handle subcommands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "validate":
			os.Exit(runValidate(issuesDir, os.Args[2:]))
		case "issue":
			os.Exit(runIssue(issuesDir, os.Args[2:]))
		}
	}

	projectRoot := data.ProjectRoot(issuesDir)
	cfg, cfgErr := config.Load(issuesDir)
	issues, err := data.LoadAllSources(issuesDir, projectRoot, cfg.Sources.Dirs...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading issues: %v\n", err)
		os.Exit(1)
	}
	model := tui.NewModel(issues, issuesDir, cfg, version)
	if cfgErr != nil {
		// The TUI owns the screen from here, so a stderr warning would be
		// wiped by the alt-screen switch. Surface it in the status bar.
		model = model.WithStatus("Config error (using defaults): " + cfgErr.Error())
	}
	p := tea.NewProgram(model)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func writeHelp(w io.Writer) {
	fmt.Fprint(w, `grapes — file-based issue tracker (TUI + CLI)

USAGE:
  grapes                    Launch the interactive TUI (needs a real terminal)
  grapes <command> [args]

COMMANDS:
  issue                     Allocate the next ID, create .grapes/<id>/, print the ID
  issue <id>                Bump the 'updated' timestamp on issue <id>
  validate [<id>...]        Validate all issues, or only the given IDs
  help, --help, -h          Show this help
  version, --version, -v    Show the version

ISSUE FILES (.grapes/<id>/):
  meta.toml                 title, status, priority, labels, dates
  content.md                issue description (markdown)
  comments.md               append-only comment log

  status:   backlog, todo, in_progress, done, cancelled
  priority: urgent, high, medium, low

TYPICAL AGENT WORKFLOW:
  id=$(grapes issue)        # create .grapes/$id/, capture the new ID
  # edit .grapes/$id/meta.toml and content.md
  grapes issue $id          # bump 'updated' after editing
  grapes validate $id       # check the issue is well-formed

NOTE:
  Bare 'grapes' opens a full-screen TUI and requires an interactive terminal.
  Automated/agent environments should use the subcommands above instead.
`)
}

func runIssue(issuesDir string, args []string) int {
	cfg, err := config.Load(issuesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error (using defaults): %v\n", err)
	}
	if len(args) == 0 {
		// No ID: allocate next ID, stamp timestamps, print ID
		id, err := data.NextID(issuesDir, cfg.Sources.Dirs...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		if err := data.StampTimestamps(issuesDir, id); err != nil {
			fmt.Fprintf(os.Stderr, "Error stamping timestamps: %v\n", err)
			return 1
		}
		fmt.Println(id)
		return 0
	}

	// With ID: stamp timestamps on existing issue
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid issue ID: %s\n", args[0])
		return 1
	}

	// Create directory if it doesn't exist
	issueDir := issuesDir + "/" + args[0]
	if err := os.MkdirAll(issueDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
		return 1
	}

	if err := data.StampTimestamps(issuesDir, id); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}

func runValidate(issuesDir string, args []string) int {
	var errs []data.ValidationError

	if len(args) > 0 {
		// Validate specific issue(s)
		for _, arg := range args {
			id, err := strconv.Atoi(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid issue ID: %s\n", arg)
				return 1
			}
			errs = append(errs, data.ValidateIssue(issuesDir, id)...)
		}
	} else {
		// Validate all issues
		errs = data.ValidateAll(issuesDir)
	}

	if len(errs) == 0 {
		fmt.Println("All issues valid.")
		return 0
	}

	for _, e := range errs {
		fmt.Fprintf(os.Stderr, "  %s\n", e.Error())
	}
	fmt.Fprintf(os.Stderr, "\n%d problem(s) found.\n", len(errs))
	return 1
}
