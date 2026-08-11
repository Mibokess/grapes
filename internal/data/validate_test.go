package data

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateRelationshipsRejectsDanglingSelfAndCycles(t *testing.T) {
	parent := 2
	one := 1
	graph := map[int]RelationshipMeta{
		1: {Parent: &parent, BlockedBy: []int{1}},
		2: {Parent: &one},
		3: {BlockedBy: []int{4}},
		4: {BlockedBy: []int{3}},
	}
	errs := ValidateRelationships(graph)
	if len(errs) == 0 {
		t.Fatal("expected relationship validation errors")
	}
	joined := make([]string, len(errs))
	for i, err := range errs {
		joined[i] = err.Error()
	}
	all := strings.Join(joined, "\n")
	for _, want := range []string{"#1 blocked_by: cannot be blocked by itself", "#1 parent: parent relationship cycle detected", "#3 blocked_by: blocked_by relationship cycle detected"} {
		if !strings.Contains(all, want) {
			t.Errorf("missing %q in errors:\n%s", want, all)
		}
	}
}

func TestValidateIssueReportsUnreadableSecondaryFile(t *testing.T) {
	dir := t.TempDir()
	issueDir := filepath.Join(dir, "1")
	if err := os.MkdirAll(issueDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(issueDir, "meta.toml"), []byte(validMeta), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(issueDir, "comments.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	errs := ValidateIssue(dir, 1)
	if len(errs) == 0 || !strings.Contains(errs[len(errs)-1].Message, "cannot read file") {
		t.Fatalf("expected unreadable comments error, got %+v", errs)
	}
}

func TestValidateAllRejectsNonpositiveDirectories(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"0", "-1"} {
		if err := os.Mkdir(filepath.Join(dir, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	errs := ValidateAll(dir)
	if len(errs) != 2 {
		t.Fatalf("ValidateAll returned %d errors, want 2: %+v", len(errs), errs)
	}
	for _, err := range errs {
		if err.Field != "id" || err.Message != "must be positive" || err.IssueID > 0 {
			t.Errorf("unexpected nonpositive-ID error: %+v", err)
		}
	}
}
