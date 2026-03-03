package filter

import (
	"strings"

	"github.com/Mibokess/grapes/internal/data"
	"github.com/Mibokess/grapes/internal/tui/common"
	"charm.land/lipgloss/v2"
)

var (
	barStyleLabel = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#e6edf3"))
	barStyleBracket = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#484f58"))
)

// BarHeight returns the number of lines the filter bar occupies (0 or 1).
func BarHeight(fs FilterSet) int {
	if fs.ActiveCount() == 0 {
		return 0
	}
	return 1
}

// RenderBar renders the filter chip bar. Returns "" when no filters are active.
func RenderBar(fs FilterSet, width int) string {
	if fs.ActiveCount() == 0 {
		return ""
	}

	var chips []string

	if len(fs.Statuses) > 0 {
		chips = append(chips, renderChip("Status", statusValues(fs.Statuses), func(v string) string {
			return common.StatusStyle(data.Status(v)).Render(v)
		}))
	}
	if len(fs.Priorities) > 0 {
		chips = append(chips, renderChip("Priority", priorityValues(fs.Priorities), func(v string) string {
			return common.PriorityStyle(data.Priority(v)).Render(v)
		}))
	}
	if len(fs.Labels) > 0 {
		chips = append(chips, renderChip("Label", fs.Labels, func(v string) string {
			return common.RenderLabel(v)
		}))
	}
	if len(fs.Sources) > 0 {
		chips = append(chips, renderChip("Source", fs.Sources, func(v string) string {
			if v == "main" {
				return v
			}
			return common.StyleWorktreeLabel.Render(common.WorktreeIcon() + " " + v)
		}))
	}
	if fs.TopLevelOnly {
		chips = append(chips, renderChip("Scope", []string{"top-level"}, func(v string) string {
			return v
		}))
	}
	if fs.HasChildren != nil {
		val := "no"
		if *fs.HasChildren {
			val = "yes"
		}
		chips = append(chips, renderChip("Sub-issues", []string{val}, func(v string) string {
			return v
		}))
	}

	prefix := "  " + barStyleLabel.Render("Filters:") + " "
	line := prefix + strings.Join(chips, " ")

	return line
}

func renderChip(category string, values []string, styleFn func(string) string) string {
	open := barStyleBracket.Render("[")
	close := barStyleBracket.Render("]")

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
