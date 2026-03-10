package filter

import (
	"strings"

	"github.com/Mibokess/grapes/internal/data"
	"github.com/Mibokess/grapes/internal/tui/common"
	"charm.land/lipgloss/v2"
)

// BarHeight returns the number of lines the filter bar occupies (always 1).
func BarHeight(_ FilterSet) int {
	return 1
}

// RenderBar renders the filter chip bar. Shows "Filters: none" when empty.
func RenderBar(fs FilterSet, width int, theme common.Theme) string {
	barLabel := lipgloss.NewStyle().Bold(true).Foreground(theme.ColorText)
	barBracket := lipgloss.NewStyle().Foreground(theme.ColorFaint)

	if fs.ActiveCount() == 0 {
		hint := lipgloss.NewStyle().Foreground(theme.ColorFaint)
		return "  " + barLabel.Render("Filters:") + " " + hint.Render("none")
	}

	var chips []string

	if len(fs.Statuses) > 0 {
		chips = append(chips, renderChip("Status", statusValues(fs.Statuses), barBracket, func(v string) string {
			return theme.StatusStyle(data.Status(v)).Render(v)
		}))
	}
	if len(fs.Priorities) > 0 {
		chips = append(chips, renderChip("Priority", priorityValues(fs.Priorities), barBracket, func(v string) string {
			return theme.PriorityStyle(data.Priority(v)).Render(v)
		}))
	}
	if len(fs.Labels) > 0 {
		chips = append(chips, renderChip("Label", fs.Labels, barBracket, func(v string) string {
			return theme.RenderLabel(v)
		}))
	}
	if len(fs.Sources) > 0 {
		chips = append(chips, renderChip("Source", fs.Sources, barBracket, func(v string) string {
			if v == "main" {
				return v
			}
			return theme.StyleWorktreeLabel.Render(common.WorktreeIcon() + " " + v)
		}))
	}
	if fs.TopLevelOnly {
		chips = append(chips, renderChip("Scope", []string{"top-level"}, barBracket, func(v string) string {
			return v
		}))
	}
	if fs.HasChildren != nil {
		val := "no"
		if *fs.HasChildren {
			val = "yes"
		}
		chips = append(chips, renderChip("Sub-issues", []string{val}, barBracket, func(v string) string {
			return v
		}))
	}

	prefix := "  " + barLabel.Render("Filters:") + " "
	line := prefix + strings.Join(chips, " ")

	return line
}

func renderChip(category string, values []string, bracketStyle lipgloss.Style, styleFn func(string) string) string {
	open := bracketStyle.Render("[")
	close := bracketStyle.Render("]")

	var styled []string
	for _, v := range values {
		styled = append(styled, styleFn(v))
	}

	return open + category + ": " + strings.Join(styled, ", ") + close
}

func statusValues(ss []data.Status) []string {
	out := make([]string, len(ss))
	for i, s := range ss {
		out[i] = string(s)
	}
	return out
}

func priorityValues(ps []data.Priority) []string {
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = string(p)
	}
	return out
}
