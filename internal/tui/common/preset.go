package common

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/Mibokess/grapes/internal/config"
	"github.com/lucasb-eyer/go-colorful"
	themes "go.withmatt.com/themes"
)

// CuratedPresets is the list of theme names shown in the settings UI.
// "Auto", "Light", and "Dark" map to the built-in color scheme with the
// corresponding mode; all other entries are external preset names.
var CuratedPresets = []string{
	"Auto",
	"Light",
	"Dark",
	"Catppuccin Mocha",
	"Catppuccin Latte",
	"Dracula",
	"Gruvbox Dark",
	"Gruvbox Light",
	"Kanagawa Wave",
	"Nord",
	"One Half Dark",
	"One Half Light",
	"Rose Pine",
	"Rose Pine Dawn",
	"iTerm2 Solarized Dark",
	"iTerm2 Solarized Light",
	"TokyoNight",
}

// PresetIsDark returns true if the preset's background is dark.
func PresetIsDark(t *themes.Theme) bool {
	return luminance(t.Background) < 0.5
}

// applyPreset populates all Theme color fields from an external ANSI palette.
func applyPreset(t *Theme, ext *themes.Theme) {
	isDark := PresetIsDark(ext)

	// The palette, expressed in the same shape as a user color set so that
	// overrides can be merged on top of it.
	t.applyColorSet(config.ColorSetConfig{
		Text:    ext.Foreground,
		Surface: ext.Background,
		Muted:   ext.BrightBlack,
		Accent:  ext.Magenta,

		// Derived colors (blended).
		Faint:    blendHex(ext.BrightBlack, ext.Background, 0.6),
		Border:   blendHex(ext.BrightBlack, ext.Background, 0.4),
		AccentBg: blendHex(ext.Magenta, ext.Background, 0.85),

		// Status colors.
		ColorBacklog:    ext.BrightBlack,
		ColorTodo:       ext.Blue,
		ColorInProgress: ext.Yellow,
		ColorDone:       ext.Green,
		ColorCancelled:  ext.BrightBlack,

		// Priority colors.
		ColorUrgent: ext.Red,
		ColorHigh:   ext.Yellow,
		ColorMedium: ext.Blue,
		ColorLow:    ext.BrightBlack,
	})

	// Contrast for pill text.
	if isDark {
		t.ColorContrast = hexToColor(blendHex(ext.Background, "#000000", 0.3))
	} else {
		t.ColorContrast = hexToColor(blendHex(ext.Background, "#ffffff", 0.3))
	}

	t.ColorWorktree = hexToColor(fallback(ext.BrightMagenta, ext.Magenta))

	// Pill backgrounds.
	t.PillBgBacklog = hexToColor(blendHex(ext.BrightBlack, ext.Background, 0.5))
	t.PillBgCancelled = hexToColor(blendHex(ext.BrightBlack, ext.Background, 0.7))

	// Label palette (10 fg/bg pairs).
	labelFgs := []string{
		ext.Magenta, ext.Blue, ext.Green, ext.Yellow, ext.Red,
		fallback(ext.BrightMagenta, ext.Magenta),
		ext.Cyan,
		fallback(ext.BrightGreen, ext.Green),
		fallback(ext.BrightBlue, ext.Blue),
		fallback(ext.BrightCyan, ext.Cyan),
	}
	t.LabelColors = make([]LabelColor, len(labelFgs))
	for i, fg := range labelFgs {
		t.LabelColors[i] = LabelColor{
			Fg: hexToColor(fg),
			Bg: hexToColor(blendHex(fg, ext.Background, 0.85)),
		}
	}

	// Worktree palette (8 colors).
	wtHexes := []string{
		fallback(ext.BrightRed, ext.Red),
		fallback(ext.BrightBlue, ext.Blue),
		fallback(ext.BrightGreen, ext.Green),
		fallback(ext.BrightMagenta, ext.Magenta),
		fallback(ext.BrightCyan, ext.Cyan),
		ext.Blue,
		ext.Yellow,
		ext.Green,
	}
	t.WorktreeColors = make([]color.Color, len(wtHexes))
	for i, hex := range wtHexes {
		t.WorktreeColors[i] = hexToColor(hex)
	}

	// Glamour markdown style.
	if isDark {
		t.GlamourStyle = "dark"
	} else {
		t.GlamourStyle = "light"
	}
}

// hexToColor converts a hex string to a lipgloss-compatible color.Color.
func hexToColor(hex string) color.Color {
	return lipgloss.Color(hex)
}

// luminance returns the relative luminance (0=black, 1=white) of a hex color.
func luminance(hex string) float64 {
	c, err := colorful.Hex(hex)
	if err != nil {
		return 0
	}
	r, g, b := c.LinearRgb()
	return 0.2126*r + 0.7152*g + 0.0722*b
}

// blendHex blends two hex colors in Lab space. t=0 returns a, t=1 returns b.
func blendHex(a, b string, t float64) string {
	ca, err1 := colorful.Hex(a)
	cb, err2 := colorful.Hex(b)
	if err1 != nil {
		return b
	}
	if err2 != nil {
		return a
	}
	return ca.BlendLab(cb, t).Hex()
}

// fallback returns a if non-empty, otherwise b.
func fallback(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
