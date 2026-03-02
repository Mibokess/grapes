package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/Mibokess/grapes/internal/data"
	"github.com/Mibokess/grapes/internal/tui"
	tea "charm.land/bubbletea/v2"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	issuesDir, err := data.FindIssuesDir(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Handle "validate" subcommand
	if len(os.Args) > 1 && os.Args[1] == "validate" {
		os.Exit(runValidate(issuesDir, os.Args[2:]))
	}

	issues, err := data.LoadAllIssues(issuesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading issues: %v\n", err)
		os.Exit(1)
	}

	model := tui.NewModel(issues, issuesDir)
	p := tea.NewProgram(model)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
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
