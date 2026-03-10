package filter

import (
	"fmt"
	"strings"

	"github.com/Mibokess/grapes/internal/tui/common"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// MenuCategory represents a filter category in the menu.
type MenuCategory struct {
	Field       string // "status", "priority", "labels", "has_children"
	Label       string // display name
	ActiveCount int    // number of selected values (0 = not active)
	ToggleLabel string // for has_children: "yes"/"no"/"" to show inline toggle state
}

// Menu is the filter category picker overlay.
type Menu struct {
	categories []MenuCategory
	cursor     int
	theme      common.Theme

	// Screen position and width, set by the app for mouse hit-testing.
	ScreenX, ScreenY, ScreenW int
}

// NewMenu creates a filter menu from the current filter state.
func NewMenu(fs FilterSet, labelCount int, theme common.Theme) Menu {
	// Has sub-issues shows current toggle state inline
	var toggleLabel string
	if fs.HasChildren != nil {
		if *fs.HasChildren {
			toggleLabel = "yes"
		} else {
			toggleLabel = "no"
		}
	}

	// Top-level only toggle
	var topLevelLabel string
	if fs.TopLevelOnly {
		topLevelLabel = "on"
	}

	cats := []MenuCategory{
		{Field: "top_level_only", Label: "Top-level only", ToggleLabel: topLevelLabel},
		{Field: "has_children", Label: "Has sub-issues", ToggleLabel: toggleLabel},
		{Field: "status", Label: "Status", ActiveCount: len(fs.Statuses)},
		{Field: "priority", Label: "Priority", ActiveCount: len(fs.Priorities)},
		{Field: "labels", Label: "Label", ActiveCount: len(fs.Labels)},
		{Field: "source", Label: "Source", ActiveCount: len(fs.Sources)},
	}

	// Clear — only shown when any filter is active
	if fs.ActiveCount() > 0 {
		cats = append(cats, MenuCategory{
			Field: "clear_all",
			Label: "Clear",
		})
	}

	return Menu{categories: cats, theme: theme}
}

func (m Menu) Init() tea.Cmd { return nil }

// boxHeight returns the total height of the rendered menu box.
func (m Menu) boxHeight() int {
	// border(1) + padding(1) + categories + padding(1) + border(1)
	return len(m.categories) + 4
}

// selectCurrent returns a command for the currently highlighted category.
func (m Menu) selectCurrent() tea.Cmd {
	cat := m.categories[m.cursor]
	switch cat.Field {
	case "clear_all":
		return func() tea.Msg { return common.ClearAllFiltersMsg{} }
	case "top_level_only":
		return func() tea.Msg { return common.FilterToggleTopLevelMsg{} }
	case "has_children":
		return func() tea.Msg { return common.FilterToggleChildrenMsg{} }
	default:
		field := cat.Field
		return func() tea.Msg { return common.FilterMenuSelectMsg{Field: field} }
	}
}

func (m Menu) Update(msg tea.Msg) (Menu, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, menuKeyUp):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, menuKeyDown):
			if m.cursor < len(m.categories)-1 {
				m.cursor++
			}
		case key.Matches(msg, menuKeySelect):
			return m, m.selectCurrent()
		case key.Matches(msg, menuKeyCancel):
			return m, func() tea.Msg { return common.FilterCancelMsg{} }
		}

	case tea.MouseClickMsg:
		if msg.Button != tea.MouseLeft {
			break
		}
		mouse := msg.Mouse()
		// Options start 2 lines below menu top (border + padding)
		relY := mouse.Y - m.ScreenY - 2
		h := m.boxHeight()
		inBox := mouse.Y >= m.ScreenY && mouse.Y < m.ScreenY+h &&
			mouse.X >= m.ScreenX && mouse.X < m.ScreenX+m.ScreenW
		if inBox && relY >= 0 && relY < len(m.categories) {
			m.cursor = relY
			return m, m.selectCurrent()
		}
		if !inBox {
			return m, func() tea.Msg { return common.FilterCancelMsg{} }
		}

	case tea.MouseMotionMsg:
		mouse := msg.Mouse()
		relY := mouse.Y - m.ScreenY - 2
		if relY >= 0 && relY < len(m.categories) {
			m.cursor = relY
		}
	}
	return m, nil
}

func (m Menu) View() string {
	t := m.theme
	const rowWidth = 28
	var rows []string

	for i, cat := range m.categories {
		isCursor := i == m.cursor

		var cursor string
		if isCursor {
			cursor = pickerStyleCursor(t).Render("›") + " "
		} else {
			cursor = "  "
		}

		label := cat.Label
		if cat.Field == "clear_all" {
			label = pickerStyleHint(t).Render(cat.Label)
		}

		// Right-aligned indicator
		var indicator string
		if cat.ToggleLabel != "" {
			indicator = pickerStyleCheck(t).Render(cat.ToggleLabel)
		} else if cat.ActiveCount > 0 {
			indicator = pickerStyleCheck(t).Render(fmt.Sprintf("(%d)", cat.ActiveCount))
		}

		row := cursor + label
		if indicator != "" {
			visRow := lipgloss.Width(row)
			visInd := lipgloss.Width(indicator)
			gap := rowWidth - visRow - visInd
			if gap < 1 {
				gap = 1
			}
			row += strings.Repeat(" ", gap) + indicator
		}

		// Pad to width
		visible := lipgloss.Width(row)
		if visible < rowWidth {
			row += strings.Repeat(" ", rowWidth-visible)
		}

		if isCursor {
			row = pickerStyleRowActive(t).Render(row)
		}

		rows = append(rows, row)
	}

	content := strings.Join(rows, "\n")

	// Box with rounded border
	title := " " + pickerStyleTitle(t).Render("Filter") + " "
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.ColorAccent).
		Padding(1, 2).
		Render(content)

	// Insert title in top border
	lines := strings.Split(box, "\n")
	if len(lines) > 0 {
		topBorder := lines[0]
		if len(topBorder) > 4 {
			runeTop := []rune(topBorder)
			titleRunes := []rune(title)
			insertAt := 2
			end := insertAt + len(titleRunes)
			if end < len(runeTop) {
				result := make([]rune, 0, len(runeTop))
				result = append(result, runeTop[:insertAt]...)
				result = append(result, titleRunes...)
				result = append(result, runeTop[end:]...)
				lines[0] = string(result)
			}
		}
		box = strings.Join(lines, "\n")
	}

	return box
}

var (
	menuKeyUp = key.NewBinding(
		key.WithKeys("k", "up"),
	)
	menuKeyDown = key.NewBinding(
		key.WithKeys("j", "down"),
	)
	menuKeySelect = key.NewBinding(
		key.WithKeys("enter"),
	)
	menuKeyCancel = key.NewBinding(
		key.WithKeys("esc"),
	)
)
