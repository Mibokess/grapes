package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestValidateCommandArgs(t *testing.T) {
	tests := []struct {
		name    string
		command string
		args    []string
		wantErr bool
	}{
		{name: "issue without ID", command: "issue"},
		{name: "issue with one ID", command: "issue", args: []string{"12"}},
		{name: "validate without IDs", command: "validate"},
		{name: "validate with IDs", command: "validate", args: []string{"1", "2"}},
		{name: "issue extra ID", command: "issue", args: []string{"1", "2"}, wantErr: true},
		{name: "issue zero", command: "issue", args: []string{"0"}, wantErr: true},
		{name: "validate zero", command: "validate", args: []string{"0"}, wantErr: true},
		{name: "validate negative", command: "validate", args: []string{"-1"}, wantErr: true},
		{name: "validate nonnumeric", command: "validate", args: []string{"abc"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCommandArgs(tt.command, tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateCommandArgs(%q, %v) error = %v, wantErr=%v", tt.command, tt.args, err, tt.wantErr)
			}
		})
	}
}

func TestRunIssueRejectsExtraArguments(t *testing.T) {
	if got := runIssue(t.TempDir(), []string{"1", "2"}); got != 2 {
		t.Fatalf("runIssue with extra arguments returned %d, want usage error 2", got)
	}
}

func TestRunValidateRejectsNonPositiveIDs(t *testing.T) {
	for _, arg := range []string{"0", "-1"} {
		if got := runValidate(t.TempDir(), []string{arg}); got != 2 {
			t.Fatalf("runValidate(%q) returned %d, want usage error 2", arg, got)
		}
	}
}

func TestWriteProgramErrorForMissingTTY(t *testing.T) {
	var got strings.Builder
	err := fmt.Errorf("%w: open /dev/tty", errInteractiveTerminalUnavailable)

	writeProgramError(&got, err)

	output := got.String()
	for _, want := range []string{
		"Error: bare `grapes` tried to launch the interactive TUI, but no interactive terminal is available.",
		"USAGE:",
		"grapes                    Launch the interactive TUI (needs a real terminal)",
		"Automated/agent environments should use the subcommands above instead.",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("writeProgramError() missing %q in output:\n%s", want, output)
		}
	}
	if strings.Contains(output, "bubbletea") {
		t.Errorf("writeProgramError() leaked Bubble Tea implementation error:\n%s", output)
	}
}
