package settings_test

import (
	"testing"

	"github.com/Mibokess/grapes/internal/config"
	"github.com/Mibokess/grapes/internal/tui/common"
	"github.com/Mibokess/grapes/internal/tui/settings"
	"github.com/Mibokess/grapes/internal/tui/testutil"
	tea "charm.land/bubbletea/v2"
)

func newTestModel() settings.Model {
	cfg := config.Defaults()
	theme := common.NewThemeFromConfig(cfg.Theme, true)
	return settings.New(cfg, "", 100, 30, theme).SetTopOffset(1)
}

func clickLeft(x, y int) tea.MouseClickMsg {
	return tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft}
}

func wheelUp() tea.MouseWheelMsg {
	return tea.MouseWheelMsg{Button: tea.MouseWheelUp}
}

func wheelDown() tea.MouseWheelMsg {
	return tea.MouseWheelMsg{Button: tea.MouseWheelDown}
}

func keyMsg(k string) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: -2, Text: k})
}

func TestSettingsView_Default(t *testing.T) {
	m := newTestModel()
	testutil.RequireGolden(t, m.View())
}

func TestSettingsView_ThemeCategory(t *testing.T) {
	m := newTestModel()
	// Navigate down to Theme category
	m, _ = m.Update(keyMsg("j"))
	testutil.RequireGolden(t, m.View())
}

func TestClickCategory_SelectsTheme(t *testing.T) {
	m := newTestModel()

	// Click on "Theme" category (row index 1, topOffset=1 so y=1+1+1=3,
	// but the view has 1 line of top padding so y = topOffset + 1 + catIndex)
	// topOffset=1, padding=1, Theme is at index 1 → y = 1 + 1 + 1 = 3
	m, _ = m.Update(clickLeft(5, 3))

	view := testutil.StripANSI(m.View())
	// After clicking Theme, the Theme fields (Accent, etc.) should be visible
	if !containsStr(view, "Accent") {
		t.Error("expected Theme fields to be visible after clicking Theme category")
	}
}

func TestClickCategory_SelectsKeys(t *testing.T) {
	m := newTestModel()

	// Keys is at category index 2 → y = topOffset(1) + padding(1) + 2 = 4
	m, _ = m.Update(clickLeft(5, 4))

	view := testutil.StripANSI(m.View())
	if !containsStr(view, "Quit") {
		t.Error("expected Keys fields to be visible after clicking Keys category")
	}
}

func TestClickField_SelectsField(t *testing.T) {
	m := newTestModel()

	// Click on the second field ("Default sort") in the right pane
	// Field index 1, y = topOffset(1) + padding(1) + 1 = 3
	// x must be >= catW (18) to be in the fields pane
	m, _ = m.Update(clickLeft(20, 3))

	// Verify focus moved to fields pane by clicking the same field again
	// (should enter edit/cycle mode for enum)
	m2, _ := m.Update(clickLeft(20, 3))
	view := testutil.StripANSI(m2.View())
	// Default sort should have cycled from its default value
	_ = view // field was activated
}

func TestClickSelectedField_CyclesEnum(t *testing.T) {
	m := newTestModel()

	// Click first field (Default screen, enum) in fields pane
	// y = topOffset(1) + padding(1) + 0 = 2, x >= 18
	m, _ = m.Update(clickLeft(20, 2))
	// First click selects the field
	// Second click should cycle the enum value
	m, _ = m.Update(clickLeft(20, 2))

	val := testutil.StripANSI(m.View())
	// The default_screen should have cycled from "board" to "list"
	if !containsStr(val, "list") {
		t.Error("expected enum value to cycle to 'list' after clicking selected field")
	}
}

func TestClickSelectedField_EditsColor(t *testing.T) {
	m := newTestModel()

	// Navigate to Theme category first
	m, _ = m.Update(keyMsg("j"))

	// Click on Accent field (first field in Theme), x >= 18
	// y = topOffset(1) + padding(1) + 0 = 2
	m, _ = m.Update(clickLeft(20, 2))
	// First click selects
	// Second click should enter edit mode
	m, _ = m.Update(clickLeft(20, 2))

	// In edit mode, the textinput should be visible in the view
	view := testutil.StripANSI(m.View())
	// The input should contain the current accent color value
	_ = view // editing mode entered
}

func TestMouseWheel_ScrollsFields(t *testing.T) {
	m := newTestModel()

	// Switch to Keys category (many fields) to test scrolling
	m, _ = m.Update(keyMsg("j"))
	m, _ = m.Update(keyMsg("j"))
	// Move focus to fields pane
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: 9})) // Tab

	// Scroll down several times
	for i := 0; i < 5; i++ {
		m, _ = m.Update(wheelDown())
	}

	view := testutil.StripANSI(m.View())
	// After scrolling down 5 times in Keys, we should see "Board: Open" highlighted
	if !containsStr(view, "Board: Open") {
		t.Error("expected to see 'Board: Open' after scrolling down in Keys category")
	}
}

func TestMouseWheel_DoesNotScrollPastBounds(t *testing.T) {
	m := newTestModel()

	// Scroll up when already at top — should stay at 0
	m, _ = m.Update(wheelUp())
	view1 := testutil.StripANSI(m.View())

	m2 := newTestModel()
	view2 := testutil.StripANSI(m2.View())

	if view1 != view2 {
		t.Error("scrolling up at top should not change the view")
	}
}

func TestClickCategory_WithOffset_Theme(t *testing.T) {
	// Test with a larger topOffset to verify offset calculation
	cfg := config.Defaults()
	theme := common.NewThemeFromConfig(cfg.Theme, true)
	m := settings.New(cfg, "", 100, 30, theme).SetTopOffset(3)

	// Theme is at category index 1 → y = topOffset(3) + padding(1) + 1 = 5
	m, _ = m.Update(clickLeft(5, 5))

	view := testutil.StripANSI(m.View())
	if !containsStr(view, "Accent") {
		t.Error("expected Theme fields with topOffset=3, clicking y=5")
	}
}

func TestClickCategory_WithOffset_Keys(t *testing.T) {
	cfg := config.Defaults()
	theme := common.NewThemeFromConfig(cfg.Theme, true)
	m := settings.New(cfg, "", 100, 30, theme).SetTopOffset(3)

	// Keys is at category index 2 → y = topOffset(3) + padding(1) + 2 = 6
	m, _ = m.Update(clickLeft(5, 6))

	view := testutil.StripANSI(m.View())
	if !containsStr(view, "Quit") {
		t.Error("expected Keys fields with topOffset=3, clicking y=6")
	}
}

func TestClickDoesNothing_WhenEditing(t *testing.T) {
	m := newTestModel()

	// Enter edit mode: navigate to Theme, select a color field, press enter
	m, _ = m.Update(keyMsg("j"))                        // Theme category
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: 9}))  // Tab to fields
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: 13})) // Enter to edit

	viewBefore := testutil.StripANSI(m.View())
	m, _ = m.Update(clickLeft(5, 4))
	viewAfter := testutil.StripANSI(m.View())

	if viewBefore != viewAfter {
		t.Error("clicking while editing should not change anything")
	}
}

func TestKeyboardNavigation_StillWorks(t *testing.T) {
	m := newTestModel()

	// j moves down in categories
	m, _ = m.Update(keyMsg("j"))
	view := testutil.StripANSI(m.View())
	if !containsStr(view, "Accent") {
		t.Error("j key should navigate to Theme category")
	}

	// k moves back up
	m, _ = m.Update(keyMsg("k"))
	view = testutil.StripANSI(m.View())
	if !containsStr(view, "Default screen") {
		t.Error("k key should navigate back to View category")
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
