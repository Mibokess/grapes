package list

import (
	"fmt"
	"strings"

	"github.com/Mibokess/grapes/internal/data"
	"github.com/Mibokess/grapes/internal/tui/common"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Model struct {
	allIssues    []data.Issue
	table        table.Model
	filter       textinput.Model
	filtering    bool
	width        int
	height       int
	visibleStart int // first visible row index, mirrors table's internal start
}

func New(issues []data.Issue) Model {
	ti := textinput.New()
	ti.Placeholder = "Filter by title..."
	ti.CharLimit = 100

	m := Model{
		allIssues: issues,
		filter:    ti,
	}
	m.table = m.buildTable(issues, 80, 20)
	return m
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Filtering() bool { return m.filtering }

// updateVisibleStart keeps visibleStart in sync with the table's internal scroll position.
func (m *Model) updateVisibleStart() {
	tableHeight := m.height - 3
	if tableHeight < 1 {
		tableHeight = 1
	}
	cursor := m.table.Cursor()
	if m.visibleStart > cursor {
		m.visibleStart = cursor
	}
	if cursor >= m.visibleStart+tableHeight {
		m.visibleStart = cursor - tableHeight + 1
	}
	if m.visibleStart < 0 {
		m.visibleStart = 0
	}
}

func (m Model) SetSize(w, h int) Model {
	m.width = w
	m.height = h
	m.table = m.buildTable(m.filteredIssues(), w, h-3)
	return m
}

func (m Model) SetIssues(issues []data.Issue) Model {
	m.allIssues = issues
	m.table = m.buildTable(m.filteredIssues(), m.width, m.height-3)
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.filtering {
			switch {
			case key.Matches(msg, common.ListKeyMap.Clear):
				m.filtering = false
				m.filter.SetValue("")
				m.filter.Blur()
				m.table = m.buildTable(m.allIssues, m.width, m.height-3)
				m.visibleStart = 0
				return m, nil
			case msg.Code == tea.KeyEnter:
				m.filtering = false
				m.filter.Blur()
				return m, nil
			default:
				var cmd tea.Cmd
				m.filter, cmd = m.filter.Update(msg)
				m.table = m.buildTable(m.filteredIssues(), m.width, m.height-3)
				m.visibleStart = 0
				return m, cmd
			}
		}

		switch {
		case key.Matches(msg, common.ListKeyMap.Open):
			if row := m.table.SelectedRow(); row != nil {
				id := 0
				fmt.Sscanf(row[0], "%d", &id)
				if id > 0 {
					return m, func() tea.Msg { return common.OpenDetailMsg{ID: id} }
				}
			}
		case key.Matches(msg, common.ListKeyMap.EditIssue):
			if row := m.table.SelectedRow(); row != nil {
				id := 0
				fmt.Sscanf(row[0], "%d", &id)
				if id > 0 {
					return m, func() tea.Msg { return common.LaunchEditMsg{ID: id} }
				}
			}
		case key.Matches(msg, common.ListKeyMap.CycleStatus):
			if row := m.table.SelectedRow(); row != nil {
				id := 0
				fmt.Sscanf(row[0], "%d", &id)
				if id > 0 {
					return m, func() tea.Msg {
						return common.ShowPickerMsg{IssueID: id, Field: "status"}
					}
				}
			}
		case key.Matches(msg, common.ListKeyMap.CyclePriority):
			if row := m.table.SelectedRow(); row != nil {
				id := 0
				fmt.Sscanf(row[0], "%d", &id)
				if id > 0 {
					return m, func() tea.Msg {
						return common.ShowPickerMsg{IssueID: id, Field: "priority"}
					}
				}
			}
		case key.Matches(msg, common.ListKeyMap.CycleSort):
			return m, func() tea.Msg { return common.CycleSortMsg{} }
		case key.Matches(msg, common.ListKeyMap.ReverseSort):
			return m, func() tea.Msg { return common.ReverseSortMsg{} }
		case key.Matches(msg, common.ListKeyMap.Filter):
			m.filtering = true
			m.filter.Focus()
			return m, textinput.Blink
		case key.Matches(msg, common.ListKeyMap.ToBoard):
			return m, func() tea.Msg { return common.SwitchScreenMsg{Screen: common.ScreenBoard} }
		case key.Matches(msg, common.ListKeyMap.Refresh):
			return m, func() tea.Msg { return common.RefreshMsg{} }
		}

	case tea.MouseWheelMsg:
		if msg.Button == tea.MouseWheelUp {
			m.table.MoveUp(1)
			m.updateVisibleStart()
			return m, nil
		} else if msg.Button == tea.MouseWheelDown {
			m.table.MoveDown(1)
			m.updateVisibleStart()
			return m, nil
		}

	case tea.MouseClickMsg:
		mouse := msg.Mouse()
		switch msg.Button {
		case tea.MouseLeft:
			if m.filtering {
				break
			}
			// 2-line app header + 2-line table header = 4; +1 if filter line shown.
			tableTopY := common.AppHeaderHeight + 2
			if m.filter.Value() != "" {
				tableTopY = common.AppHeaderHeight + 3
			}
			if mouse.Y >= tableTopY {
				clickedRow := m.visibleStart + (mouse.Y - tableTopY)
				issues := m.filteredIssues()
				if clickedRow >= 0 && clickedRow < len(issues) {
					m.table.SetCursor(clickedRow)
					m.updateVisibleStart()
					if row := m.table.SelectedRow(); row != nil {
						id := 0
						fmt.Sscanf(row[0], "%d", &id)
						if id > 0 {
							return m, func() tea.Msg { return common.OpenDetailMsg{ID: id} }
						}
					}
				}
			}
		case tea.MouseBackward:
			return m, func() tea.Msg { return common.SwitchScreenMsg{Screen: common.ScreenBoard} }
		case tea.MouseForward:
			if row := m.table.SelectedRow(); row != nil {
				id := 0
				fmt.Sscanf(row[0], "%d", &id)
				if id > 0 {
					return m, func() tea.Msg { return common.OpenDetailMsg{ID: id} }
				}
			}
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	m.updateVisibleStart()
	return m, cmd
}

func (m Model) View() string {
	var filterLine string
	if m.filtering {
		filterLine = "  / " + m.filter.View()
	} else if m.filter.Value() != "" {
		filterLine = common.StyleSubtitle.Render(fmt.Sprintf("  Filter: %s", m.filter.Value()))
	}

	tableView := m.table.View()

	if filterLine != "" {
		return lipgloss.JoinVertical(lipgloss.Left, filterLine, tableView)
	}
	return tableView
}

func (m Model) filteredIssues() []data.Issue {
	query := strings.ToLower(m.filter.Value())
	if query == "" {
		return m.allIssues
	}
	var out []data.Issue
	for _, iss := range m.allIssues {
		if strings.Contains(strings.ToLower(iss.Title), query) {
			out = append(out, iss)
		}
	}
	return out
}

func (m Model) buildTable(issues []data.Issue, width, height int) table.Model {
	if width < 40 {
		width = 40
	}

	titleW := width - 48
	if titleW < 20 {
		titleW = 20
	}

	cols := []table.Column{
		{Title: "ID", Width: 5},
		{Title: "Title", Width: titleW},
		{Title: "Status", Width: 13},
		{Title: "Priority", Width: 9},
		{Title: "Labels", Width: 15},
	}

	var rows []table.Row
	for _, iss := range issues {
		var labelParts []string
		for _, l := range iss.Labels {
			labelParts = append(labelParts, common.RenderLabel(l))
		}
		labels := strings.Join(labelParts, " ")
		statusCell := common.StatusStyle(iss.Status).Render(common.StatusIcon(iss.Status) + " " + iss.Status.Label())
		prioCell := common.PriorityStyle(iss.Priority).Render(common.PriorityIcon(iss.Priority) + " " + iss.Priority.Label())
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", iss.ID),
			iss.Title,
			statusCell,
			prioCell,
			labels,
		})
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithWidth(width),
		table.WithHeight(height),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(common.ColorBorder).
		BorderBottom(true).
		Foreground(common.ColorMuted).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#e6edf3")).
		Background(common.ColorAccent).
		Bold(false)
	t.SetStyles(s)

	return t
}
