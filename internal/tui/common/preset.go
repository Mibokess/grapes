package common

import (
	"image/color"

	"github.com/lucasb-eyer/go-colorful"
	"charm.land/lipgloss/v2"
	themes "go.withmatt.com/themes"
)

// CuratedPresets is the list of popular theme names shown in the settings UI.
var CuratedPresets = []string{
	"default",
	"Catppuccin Mocha",
	"Catppuccin Latte",
	"Dracula",
	"Gruvbox Dark",
	"Gruvbox Light",
	"Kanagawa",
	"Nord",
	"One Dark",
	"One Light",
	"Rosé Pine",
	"Rosé Pine Dawn",
	"Solarized Dark",
	"Solarized Light",
	"Tokyo Night",
}

// PresetIsDark returns true if the preset's background is dark.
func PresetIsDark(t *themes.Theme) bool {
	return luminance(t.Background) < 0.5
}

// applyPreset populates all Theme color fields from an external ANSI palette.
func applyPreset(t *Theme, ext *themes.Theme) {
	isDark := PresetIsDark(ext)

	// Core colors.
	t.ColorText = hexToColor(ext.Foreground)
	t.ColorSurface = hexToColor(ext.Background)
	t.ColorMuted = hexToColor(ext.BrightBlack)
	t.ColorAccent = hexToColor(ext.Magenta)

	// Derived colors (blended).
	t.ColorFaint = hexToColor(blendHex(ext.BrightBlack, ext.Background, 0.6))
	t.ColorBorder = hexToColor(blendHex(ext.BrightBlack, ext.Background, 0.4))
	t.ColorAccentBg = hexToColor(blendHex(ext.Magenta, ext.Background, 0.85))

	// Contrast for pill text.
	if isDark {
		t.ColorContrast = hexToColor(blendHex(ext.Background, "#000000", 0.3))
	} else {
		t.ColorContrast = hexToColor(blendHex(ext.Background, "#ffffff", 0.3))
	}

	// Status colors.
	t.ColorBacklog = hexToColor(ext.BrightBlack)
	t.ColorTodo = hexToColor(ext.Blue)
	t.ColorInProgress = hexToColor(ext.Yellow)
	t.ColorDone = hexToColor(ext.Green)
	t.ColorCancelled = hexToColor(ext.BrightBlack)

	// Priority colors.
	t.ColorUrgent = hexToColor(ext.Red)
	t.ColorHigh = hexToColor(ext.Yellow)
	t.ColorMedium = hexToColor(ext.Blue)
	t.ColorLow = hexToColor(ext.BrightBlack)

	// Error & worktree.
	t.ColorError = hexToColor(ext.Red)
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
