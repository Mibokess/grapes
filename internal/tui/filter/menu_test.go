package filter

import (
	"testing"

	"github.com/Mibokess/grapes/internal/config"
	"github.com/Mibokess/grapes/internal/tui/common"
	tea "charm.land/bubbletea/v2"
)

func newTestMenu() Menu {
	theme := common.NewThemeFromConfig(config.Defaults().Theme, true)
	fs := FilterSet{}
	return NewMenu(fs, 0, theme)
}

func newTestMenuWithFilters() Menu {
	theme := common.NewThemeFromConfig(config.Defaults().Theme, true)
	fs := FilterSet{TopLevelOnly: true}
	return NewMenu(fs, 0, theme)
}

func extractMsg(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}
	return cmd()
}

// --- Menu keyboard tests ---

func TestMenu_KeyNavigation(t *testing.T) {
	m := newTestMenu()
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

func TestMenu_KeyEsc_Cancels(t *testing.T) {
	m := newTestMenu()
	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: 27})) // esc
	if _, ok := extractMsg(cmd).(common.FilterCancelMsg); !ok {
		t.Error("esc should send FilterCancelMsg")
	}
}

func TestMenu_KeyEnter_SelectsCategory(t *testing.T) {
	m := newTestMenu()
	// First category is "top_level_only"
	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: 13})) // enter
	if _, ok := extractMsg(cmd).(common.FilterToggleTopLevelMsg); !ok {
		t.Errorf("enter on top_level_only should send FilterToggleTopLevelMsg, got %T", extractMsg(cmd))
	}
}

func TestMenu_KeyEnter_StatusCategory(t *testing.T) {
	m := newTestMenu()
	// Navigate to status (index 2)
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: -2, Text: "j"}))
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: -2, Text: "j"}))
	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: 13}))
	msg, ok := extractMsg(cmd).(common.FilterMenuSelectMsg)
	if !ok {
		t.Fatalf("enter on status should send FilterMenuSelectMsg, got %T", extractMsg(cmd))
	}
	if msg.Field != "status" {
		t.Errorf("expected field 'status', got %q", msg.Field)
	}
}

// --- Menu mouse tests ---

func TestMenu_MouseClick_SelectsItem(t *testing.T) {
	m := newTestMenu()
	// Position the menu at a known location
	m.ScreenX = 10
	m.ScreenY = 5
	// Click on the third option (index 2 = status): Y = ScreenY + 2 (border+padding) + 2
	_, cmd := m.Update(tea.MouseClickMsg{X: 15, Y: 9, Button: tea.MouseLeft})
	msg, ok := extractMsg(cmd).(common.FilterMenuSelectMsg)
	if !ok {
		t.Fatalf("clicking status row should send FilterMenuSelectMsg, got %T", extractMsg(cmd))
	}
	if msg.Field != "status" {
		t.Errorf("expected field 'status', got %q", msg.Field)
	}
}

func TestMenu_MouseClick_ToggleItem(t *testing.T) {
	m := newTestMenu()
	m.ScreenX = 10
	m.ScreenY = 5
	// Click on first option (index 0 = top_level_only): Y = ScreenY + 2
	_, cmd := m.Update(tea.MouseClickMsg{X: 15, Y: 7, Button: tea.MouseLeft})
	if _, ok := extractMsg(cmd).(common.FilterToggleTopLevelMsg); !ok {
		t.Errorf("clicking top_level_only should send FilterToggleTopLevelMsg, got %T", extractMsg(cmd))
	}
}

func TestMenu_MouseClick_OutsideCancels(t *testing.T) {
	m := newTestMenu()
	m.ScreenX = 10
	m.ScreenY = 5
	// Click well outside the box
	_, cmd := m.Update(tea.MouseClickMsg{X: 0, Y: 0, Button: tea.MouseLeft})
	if _, ok := extractMsg(cmd).(common.FilterCancelMsg); !ok {
		t.Error("clicking outside should send FilterCancelMsg")
	}
}

func TestMenu_MouseMotion_MovesCursor(t *testing.T) {
	m := newTestMenu()
	m.ScreenX = 10
	m.ScreenY = 5
	if m.cursor != 0 {
		t.Fatal("cursor should start at 0")
	}
	// Hover over third option (index 2): Y = ScreenY + 2 + 2
	m, _ = m.Update(tea.MouseMotionMsg{X: 15, Y: 9})
	if m.cursor != 2 {
		t.Errorf("mouse motion should move cursor to 2, got %d", m.cursor)
	}
}

func TestMenu_MouseClick_ClearAll(t *testing.T) {
	m := newTestMenuWithFilters()
	m.ScreenX = 10
	m.ScreenY = 5
	// With TopLevelOnly active, "Clear" is shown as the last item (index 6)
	lastIdx := len(m.categories) - 1
	clickY := m.ScreenY + 2 + lastIdx
	_, cmd := m.Update(tea.MouseClickMsg{X: 15, Y: clickY, Button: tea.MouseLeft})
	if _, ok := extractMsg(cmd).(common.ClearAllFiltersMsg); !ok {
		t.Errorf("clicking Clear should send ClearAllFiltersMsg, got %T", extractMsg(cmd))
	}
}
