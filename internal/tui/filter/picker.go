package filter

import (
	"strings"

	"github.com/Mibokess/grapes/internal/tui/common"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// PickerOption represents a selectable item in the multi-picker.
type PickerOption struct {
	Value string
	Label string
	Icon  string
	Style lipgloss.Style
}

// MultiPicker is an overlay for selecting multiple values from a list.
type MultiPicker struct {
	title    string
	field    string
	options  []PickerOption
	selected map[string]bool
	cursor   int
	theme    common.Theme

	// Screen position and width, set by the app for mouse hit-testing.
	ScreenX, ScreenY, ScreenW int
}

func pickerStyleTitle(t common.Theme) lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(t.ColorAccent)
}
func pickerStyleCursor(t common.Theme) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.ColorAccent).Bold(true)
}
func pickerStyleCheck(t common.Theme) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.ColorDone)
}
func pickerStyleRowActive(t common.Theme) lipgloss.Style {
	return lipgloss.NewStyle().Background(t.ColorAccentBg)
}
func pickerStyleHint(t common.Theme) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.ColorFaint)
}

// NewMultiPicker creates a multi-select picker for a filter field.
func NewMultiPicker(title, field string, options []PickerOption, preSelected []string, theme common.Theme) MultiPicker {
	sel := make(map[string]bool, len(preSelected))
	for _, v := range preSelected {
		sel[v] = true
	}
	return MultiPicker{
		title:    title,
		field:    field,
		options:  options,
		selected: sel,
		theme:    theme,
	}
}

func (m MultiPicker) Init() tea.Cmd { return nil }

// boxHeight returns the total height of the rendered picker box.
func (m MultiPicker) boxHeight() int {
	// border(1) + padding(1) + options + blank + hint + padding(1) + border(1)
	return len(m.options) + 6
}

// applyCmd returns a command that sends the current selection.
func (m MultiPicker) applyCmd() tea.Cmd {
	var vals []string
	for _, opt := range m.options {
		if m.selected[opt.Value] {
			vals = append(vals, opt.Value)
		}
	}
	field := m.field
	return func() tea.Msg {
		return common.FilterPickerResultMsg{Field: field, Selected: vals}
	}
}

func (m MultiPicker) Update(msg tea.Msg) (MultiPicker, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, pickerKeyUp):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, pickerKeyDown):
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case key.Matches(msg, pickerKeyToggle):
			if m.cursor < len(m.options) {
				v := m.options[m.cursor].Value
				if m.selected[v] {
					delete(m.selected, v)
				} else {
					m.selected[v] = true
				}
			}
		case key.Matches(msg, pickerKeyConfirm):
			return m, m.applyCmd()
		case key.Matches(msg, pickerKeyCancel):
			return m, func() tea.Msg { return common.FilterCancelMsg{} }
		}

	case tea.MouseClickMsg:
		if msg.Button != tea.MouseLeft {
			break
		}
		mouse := msg.Mouse()
		// Options start 2 lines below picker top (border + padding)
		relY := mouse.Y - m.ScreenY - 2
		h := m.boxHeight()
		inBox := mouse.Y >= m.ScreenY && mouse.Y < m.ScreenY+h &&
			mouse.X >= m.ScreenX && mouse.X < m.ScreenX+m.ScreenW
		if inBox && relY >= 0 && relY < len(m.options) {
			// Click on an option → toggle it
			m.cursor = relY
			v := m.options[m.cursor].Value
			if m.selected[v] {
				delete(m.selected, v)
			} else {
				m.selected[v] = true
			}
		} else if inBox {
			// Click inside box but not on an option (hint area) → apply
			hintY := len(m.options) + 1 // blank line + hint line
			if relY >= hintY {
				return m, m.applyCmd()
			}
		} else {
			// Click outside → cancel
			return m, func() tea.Msg { return common.FilterCancelMsg{} }
		}

	case tea.MouseMotionMsg:
		mouse := msg.Mouse()
		relY := mouse.Y - m.ScreenY - 2
		if relY >= 0 && relY < len(m.options) {
			m.cursor = relY
		}
	}
	return m, nil
}

func (m MultiPicker) View() string {
	t := m.theme
	const rowWidth = 30
	var rows []string

	for i, opt := range m.options {
		isCursor := i == m.cursor
		isSelected := m.selected[opt.Value]

		// Checkbox
		var check string
		if isSelected {
			check = pickerStyleCheck(t).Render("[✓]")
		} else {
			check = pickerStyleHint(t).Render("[ ]")
		}

		// Cursor indicator
		var cursor string
		if isCursor {
			cursor = pickerStyleCursor(t).Render("›") + " "
		} else {
			cursor = "  "
		}

		icon := opt.Style.Render(opt.Icon)
		label := opt.Style.Render(opt.Label)

		row := cursor + check + " " + icon + "  " + label

		// Pad to consistent width
		visible := lipgloss.Width(row)
		if visible < rowWidth {
			row += strings.Repeat(" ", rowWidth-visible)
		}

		if isCursor {
			row = pickerStyleRowActive(t).Render(row)
		}

		rows = append(rows, row)
	}

	// Footer hint
	hint := pickerStyleHint(t).Render("space toggle · enter apply · esc cancel")
	rows = append(rows, "")
	rows = append(rows, hint)

	content := strings.Join(rows, "\n")

	// Box with rounded border and accent color
	title := " " + pickerStyleTitle(t).Render(m.title) + " "
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
	pickerKeyUp = key.NewBinding(
		key.WithKeys("k", "up"),
	)
	pickerKeyDown = key.NewBinding(
		key.WithKeys("j", "down"),
	)
	pickerKeyToggle = key.NewBinding(
		key.WithKeys("space", "x"),
	)
	pickerKeyConfirm = key.NewBinding(
		key.WithKeys("enter"),
	)
	pickerKeyCancel = key.NewBinding(
		key.WithKeys("esc"),
	)
)
