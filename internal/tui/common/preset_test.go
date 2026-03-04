package common

import (
	"math"
	"testing"

	"github.com/Mibokess/grapes/internal/config"
	themes "go.withmatt.com/themes"
)

func TestLuminance(t *testing.T) {
	tests := []struct {
		hex  string
		want float64
		tol  float64
	}{
		{"#000000", 0.0, 0.01},
		{"#ffffff", 1.0, 0.01},
		{"#808080", 0.22, 0.05}, // mid-gray, ~0.22 relative luminance
	}
	for _, tt := range tests {
		got := luminance(tt.hex)
		if math.Abs(got-tt.want) > tt.tol {
			t.Errorf("luminance(%q) = %f, want %f (±%f)", tt.hex, got, tt.want, tt.tol)
		}
	}
}

func TestLuminance_InvalidHex(t *testing.T) {
	got := luminance("not-a-color")
	if got != 0 {
		t.Errorf("luminance of invalid hex should be 0, got %f", got)
	}
}

func TestBlendHex(t *testing.T) {
	// t=0 returns a
	if got := blendHex("#ff0000", "#0000ff", 0); got != "#ff0000" {
		t.Errorf("blendHex at t=0 should return a, got %s", got)
	}
	// t=1 returns b
	if got := blendHex("#ff0000", "#0000ff", 1); got != "#0000ff" {
		t.Errorf("blendHex at t=1 should return b, got %s", got)
	}
	// midpoint should produce something in between
	mid := blendHex("#000000", "#ffffff", 0.5)
	lum := luminance(mid)
	if lum < 0.1 || lum > 0.9 {
		t.Errorf("blendHex midpoint luminance should be between 0.1 and 0.9, got %f (color: %s)", lum, mid)
	}
}

func TestPresetIsDark(t *testing.T) {
	dark := &themes.Theme{Background: "#282a36"} // Dracula-like
	if !PresetIsDark(dark) {
		t.Error("expected dark background to be detected as dark")
	}
	light := &themes.Theme{Background: "#fafafa"}
	if PresetIsDark(light) {
		t.Error("expected light background to be detected as light")
	}
}

func TestApplyPreset_SetsAllFields(t *testing.T) {
	ext, err := themes.GetTheme("Dracula")
	if err != nil {
		t.Fatalf("failed to get Dracula theme: %v", err)
	}
	var theme Theme
	applyPreset(&theme, ext)

	if theme.ColorText == nil {
		t.Error("ColorText should be set")
	}
	if theme.ColorSurface == nil {
		t.Error("ColorSurface should be set")
	}
	if theme.ColorAccent == nil {
		t.Error("ColorAccent should be set")
	}
	if theme.ColorMuted == nil {
		t.Error("ColorMuted should be set")
	}
	if theme.ColorDone == nil {
		t.Error("ColorDone should be set")
	}
	if theme.ColorUrgent == nil {
		t.Error("ColorUrgent should be set")
	}
	if len(theme.LabelColors) != 10 {
		t.Errorf("expected 10 label colors, got %d", len(theme.LabelColors))
	}
	if len(theme.WorktreeColors) != 8 {
		t.Errorf("expected 8 worktree colors, got %d", len(theme.WorktreeColors))
	}
	if theme.GlamourStyle != "dark" {
		t.Errorf("Dracula should produce dark glamour style, got %q", theme.GlamourStyle)
	}
}

func TestNewThemeFromConfig_Preset(t *testing.T) {
	cfg := config.ThemeConfig{Preset: "Dracula"}
	preset := NewThemeFromConfig(cfg, true)
	dflt := NewThemeFromConfig(config.ThemeConfig{}, true)

	// The preset theme should differ from the default.
	if preset.GlamourStyle != dflt.GlamourStyle {
		// Both are dark, so this is fine — but colors should differ.
	}
	if preset.ColorAccent == dflt.ColorAccent {
		t.Error("preset theme accent should differ from default")
	}
}

func TestNewThemeFromConfig_UnknownPreset(t *testing.T) {
	cfg := config.ThemeConfig{Preset: "definitely_not_a_real_theme_xyz"}
	theme := NewThemeFromConfig(cfg, true)
	dflt := NewThemeFromConfig(config.ThemeConfig{}, true)

	// Should fall back to default.
	if theme.GlamourStyle != dflt.GlamourStyle {
		t.Error("unknown preset should fall back to default glamour style")
	}
}

func TestNewThemeFromConfig_DefaultPreset(t *testing.T) {
	empty := NewThemeFromConfig(config.ThemeConfig{}, true)
	explicit := NewThemeFromConfig(config.ThemeConfig{Preset: "default"}, true)

	if empty.GlamourStyle != explicit.GlamourStyle {
		t.Error("empty and 'default' preset should produce the same theme")
	}
}

func TestNewThemeFromConfig_PresetWithModeOverride(t *testing.T) {
	// Dracula is dark, but force light mode.
	cfg := config.ThemeConfig{Preset: "Dracula", Mode: "light"}
	theme := NewThemeFromConfig(cfg, true)

	if theme.GlamourStyle != "light" {
		t.Errorf("mode override should force light, got %q", theme.GlamourStyle)
	}
}

func TestFallback(t *testing.T) {
	if got := fallback("a", "b"); got != "a" {
		t.Errorf("fallback should return a, got %s", got)
	}
	if got := fallback("", "b"); got != "b" {
		t.Errorf("fallback should return b when a is empty, got %s", got)
	}
}
