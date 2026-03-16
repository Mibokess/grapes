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
			{Name: "", Title: "Child from main", Status: data.StatusTodo, Parent: &parentID},
			{Name: "feature-wt", Title: "Child from worktree", Status: data.StatusInProgress, Parent: &parentID},
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

func TestDetailView_MultiSource_ChildrenComputedPerSource(t *testing.T) {
	parentID := 10
	// Child 20: parent=10 in main, no parent in worktree
	child20 := data.Issue{
		ID:     20,
		Title:  "Child20 main",
		Status: data.StatusTodo,
		Parent: &parentID, // active source is main
		Sources: []data.IssueSource{
			{Name: "", Title: "Child20 main", Status: data.StatusTodo, Parent: &parentID},
			{Name: "feature-wt", Title: "Child20 wt", Status: data.StatusTodo, Parent: nil},
		},
		ActiveSource: 0,
	}
	// Child 30: no parent in main, parent=10 in worktree
	child30 := data.Issue{
		ID:     30,
		Title:  "Child30 main",
		Status: data.StatusInProgress,
		Parent: nil, // active source is main
		Sources: []data.IssueSource{
			{Name: "", Title: "Child30 main", Status: data.StatusInProgress, Parent: nil},
			{Name: "feature-wt", Title: "Child30 wt", Status: data.StatusInProgress, Parent: &parentID},
		},
		ActiveSource: 0,
	}
	parent := data.Issue{
		ID:       10,
		Title:    "Parent in worktree",
		Status:   data.StatusInProgress,
		Children: []int{20}, // globally computed from active sources
		Worktree: "feature-wt",
		Sources: []data.IssueSource{
			{Name: "", Title: "Parent in main"},
			{Name: "feature-wt", Title: "Parent in worktree"},
		},
		ActiveSource: 1,
	}
	allIssues := []data.Issue{parent, child20, child30}

	// Viewing parent from worktree: should show child30 (parent=10 in wt), NOT child20 (no parent in wt)
	m := detail.New(parent, allIssues, 100, 40, common.NewTheme(true))
	view := testutil.StripANSI(m.View())

	if !strings.Contains(view, "Child30 wt") {
		t.Error("should show child30 (has parent=10 in worktree source)")
	}
	if strings.Contains(view, "Child20") {
		t.Error("should NOT show child20 (has no parent in worktree source)")
	}

	// Now view parent from main: should show child20, NOT child30
	parentMain := parent
	parentMain.Worktree = ""
	parentMain.ActiveSource = 0
	parentMain.Children = []int{20} // globally computed
	m2 := detail.New(parentMain, allIssues, 100, 40, common.NewTheme(true))
	view2 := testutil.StripANSI(m2.View())

	if !strings.Contains(view2, "Child20 main") {
		t.Error("should show child20 from main (has parent=10 in main)")
	}
	if strings.Contains(view2, "Child30") {
		t.Error("should NOT show child30 from main (has no parent in main)")
	}
}

func TestDetailView_MultiSource_BlocksComputedPerSource(t *testing.T) {
	blockerID := 10
	// Issue 20: blocked_by=[10] in main, no blocker in worktree
	issue20 := data.Issue{
		ID:        20,
		Title:     "Issue20 main",
		Status:    data.StatusTodo,
		BlockedBy: []int{blockerID},
		Sources: []data.IssueSource{
			{Name: "", Title: "Issue20 main", Status: data.StatusTodo, BlockedBy: []int{blockerID}},
			{Name: "feature-wt", Title: "Issue20 wt", Status: data.StatusTodo, BlockedBy: nil},
		},
		ActiveSource: 0,
	}
	// Issue 30: no blocker in main, blocked_by=[10] in worktree
	issue30 := data.Issue{
		ID:        30,
		Title:     "Issue30 main",
		Status:    data.StatusInProgress,
		BlockedBy: nil,
		Sources: []data.IssueSource{
			{Name: "", Title: "Issue30 main", Status: data.StatusInProgress, BlockedBy: nil},
			{Name: "feature-wt", Title: "Issue30 wt", Status: data.StatusInProgress, BlockedBy: []int{blockerID}},
		},
		ActiveSource: 0,
	}
	blocker := data.Issue{
		ID:       10,
		Title:    "Blocker in worktree",
		Status:   data.StatusInProgress,
		Blocks:   []int{20}, // globally computed
		Worktree: "feature-wt",
		Sources: []data.IssueSource{
			{Name: "", Title: "Blocker in main"},
			{Name: "feature-wt", Title: "Blocker in worktree"},
		},
		ActiveSource: 1,
	}
	allIssues := []data.Issue{blocker, issue20, issue30}

	// Viewing blocker from worktree: should show blocks issue30, NOT issue20
	m := detail.New(blocker, allIssues, 100, 40, common.NewTheme(true))
	view := testutil.StripANSI(m.View())

	if !strings.Contains(view, "Issue30 wt") {
		t.Error("should show issue30 in blocks (has blocked_by=10 in worktree)")
	}
	if strings.Contains(view, "Issue20") {
		t.Error("should NOT show issue20 in blocks (has no blocked_by=10 in worktree)")
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
