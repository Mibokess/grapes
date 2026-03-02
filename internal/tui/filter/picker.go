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
}

var (
	colorAccent   = lipgloss.Color("#a371f7")
	colorAccentBg = lipgloss.Color("#2d1b69")
	colorGreen    = lipgloss.Color("#3fb950")
	colorMuted    = lipgloss.Color("#8b949e")
	colorFaint    = lipgloss.Color("#484f58")

	pickerStyleTitle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorAccent)

	pickerStyleCursor = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true)

	pickerStyleCheck = lipgloss.NewStyle().
				Foreground(colorGreen)

	pickerStyleRowActive = lipgloss.NewStyle().
				Background(colorAccentBg)

	pickerStyleHint = lipgloss.NewStyle().
			Foreground(colorFaint)
)

// NewMultiPicker creates a multi-select picker for a filter field.
func NewMultiPicker(title, field string, options []PickerOption, preSelected []string) MultiPicker {
	sel := make(map[string]bool, len(preSelected))
	for _, v := range preSelected {
		sel[v] = true
	}
	return MultiPicker{
		title:    title,
		field:    field,
		options:  options,
		selected: sel,
	}
}

func (m MultiPicker) Init() tea.Cmd { return nil }

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
			var vals []string
			for _, opt := range m.options {
				if m.selected[opt.Value] {
					vals = append(vals, opt.Value)
				}
			}
			field := m.field
			return m, func() tea.Msg {
				return common.FilterPickerResultMsg{Field: field, Selected: vals}
			}
		case key.Matches(msg, pickerKeyCancel):
			return m, func() tea.Msg { return common.FilterCancelMsg{} }
		}
	}
	return m, nil
}

func (m MultiPicker) View() string {
	const rowWidth = 30
	var rows []string

	for i, opt := range m.options {
		isCursor := i == m.cursor
		isSelected := m.selected[opt.Value]

		// Checkbox
		var check string
		if isSelected {
			check = pickerStyleCheck.Render("[✓]")
		} else {
			check = pickerStyleHint.Render("[ ]")
		}

		// Cursor indicator
		var cursor string
		if isCursor {
			cursor = pickerStyleCursor.Render("›") + " "
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
			row = pickerStyleRowActive.Render(row)
		}

		rows = append(rows, row)
	}

	// Footer hint
	hint := pickerStyleHint.Render("space toggle · enter apply · esc cancel")
	rows = append(rows, "")
	rows = append(rows, hint)

	content := strings.Join(rows, "\n")

	// Box with rounded border and accent color
	title := " " + pickerStyleTitle.Render(m.title) + " "
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
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
