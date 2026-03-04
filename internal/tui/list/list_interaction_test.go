package list_test

import (
	"testing"

	"github.com/Mibokess/grapes/internal/tui/common"
	"github.com/Mibokess/grapes/internal/tui/list"
	"github.com/Mibokess/grapes/internal/tui/testutil"
	tea "charm.land/bubbletea/v2"
)

func newListModel() list.Model {
	issues := testutil.SampleIssues()
	return list.New(issues).SetTopOffset(1).SetSize(100, 30)
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

func TestList_KeyDown_SelectsDifferentIssue(t *testing.T) {
	m := newListModel()
	_, cmd1 := m.Update(tea.KeyPressMsg(tea.Key{Code: 13}))
	id1 := extractMsg(cmd1).(common.OpenDetailMsg).ID

	m2 := newListModel()
	m2, _ = m2.Update(keyMsg("j"))
	_, cmd2 := m2.Update(tea.KeyPressMsg(tea.Key{Code: 13}))
	id2 := extractMsg(cmd2).(common.OpenDetailMsg).ID

	if id1 == id2 {
		t.Error("j should move to a different issue")
	}
}

func TestList_KeyUpDown_ReturnsToSame(t *testing.T) {
	m := newListModel()
	_, cmd1 := m.Update(tea.KeyPressMsg(tea.Key{Code: 13}))
	id1 := extractMsg(cmd1).(common.OpenDetailMsg).ID

	m2 := newListModel()
	m2, _ = m2.Update(keyMsg("j"))
	m2, _ = m2.Update(keyMsg("k"))
	_, cmd2 := m2.Update(tea.KeyPressMsg(tea.Key{Code: 13}))
	id2 := extractMsg(cmd2).(common.OpenDetailMsg).ID

	if id1 != id2 {
		t.Error("down then up should return to same issue")
	}
}

func TestList_KeyEnter_OpensDetail(t *testing.T) {
	m := newListModel()
	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: 13}))
	if _, ok := extractMsg(cmd).(common.OpenDetailMsg); !ok {
		t.Error("enter should send OpenDetailMsg")
	}
}

func TestList_KeyE_LaunchesEdit(t *testing.T) {
	m := newListModel()
	_, cmd := m.Update(keyMsg("e"))
	if _, ok := extractMsg(cmd).(common.LaunchEditMsg); !ok {
		t.Error("e should send LaunchEditMsg")
	}
}

func TestList_KeyS_ShowsStatusPicker(t *testing.T) {
	m := newListModel()
	_, cmd := m.Update(keyMsg("s"))
	picker, ok := extractMsg(cmd).(common.ShowPickerMsg)
	if !ok {
		t.Fatal("s should send ShowPickerMsg")
	}
	if picker.Field != "status" {
		t.Errorf("expected field 'status', got %q", picker.Field)
	}
}

func TestList_KeyP_ShowsPriorityPicker(t *testing.T) {
	m := newListModel()
	_, cmd := m.Update(keyMsg("p"))
	picker, ok := extractMsg(cmd).(common.ShowPickerMsg)
	if !ok {
		t.Fatal("p should send ShowPickerMsg")
	}
	if picker.Field != "priority" {
		t.Errorf("expected field 'priority', got %q", picker.Field)
	}
}

func TestList_KeyO_CyclesSort(t *testing.T) {
	m := newListModel()
	_, cmd := m.Update(keyMsg("o"))
	if _, ok := extractMsg(cmd).(common.CycleSortMsg); !ok {
		t.Error("o should send CycleSortMsg")
	}
}

func TestList_KeyShiftO_ReversesSort(t *testing.T) {
	m := newListModel()
	_, cmd := m.Update(keyMsg("O"))
	if _, ok := extractMsg(cmd).(common.ReverseSortMsg); !ok {
		t.Error("O should send ReverseSortMsg")
	}
}

func TestList_KeyF_ShowsFilter(t *testing.T) {
	m := newListModel()
	_, cmd := m.Update(keyMsg("f"))
	if _, ok := extractMsg(cmd).(common.ShowFilterMenuMsg); !ok {
		t.Error("f should send ShowFilterMenuMsg")
	}
}

func TestList_KeyShiftB_SwitchesToBoard(t *testing.T) {
	m := newListModel()
	_, cmd := m.Update(keyMsg("B"))
	sw, ok := extractMsg(cmd).(common.SwitchScreenMsg)
	if !ok {
		t.Fatal("B should send SwitchScreenMsg")
	}
	if sw.Screen != common.ScreenBoard {
		t.Error("expected switch to board screen")
	}
}

func TestList_KeyR_Refreshes(t *testing.T) {
	m := newListModel()
	_, cmd := m.Update(keyMsg("r"))
	if _, ok := extractMsg(cmd).(common.RefreshMsg); !ok {
		t.Error("r should send RefreshMsg")
	}
}

func TestList_KeySlash_EntersFilterMode(t *testing.T) {
	m := newListModel()
	m, _ = m.Update(keyMsg("/"))
	if !m.Filtering() {
		t.Error("/ should enter filter mode")
	}
}

func TestList_FilterMode_EscClears(t *testing.T) {
	m := newListModel()
	m, _ = m.Update(keyMsg("/"))
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: 27})) // esc
	if m.Filtering() {
		t.Error("esc should exit filter mode")
	}
}

// --- Mouse interactions ---

func TestList_MouseWheel_MovesSelection(t *testing.T) {
	m := newListModel()
	_, cmd1 := m.Update(tea.KeyPressMsg(tea.Key{Code: 13}))
	id1 := extractMsg(cmd1).(common.OpenDetailMsg).ID

	m2 := newListModel()
	m2, _ = m2.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	_, cmd2 := m2.Update(tea.KeyPressMsg(tea.Key{Code: 13}))
	id2 := extractMsg(cmd2).(common.OpenDetailMsg).ID

	if id1 == id2 {
		t.Error("mouse wheel down should move to a different issue")
	}
}

func TestList_MouseWheelUp_AtTop_NoChange(t *testing.T) {
	m := newListModel()
	_, cmd1 := m.Update(tea.KeyPressMsg(tea.Key{Code: 13}))
	id1 := extractMsg(cmd1).(common.OpenDetailMsg).ID

	m2 := newListModel()
	m2, _ = m2.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	_, cmd2 := m2.Update(tea.KeyPressMsg(tea.Key{Code: 13}))
	id2 := extractMsg(cmd2).(common.OpenDetailMsg).ID

	if id1 != id2 {
		t.Error("mouse wheel up at top should not change selection")
	}
}

func TestList_ClickRow_OpensDetail(t *testing.T) {
	m := newListModel()
	// Click on first data row, in the Title column (col index 1).
	// topOffset=1, header=1, border=1 → data starts at y=3.
	// Title column: x ~6-17 (ID is 4 chars + 2 padding = 6 offset).
	_, cmd := m.Update(tea.MouseClickMsg{X: 10, Y: 3, Button: tea.MouseLeft})
	msg := extractMsg(cmd)
	if _, ok := msg.(common.OpenDetailMsg); !ok {
		t.Errorf("clicking title column should send OpenDetailMsg, got %T", msg)
	}
}

func TestList_ClickStatusColumn_ShowsPicker(t *testing.T) {
	m := newListModel()
	// From golden file: Status column starts around x=30 ("◑ in_progress")
	_, cmd := m.Update(tea.MouseClickMsg{X: 35, Y: 3, Button: tea.MouseLeft})
	msg := extractMsg(cmd)
	picker, ok := msg.(common.ShowPickerMsg)
	if !ok {
		t.Fatalf("clicking status column should send ShowPickerMsg, got %T", msg)
	}
	if picker.Field != "status" {
		t.Errorf("expected field 'status', got %q", picker.Field)
	}
}

func TestList_ClickPriorityColumn_ShowsPicker(t *testing.T) {
	m := newListModel()
	// From golden file: Priority column starts around x=45 ("! high")
	_, cmd := m.Update(tea.MouseClickMsg{X: 48, Y: 3, Button: tea.MouseLeft})
	msg := extractMsg(cmd)
	picker, ok := msg.(common.ShowPickerMsg)
	if !ok {
		t.Fatalf("clicking priority column should send ShowPickerMsg, got %T", msg)
	}
	if picker.Field != "priority" {
		t.Errorf("expected field 'priority', got %q", picker.Field)
	}
}

func TestList_BackwardButton_SwitchesToBoard(t *testing.T) {
	m := newListModel()
	_, cmd := m.Update(tea.MouseClickMsg{X: 5, Y: 5, Button: tea.MouseBackward})
	sw, ok := extractMsg(cmd).(common.SwitchScreenMsg)
	if !ok {
		t.Fatal("backward button should send SwitchScreenMsg")
	}
	if sw.Screen != common.ScreenBoard {
		t.Error("expected switch to board screen")
	}
}

func TestList_ForwardButton_OpensDetail(t *testing.T) {
	m := newListModel()
	_, cmd := m.Update(tea.MouseClickMsg{X: 5, Y: 5, Button: tea.MouseForward})
	if _, ok := extractMsg(cmd).(common.OpenDetailMsg); !ok {
		t.Error("forward button should send OpenDetailMsg")
	}
}

func TestList_ClickWhileFiltering_NoOp(t *testing.T) {
	m := newListModel()
	m, _ = m.Update(keyMsg("/"))
	_, cmd := m.Update(tea.MouseClickMsg{X: 10, Y: 3, Button: tea.MouseLeft})
	msg := extractMsg(cmd)
	if msg != nil {
		t.Error("clicking while filtering should not send a message")
	}
}

// --- Empty list ---

func TestList_Empty_EnterNoOp(t *testing.T) {
	m := list.New(nil).SetTopOffset(1).SetSize(100, 30)
	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: 13}))
	// Should not panic
	_ = cmd
}
