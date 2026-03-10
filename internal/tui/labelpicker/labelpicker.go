package labelpicker

import (
	"strings"

	"github.com/Mibokess/grapes/internal/tui/common"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Model is an overlay for editing an issue's labels.
// It shows all known labels with checkboxes and a text input for new labels.
type Model struct {
	issueID  int
	labels   []string       // all known labels (display order)
	selected map[string]bool // currently checked labels
	cursor   int            // cursor in label list (len(labels) = input field)
	input    textinput.Model
	theme    common.Theme

	// Screen position, set by app for mouse hit-testing.
	ScreenX, ScreenY int
}

// New creates a label picker for the given issue.
// allLabels is the list of all known labels across the project.
// issueLabels is the set currently on this issue.
func New(issueID int, allLabels, issueLabels []string, theme common.Theme) Model {
	sel := make(map[string]bool, len(issueLabels))
	for _, l := range issueLabels {
		sel[l] = true
	}

	ti := textinput.New()
	ti.Placeholder = "new label..."
	ti.CharLimit = 50
	ti.SetWidth(20)

	return Model{
		issueID:  issueID,
		labels:   allLabels,
		selected: sel,
		theme:    theme,
		input:    ti,
	}
}

func (m Model) Init() tea.Cmd { return nil }

// boxHeight returns the total rendered height of the picker box.
func (m Model) boxHeight() int {
	// border(1) + padding(1) + labels + blank + input + blank + hint + padding(1) + border(1)
	return len(m.labels) + 7
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// When input is focused, handle text input first
		if m.cursor == len(m.labels) {
			switch {
			case key.Matches(msg, keyUp):
				m.input.Blur()
				if len(m.labels) > 0 {
					m.cursor = len(m.labels) - 1
				}
				return m, nil
			case key.Matches(msg, keyConfirm):
				// Add new label if input is non-empty
				val := strings.TrimSpace(m.input.Value())
				if val != "" && !m.hasLabel(val) {
					m.labels = append(m.labels, val)
					m.selected[val] = true
					m.input.SetValue("")
				} else if val == "" {
					// Empty input + enter = apply
					return m, m.applyCmd()
				}
				return m, nil
			case key.Matches(msg, keyCancel):
				return m, func() tea.Msg { return common.LabelPickerCancelMsg{} }
			default:
				var cmd tea.Cmd
				m.input, cmd = m.input.Update(msg)
				return m, cmd
			}
		}

		switch {
		case key.Matches(msg, keyUp):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, keyDown):
			if m.cursor < len(m.labels) {
				m.cursor++
				if m.cursor == len(m.labels) {
					m.input.Focus()
				}
			}
		case key.Matches(msg, keyToggle):
			if m.cursor < len(m.labels) {
				v := m.labels[m.cursor]
				if m.selected[v] {
					delete(m.selected, v)
				} else {
					m.selected[v] = true
				}
			}
		case key.Matches(msg, keyConfirm):
			return m, m.applyCmd()
		case key.Matches(msg, keyCancel):
			return m, func() tea.Msg { return common.LabelPickerCancelMsg{} }
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
			mouse.X >= m.ScreenX
		if !inBox {
			return m, func() tea.Msg { return common.LabelPickerCancelMsg{} }
		}
		if relY >= 0 && relY < len(m.labels) {
			// Click on a label → toggle it
			m.cursor = relY
			m.input.Blur()
			v := m.labels[m.cursor]
			if m.selected[v] {
				delete(m.selected, v)
			} else {
				m.selected[v] = true
			}
		} else {
			// Click inside box but not on a label.
			// Layout after labels: blank(1) + input(1) + blank(1) + hint(1)
			inputY := len(m.labels) + 1 // blank + input row
			hintY := len(m.labels) + 3  // blank + input + blank + hint
			if relY == inputY {
				// Click on input field → focus it
				m.cursor = len(m.labels)
				m.input.Focus()
			} else if relY >= hintY {
				// Click on hint area → apply
				return m, m.applyCmd()
			}
		}

	case tea.MouseMotionMsg:
		mouse := msg.Mouse()
		relY := mouse.Y - m.ScreenY - 2
		if relY >= 0 && relY <= len(m.labels) {
			m.cursor = relY
			if m.cursor == len(m.labels) {
				m.input.Focus()
			} else {
				m.input.Blur()
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	t := m.theme
	const rowWidth = 34
	cursorStyle := lipgloss.NewStyle().Foreground(t.ColorAccent).Bold(true)
	checkStyle := lipgloss.NewStyle().Foreground(t.ColorDone)
	hintStyle := lipgloss.NewStyle().Foreground(t.ColorFaint)
	rowActiveStyle := lipgloss.NewStyle().Background(t.ColorAccentBg)

	var rows []string
	for i, label := range m.labels {
		isCursor := i == m.cursor
		isSelected := m.selected[label]

		var check string
		if isSelected {
			check = checkStyle.Render("[✓]")
		} else {
			check = hintStyle.Render("[ ]")
		}

		var cursor string
		if isCursor {
			cursor = cursorStyle.Render("›") + " "
		} else {
			cursor = "  "
		}

		labelStr := t.RenderLabel(label)
		row := cursor + check + " " + labelStr

		visible := lipgloss.Width(row)
		if visible < rowWidth {
			row += strings.Repeat(" ", rowWidth-visible)
		}

		if isCursor {
			row = rowActiveStyle.Render(row)
		}

		rows = append(rows, row)
	}

	// Input field for new labels
	rows = append(rows, "")
	inputPrefix := "  "
	if m.cursor == len(m.labels) {
		inputPrefix = cursorStyle.Render("›") + " "
	}
	rows = append(rows, inputPrefix+"+ "+m.input.View())

	// Footer hint
	rows = append(rows, "")
	rows = append(rows, hintStyle.Render("space toggle · enter apply · esc cancel"))

	content := strings.Join(rows, "\n")

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(t.ColorAccent)
	title := " " + titleStyle.Render("Labels") + " "
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

func (m Model) hasLabel(label string) bool {
	for _, l := range m.labels {
		if l == label {
			return true
		}
	}
	return false
}

func (m Model) applyCmd() tea.Cmd {
	var labels []string
	for _, l := range m.labels {
		if m.selected[l] {
			labels = append(labels, l)
		}
	}
	issueID := m.issueID
	return func() tea.Msg {
		return common.LabelPickerResultMsg{IssueID: issueID, Labels: labels}
	}
}

var (
	keyUp = key.NewBinding(
		key.WithKeys("k", "up"),
	)
	keyDown = key.NewBinding(
		key.WithKeys("j", "down"),
	)
	keyToggle = key.NewBinding(
		key.WithKeys("space", "x"),
	)
	keyConfirm = key.NewBinding(
		key.WithKeys("enter"),
	)
	keyCancel = key.NewBinding(
		key.WithKeys("esc"),
	)
)
