package board

import (
	"fmt"
	"strings"

	"github.com/Mibokess/grapes/internal/data"
	"github.com/Mibokess/grapes/internal/tui/common"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type column struct {
	status data.Status
	issues []data.Issue
}

type Model struct {
	columns   []column
	curCol    int
	curRow    int
	scrollCol int // first visible column index
	width     int
	height    int
	visCols   int // number of visible columns
}

func New(issues []data.Issue) Model {
	m := Model{visCols: 4}
	m.columns = groupByStatus(issues)
	return m
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) SetSize(w, h int) Model {
	m.width = w
	m.height = h
	if m.width < 80 {
		m.visCols = 3
	} else {
		m.visCols = min(5, len(m.columns))
	}
	return m
}

func (m Model) SetIssues(issues []data.Issue) Model {
	m.columns = groupByStatus(issues)
	if m.curCol >= len(m.columns) {
		m.curCol = max(0, len(m.columns)-1)
	}
	if len(m.columns) > 0 && m.curRow >= len(m.columns[m.curCol].issues) {
		m.curRow = max(0, len(m.columns[m.curCol].issues)-1)
	}
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, common.BoardKeyMap.Left):
			if m.curCol > 0 {
				m.curCol--
				m.clampRow()
				m.ensureVisible()
			}
		case key.Matches(msg, common.BoardKeyMap.Right):
			if m.curCol < len(m.columns)-1 {
				m.curCol++
				m.clampRow()
				m.ensureVisible()
			}
		case key.Matches(msg, common.BoardKeyMap.Up):
			if m.curRow > 0 {
				m.curRow--
			}
		case key.Matches(msg, common.BoardKeyMap.Down):
			col := m.columns[m.curCol]
			if m.curRow < len(col.issues)-1 {
				m.curRow++
			}
		case key.Matches(msg, common.BoardKeyMap.Open):
			if len(m.columns) > 0 && len(m.columns[m.curCol].issues) > 0 {
				issue := m.columns[m.curCol].issues[m.curRow]
				return m, func() tea.Msg { return common.OpenDetailMsg{ID: issue.ID} }
			}
		case key.Matches(msg, common.BoardKeyMap.ToList):
			return m, func() tea.Msg { return common.SwitchScreenMsg{Screen: common.ScreenList} }
		case key.Matches(msg, common.BoardKeyMap.Refresh):
			return m, func() tea.Msg { return common.RefreshMsg{} }
		}
	}
	return m, nil
}

func (m *Model) clampRow() {
	if len(m.columns) == 0 {
		return
	}
	col := m.columns[m.curCol]
	if m.curRow >= len(col.issues) {
		m.curRow = max(0, len(col.issues)-1)
	}
}

func (m *Model) ensureVisible() {
	if m.curCol < m.scrollCol {
		m.scrollCol = m.curCol
	}
	if m.curCol >= m.scrollCol+m.visCols {
		m.scrollCol = m.curCol - m.visCols + 1
	}
}

func (m Model) View() string {
	if m.width == 0 || len(m.columns) == 0 {
		return "No issues found."
	}

	visible := m.visCols
	if visible > len(m.columns)-m.scrollCol {
		visible = len(m.columns) - m.scrollCol
	}
	colWidth := m.width/visible - 2
	if colWidth < 20 {
		colWidth = 20
	}

	renderedCols := make([]string, visible)
	for i := 0; i < visible; i++ {
		ci := m.scrollCol + i
		col := m.columns[ci]
		isActiveCol := ci == m.curCol
		renderedCols[i] = m.renderColumn(col, colWidth, isActiveCol)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, renderedCols...)
}

func (m Model) renderColumn(col column, width int, isActive bool) string {
	header := common.StatusHeaderStyle(col.status).
		Width(width).
		Render(fmt.Sprintf(" %s (%d)", col.status.Label(), len(col.issues)))

	var cards []string
	maxCards := m.height - 4
	if maxCards < 1 {
		maxCards = 1
	}

	for i, issue := range col.issues {
		if i >= maxCards {
			remaining := len(col.issues) - maxCards
			cards = append(cards, common.StyleSubtitle.Render(fmt.Sprintf("  +%d more", remaining)))
			break
		}
		active := isActive && i == m.curRow
		cards = append(cards, m.renderCard(issue, width-2, active))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, cards...)
	return lipgloss.JoinVertical(lipgloss.Left, header, content)
}

func (m Model) renderCard(issue data.Issue, width int, active bool) string {
	style := common.StyleCard.Width(width)
	if active {
		style = common.StyleActiveCard.Width(width)
	}

	idStr := common.StyleSubtitle.Render(fmt.Sprintf("#%d", issue.ID))
	title := truncate(issue.Title, width-6)
	line1 := fmt.Sprintf("%s %s", idStr, title)

	var parts []string
	prioStr := common.PriorityStyle(issue.Priority).Render(issue.Priority.Label())
	parts = append(parts, prioStr)
	if issue.Assignee != "" {
		parts = append(parts, common.StyleSubtitle.Render("@"+issue.Assignee))
	}
	line2 := strings.Join(parts, " ")

	var line3 string
	if len(issue.Labels) > 0 {
		var labels []string
		for _, l := range issue.Labels {
			labels = append(labels, common.StyleLabel.Render(l))
		}
		line3 = strings.Join(labels, " ")
	}

	content := line1 + "\n" + line2
	if line3 != "" {
		content += "\n" + line3
	}

	return style.Render(content)
}

func groupByStatus(issues []data.Issue) []column {
	byStatus := make(map[data.Status][]data.Issue)
	for _, iss := range issues {
		byStatus[iss.Status] = append(byStatus[iss.Status], iss)
	}

	var cols []column
	for _, s := range data.AllStatuses {
		cols = append(cols, column{
			status: s,
			issues: byStatus[s],
		})
	}
	return cols
}

func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
