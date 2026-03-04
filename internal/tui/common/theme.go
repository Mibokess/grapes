package common

import (
	"image/color"

	"github.com/Mibokess/grapes/internal/config"
	"github.com/Mibokess/grapes/internal/data"
	"charm.land/lipgloss/v2"
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

// T is the global theme instance used for rendering. It defaults to dark.
var T = NewTheme(true)

// ApplyTheme rebuilds the global theme from user config overrides.
func ApplyTheme(cfg config.ThemeConfig) {
	T = NewThemeFromConfig(cfg)
}

// ThemeMsg is sent when the terminal background is detected and the theme changes.
type ThemeMsg struct{ Theme Theme }

// LabelColor holds a foreground/background pair for label rendering.
type LabelColor struct{ Fg, Bg color.Color }

// Theme holds all colors and pre-built styles for the TUI.
type Theme struct {
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
	StyleWorktreeCard  lipgloss.Style
	StyleWorktreeLabel lipgloss.Style
	StyleWorktreeBadge lipgloss.Style

	// Glamour markdown style name ("dark" or "light").
	GlamourStyle string
}

// NewTheme creates a theme appropriate for the terminal background.
func NewTheme(isDark bool) Theme {
	var t Theme
	if isDark {
		t.setDarkColors()
	} else {
		t.setLightColors()
	}
	t.buildStyles()
	return t
}

// NewThemeFromConfig creates a dark theme overridden by user config values.
func NewThemeFromConfig(cfg config.ThemeConfig) Theme {
	t := NewTheme(true)
	if cfg.Text != "" {
		t.ColorText = lipgloss.Color(cfg.Text)
	}
	if cfg.Muted != "" {
		t.ColorMuted = lipgloss.Color(cfg.Muted)
	}
	if cfg.Faint != "" {
		t.ColorFaint = lipgloss.Color(cfg.Faint)
	}
	if cfg.Border != "" {
		t.ColorBorder = lipgloss.Color(cfg.Border)
	}
	if cfg.Surface != "" {
		t.ColorSurface = lipgloss.Color(cfg.Surface)
	}
	if cfg.Accent != "" {
		t.ColorAccent = lipgloss.Color(cfg.Accent)
	}
	if cfg.AccentBg != "" {
		t.ColorAccentBg = lipgloss.Color(cfg.AccentBg)
	}
	if cfg.ColorUrgent != "" {
		t.ColorUrgent = lipgloss.Color(cfg.ColorUrgent)
		t.ColorError = t.ColorUrgent
	}
	if cfg.ColorHigh != "" {
		t.ColorHigh = lipgloss.Color(cfg.ColorHigh)
	}
	if cfg.ColorMedium != "" {
		t.ColorMedium = lipgloss.Color(cfg.ColorMedium)
	}
	if cfg.ColorLow != "" {
		t.ColorLow = lipgloss.Color(cfg.ColorLow)
	}
	if cfg.ColorBacklog != "" {
		t.ColorBacklog = lipgloss.Color(cfg.ColorBacklog)
	}
	if cfg.ColorTodo != "" {
		t.ColorTodo = lipgloss.Color(cfg.ColorTodo)
	}
	if cfg.ColorInProgress != "" {
		t.ColorInProgress = lipgloss.Color(cfg.ColorInProgress)
	}
	if cfg.ColorDone != "" {
		t.ColorDone = lipgloss.Color(cfg.ColorDone)
	}
	if cfg.ColorCancelled != "" {
		t.ColorCancelled = lipgloss.Color(cfg.ColorCancelled)
	}
	t.buildStyles()
	return t
}

func (t *Theme) setDarkColors() {
	t.ColorText = lipgloss.Color("#e6edf3")
	t.ColorMuted = lipgloss.Color("#8b949e")
	t.ColorFaint = lipgloss.Color("#484f58")
	t.ColorBorder = lipgloss.Color("#30363d")
	t.ColorSurface = lipgloss.Color("#161b22")
	t.ColorAccent = lipgloss.Color("#a371f7")
	t.ColorAccentBg = lipgloss.Color("#2d1b69")
	t.ColorContrast = lipgloss.Color("#0d1117")
	t.ColorError = lipgloss.Color("#f85149")
	t.ColorWorktree = lipgloss.Color("#f0883e")

	t.ColorUrgent = lipgloss.Color("#f85149")
	t.ColorHigh = lipgloss.Color("#d29922")
	t.ColorMedium = lipgloss.Color("#388bfd")
	t.ColorLow = lipgloss.Color("#6e7681")

	t.ColorBacklog = lipgloss.Color("#8b949e")
	t.ColorTodo = lipgloss.Color("#388bfd")
	t.ColorInProgress = lipgloss.Color("#d29922")
	t.ColorDone = lipgloss.Color("#3fb950")
	t.ColorCancelled = lipgloss.Color("#6e7681")

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

	t.GlamourStyle = "dark"
}

func (t *Theme) setLightColors() {
	t.ColorText = lipgloss.Color("#1f2328")
	t.ColorMuted = lipgloss.Color("#656d76")
	t.ColorFaint = lipgloss.Color("#afb8c1")
	t.ColorBorder = lipgloss.Color("#d0d7de")
	t.ColorSurface = lipgloss.Color("#f6f8fa")
	t.ColorAccent = lipgloss.Color("#8250df")
	t.ColorAccentBg = lipgloss.Color("#eddeff")
	t.ColorContrast = lipgloss.Color("#ffffff")
	t.ColorError = lipgloss.Color("#cf222e")
	t.ColorWorktree = lipgloss.Color("#bc4c00")

	t.ColorUrgent = lipgloss.Color("#cf222e")
	t.ColorHigh = lipgloss.Color("#9a6700")
	t.ColorMedium = lipgloss.Color("#0969da")
	t.ColorLow = lipgloss.Color("#8c959f")

	t.ColorBacklog = lipgloss.Color("#656d76")
	t.ColorTodo = lipgloss.Color("#0969da")
	t.ColorInProgress = lipgloss.Color("#9a6700")
	t.ColorDone = lipgloss.Color("#1a7f37")
	t.ColorCancelled = lipgloss.Color("#8c959f")

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

	t.StyleWorktreeCard = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.ColorWorktree).
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
