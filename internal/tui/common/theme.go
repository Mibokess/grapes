package common

import (
	"github.com/Mibokess/grapes/internal/data"
	"github.com/charmbracelet/lipgloss"
)

// gh-dash inspired color palette
var (
	colorSubtle  = lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"}
	colorText    = lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#FAFAFA"}
	colorDimText = lipgloss.AdaptiveColor{Light: "#666666", Dark: "#999999"}
	colorBorder  = lipgloss.AdaptiveColor{Light: "#DBDBDB", Dark: "#3C3C3C"}
	colorAccent  = lipgloss.AdaptiveColor{Light: "#6C40BF", Dark: "#B48EF7"}

	// Priority colors
	colorUrgent = lipgloss.AdaptiveColor{Light: "#E03131", Dark: "#FF6B6B"}
	colorHigh   = lipgloss.AdaptiveColor{Light: "#E8590C", Dark: "#FF922B"}
	colorMedium = lipgloss.AdaptiveColor{Light: "#E67700", Dark: "#FFD43B"}
	colorLow    = lipgloss.AdaptiveColor{Light: "#868E96", Dark: "#909296"}

	// Status colors
	colorBacklog    = lipgloss.AdaptiveColor{Light: "#868E96", Dark: "#909296"}
	colorTodo       = lipgloss.AdaptiveColor{Light: "#1C7ED6", Dark: "#74C0FC"}
	colorInProgress = lipgloss.AdaptiveColor{Light: "#E67700", Dark: "#FFD43B"}
	colorDone       = lipgloss.AdaptiveColor{Light: "#2B8A3E", Dark: "#69DB7C"}
	colorCancelled  = lipgloss.AdaptiveColor{Light: "#ADB5BD", Dark: "#5C5C5C"}
)

// Shared styles
var (
	StyleStatusBar = lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{Light: "#E9ECEF", Dark: "#2C2C2C"}).
			Foreground(colorDimText).
			Padding(0, 1)

	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorText)

	StyleSubtitle = lipgloss.NewStyle().
			Foreground(colorDimText)

	StyleLabel = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#1A1A1A"}).
			Background(colorAccent).
			Padding(0, 1)

	StyleCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1).
			MarginBottom(1)

	StyleActiveCard = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorAccent).
				Padding(0, 1).
				MarginBottom(1)

	StyleColumnHeader = lipgloss.NewStyle().
				Bold(true).
				Padding(0, 1).
				MarginBottom(1)
)

func PriorityStyle(p data.Priority) lipgloss.Style {
	var c lipgloss.AdaptiveColor
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
		c = colorDimText
	}
	return lipgloss.NewStyle().Foreground(c)
}

func StatusStyle(s data.Status) lipgloss.Style {
	var c lipgloss.AdaptiveColor
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
		c = colorDimText
	}
	return lipgloss.NewStyle().Foreground(c)
}

func StatusHeaderStyle(s data.Status) lipgloss.Style {
	var c lipgloss.AdaptiveColor
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
		c = colorDimText
	}
	return StyleColumnHeader.Foreground(c)
}
