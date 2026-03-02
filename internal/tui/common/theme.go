package common

import (
	"image/color"

	"github.com/Mibokess/grapes/internal/data"
	"charm.land/lipgloss/v2"
)

// GitHub dark-inspired color palette.
var (
	colorBorder  = lipgloss.Color("#30363d")
	colorText    = lipgloss.Color("#e6edf3")
	colorMuted   = lipgloss.Color("#8b949e")
	colorFaint   = lipgloss.Color("#484f58")
	colorSurface = lipgloss.Color("#161b22")
	colorAccent  = lipgloss.Color("#a371f7")
	colorAccentBg = lipgloss.Color("#2d1b69")

	// Priority colors
	colorUrgent = lipgloss.Color("#f85149")
	colorHigh   = lipgloss.Color("#d29922")
	colorMedium = lipgloss.Color("#388bfd")
	colorLow    = lipgloss.Color("#6e7681")

	// Status colors
	colorBacklog    = lipgloss.Color("#8b949e")
	colorTodo       = lipgloss.Color("#388bfd")
	colorInProgress = lipgloss.Color("#d29922")
	colorDone       = lipgloss.Color("#3fb950")
	colorCancelled  = lipgloss.Color("#6e7681")
)

// Exported raw colors needed by sub-packages (e.g. bubbles/table styling).
var (
	ColorBorder = colorBorder
	ColorMuted  = colorMuted
	ColorAccent = colorAccent
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

// Shared styles.
var (
	StyleAppTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent).
			Padding(0, 1, 0, 2)

	StyleTabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent).
			Background(colorAccentBg).
			Padding(0, 1)

	StyleTabInactive = lipgloss.NewStyle().
				Foreground(colorMuted).
				Padding(0, 1)

	StyleSeparator = lipgloss.NewStyle().
			Foreground(colorBorder)

	StyleStatusBar = lipgloss.NewStyle().
			Background(colorSurface).
			Foreground(colorMuted).
			Padding(0, 1)

	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorText)

	StyleSubtitle = lipgloss.NewStyle().
			Foreground(colorMuted)

	StyleFaint = lipgloss.NewStyle().
			Foreground(colorFaint)

	StyleSectionHeader = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorAccent)

	StyleLabel = lipgloss.NewStyle()

	StyleLabelPill = lipgloss.NewStyle().
			Padding(0, 1)

	StyleCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	StyleActiveCard = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorAccent).
				Padding(0, 1)

	StyleColumnHeader = lipgloss.NewStyle().
				Bold(true).
				Padding(0, 1)

	StyleStatusKey = lipgloss.NewStyle().
			Foreground(colorText).
			Bold(true)

	StyleStatusSep = lipgloss.NewStyle().
			Foreground(colorFaint)

	StyleCommentBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	StyleMetaBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)
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

func PriorityStyle(p data.Priority) lipgloss.Style {
	var c color.Color
	switch p {
	case data.PriorityUrgent:
		c = colorUrgent
	case data.PriorityHigh:
		c = colorHigh
	case data.PriorityMedium:
		c = colorMedium
	case data.PriorityLow:
		c = colorLow
	default:
		c = colorFaint
	}
	return lipgloss.NewStyle().Foreground(c)
}

func StatusStyle(s data.Status) lipgloss.Style {
	var c color.Color
	switch s {
	case data.StatusBacklog:
		c = colorBacklog
	case data.StatusTodo:
		c = colorTodo
	case data.StatusInProgress:
		c = colorInProgress
	case data.StatusDone:
		c = colorDone
	case data.StatusCancelled:
		c = colorCancelled
	default:
		c = colorMuted
	}
	return lipgloss.NewStyle().Foreground(c)
}

// StatusColorFor returns the raw color for a given status.
func StatusColorFor(s data.Status) color.Color {
	switch s {
	case data.StatusBacklog:
		return colorBacklog
	case data.StatusTodo:
		return colorTodo
	case data.StatusInProgress:
		return colorInProgress
	case data.StatusDone:
		return colorDone
	case data.StatusCancelled:
		return colorCancelled
	default:
		return colorMuted
	}
}

// FormatKeyHint renders a styled "key action" pair for the status bar.
func FormatKeyHint(k, action string) string {
	return StyleStatusKey.Render(k) + " " + action
}

func StatusHeaderStyle(s data.Status) lipgloss.Style {
	var c color.Color
	switch s {
	case data.StatusBacklog:
		c = colorBacklog
	case data.StatusTodo:
		c = colorTodo
	case data.StatusInProgress:
		c = colorInProgress
	case data.StatusDone:
		c = colorDone
	case data.StatusCancelled:
		c = colorCancelled
	default:
		c = colorMuted
	}
	return StyleColumnHeader.Foreground(c)
}

// StatusPillStyle returns a colored-background pill style for the detail view.
func StatusPillStyle(s data.Status) lipgloss.Style {
	var fg, bg color.Color
	switch s {
	case data.StatusBacklog:
		fg, bg = colorText, lipgloss.Color("#3d4148")
	case data.StatusTodo:
		fg, bg = lipgloss.Color("#0d1117"), colorTodo
	case data.StatusInProgress:
		fg, bg = lipgloss.Color("#0d1117"), colorInProgress
	case data.StatusDone:
		fg, bg = lipgloss.Color("#0d1117"), colorDone
	case data.StatusCancelled:
		fg, bg = colorMuted, lipgloss.Color("#21262d")
	default:
		fg, bg = colorText, colorBorder
	}
	return lipgloss.NewStyle().
		Foreground(fg).
		Background(bg).
		Padding(0, 1).
		Bold(true)
}

// Label color palette — distinct, readable colors for dark backgrounds.
var labelColors = []struct{ fg, bg color.Color }{
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

// labelColorIndex returns a stable index for a label string.
func labelColorIndex(label string) int {
	h := uint32(0)
	for _, r := range label {
		h = h*31 + uint32(r)
	}
	return int(h % uint32(len(labelColors)))
}

// RenderLabel renders a label with a deterministic color (compact, for board cards).
func RenderLabel(label string) string {
	c := labelColors[labelColorIndex(label)]
	return StyleLabel.Foreground(c.fg).Render(label)
}

// RenderLabelPill renders a label pill with background (for detail view).
func RenderLabelPill(label string) string {
	c := labelColors[labelColorIndex(label)]
	return StyleLabelPill.Foreground(c.fg).Background(c.bg).Render(label)
}
