package labelpicker

import (
	"testing"

	"github.com/Mibokess/grapes/internal/config"
	"github.com/Mibokess/grapes/internal/tui/common"
	tea "charm.land/bubbletea/v2"
)

func newTestLabelPicker() Model {
	theme := common.NewThemeFromConfig(config.Defaults().Theme, true)
	return New(1, []string{"bug", "feature", "docs"}, []string{"bug"}, theme)
}

func extractMsg(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}
	return cmd()
}

// --- Keyboard tests ---

func TestLabelPicker_KeyNavigation(t *testing.T) {
	m := newTestLabelPicker()
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

func TestLabelPicker_KeyToggle(t *testing.T) {
	m := newTestLabelPicker()
	// "bug" starts selected
	if !m.selected["bug"] {
		t.Fatal("bug should start selected")
	}
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: ' '})) // space
	if m.selected["bug"] {
		t.Error("space should toggle bug off")
	}
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: ' '}))
	if !m.selected["bug"] {
		t.Error("space again should toggle bug on")
	}
}

func TestLabelPicker_KeyConfirm(t *testing.T) {
	m := newTestLabelPicker()
	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: 13})) // enter
	msg, ok := extractMsg(cmd).(common.LabelPickerResultMsg)
	if !ok {
		t.Fatalf("enter should send LabelPickerResultMsg, got %T", extractMsg(cmd))
	}
	if msg.IssueID != 1 {
		t.Errorf("expected issue ID 1, got %d", msg.IssueID)
	}
	if len(msg.Labels) != 1 || msg.Labels[0] != "bug" {
		t.Errorf("expected labels [bug], got %v", msg.Labels)
	}
}

func TestLabelPicker_KeyCancel(t *testing.T) {
	m := newTestLabelPicker()
	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: 27})) // esc
	if _, ok := extractMsg(cmd).(common.LabelPickerCancelMsg); !ok {
		t.Error("esc should send LabelPickerCancelMsg")
	}
}

// --- Mouse tests ---

func TestLabelPicker_MouseClick_TogglesLabel(t *testing.T) {
	m := newTestLabelPicker()
	m.ScreenX = 10
	m.ScreenY = 5
	// Click on first label (index 0 = "bug"): Y = ScreenY + 2 (border+padding) + 0
	m, _ = m.Update(tea.MouseClickMsg{X: 15, Y: 7, Button: tea.MouseLeft})
	if m.selected["bug"] {
		t.Error("clicking 'bug' should toggle it off (was selected)")
	}
	// Click again to toggle on
	m, _ = m.Update(tea.MouseClickMsg{X: 15, Y: 7, Button: tea.MouseLeft})
	if !m.selected["bug"] {
		t.Error("clicking again should toggle 'bug' on")
	}
}

func TestLabelPicker_MouseClick_SecondLabel(t *testing.T) {
	m := newTestLabelPicker()
	m.ScreenX = 10
	m.ScreenY = 5
	// Click on second label (index 1 = "feature"): Y = ScreenY + 2 + 1
	m, _ = m.Update(tea.MouseClickMsg{X: 15, Y: 8, Button: tea.MouseLeft})
	if !m.selected["feature"] {
		t.Error("clicking 'feature' should toggle it on")
	}
}

func TestLabelPicker_MouseClick_OutsideCancels(t *testing.T) {
	m := newTestLabelPicker()
	m.ScreenX = 10
	m.ScreenY = 5
	_, cmd := m.Update(tea.MouseClickMsg{X: 0, Y: 0, Button: tea.MouseLeft})
	if _, ok := extractMsg(cmd).(common.LabelPickerCancelMsg); !ok {
		t.Error("clicking outside should send LabelPickerCancelMsg")
	}
}

func TestLabelPicker_MouseClick_HintAreaApplies(t *testing.T) {
	m := newTestLabelPicker()
	m.ScreenX = 10
	m.ScreenY = 5
	// Hint area: Y = ScreenY + 2 + len(labels) + 3 (blank + input + blank + hint)
	hintY := m.ScreenY + 2 + len(m.labels) + 3
	_, cmd := m.Update(tea.MouseClickMsg{X: 15, Y: hintY, Button: tea.MouseLeft})
	msg, ok := extractMsg(cmd).(common.LabelPickerResultMsg)
	if !ok {
		t.Fatalf("clicking hint area should apply, got %T", extractMsg(cmd))
	}
	if len(msg.Labels) != 1 || msg.Labels[0] != "bug" {
		t.Errorf("expected labels [bug], got %v", msg.Labels)
	}
}

func TestLabelPicker_MouseMotion_MovesCursor(t *testing.T) {
	m := newTestLabelPicker()
	m.ScreenX = 10
	m.ScreenY = 5
	if m.cursor != 0 {
		t.Fatal("cursor should start at 0")
	}
	// Hover over third label (index 2): Y = ScreenY + 2 + 2
	m, _ = m.Update(tea.MouseMotionMsg{X: 15, Y: 9})
	if m.cursor != 2 {
		t.Errorf("mouse motion should move cursor to 2, got %d", m.cursor)
	}
}
