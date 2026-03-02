package board_test

import (
	"testing"

	"github.com/Mibokess/grapes/internal/tui/board"
	"github.com/Mibokess/grapes/internal/tui/testutil"
)

func TestBoardView_Default(t *testing.T) {
	issues := testutil.SampleIssues()
	m := board.New(issues)
	m = m.SetSize(100, 30)
	testutil.RequireGolden(t, m.View())
}

func TestBoardView_Wide(t *testing.T) {
	issues := testutil.SampleIssues()
	m := board.New(issues)
	m = m.SetSize(160, 30)
	testutil.RequireGolden(t, m.View())
}

func TestBoardView_Narrow(t *testing.T) {
	issues := testutil.SampleIssues()
	m := board.New(issues)
	m = m.SetSize(70, 30)
	testutil.RequireGolden(t, m.View())
}

func TestBoardView_Short(t *testing.T) {
	issues := testutil.SampleIssues()
	m := board.New(issues)
	m = m.SetSize(100, 12)
	testutil.RequireGolden(t, m.View())
}

func TestBoardView_Empty(t *testing.T) {
	m := board.New(nil)
	m = m.SetSize(100, 30)
	testutil.RequireGolden(t, m.View())
}
