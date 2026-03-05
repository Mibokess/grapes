package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/Mibokess/grapes/internal/config"
	"github.com/Mibokess/grapes/internal/data"
	"github.com/Mibokess/grapes/internal/tui"
	tea "charm.land/bubbletea/v2"
)

var version = "0.1.5"

func main() {
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
	issues, err := data.LoadAllSources(issuesDir, projectRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading issues: %v\n", err)
		os.Exit(1)
	}

	cfg := config.Load(issuesDir)
	model := tui.NewModel(issues, issuesDir, cfg, version)
	p := tea.NewProgram(model)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runIssue(issuesDir string, args []string) int {
	if len(args) == 0 {
		// No ID: allocate next ID, stamp timestamps, print ID
		id, err := data.NextID(issuesDir)
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
