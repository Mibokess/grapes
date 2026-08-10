package common

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/Mibokess/grapes/internal/config"
)

func TestNewThemeFromConfig_BuiltinDefaults(t *testing.T) {
	cfg := config.Defaults().Theme

	dark := NewThemeFromConfig(cfg, true)
	if dark.Colors != BuiltinDark {
		t.Errorf("dark palette = %+v, want the built-in dark palette", dark.Colors)
	}
	light := NewThemeFromConfig(cfg, false)
	if light.Colors != BuiltinLight {
		t.Errorf("light palette = %+v, want the built-in light palette", light.Colors)
	}
}

func TestNewThemeFromConfig_OverridesBuiltin(t *testing.T) {
	cfg := config.Defaults().Theme
	cfg.Accent = "#123456"

	got := NewThemeFromConfig(cfg, true)
	if got.Colors.Accent != "#123456" {
		t.Errorf("accent = %q, want the override", got.Colors.Accent)
	}
	if got.ColorAccent != lipgloss.Color("#123456") {
		t.Error("the override did not reach the rendered color")
	}
	if got.Colors.Text != BuiltinDark.Text {
		t.Errorf("unset colors should stay at the built-in value, got %q", got.Colors.Text)
	}
}

// Color overrides used to be dropped entirely when a preset was selected: the
// settings screen still showed and edited them, and nothing happened.
func TestNewThemeFromConfig_OverridesApplyOnTopOfPreset(t *testing.T) {
	cfg := config.Defaults().Theme
	cfg.Preset = "Dracula"

	base := NewThemeFromConfig(cfg, true)
	if base.Colors.Accent == "" {
		t.Fatal("preset did not populate the palette")
	}

	cfg.Accent = "#00ff00"
	got := NewThemeFromConfig(cfg, true)

	if got.Colors.Accent != "#00ff00" {
		t.Errorf("accent = %q, want the override to win over the preset", got.Colors.Accent)
	}
	if got.Colors.Text != base.Colors.Text {
		t.Errorf("unset colors should still come from the preset: %q != %q", got.Colors.Text, base.Colors.Text)
	}
}

func TestNewThemeFromConfig_UnknownPresetFallsBackToBuiltin(t *testing.T) {
	cfg := config.Defaults().Theme
	cfg.Preset = "No Such Theme"

	got := NewThemeFromConfig(cfg, true)
	if got.Colors != BuiltinDark {
		t.Errorf("unknown preset should fall back to the built-in palette, got %+v", got.Colors)
	}
}

func TestColorRegistry_RoundTrip(t *testing.T) {
	var set config.ColorSetConfig
	for i, key := range ColorKeys {
		if !IsColorKey(key) {
			t.Errorf("%q is listed in ColorKeys but not recognised", key)
		}
		if ColorLabels[key] == "" {
			t.Errorf("%q has no settings label", key)
		}
		val := string(rune('a'+i)) + "-value"
		SetColor(&set, key, val)
		if got := GetColor(set, key); got != val {
			t.Errorf("GetColor(%q) = %q, want %q", key, got, val)
		}
	}
	if IsColorKey("not_a_color") {
		t.Error("unknown keys must not be treated as colors")
	}
}
