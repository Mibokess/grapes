package board_test

import (
	"testing"

	"github.com/Mibokess/grapes/internal/tui/board"
	"github.com/Mibokess/grapes/internal/tui/common"
	"github.com/Mibokess/grapes/internal/tui/testutil"
	tea "charm.land/bubbletea/v2"
)

func newBoardModel() board.Model {
	issues := testutil.SampleIssues()
	return board.New(issues).SetTopOffset(1).SetSize(100, 30)
}

func keyMsg(k string) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: -2, Text: k})
}

func extractMsg(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}
	return cmd()
}

// --- Keyboard navigation ---

func TestBoard_KeyNavigation_RightThenEnter(t *testing.T) {
	m := newBoardModel()
	// Get issue from first column
	_, cmd1 := m.Update(tea.KeyPressMsg(tea.Key{Code: 13}))
	msg1 := extractMsg(cmd1).(common.OpenDetailMsg)

	// Move right and get issue from second column
	m2 := newBoardModel()
	m2, _ = m2.Update(keyMsg("l"))
	_, cmd2 := m2.Update(tea.KeyPressMsg(tea.Key{Code: 13}))
	msg2 := extractMsg(cmd2).(common.OpenDetailMsg)

	if msg1.ID == msg2.ID {
		t.Error("moving right should select a different column's issue")
	}
}

func TestBoard_KeyNavigation_DownThenEnter(t *testing.T) {
	// Navigate to in_progress column (index 2) which has 2 issues (1 and 6)
	m := newBoardModel()
	m, _ = m.Update(keyMsg("l")) // to todo
	m, _ = m.Update(keyMsg("l")) // to in_progress
	_, cmd1 := m.Update(tea.KeyPressMsg(tea.Key{Code: 13}))
	id1 := extractMsg(cmd1).(common.OpenDetailMsg).ID

	m2 := newBoardModel()
	m2, _ = m2.Update(keyMsg("l"))
	m2, _ = m2.Update(keyMsg("l"))
	m2, _ = m2.Update(keyMsg("j"))
	_, cmd2 := m2.Update(tea.KeyPressMsg(tea.Key{Code: 13}))
	id2 := extractMsg(cmd2).(common.OpenDetailMsg).ID

	if id1 == id2 {
		t.Error("moving down should select a different issue")
	}
}

func TestBoard_KeyNavigation_DownUpReturnsToSame(t *testing.T) {
	m := newBoardModel()
	_, cmd1 := m.Update(tea.KeyPressMsg(tea.Key{Code: 13}))
	id1 := extractMsg(cmd1).(common.OpenDetailMsg).ID

	m2 := newBoardModel()
	m2, _ = m2.Update(keyMsg("j"))
	m2, _ = m2.Update(keyMsg("k"))
	_, cmd2 := m2.Update(tea.KeyPressMsg(tea.Key{Code: 13}))
	id2 := extractMsg(cmd2).(common.OpenDetailMsg).ID

	if id1 != id2 {
		t.Error("down then up should return to the same issue")
	}
}

func TestBoard_KeyNavigation_LeftRightReturnsToSame(t *testing.T) {
	m := newBoardModel()
	_, cmd1 := m.Update(tea.KeyPressMsg(tea.Key{Code: 13}))
	id1 := extractMsg(cmd1).(common.OpenDetailMsg).ID

	m2 := newBoardModel()
	m2, _ = m2.Update(keyMsg("l"))
	m2, _ = m2.Update(keyMsg("h"))
	_, cmd2 := m2.Update(tea.KeyPressMsg(tea.Key{Code: 13}))
	id2 := extractMsg(cmd2).(common.OpenDetailMsg).ID

	if id1 != id2 {
		t.Error("right then left should return to the same issue")
	}
}

func TestBoard_KeyEnter_OpensDetail(t *testing.T) {
	m := newBoardModel()
	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: 13}))
	msg := extractMsg(cmd)
	if _, ok := msg.(common.OpenDetailMsg); !ok {
		t.Errorf("enter should send OpenDetailMsg, got %T", msg)
	}
}

func TestBoard_KeyE_LaunchesEdit(t *testing.T) {
	m := newBoardModel()
	_, cmd := m.Update(keyMsg("e"))
	msg := extractMsg(cmd)
	if _, ok := msg.(common.LaunchEditMsg); !ok {
		t.Errorf("e should send LaunchEditMsg, got %T", msg)
	}
}

func TestBoard_KeyS_ShowsStatusPicker(t *testing.T) {
	m := newBoardModel()
	_, cmd := m.Update(keyMsg("s"))
	msg := extractMsg(cmd)
	picker, ok := msg.(common.ShowPickerMsg)
	if !ok {
		t.Fatalf("s should send ShowPickerMsg, got %T", msg)
	}
	if picker.Field != "status" {
		t.Errorf("expected field 'status', got %q", picker.Field)
	}
}

func TestBoard_KeyP_ShowsPriorityPicker(t *testing.T) {
	m := newBoardModel()
	_, cmd := m.Update(keyMsg("p"))
	msg := extractMsg(cmd)
	picker, ok := msg.(common.ShowPickerMsg)
	if !ok {
		t.Fatalf("p should send ShowPickerMsg, got %T", msg)
	}
	if picker.Field != "priority" {
		t.Errorf("expected field 'priority', got %q", picker.Field)
	}
}

func TestBoard_KeyO_CyclesSort(t *testing.T) {
	m := newBoardModel()
	_, cmd := m.Update(keyMsg("o"))
	if _, ok := extractMsg(cmd).(common.CycleSortMsg); !ok {
		t.Error("o should send CycleSortMsg")
	}
}

func TestBoard_KeyShiftO_ReversesSort(t *testing.T) {
	m := newBoardModel()
	_, cmd := m.Update(keyMsg("O"))
	if _, ok := extractMsg(cmd).(common.ReverseSortMsg); !ok {
		t.Error("O should send ReverseSortMsg")
	}
}

func TestBoard_KeyF_ShowsFilter(t *testing.T) {
	m := newBoardModel()
	_, cmd := m.Update(keyMsg("f"))
	if _, ok := extractMsg(cmd).(common.ShowFilterMenuMsg); !ok {
		t.Error("f should send ShowFilterMenuMsg")
	}
}

func TestBoard_KeyShiftL_SwitchesToList(t *testing.T) {
	m := newBoardModel()
	_, cmd := m.Update(keyMsg("L"))
	sw, ok := extractMsg(cmd).(common.SwitchScreenMsg)
	if !ok {
		t.Fatal("L should send SwitchScreenMsg")
	}
	if sw.Screen != common.ScreenList {
		t.Error("expected switch to list screen")
	}
}

func TestBoard_KeyR_Refreshes(t *testing.T) {
	m := newBoardModel()
	_, cmd := m.Update(keyMsg("r"))
	if _, ok := extractMsg(cmd).(common.RefreshMsg); !ok {
		t.Error("r should send RefreshMsg")
	}
}

// --- Mouse interactions ---

func TestBoard_MouseWheel_NavigatesColumns(t *testing.T) {
	m := newBoardModel()
	_, cmd1 := m.Update(tea.KeyPressMsg(tea.Key{Code: 13}))
	id1 := extractMsg(cmd1).(common.OpenDetailMsg).ID

	// Scroll down (moves to next column)
	m2 := newBoardModel()
	m2, _ = m2.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	_, cmd2 := m2.Update(tea.KeyPressMsg(tea.Key{Code: 13}))
	id2 := extractMsg(cmd2).(common.OpenDetailMsg).ID

	if id1 == id2 {
		t.Error("mouse wheel down should move to next column")
	}
}

func TestBoard_MouseWheelUp_AtStart_NoChange(t *testing.T) {
	m := newBoardModel()
	_, cmd1 := m.Update(tea.KeyPressMsg(tea.Key{Code: 13}))
	id1 := extractMsg(cmd1).(common.OpenDetailMsg).ID

	m2 := newBoardModel()
	m2, _ = m2.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	_, cmd2 := m2.Update(tea.KeyPressMsg(tea.Key{Code: 13}))
	id2 := extractMsg(cmd2).(common.OpenDetailMsg).ID

	if id1 != id2 {
		t.Error("mouse wheel up at start should not change column")
	}
}

func TestBoard_ClickCard_SelectsIt(t *testing.T) {
	m := newBoardModel()
	_, cmd1 := m.Update(tea.KeyPressMsg(tea.Key{Code: 13}))
	id1 := extractMsg(cmd1).(common.OpenDetailMsg).ID

	// Click on a card in the second visible column.
	// Width=100, 3 visible cols → ~33 chars each. Col 1 starts at x=34.
	// Cards start at y = topOffset(1) + headerH(2) = 3, cardH=7.
	m2 := newBoardModel()
	m2, _ = m2.Update(tea.MouseClickMsg{X: 40, Y: 5, Button: tea.MouseLeft})
	_, cmd2 := m2.Update(tea.KeyPressMsg(tea.Key{Code: 13}))
	msg2 := extractMsg(cmd2)
	if msg2 == nil {
		t.Fatal("expected OpenDetailMsg after clicking card")
	}
	id2 := msg2.(common.OpenDetailMsg).ID

	if id1 == id2 {
		t.Error("clicking card in different column should select a different issue")
	}
}

func TestBoard_ClickRelease_OpensDetail(t *testing.T) {
	m := newBoardModel()
	// Click on first card: topOffset=1, headerH=2, so cards start at y=3
	m, _ = m.Update(tea.MouseClickMsg{X: 5, Y: 5, Button: tea.MouseLeft})
	_, cmd := m.Update(tea.MouseReleaseMsg{X: 5, Y: 5, Button: tea.MouseLeft})
	if _, ok := extractMsg(cmd).(common.OpenDetailMsg); !ok {
		t.Error("click-release on card should send OpenDetailMsg")
	}
}

func TestBoard_DragToColumn_MovesIssue(t *testing.T) {
	m := newBoardModel()
	m, _ = m.Update(tea.MouseClickMsg{X: 5, Y: 5, Button: tea.MouseLeft})
	m, _ = m.Update(tea.MouseMotionMsg{X: 40, Y: 5, Button: tea.MouseLeft})
	_, cmd := m.Update(tea.MouseReleaseMsg{X: 40, Y: 5, Button: tea.MouseLeft})
	if _, ok := extractMsg(cmd).(common.MoveIssueMsg); !ok {
		t.Error("drag to different column should send MoveIssueMsg")
	}
}

func TestBoard_DragSameColumn_NoMove(t *testing.T) {
	m := newBoardModel()
	m, _ = m.Update(tea.MouseClickMsg{X: 5, Y: 5, Button: tea.MouseLeft})
	m, _ = m.Update(tea.MouseMotionMsg{X: 10, Y: 8, Button: tea.MouseLeft})
	_, cmd := m.Update(tea.MouseReleaseMsg{X: 10, Y: 8, Button: tea.MouseLeft})
	if _, ok := extractMsg(cmd).(common.MoveIssueMsg); ok {
		t.Error("drag within same column should not send MoveIssueMsg")
	}
}

func TestBoard_ForwardButton_OpensDetail(t *testing.T) {
	m := newBoardModel()
	_, cmd := m.Update(tea.MouseClickMsg{X: 5, Y: 5, Button: tea.MouseForward})
	if _, ok := extractMsg(cmd).(common.OpenDetailMsg); !ok {
		t.Error("forward button should send OpenDetailMsg")
	}
}

// --- Empty board ---

func TestBoard_Empty_KeysNoOp(t *testing.T) {
	m := board.New(nil).SetTopOffset(1).SetSize(100, 30)
	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: 13}))
	if extractMsg(cmd) != nil {
		t.Error("enter on empty board should not send a message")
	}
	_, cmd = m.Update(keyMsg("s"))
	if extractMsg(cmd) != nil {
		t.Error("s on empty board should not send a message")
	}
}

func TestBoard_Empty_ClickNoOp(t *testing.T) {
	m := board.New(nil).SetTopOffset(1).SetSize(100, 30)
	m, _ = m.Update(tea.MouseClickMsg{X: 5, Y: 5, Button: tea.MouseLeft})
	m, _ = m.Update(tea.MouseReleaseMsg{X: 5, Y: 5, Button: tea.MouseLeft})
	// Should not panic
}
