package settings_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Mibokess/grapes/internal/config"
	"github.com/Mibokess/grapes/internal/tui/common"
	"github.com/Mibokess/grapes/internal/tui/settings"
	"github.com/Mibokess/grapes/internal/tui/testutil"
)

// selectAccent moves focus to the Theme category and lands on the Accent row,
// which is the first color field after the preset row.
func selectAccent(t *testing.T, m settings.Model) settings.Model {
	t.Helper()
	m, _ = m.Update(keyMsg("j"))   // View → Theme
	m, _ = m.Update(keyMsg("tab")) // categories → fields
	m, _ = m.Update(keyMsg("j"))   // preset row → Accent
	return m
}

func typeText(m settings.Model, s string) settings.Model {
	for _, r := range s {
		m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: r, Text: string(r)}))
	}
	return m
}

// Editing a color while a preset is active used to do nothing at all: the
// preset short-circuited every override.
func TestSettings_ColorOverrideAppliesUnderPreset(t *testing.T) {
	cfg := config.Defaults()
	cfg.Theme.Preset = "Dracula"
	theme := common.NewThemeFromConfig(cfg.Theme, true)

	m := settings.New(cfg, t.TempDir(), 120, 40, theme)
	m = selectAccent(t, m)

	// Enter edit mode, replace the value, confirm.
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	for i := 0; i < 10; i++ {
		m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyBackspace}))
	}
	m = typeText(m, "#00ff00")
	m, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))

	if cmd == nil {
		t.Fatal("confirming a color edit should emit a theme update")
	}
	msg, ok := cmd().(common.ThemeMsg)
	if !ok {
		t.Fatalf("expected a ThemeMsg, got %T", cmd())
	}
	if msg.Theme.Colors.Accent != "#00ff00" {
		t.Errorf("accent = %q, want the override to win over the preset", msg.Theme.Colors.Accent)
	}
	if !strings.Contains(testutil.StripANSI(m.View()), "#00ff00") {
		t.Error("the settings row should show the new value")
	}
}

// An invalid hex value is rejected with a message rather than stored.
func TestSettings_ColorRejectsInvalidHex(t *testing.T) {
	cfg := config.Defaults()
	m := settings.New(cfg, t.TempDir(), 120, 40, common.NewTheme(true)).SetDark(true)
	m = selectAccent(t, m)

	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	for i := 0; i < 10; i++ {
		m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyBackspace}))
	}
	m = typeText(m, "green")
	m, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))

	if cmd != nil {
		t.Error("an invalid color should not trigger a theme update")
	}
	if !strings.Contains(testutil.StripANSI(m.View()), "Invalid hex color") {
		t.Error("an invalid color should be reported in the status line")
	}
}

// A fresh config has no overrides, so the reset row only appears once the user
// has actually changed something.
func TestSettings_ResetRowTracksOverrides(t *testing.T) {
	cfg := config.Defaults()
	// SetDark mirrors the app, which resolves the mode from the terminal
	// background before the screen is shown.
	m := settings.New(cfg, t.TempDir(), 120, 40, common.NewTheme(true)).SetDark(true)
	m, _ = m.Update(keyMsg("j")) // View → Theme

	if strings.Contains(testutil.StripANSI(m.View()), "Reset colors") {
		t.Error("a default config should not offer a colors reset")
	}

	cfg.Theme.Accent = "#123456"
	m = settings.New(cfg, t.TempDir(), 120, 40, common.NewThemeFromConfig(cfg.Theme, true)).SetDark(true)
	m, _ = m.Update(keyMsg("j"))

	if !strings.Contains(testutil.StripANSI(m.View()), "Reset colors") {
		t.Error("an overridden color should offer a reset")
	}
}
