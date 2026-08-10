package common

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/Mibokess/grapes/internal/config"
	"github.com/Mibokess/grapes/internal/data"
	themes "go.withmatt.com/themes"
)

// AppHeaderHeight is the number of terminal lines occupied by the app header.
const AppHeaderHeight = 2

// Status icons.
const (
	IconBacklog    = "○"
	IconTodo       = "◌"
	IconInProgress = "◑"
	IconDone       = "●"
	IconCancelled  = "×"
)

// Priority icons — always 2 visible chars wide for alignment.
const (
	IconUrgent = "!!"
	IconHigh   = " !"
	IconMedium = " ·"
	IconLow    = "  "
)

// StatusIcon returns the icon character for a given status.
func StatusIcon(s data.Status) string {
	switch s {
	case data.StatusBacklog:
		return IconBacklog
	case data.StatusTodo:
		return IconTodo
	case data.StatusInProgress:
		return IconInProgress
	case data.StatusDone:
		return IconDone
	case data.StatusCancelled:
		return IconCancelled
	default:
		return "?"
	}
}

// PriorityIcon returns the 2-char icon for a given priority.
func PriorityIcon(p data.Priority) string {
	switch p {
	case data.PriorityUrgent:
		return IconUrgent
	case data.PriorityHigh:
		return IconHigh
	case data.PriorityMedium:
		return IconMedium
	default:
		return IconLow
	}
}

// WorktreeIcon returns the icon for worktree issues.
func WorktreeIcon() string { return "⑂" }

// MainIcon returns the icon for main source.
func MainIcon() string { return "◆" }

// worktreeColorsDark is a fixed palette for worktree source indicators (dark theme).
var worktreeColorsDark = []color.Color{
	lipgloss.Color("#f0883e"), // orange
	lipgloss.Color("#58a6ff"), // blue
	lipgloss.Color("#3fb950"), // green
	lipgloss.Color("#d2a8ff"), // lavender
	lipgloss.Color("#f692ce"), // pink
	lipgloss.Color("#79c0ff"), // light blue
	lipgloss.Color("#ffa657"), // amber
	lipgloss.Color("#7ee787"), // light green
}

// worktreeColorsLight is a fixed palette for worktree source indicators (light theme).
var worktreeColorsLight = []color.Color{
	lipgloss.Color("#bc4c00"), // orange
	lipgloss.Color("#0550ae"), // blue
	lipgloss.Color("#116329"), // green
	lipgloss.Color("#6639ba"), // lavender
	lipgloss.Color("#bf3989"), // pink
	lipgloss.Color("#0969da"), // light blue
	lipgloss.Color("#953800"), // amber
	lipgloss.Color("#1a7f37"), // light green
}

// BuiltinDark is the default dark palette. It is the single source of truth for
// the built-in colors: config holds only the user's overrides.
var BuiltinDark = config.ColorSetConfig{
	Accent:          "#a371f7",
	AccentBg:        "#2d1b69",
	Border:          "#30363d",
	Text:            "#e6edf3",
	Muted:           "#8b949e",
	Faint:           "#484f58",
	Surface:         "#161b22",
	ColorBacklog:    "#8b949e",
	ColorTodo:       "#388bfd",
	ColorInProgress: "#d29922",
	ColorDone:       "#3fb950",
	ColorCancelled:  "#6e7681",
	ColorUrgent:     "#f85149",
	ColorHigh:       "#d29922",
	ColorMedium:     "#388bfd",
	ColorLow:        "#6e7681",
}

// BuiltinLight is the default light palette.
var BuiltinLight = config.ColorSetConfig{
	Accent:          "#8250df",
	AccentBg:        "#eddeff",
	Border:          "#d0d7de",
	Text:            "#1f2328",
	Muted:           "#656d76",
	Faint:           "#afb8c1",
	Surface:         "#f6f8fa",
	ColorBacklog:    "#656d76",
	ColorTodo:       "#0969da",
	ColorInProgress: "#9a6700",
	ColorDone:       "#1a7f37",
	ColorCancelled:  "#8c959f",
	ColorUrgent:     "#cf222e",
	ColorHigh:       "#9a6700",
	ColorMedium:     "#0969da",
	ColorLow:        "#8c959f",
}

// BuiltinColors returns the built-in palette for the given mode.
func BuiltinColors(isDark bool) config.ColorSetConfig {
	if isDark {
		return BuiltinDark
	}
	return BuiltinLight
}

// ThemeMsg is sent when the terminal background is detected and the theme changes.
type ThemeMsg struct{ Theme Theme }

// LabelColor holds a foreground/background pair for label rendering.
type LabelColor struct{ Fg, Bg color.Color }

// Theme holds all colors and pre-built styles for the TUI.
type Theme struct {
	// Colors is the resolved palette behind the color fields below: the active
	// base (built-in or preset) with the user's overrides applied. The settings
	// screen reads it to show the hex value each field actually renders with.
	Colors config.ColorSetConfig

	// Raw colors — available for dynamic style construction.
	ColorText     color.Color
	ColorMuted    color.Color
	ColorFaint    color.Color
	ColorBorder   color.Color
	ColorSurface  color.Color
	ColorAccent   color.Color
	ColorAccentBg color.Color
	ColorContrast color.Color // high-contrast text for colored pill backgrounds
	ColorError    color.Color
	ColorWorktree color.Color

	// Priority colors.
	ColorUrgent color.Color
	ColorHigh   color.Color
	ColorMedium color.Color
	ColorLow    color.Color

	// Status colors.
	ColorBacklog    color.Color
	ColorTodo       color.Color
	ColorInProgress color.Color
	ColorDone       color.Color
	ColorCancelled  color.Color

	// Status pill backgrounds (detail view).
	PillBgBacklog   color.Color
	PillBgCancelled color.Color

	// Label palette (10 fg/bg pairs).
	LabelColors []LabelColor

	// Pre-built styles.
	StyleAppTitle      lipgloss.Style
	StyleTabActive     lipgloss.Style
	StyleTabInactive   lipgloss.Style
	StyleSeparator     lipgloss.Style
	StyleStatusBar     lipgloss.Style
	StyleTitle         lipgloss.Style
	StyleSubtitle      lipgloss.Style
	StyleFaint         lipgloss.Style
	StyleSectionHeader lipgloss.Style
	StyleLabel         lipgloss.Style
	StyleLabelPill     lipgloss.Style
	StyleCard          lipgloss.Style
	StyleActiveCard    lipgloss.Style
	StyleColumnHeader  lipgloss.Style
	StyleStatusKey     lipgloss.Style
	StyleStatusSep     lipgloss.Style
	StyleCommentBox    lipgloss.Style
	StyleMetaBox       lipgloss.Style
	StyleDragCard      lipgloss.Style
	StyleDropTarget    lipgloss.Style
	StyleWorktreeLabel lipgloss.Style
	StyleWorktreeBadge lipgloss.Style

	// Worktree color palette for multi-source indicators.
	WorktreeColors []color.Color

	// Glamour markdown style name ("dark" or "light").
	GlamourStyle string
}

// NewTheme creates a theme appropriate for the terminal background.
func NewTheme(isDark bool) Theme {
	var t Theme
	t.setBuiltin(isDark)
	t.buildStyles()
	return t
}

// NewThemeFromConfig creates a theme for the resolved mode: an external preset
// or the built-in palette as the base, with the user's color overrides applied
// on top of whichever base was chosen.
func NewThemeFromConfig(cfg config.ThemeConfig, termIsDark bool) Theme {
	var t Theme
	isDark := cfg.EffectiveIsDark(termIsDark)

	ext, usePreset := resolvePreset(cfg.Preset)
	if usePreset {
		isDark = PresetIsDark(ext)
		switch cfg.Mode {
		case "light":
			isDark = false
		case "dark":
			isDark = true
		}
		applyPreset(&t, ext)
		// Override glamour style if mode was explicitly set.
		if cfg.Mode == "light" || cfg.Mode == "dark" {
			if isDark {
				t.GlamourStyle = "dark"
			} else {
				t.GlamourStyle = "light"
			}
		}
	} else {
		t.setBuiltin(isDark)
	}

	t.applyColorSet(mergeColorSet(t.Colors, cfg.ColorsFor(isDark)))
	t.buildStyles()
	return t
}

// resolvePreset looks up an external theme by name. An empty name, "default",
// or an unknown name means "use the built-in palette".
func resolvePreset(name string) (*themes.Theme, bool) {
	if name == "" || name == "default" {
		return nil, false
	}
	ext, err := themes.GetTheme(name)
	if err != nil {
		return nil, false
	}
	return ext, true
}

// mergeColorSet returns base with every non-empty field of override applied.
func mergeColorSet(base, override config.ColorSetConfig) config.ColorSetConfig {
	for _, key := range ColorKeys {
		if v := GetColor(override, key); v != "" {
			SetColor(&base, key, v)
		}
	}
	return base
}

// applyColorSet assigns the palette colors from a resolved color set and
// records it as the theme's effective palette.
func (t *Theme) applyColorSet(c config.ColorSetConfig) {
	t.Colors = c
	t.ColorText = lipgloss.Color(c.Text)
	t.ColorMuted = lipgloss.Color(c.Muted)
	t.ColorFaint = lipgloss.Color(c.Faint)
	t.ColorBorder = lipgloss.Color(c.Border)
	t.ColorSurface = lipgloss.Color(c.Surface)
	t.ColorAccent = lipgloss.Color(c.Accent)
	t.ColorAccentBg = lipgloss.Color(c.AccentBg)
	t.ColorUrgent = lipgloss.Color(c.ColorUrgent)
	t.ColorHigh = lipgloss.Color(c.ColorHigh)
	t.ColorMedium = lipgloss.Color(c.ColorMedium)
	t.ColorLow = lipgloss.Color(c.ColorLow)
	t.ColorBacklog = lipgloss.Color(c.ColorBacklog)
	t.ColorTodo = lipgloss.Color(c.ColorTodo)
	t.ColorInProgress = lipgloss.Color(c.ColorInProgress)
	t.ColorDone = lipgloss.Color(c.ColorDone)
	t.ColorCancelled = lipgloss.Color(c.ColorCancelled)
	// Errors are shown in the urgent color, in every palette.
	t.ColorError = t.ColorUrgent
}

// setBuiltin loads the built-in palette and the colors derived from it.
func (t *Theme) setBuiltin(isDark bool) {
	if isDark {
		t.setDarkColors()
	} else {
		t.setLightColors()
	}
}

func (t *Theme) setDarkColors() {
	t.applyColorSet(BuiltinDark)
	t.ColorContrast = lipgloss.Color("#0d1117")
	t.ColorWorktree = lipgloss.Color("#f0883e")

	t.PillBgBacklog = lipgloss.Color("#3d4148")
	t.PillBgCancelled = lipgloss.Color("#21262d")

	t.LabelColors = []LabelColor{
		{lipgloss.Color("#a371f7"), lipgloss.Color("#2d1b69")}, // purple
		{lipgloss.Color("#58a6ff"), lipgloss.Color("#0d2240")}, // blue
		{lipgloss.Color("#3fb950"), lipgloss.Color("#0f2d1a")}, // green
		{lipgloss.Color("#d29922"), lipgloss.Color("#2d2006")}, // yellow
		{lipgloss.Color("#f78166"), lipgloss.Color("#2d1710")}, // orange
		{lipgloss.Color("#f692ce"), lipgloss.Color("#2d1226")}, // pink
		{lipgloss.Color("#79c0ff"), lipgloss.Color("#0d2240")}, // light blue
		{lipgloss.Color("#7ee787"), lipgloss.Color("#0f2d1a")}, // light green
		{lipgloss.Color("#d2a8ff"), lipgloss.Color("#2d1b69")}, // lavender
		{lipgloss.Color("#ffa657"), lipgloss.Color("#2d1c0a")}, // amber
	}

	t.WorktreeColors = worktreeColorsDark

	t.GlamourStyle = "dark"
}

func (t *Theme) setLightColors() {
	t.applyColorSet(BuiltinLight)
	t.ColorContrast = lipgloss.Color("#ffffff")
	t.ColorWorktree = lipgloss.Color("#bc4c00")

	t.PillBgBacklog = lipgloss.Color("#d0d7de")
	t.PillBgCancelled = lipgloss.Color("#eaeef2")

	t.LabelColors = []LabelColor{
		{lipgloss.Color("#8250df"), lipgloss.Color("#eddeff")}, // purple
		{lipgloss.Color("#0969da"), lipgloss.Color("#ddf4ff")}, // blue
		{lipgloss.Color("#1a7f37"), lipgloss.Color("#dafbe1")}, // green
		{lipgloss.Color("#9a6700"), lipgloss.Color("#fff8c5")}, // yellow
		{lipgloss.Color("#bc4c00"), lipgloss.Color("#fff1e5")}, // orange
		{lipgloss.Color("#bf3989"), lipgloss.Color("#ffeff7")}, // pink
		{lipgloss.Color("#0550ae"), lipgloss.Color("#ddf4ff")}, // light blue
		{lipgloss.Color("#116329"), lipgloss.Color("#dafbe1")}, // light green
		{lipgloss.Color("#6639ba"), lipgloss.Color("#eddeff")}, // lavender
		{lipgloss.Color("#953800"), lipgloss.Color("#fff1e5")}, // amber
	}

	t.WorktreeColors = worktreeColorsLight

	t.GlamourStyle = "light"
}

func (t *Theme) buildStyles() {
	t.StyleAppTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.ColorAccent).
		Padding(0, 1, 0, 2)

	t.StyleTabActive = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.ColorAccent).
		Background(t.ColorAccentBg).
		Padding(0, 1)

	t.StyleTabInactive = lipgloss.NewStyle().
		Foreground(t.ColorMuted).
		Padding(0, 1)

	t.StyleSeparator = lipgloss.NewStyle().
		Foreground(t.ColorBorder)

	t.StyleStatusBar = lipgloss.NewStyle().
		Background(t.ColorSurface).
		Foreground(t.ColorMuted).
		Padding(0, 1)

	t.StyleTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.ColorText)

	t.StyleSubtitle = lipgloss.NewStyle().
		Foreground(t.ColorMuted)

	t.StyleFaint = lipgloss.NewStyle().
		Foreground(t.ColorFaint)

	t.StyleSectionHeader = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.ColorAccent)

	t.StyleLabel = lipgloss.NewStyle()

	t.StyleLabelPill = lipgloss.NewStyle().
		Padding(0, 1)

	t.StyleCard = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.ColorBorder).
		Padding(0, 1)

	t.StyleActiveCard = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.ColorAccent).
		Padding(0, 1)

	t.StyleColumnHeader = lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1)

	t.StyleStatusKey = lipgloss.NewStyle().
		Foreground(t.ColorText).
		Bold(true)

	t.StyleStatusSep = lipgloss.NewStyle().
		Foreground(t.ColorFaint)

	t.StyleCommentBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.ColorBorder).
		Padding(0, 1).
		MarginLeft(1)

	t.StyleMetaBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.ColorBorder).
		Padding(0, 1).
		MarginLeft(1)

	t.StyleDragCard = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.ColorFaint).
		Foreground(t.ColorFaint).
		Padding(0, 1)

	t.StyleDropTarget = lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1)

	t.StyleWorktreeLabel = lipgloss.NewStyle().
		Foreground(t.ColorWorktree)

	t.StyleWorktreeBadge = lipgloss.NewStyle().
		Foreground(t.ColorContrast).
		Background(t.ColorWorktree).
		Padding(0, 1).
		Bold(true)
}

// PriorityStyle returns a foreground-colored style for a priority level.
func (t Theme) PriorityStyle(p data.Priority) lipgloss.Style {
	var c color.Color
	switch p {
	case data.PriorityUrgent:
		c = t.ColorUrgent
	case data.PriorityHigh:
		c = t.ColorHigh
	case data.PriorityMedium:
		c = t.ColorMedium
	case data.PriorityLow:
		c = t.ColorLow
	default:
		c = t.ColorFaint
	}
	return lipgloss.NewStyle().Foreground(c)
}

// StatusStyle returns a foreground-colored style for a status.
func (t Theme) StatusStyle(s data.Status) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.StatusColorFor(s))
}

// StatusColorFor returns the raw color for a given status.
func (t Theme) StatusColorFor(s data.Status) color.Color {
	switch s {
	case data.StatusBacklog:
		return t.ColorBacklog
	case data.StatusTodo:
		return t.ColorTodo
	case data.StatusInProgress:
		return t.ColorInProgress
	case data.StatusDone:
		return t.ColorDone
	case data.StatusCancelled:
		return t.ColorCancelled
	default:
		return t.ColorMuted
	}
}

// StatusHeaderStyle returns a column header style colored by status.
func (t Theme) StatusHeaderStyle(s data.Status) lipgloss.Style {
	return t.StyleColumnHeader.Foreground(t.StatusColorFor(s))
}

// StatusPillStyle returns a colored-background pill style for the detail view.
func (t Theme) StatusPillStyle(s data.Status) lipgloss.Style {
	var fg, bg color.Color
	switch s {
	case data.StatusBacklog:
		fg, bg = t.ColorText, t.PillBgBacklog
	case data.StatusTodo:
		fg, bg = t.ColorContrast, t.ColorTodo
	case data.StatusInProgress:
		fg, bg = t.ColorContrast, t.ColorInProgress
	case data.StatusDone:
		fg, bg = t.ColorContrast, t.ColorDone
	case data.StatusCancelled:
		fg, bg = t.ColorMuted, t.PillBgCancelled
	default:
		fg, bg = t.ColorText, t.ColorBorder
	}
	return lipgloss.NewStyle().
		Foreground(fg).
		Background(bg).
		Padding(0, 1).
		Bold(true)
}

// FormatKeyHint renders a styled "key action" pair for the status bar.
func (t Theme) FormatKeyHint(k, action string) string {
	return t.StyleStatusKey.Render(k) + " " + action
}

// labelColorIndex returns a stable index for a label string.
func labelColorIndex(label string, n int) int {
	h := uint32(0)
	for _, r := range label {
		h = h*31 + uint32(r)
	}
	return int(h % uint32(n))
}

// RenderLabel renders a label with a deterministic color (compact, for board cards).
func (t Theme) RenderLabel(label string) string {
	c := t.LabelColors[labelColorIndex(label, len(t.LabelColors))]
	return t.StyleLabel.Foreground(c.Fg).Render(label)
}

// RenderLabelPill renders a label pill with background (for detail view).
func (t Theme) RenderLabelPill(label string) string {
	c := t.LabelColors[labelColorIndex(label, len(t.LabelColors))]
	return t.StyleLabelPill.Foreground(c.Fg).Background(c.Bg).Render(label)
}

// WorktreeColorFor returns the color for a worktree name given the sorted list
// of all worktree names. The color is determined by the name's index.
func (t Theme) WorktreeColorFor(name string, allWorktrees []string) color.Color {
	for i, n := range allWorktrees {
		if n == name {
			return t.WorktreeColors[i%len(t.WorktreeColors)]
		}
	}
	return t.ColorWorktree // fallback
}

// RenderSourceIndicators returns a compact string showing where an issue exists.
// Example: "◆ ⑂⑂" with main diamond and colored fork icons.
func (t Theme) RenderSourceIndicators(sources []data.IssueSource, wtNames []string) string {
	var parts []string
	for _, s := range sources {
		if s.Name == "" {
			parts = append(parts, lipgloss.NewStyle().Foreground(t.ColorMuted).Render(MainIcon()))
		} else {
			c := t.WorktreeColorFor(s.Name, wtNames)
			parts = append(parts, lipgloss.NewStyle().Foreground(c).Render(WorktreeIcon()))
		}
	}
	return strings.Join(parts, "")
}
