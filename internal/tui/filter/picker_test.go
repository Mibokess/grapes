package filter

import (
	"testing"

	"github.com/Mibokess/grapes/internal/config"
	"github.com/Mibokess/grapes/internal/tui/common"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func newTestMultiPicker() MultiPicker {
	theme := common.NewThemeFromConfig(config.Defaults().Theme, true)
	opts := []PickerOption{
		{Value: "todo", Label: "Todo", Icon: "◌", Style: lipgloss.NewStyle()},
		{Value: "in_progress", Label: "In Progress", Icon: "◑", Style: lipgloss.NewStyle()},
		{Value: "done", Label: "Done", Icon: "●", Style: lipgloss.NewStyle()},
	}
	return NewMultiPicker("Status", "status", opts, nil, theme)
}

// --- MultiPicker keyboard tests ---

func TestMultiPicker_KeyNavigation(t *testing.T) {
	m := newTestMultiPicker()
	if m.cursor != 0 {
		t.Fatal("cursor should start at 0")
	}
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: -2, Text: "j"}))
	if m.cursor != 1 {
		t.Error("j should move cursor down")
	}
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: -2, Text: "k"}))
	if m.cursor != 0 {
		t.Error("k should move cursor up")
	}
}

func TestMultiPicker_KeyToggle(t *testing.T) {
	m := newTestMultiPicker()
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: ' '})) // space
	if !m.selected["todo"] {
		t.Error("space should toggle todo on")
	}
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: ' '}))
	if m.selected["todo"] {
		t.Error("space again should toggle todo off")
	}
}

func TestMultiPicker_KeyConfirm(t *testing.T) {
	m := newTestMultiPicker()
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: ' '})) // toggle first
	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: 13}))         // enter
	msg, ok := extractMsg(cmd).(common.FilterPickerResultMsg)
	if !ok {
		t.Fatalf("enter should send FilterPickerResultMsg, got %T", extractMsg(cmd))
	}
	if msg.Field != "status" {
		t.Errorf("expected field 'status', got %q", msg.Field)
	}
	if len(msg.Selected) != 1 || msg.Selected[0] != "todo" {
		t.Errorf("expected selected [todo], got %v", msg.Selected)
	}
}

func TestMultiPicker_KeyCancel(t *testing.T) {
	m := newTestMultiPicker()
	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: 27})) // esc
	if _, ok := extractMsg(cmd).(common.FilterCancelMsg); !ok {
		t.Error("esc should send FilterCancelMsg")
	}
}

// --- MultiPicker mouse tests ---

func TestMultiPicker_MouseClick_TogglesOption(t *testing.T) {
	m := newTestMultiPicker()
	m.ScreenX = 10
	m.ScreenY = 5
	m.ScreenW = 40
	// Click on first option (index 0): Y = ScreenY + 2 (border+padding) + 0
	m, _ = m.Update(tea.MouseClickMsg{X: 15, Y: 7, Button: tea.MouseLeft})
	if !m.selected["todo"] {
		t.Error("clicking first option should toggle 'todo' on")
	}
	// Click again to toggle off
	m, _ = m.Update(tea.MouseClickMsg{X: 15, Y: 7, Button: tea.MouseLeft})
	if m.selected["todo"] {
		t.Error("clicking again should toggle 'todo' off")
	}
}

func TestMultiPicker_MouseClick_SecondOption(t *testing.T) {
	m := newTestMultiPicker()
	m.ScreenX = 10
	m.ScreenY = 5
	m.ScreenW = 40
	// Click on second option (index 1): Y = ScreenY + 2 + 1
	m, _ = m.Update(tea.MouseClickMsg{X: 15, Y: 8, Button: tea.MouseLeft})
	if !m.selected["in_progress"] {
		t.Error("clicking second option should toggle 'in_progress' on")
	}
}

func TestMultiPicker_MouseClick_OutsideCancels(t *testing.T) {
	m := newTestMultiPicker()
	m.ScreenX = 10
	m.ScreenY = 5
	m.ScreenW = 40
	_, cmd := m.Update(tea.MouseClickMsg{X: 0, Y: 0, Button: tea.MouseLeft})
	if _, ok := extractMsg(cmd).(common.FilterCancelMsg); !ok {
		t.Error("clicking outside should send FilterCancelMsg")
	}
}

func TestMultiPicker_MouseClick_HintAreaApplies(t *testing.T) {
	m := newTestMultiPicker()
	m.ScreenX = 10
	m.ScreenY = 5
	m.ScreenW = 40
	// Toggle a value first
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: ' '}))
	// Click on hint area: Y = ScreenY + 2 + len(options) + 1 (blank) + 1 (hint)
	hintY := m.ScreenY + 2 + len(m.options) + 1
	_, cmd := m.Update(tea.MouseClickMsg{X: 15, Y: hintY, Button: tea.MouseLeft})
	msg, ok := extractMsg(cmd).(common.FilterPickerResultMsg)
	if !ok {
		t.Fatalf("clicking hint area should apply, got %T", extractMsg(cmd))
	}
	if len(msg.Selected) != 1 || msg.Selected[0] != "todo" {
		t.Errorf("expected selected [todo], got %v", msg.Selected)
	}
}

func TestMultiPicker_MouseMotion_MovesCursor(t *testing.T) {
	m := newTestMultiPicker()
	m.ScreenX = 10
	m.ScreenY = 5
	m.ScreenW = 40
	if m.cursor != 0 {
		t.Fatal("cursor should start at 0")
	}
	// Hover over third option (index 2): Y = ScreenY + 2 + 2
	m, _ = m.Update(tea.MouseMotionMsg{X: 15, Y: 9})
	if m.cursor != 2 {
		t.Errorf("mouse motion should move cursor to 2, got %d", m.cursor)
	}
}

func TestMultiPicker_PreSelected(t *testing.T) {
	theme := common.NewThemeFromConfig(config.Defaults().Theme, true)
	opts := []PickerOption{
		{Value: "todo", Label: "Todo", Style: lipgloss.NewStyle()},
		{Value: "done", Label: "Done", Style: lipgloss.NewStyle()},
	}
	m := NewMultiPicker("Status", "status", opts, []string{"done"}, theme)
	if !m.selected["done"] {
		t.Error("'done' should be pre-selected")
	}
	if m.selected["todo"] {
		t.Error("'todo' should not be pre-selected")
	}
}
