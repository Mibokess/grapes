package list

import (
	"fmt"
	"strings"

	"github.com/Mibokess/grapes/internal/data"
	"github.com/Mibokess/grapes/internal/tui/common"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	allIssues []data.Issue
	table     table.Model
	filter    textinput.Model
	filtering bool
	width     int
	height    int
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
	case tea.KeyMsg:
		if m.filtering {
			switch {
			case key.Matches(msg, common.ListKeyMap.Clear):
				m.filtering = false
				m.filter.SetValue("")
				m.filter.Blur()
				m.table = m.buildTable(m.allIssues, m.width, m.height-3)
				return m, nil
			case msg.Type == tea.KeyEnter:
				m.filtering = false
				m.filter.Blur()
				return m, nil
			default:
				var cmd tea.Cmd
				m.filter, cmd = m.filter.Update(msg)
				m.table = m.buildTable(m.filteredIssues(), m.width, m.height-3)
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
		case key.Matches(msg, common.ListKeyMap.Filter):
			m.filtering = true
			m.filter.Focus()
			return m, textinput.Blink
		case key.Matches(msg, common.ListKeyMap.ToBoard):
			return m, func() tea.Msg { return common.SwitchScreenMsg{Screen: common.ScreenBoard} }
		case key.Matches(msg, common.ListKeyMap.Refresh):
			return m, func() tea.Msg { return common.RefreshMsg{} }
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	var filterLine string
	if m.filtering {
		filterLine = "  / " + m.filter.View()
	} else if m.filter.Value() != "" {
		filterLine = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#666", Dark: "#999"}).
			Render(fmt.Sprintf("  Filter: %s", m.filter.Value()))
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

	titleW := width - 60
	if titleW < 20 {
		titleW = 20
	}

	cols := []table.Column{
		{Title: "ID", Width: 5},
		{Title: "Title", Width: titleW},
		{Title: "Status", Width: 13},
		{Title: "Priority", Width: 9},
		{Title: "Assignee", Width: 12},
		{Title: "Labels", Width: 15},
	}

	var rows []table.Row
	for _, iss := range issues {
		labels := strings.Join(iss.Labels, ", ")
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", iss.ID),
			iss.Title,
			iss.Status.Label(),
			iss.Priority.Label(),
			iss.Assignee,
			labels,
		})
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#DBDBDB", Dark: "#3C3C3C"}).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}).
		Background(lipgloss.AdaptiveColor{Light: "#6C40BF", Dark: "#6C40BF"}).
		Bold(false)
	t.SetStyles(s)

	return t
}
