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
	scrollCol int // first visible column index (horizontal scroll)
	scrollRow int // first visible row in current column (vertical scroll)
	width     int
	height    int
	visCols   int // number of visible columns
}

func New(issues []data.Issue) Model {
	m := Model{visCols: 3}
	m.columns = groupByStatus(issues)
	return m
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) SetSize(w, h int) Model {
	m.width = w
	m.height = h
	if m.width >= 160 {
		m.visCols = min(5, len(m.columns))
	} else if m.width >= 120 {
		m.visCols = min(4, len(m.columns))
	} else {
		m.visCols = min(3, len(m.columns))
	}
	m.ensureRowVisible()
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
	m.ensureRowVisible()
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
				m.scrollRow = 0
				m.ensureRowVisible()
				m.ensureColVisible()
			}
		case key.Matches(msg, common.BoardKeyMap.Right):
			if m.curCol < len(m.columns)-1 {
				m.curCol++
				m.clampRow()
				m.scrollRow = 0
				m.ensureRowVisible()
				m.ensureColVisible()
			}
		case key.Matches(msg, common.BoardKeyMap.Up):
			if m.curRow > 0 {
				m.curRow--
				m.ensureRowVisible()
			}
		case key.Matches(msg, common.BoardKeyMap.Down):
			col := m.columns[m.curCol]
			if m.curRow < len(col.issues)-1 {
				m.curRow++
				m.ensureRowVisible()
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

	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			// Scroll left through columns (matches Python SnapHorizontalScroll behaviour)
			if m.curCol > 0 {
				m.curCol--
				m.clampRow()
				m.scrollRow = 0
				m.ensureRowVisible()
				m.ensureColVisible()
			}
		case tea.MouseButtonWheelDown:
			// Scroll right through columns
			if m.curCol < len(m.columns)-1 {
				m.curCol++
				m.clampRow()
				m.scrollRow = 0
				m.ensureRowVisible()
				m.ensureColVisible()
			}
		case tea.MouseButtonLeft:
			if msg.Action != tea.MouseActionPress {
				break
			}
			if colIdx, rowIdx, ok := m.cardAt(msg.X, msg.Y); ok {
				m.curCol = colIdx
				m.curRow = rowIdx
				m.ensureColVisible()
				m.ensureRowVisible()
				issue := m.columns[colIdx].issues[rowIdx]
				return m, func() tea.Msg { return common.OpenDetailMsg{ID: issue.ID} }
			}
			// Click in column area without hitting a card — select the column
			if colIdx, ok := m.columnAt(msg.X); ok && colIdx != m.curCol {
				m.curCol = colIdx
				m.clampRow()
				m.scrollRow = 0
				m.ensureRowVisible()
				m.ensureColVisible()
			}
		case tea.MouseButtonForward:
			if msg.Action != tea.MouseActionPress {
				break
			}
			if len(m.columns) > 0 && len(m.columns[m.curCol].issues) > 0 {
				issue := m.columns[m.curCol].issues[m.curRow]
				return m, func() tea.Msg { return common.OpenDetailMsg{ID: issue.ID} }
			}
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

// ensureColVisible adjusts horizontal scroll so the current column is on screen.
func (m *Model) ensureColVisible() {
	if m.curCol < m.scrollCol {
		m.scrollCol = m.curCol
	}
	if m.curCol >= m.scrollCol+m.visCols {
		m.scrollCol = m.curCol - m.visCols + 1
	}
}

// maxVisibleCards returns how many cards fit vertically in a column.
func (m Model) maxVisibleCards() int {
	// Each card: border(2) + ID(1) + title(2) + meta(1) + date(1) = 7 lines.
	// Column header takes 2 lines (text + separator).
	// Reserve 1 extra line for a "more" indicator.
	const cardHeight = 7
	const overhead = 3

	available := m.height - overhead
	if available < cardHeight {
		return 1
	}
	return available / cardHeight
}

// ensureRowVisible adjusts scrollRow so curRow is visible in the current column.
func (m *Model) ensureRowVisible() {
	maxCards := m.maxVisibleCards()
	if m.scrollRow > m.curRow {
		m.scrollRow = m.curRow
	}
	if m.curRow >= m.scrollRow+maxCards {
		m.scrollRow = m.curRow - maxCards + 1
	}
	if m.scrollRow < 0 {
		m.scrollRow = 0
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

	// Shrink visible column count until each column is at least minColWidth wide.
	// Account for inter-column gaps (1 char each, except after last column).
	const minColWidth = 22
	totalGaps := visible - 1
	for visible > 1 && (m.width-totalGaps)/visible < minColWidth {
		visible--
		totalGaps = visible - 1
	}
	colWidth := (m.width - totalGaps) / visible
	if colWidth < minColWidth {
		colWidth = minColWidth
	}

	renderedCols := make([]string, visible)
	for i := 0; i < visible; i++ {
		ci := m.scrollCol + i
		col := m.columns[ci]
		isActiveCol := ci == m.curCol
		colContent := m.renderColumn(col, colWidth, isActiveCol)
		if i < visible-1 {
			colContent = lipgloss.NewStyle().MarginRight(1).Render(colContent)
		}
		renderedCols[i] = colContent
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, renderedCols...)
}

func (m Model) renderColumn(col column, width int, isActive bool) string {
	icon := common.StatusIcon(col.status)
	label := strings.ToUpper(col.status.Label())
	count := common.StyleFaint.Render(fmt.Sprintf("(%d)", len(col.issues)))
	headerText := common.StatusHeaderStyle(col.status).Width(width).
		Render(fmt.Sprintf(" %s %s ", icon, label) + count)
	sepStyle := lipgloss.NewStyle().Foreground(common.StatusColorFor(col.status))
	separator := sepStyle.Render(strings.Repeat("━", width))
	header := lipgloss.JoinVertical(lipgloss.Left, headerText, separator)

	if len(col.issues) == 0 {
		return header
	}

	maxCards := m.maxVisibleCards()

	// Determine scroll window: active column scrolls with cursor, others show from top
	startIdx := 0
	if isActive {
		startIdx = m.scrollRow
	}
	endIdx := startIdx + maxCards
	if endIdx > len(col.issues) {
		endIdx = len(col.issues)
	}

	var cards []string

	if startIdx > 0 {
		cards = append(cards, common.StyleSubtitle.Render(
			fmt.Sprintf("  ↑ %d more", startIdx)))
	}

	for i := startIdx; i < endIdx; i++ {
		active := isActive && i == m.curRow
		cards = append(cards, m.renderCard(col.issues[i], width, active))
	}

	if endIdx < len(col.issues) {
		remaining := len(col.issues) - endIdx
		cards = append(cards, common.StyleSubtitle.Render(
			fmt.Sprintf("  +%d more", remaining)))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, cards...)
	return lipgloss.JoinVertical(lipgloss.Left, header, content)
}

func (m Model) renderCard(issue data.Issue, width int, active bool) string {
	style := common.StyleCard.Width(width - 2) // -2 for border chars
	if active {
		style = common.StyleActiveCard.Width(width - 2)
	}

	// Inner content width = card width - 2 (border) - 2 (padding)
	innerW := width - 4

	// Line 1: #ID + priority icon (small, muted — like Linear's "ETA-502")
	idStr := common.StyleFaint.Render(fmt.Sprintf("#%d", issue.ID))
	prioIcon := common.PriorityStyle(issue.Priority).Render(
		strings.TrimSpace(common.PriorityIcon(issue.Priority)))
	line1 := idStr
	if issue.Priority <= data.PriorityHigh {
		line1 += " " + prioIcon
	}

	// Lines 2-3: Title wraps up to 2 lines, word-wrapping line 1
	titleRunes := []rune(issue.Title)
	var titleLine1, titleLine2 string
	if len(titleRunes) <= innerW {
		titleLine1 = issue.Title
	} else {
		// Find last space within innerW to break on a word boundary
		breakAt := innerW
		for i := innerW - 1; i > 0; i-- {
			if titleRunes[i] == ' ' {
				breakAt = i
				break
			}
		}
		titleLine1 = string(titleRunes[:breakAt])
		rest := titleRunes[breakAt:]
		// Trim leading space from the wrapped portion
		if len(rest) > 0 && rest[0] == ' ' {
			rest = rest[1:]
		}
		titleLine2 = truncate(string(rest), innerW)
	}
	title := titleLine1
	if titleLine2 != "" {
		title += "\n" + titleLine2
	}
	if active {
		title = common.StyleTitle.Render(title)
	}

	// Line 3: @assignee + labels (compact, muted)
	var meta []string
	if issue.Assignee != "" {
		meta = append(meta, common.StyleSubtitle.Render("@"+issue.Assignee))
	}
	used := 0
	if issue.Assignee != "" {
		used = len([]rune(issue.Assignee)) + 1
	}
	for _, l := range issue.Labels {
		lw := len([]rune(l))
		sep := 0
		if len(meta) > 0 {
			sep = 2 // "  "
		}
		if used+sep+lw > innerW {
			break
		}
		meta = append(meta, common.RenderLabel(l))
		used += sep + lw
	}

	// Bottom line: Created date (faint, like Linear)
	var dateLine string
	if !issue.Created.IsZero() {
		dateLine = common.StyleFaint.Render("Created " + issue.Created.Format("Jan 2"))
	}

	content := line1 + "\n" + title
	if len(meta) > 0 {
		content += "\n" + strings.Join(meta, "  ")
	}
	if dateLine != "" {
		content += "\n" + dateLine
	}

	return style.Render(content)
}

// visibleColWidth computes the number of visible columns (after narrowing
// for minimum width) and the content width of each column.
func (m Model) visibleColWidth() (visible, colWidth int) {
	visible = m.visCols
	if visible > len(m.columns)-m.scrollCol {
		visible = len(m.columns) - m.scrollCol
	}
	const minColWidth = 22
	totalGaps := visible - 1
	for visible > 1 && (m.width-totalGaps)/visible < minColWidth {
		visible--
		totalGaps = visible - 1
	}
	totalGaps = visible - 1
	colWidth = (m.width - totalGaps) / visible
	if colWidth < minColWidth {
		colWidth = minColWidth
	}
	return visible, colWidth
}

// columnAt maps a screen x coordinate to a visible column index.
func (m Model) columnAt(x int) (colIdx int, ok bool) {
	if len(m.columns) == 0 || m.width == 0 {
		return 0, false
	}
	visible, colWidth := m.visibleColWidth()
	renderWidth := colWidth + 1 // column width + inter-column gap
	ci := x/renderWidth + m.scrollCol
	if ci < m.scrollCol || ci >= m.scrollCol+visible || ci >= len(m.columns) {
		return 0, false
	}
	return ci, true
}

// cardAt maps a screen (x, y) coordinate to a column and row index.
// Returns ok=false if the click didn't land on a card.
func (m Model) cardAt(x, y int) (colIdx, rowIdx int, ok bool) {
	if len(m.columns) == 0 || m.width == 0 {
		return 0, 0, false
	}
	visible, colWidth := m.visibleColWidth()
	renderWidth := colWidth + 1 // column width + inter-column gap

	ci := x/renderWidth + m.scrollCol
	if ci < m.scrollCol || ci >= m.scrollCol+visible || ci >= len(m.columns) {
		return 0, 0, false
	}
	col := m.columns[ci]
	if len(col.issues) == 0 {
		return 0, 0, false
	}

	// Skip app header lines + column header lines (2 each = 4 total).
	const appH = common.AppHeaderHeight
	const headerH = 2
	const totalSkip = appH + headerH
	if y < totalSkip {
		return 0, 0, false
	}

	// Scroll offset for this column (active column may be scrolled).
	scrollOff := 0
	if ci == m.curCol {
		scrollOff = m.scrollRow
	}

	yOffset := y - totalSkip
	// "↑ N more" indicator occupies the first line when scrolled.
	if scrollOff > 0 {
		if yOffset == 0 {
			return 0, 0, false
		}
		yOffset--
	}

	const cardH = 7
	ri := yOffset/cardH + scrollOff
	if ri < 0 || ri >= len(col.issues) {
		return 0, 0, false
	}
	return ci, ri, true
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
