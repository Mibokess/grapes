package list_test

import (
	"testing"

	"github.com/Mibokess/grapes/internal/tui/list"
	"github.com/Mibokess/grapes/internal/tui/testutil"
)

func TestListView_Default(t *testing.T) {
	issues := testutil.SampleIssues()
	m := list.New(issues)
	m = m.SetSize(100, 30)
	testutil.RequireGolden(t, m.View())
}

func TestListView_Wide(t *testing.T) {
	issues := testutil.SampleIssues()
	m := list.New(issues)
	m = m.SetSize(160, 30)
	testutil.RequireGolden(t, m.View())
}

func TestListView_Narrow(t *testing.T) {
	issues := testutil.SampleIssues()
	m := list.New(issues)
	m = m.SetSize(70, 20)
	testutil.RequireGolden(t, m.View())
}

func TestListView_Empty(t *testing.T) {
	m := list.New(nil)
	m = m.SetSize(100, 30)
	testutil.RequireGolden(t, m.View())
}
