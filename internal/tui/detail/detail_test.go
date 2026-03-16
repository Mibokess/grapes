package detail_test

import (
	"strings"
	"testing"

	"github.com/Mibokess/grapes/internal/data"
	"github.com/Mibokess/grapes/internal/tui/common"
	"github.com/Mibokess/grapes/internal/tui/detail"
	"github.com/Mibokess/grapes/internal/tui/testutil"
)

func TestDetailView_FullIssue(t *testing.T) {
	issues := testutil.SampleIssues()
	// Issue 1 has children, comments, content, labels — the works.
	m := detail.New(issues[0], issues, 100, 40, common.NewTheme(true))
	testutil.RequireGolden(t, m.View())
}

func TestDetailView_SimpleIssue(t *testing.T) {
	issues := testutil.SampleIssues()
	// Issue 3 has no comments, no content, no children.
	m := detail.New(issues[2], issues, 100, 30, common.NewTheme(true))
	testutil.RequireGolden(t, m.View())
}

func TestDetailView_ChildIssue(t *testing.T) {
	issues := testutil.SampleIssues()
	// Issue 5 has a parent link.
	m := detail.New(issues[4], issues, 100, 30, common.NewTheme(true))
	testutil.RequireGolden(t, m.View())
}

func TestDetailView_Narrow(t *testing.T) {
	issues := testutil.SampleIssues()
	m := detail.New(issues[0], issues, 60, 40, common.NewTheme(true))
	testutil.RequireGolden(t, m.View())
}

func TestDetailView_MultiSource_ChildrenUseViewingSource(t *testing.T) {
	parentID := 10
	child := data.Issue{
		ID:       20,
		Title:    "Child from main",
		Status:   data.StatusTodo,
		Parent:   &parentID,
		Worktree: "",
		Sources: []data.IssueSource{
			{Name: "", Title: "Child from main", Status: data.StatusTodo},
			{Name: "feature-wt", Title: "Child from worktree", Status: data.StatusInProgress},
		},
		ActiveSource: 0,
	}
	parent := data.Issue{
		ID:       10,
		Title:    "Parent in worktree",
		Status:   data.StatusInProgress,
		Children: []int{20},
		Worktree: "feature-wt",
		Sources: []data.IssueSource{
			{Name: "", Title: "Parent in main"},
			{Name: "feature-wt", Title: "Parent in worktree"},
		},
		ActiveSource: 1,
	}
	allIssues := []data.Issue{parent, child}
	m := detail.New(parent, allIssues, 100, 40, common.NewTheme(true))
	view := testutil.StripANSI(m.View())

	if !strings.Contains(view, "Child from worktree") {
		t.Error("sub-issue should show title from the viewing issue's source (worktree)")
	}
	if strings.Contains(view, "Child from main") {
		t.Error("sub-issue should NOT show title from main when viewing from worktree")
	}
}

func TestDetailView_MultiSource_ParentUsesViewingSource(t *testing.T) {
	parentID := 10
	parent := data.Issue{
		ID:     10,
		Title:  "Parent from main",
		Status: data.StatusInProgress,
		Sources: []data.IssueSource{
			{Name: "", Title: "Parent from main"},
			{Name: "feature-wt", Title: "Parent from worktree"},
		},
		ActiveSource: 0,
	}
	child := data.Issue{
		ID:       20,
		Title:    "Child in worktree",
		Status:   data.StatusTodo,
		Parent:   &parentID,
		Worktree: "feature-wt",
		Sources: []data.IssueSource{
			{Name: "", Title: "Child in main"},
			{Name: "feature-wt", Title: "Child in worktree"},
		},
		ActiveSource: 1,
	}
	allIssues := []data.Issue{parent, child}
	m := detail.New(child, allIssues, 100, 40, common.NewTheme(true))
	view := testutil.StripANSI(m.View())

	if !strings.Contains(view, "Parent from worktree") {
		t.Error("parent link should show title from the viewing issue's source (worktree)")
	}
	if strings.Contains(view, "Parent from main") {
		t.Error("parent link should NOT show title from main when viewing from worktree")
	}
}

func TestDetailView_MultiSource_FallbackToActiveSource(t *testing.T) {
	parentID := 10
	// Parent only exists in main, child is viewing from worktree
	parent := data.Issue{
		ID:     10,
		Title:  "Parent main only",
		Status: data.StatusDone,
		Sources: []data.IssueSource{
			{Name: "", Title: "Parent main only"},
		},
		ActiveSource: 0,
	}
	child := data.Issue{
		ID:       20,
		Title:    "Child in worktree",
		Status:   data.StatusTodo,
		Parent:   &parentID,
		Worktree: "feature-wt",
		Sources: []data.IssueSource{
			{Name: "feature-wt", Title: "Child in worktree"},
		},
		ActiveSource: 0,
	}
	allIssues := []data.Issue{parent, child}
	m := detail.New(child, allIssues, 100, 40, common.NewTheme(true))
	view := testutil.StripANSI(m.View())

	// Should fall back to parent's active source since it doesn't exist in "feature-wt"
	if !strings.Contains(view, "Parent main only") {
		t.Error("parent link should fall back to active source when viewing source doesn't exist")
	}
}
