package picker

import (
	"strings"

	"github.com/Mibokess/grapes/internal/tui/common"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Option represents a single selectable item in the picker.
type Option struct {
	Value string         // raw value ("todo", "high")
	Label string         // display text ("Todo", "High")
	Icon  string         // icon character ("◌", "!")
	Style lipgloss.Style // color style for icon and label
}

// Model is a small overlay picker for selecting from a list of options.
type Model struct {
	title   string
	options []Option
	cursor  int    // currently highlighted option
	current int    // index of the issue's current value (shown with ✓)
	issueID int    // which issue this picker is for
	field   string // "status" or "priority"
	theme   common.Theme

	// Screen position of the picker box, set by the app for mouse handling.
	ScreenX, ScreenY int
}

// New creates a picker model. current is the index of the currently active value.
func New(title string, options []Option, current, issueID int, field string, theme common.Theme) Model {
	cursor := current
	if cursor < 0 || cursor >= len(options) {
		cursor = 0
	}
	return Model{
		title:   title,
		options: options,
		cursor:  cursor,
		current: current,
		issueID: issueID,
		field:   field,
		theme:   theme,
	}
}

func (m Model) Init() tea.Cmd { return nil }

// boxHeight returns the total height of the rendered picker box.
func (m Model) boxHeight() int {
	// border(1) + padding(1) + options + padding(1) + border(1)
	return len(m.options) + 4
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, keyUp):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, keyDown):
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case key.Matches(msg, keySelect):
			opt := m.options[m.cursor]
			return m, func() tea.Msg {
				return common.PickerResultMsg{
					IssueID: m.issueID,
					Field:   m.field,
					Value:   opt.Value,
				}
			}
		case key.Matches(msg, keyCancel):
			return m, func() tea.Msg { return common.PickerCancelMsg{} }
		}

	case tea.MouseClickMsg:
		if msg.Button != tea.MouseLeft {
			break
		}
		mouse := msg.Mouse()
		// Options start 2 lines below picker top (border + padding)
		relY := mouse.Y - m.ScreenY - 2
		inBox := mouse.Y >= m.ScreenY && mouse.Y < m.ScreenY+m.boxHeight() &&
			mouse.X >= m.ScreenX
		if inBox && relY >= 0 && relY < len(m.options) {
			m.cursor = relY
			opt := m.options[m.cursor]
			return m, func() tea.Msg {
				return common.PickerResultMsg{
					IssueID: m.issueID,
					Field:   m.field,
					Value:   opt.Value,
				}
			}
		}
		// Click outside → cancel
		if !inBox {
			return m, func() tea.Msg { return common.PickerCancelMsg{} }
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

func (m Model) View() string {
	cursorStyle := lipgloss.NewStyle().Foreground(m.theme.ColorAccent).Bold(true)
	checkStyle := lipgloss.NewStyle().Foreground(m.theme.ColorDone)
	rowActiveStyle := lipgloss.NewStyle().Background(m.theme.ColorAccentBg)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(m.theme.ColorAccent)

	// Build each option row
	const rowWidth = 24
	var rows []string
	for i, opt := range m.options {
		var prefix string
		isCursor := i == m.cursor
		isCurrent := i == m.current

		switch {
		case isCursor:
			prefix = cursorStyle.Render("›") + " "
		case isCurrent:
			prefix = checkStyle.Render("✓") + " "
		default:
			prefix = "  "
		}

		icon := opt.Style.Render(opt.Icon)
		label := opt.Style.Render(opt.Label)
		row := prefix + icon + "  " + label

		// Pad to consistent width
		visible := lipgloss.Width(row)
		if visible < rowWidth {
			row += strings.Repeat(" ", rowWidth-visible)
		}

		if isCursor {
			row = rowActiveStyle.Render(row)
		}

		rows = append(rows, row)
	}

	content := strings.Join(rows, "\n")

	// Box with rounded border and accent color
	title := " " + titleStyle.Render(m.title) + " "
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.ColorAccent).
		Padding(1, 2).
		Render(content)

	// Replace the top border segment with the title
	lines := strings.Split(box, "\n")
	if len(lines) > 0 {
		topBorder := lines[0]
		// Insert title after the first 2 border chars
		if len(topBorder) > 4 {
			runeTop := []rune(topBorder)
			titleRunes := []rune(title)
			insertAt := 2 // after "╭─"
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

// Picker-local keybindings (not exported, only used within the picker).
var (
	keyUp = key.NewBinding(
		key.WithKeys("k", "up"),
	)
	keyDown = key.NewBinding(
		key.WithKeys("j", "down"),
	)
	keySelect = key.NewBinding(
		key.WithKeys("enter"),
	)
	keyCancel = key.NewBinding(
		key.WithKeys("esc"),
	)
)
