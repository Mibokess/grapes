package detail_test

import (
	"testing"

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
