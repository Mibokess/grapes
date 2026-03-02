package board

import (
	"fmt"
	"strings"

	"github.com/Mibokess/grapes/internal/data"
	"github.com/Mibokess/grapes/internal/tui/common"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
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
	sortMode  data.SortMode

	// Drag-and-drop state
	dragging    bool
	dragMoved   bool // true once the mouse moves after click (real drag)
	dragIssueID int
	dragFromCol int
	dragOverCol int // column cursor is hovering over (-1 = none)
	dragX, dragY int // current cursor position (screen coords)
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

func (m Model) SetSortMode(mode data.SortMode) Model {
	m.sortMode = mode
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
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
		case key.Matches(msg, common.BoardKeyMap.EditIssue):
			if len(m.columns) > 0 && len(m.columns[m.curCol].issues) > 0 {
				issue := m.columns[m.curCol].issues[m.curRow]
				return m, func() tea.Msg { return common.LaunchEditMsg{ID: issue.ID} }
			}
		case key.Matches(msg, common.BoardKeyMap.CycleStatus):
			if len(m.columns) > 0 && len(m.columns[m.curCol].issues) > 0 {
				issue := m.columns[m.curCol].issues[m.curRow]
				return m, func() tea.Msg {
					return common.ShowPickerMsg{IssueID: issue.ID, Field: "status"}
				}
			}
		case key.Matches(msg, common.BoardKeyMap.CyclePriority):
			if len(m.columns) > 0 && len(m.columns[m.curCol].issues) > 0 {
				issue := m.columns[m.curCol].issues[m.curRow]
				return m, func() tea.Msg {
					return common.ShowPickerMsg{IssueID: issue.ID, Field: "priority"}
				}
			}
		case key.Matches(msg, common.BoardKeyMap.CycleSort):
			return m, func() tea.Msg { return common.CycleSortMsg{} }
		case key.Matches(msg, common.BoardKeyMap.ReverseSort):
			return m, func() tea.Msg { return common.ReverseSortMsg{} }
		case key.Matches(msg, common.BoardKeyMap.ToList):
			return m, func() tea.Msg { return common.SwitchScreenMsg{Screen: common.ScreenList} }
		case key.Matches(msg, common.BoardKeyMap.Refresh):
			return m, func() tea.Msg { return common.RefreshMsg{} }
		}

	case tea.MouseWheelMsg:
		if msg.Button == tea.MouseWheelUp {
			// Scroll left through columns
			if m.curCol > 0 {
				m.curCol--
				m.clampRow()
				m.scrollRow = 0
				m.ensureRowVisible()
				m.ensureColVisible()
			}
		} else if msg.Button == tea.MouseWheelDown {
			// Scroll right through columns
			if m.curCol < len(m.columns)-1 {
				m.curCol++
				m.clampRow()
				m.scrollRow = 0
				m.ensureRowVisible()
				m.ensureColVisible()
			}
		}

	case tea.MouseClickMsg:
		mouse := msg.Mouse()
		switch msg.Button {
		case tea.MouseLeft:
			if colIdx, rowIdx, ok := m.cardAt(mouse.X, mouse.Y); ok {
				// Start drag from this card
				m.curCol = colIdx
				m.curRow = rowIdx
				m.ensureColVisible()
				m.ensureRowVisible()
				issue := m.columns[colIdx].issues[rowIdx]
				m.dragging = true
				m.dragMoved = false
				m.dragIssueID = issue.ID
				m.dragFromCol = colIdx
				m.dragOverCol = colIdx
			} else if colIdx, ok := m.columnAt(mouse.X); ok && colIdx != m.curCol {
				// Click in column area without hitting a card — select the column
				m.curCol = colIdx
				m.clampRow()
				m.scrollRow = 0
				m.ensureRowVisible()
				m.ensureColVisible()
			}
		case tea.MouseForward:
			if len(m.columns) > 0 && len(m.columns[m.curCol].issues) > 0 {
				issue := m.columns[m.curCol].issues[m.curRow]
				return m, func() tea.Msg { return common.OpenDetailMsg{ID: issue.ID} }
			}
		}

	case tea.MouseMotionMsg:
		if m.dragging {
			m.dragMoved = true
			mouse := msg.Mouse()
			m.dragX = mouse.X
			m.dragY = mouse.Y
			if colIdx, ok := m.columnAt(mouse.X); ok {
				m.dragOverCol = colIdx
			}
		}

	case tea.MouseReleaseMsg:
		if m.dragging {
			m.dragging = false
			fromCol := m.dragFromCol
			overCol := m.dragOverCol
			issueID := m.dragIssueID
			m.dragOverCol = -1

			if overCol != fromCol && overCol >= 0 && overCol < len(m.columns) {
				// Dropped on a different column — move the issue
				newStatus := m.columns[overCol].status
				return m, func() tea.Msg {
					return common.MoveIssueMsg{IssueID: issueID, NewStatus: newStatus}
				}
			}
			// Released without moving — treat as a click (open detail)
			if !m.dragMoved && len(m.columns) > 0 && len(m.columns[m.curCol].issues) > 0 {
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

	board := lipgloss.JoinHorizontal(lipgloss.Top, renderedCols...)

	// Overlay floating card at cursor during drag — temporarily clear dragging
	// so renderCard uses the active style instead of the ghost style.
	if m.dragging {
		if issue, ok := m.findIssue(m.dragIssueID); ok {
			saved := m.dragging
			m.dragging = false
			floating := m.renderCard(issue, colWidth, true)
			m.dragging = saved
			board = m.overlayAt(board, floating, m.dragX, m.dragY-common.AppHeaderHeight)
		}
	}

	return board
}

func (m Model) renderColumn(col column, width int, isActive bool) string {
	icon := common.StatusIcon(col.status)
	label := strings.ToUpper(col.status.Label())
	count := common.StyleFaint.Render(fmt.Sprintf("(%d)", len(col.issues)))

	// Highlight column header when it's the drop target during a drag
	isDropTarget := m.dragging && m.dragOverCol >= 0 && m.dragOverCol < len(m.columns) &&
		m.columns[m.dragOverCol].status == col.status && m.dragOverCol != m.dragFromCol

	var headerText, separator string
	if isDropTarget {
		headerText = common.StyleDropTarget.
			Foreground(common.StatusColorFor(col.status)).
			Background(common.StatusColorFor(col.status)).
			Width(width).
			Render(fmt.Sprintf(" %s %s ", icon, label) + count)
		separator = lipgloss.NewStyle().
			Foreground(common.StatusColorFor(col.status)).
			Render(strings.Repeat("━", width))
	} else {
		headerText = common.StatusHeaderStyle(col.status).Width(width).
			Render(fmt.Sprintf(" %s %s ", icon, label) + count)
		sepStyle := lipgloss.NewStyle().Foreground(common.StatusColorFor(col.status))
		separator = sepStyle.Render(strings.Repeat("━", width))
	}
	header := lipgloss.JoinVertical(lipgloss.Left, headerText, separator)

	// Insert a ghost preview card when dragging over this column
	var ghostCard string
	ghostInsertIdx := -1
	if isDropTarget {
		if issue, ok := m.findIssue(m.dragIssueID); ok {
			ghostCard = m.renderGhostCard(issue, width)
			// Figure out which slot the cursor is near based on Y position
			const headerH = 2
			const cardH = 7
			yInCol := m.dragY - common.AppHeaderHeight - headerH
			if yInCol < 0 {
				ghostInsertIdx = 0
			} else {
				ghostInsertIdx = yInCol / cardH
			}
		}
	}

	if len(col.issues) == 0 {
		if ghostCard != "" {
			return lipgloss.JoinVertical(lipgloss.Left, header, ghostCard)
		}
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

	visibleIdx := 0
	for i := startIdx; i < endIdx; i++ {
		if ghostCard != "" && visibleIdx == ghostInsertIdx {
			cards = append(cards, ghostCard)
		}
		active := isActive && i == m.curRow
		cards = append(cards, m.renderCard(col.issues[i], width, active))
		visibleIdx++
	}
	// If cursor is past the last card, append ghost at the end
	if ghostCard != "" && ghostInsertIdx >= visibleIdx {
		cards = append(cards, ghostCard)
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
	isDragged := m.dragging && issue.ID == m.dragIssueID

	style := common.StyleCard.Width(width - 2) // -2 for border chars
	if isDragged {
		style = common.StyleDragCard.Width(width - 2)
	} else if active {
		style = common.StyleActiveCard.Width(width - 2)
	}

	// Inner text width = card width - 2 (border) - 2 (border in Width) - 2 (padding)
	// lipgloss v2 Width() sets the total outer width including border.
	innerW := width - 6

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
	// Always render title as exactly 2 lines so all cards have the same height.
	title := titleLine1 + "\n" + titleLine2
	if isDragged {
		title = common.StyleFaint.Render(titleLine1 + "\n" + titleLine2)
	} else if active {
		title = common.StyleTitle.Render(title)
	}

	// Line 4: labels (compact, muted)
	var metaLine string
	var meta []string
	used := 0
	for _, l := range issue.Labels {
		lw := len([]rune(l))
		sep := 0
		if len(meta) > 0 {
			sep = 2 // "  "
		}
		if used+sep+lw > innerW {
			break
		}
		if isDragged {
			meta = append(meta, common.StyleFaint.Render(l))
		} else {
			meta = append(meta, common.RenderLabel(l))
		}
		used += sep + lw
	}
	if len(meta) > 0 {
		metaLine = strings.Join(meta, "  ")
	}

	// Line 5: Date — show "Updated" when sorting by updated, otherwise "Created"
	var dateLine string
	if m.sortMode == data.SortByUpdated && !issue.Updated.IsZero() {
		dateLine = common.StyleFaint.Render("Updated " + issue.Updated.Format("Jan 2 15:04"))
	} else if !issue.Created.IsZero() {
		dateLine = common.StyleFaint.Render("Created " + issue.Created.Format("Jan 2 15:04"))
	}

	// When dragged, force all content to faint
	if isDragged {
		line1 = common.StyleFaint.Render(fmt.Sprintf("#%d", issue.ID))
	}

	// Always include all 5 lines: ID, title (2), meta, date — for uniform card height.
	content := line1 + "\n" + title + "\n" + metaLine + "\n" + dateLine

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

// renderGhostCard renders a card in the ghost/dim style for the drop preview.
func (m Model) renderGhostCard(issue data.Issue, width int) string {
	style := common.StyleDragCard.Width(width - 2)
	innerW := width - 6

	idStr := common.StyleFaint.Render(fmt.Sprintf("#%d", issue.ID))
	prioIcon := common.PriorityStyle(issue.Priority).Render(
		strings.TrimSpace(common.PriorityIcon(issue.Priority)))
	line1 := idStr
	if issue.Priority <= data.PriorityHigh {
		line1 += " " + prioIcon
	}

	titleRunes := []rune(issue.Title)
	var titleLine1, titleLine2 string
	if len(titleRunes) <= innerW {
		titleLine1 = issue.Title
	} else {
		breakAt := innerW
		for i := innerW - 1; i > 0; i-- {
			if titleRunes[i] == ' ' {
				breakAt = i
				break
			}
		}
		titleLine1 = string(titleRunes[:breakAt])
		rest := titleRunes[breakAt:]
		if len(rest) > 0 && rest[0] == ' ' {
			rest = rest[1:]
		}
		titleLine2 = truncate(string(rest), innerW)
	}
	title := common.StyleFaint.Render(titleLine1 + "\n" + titleLine2)

	var metaLine string
	var meta []string
	used := 0
	for _, l := range issue.Labels {
		lw := len([]rune(l))
		sep := 0
		if len(meta) > 0 {
			sep = 2
		}
		if used+sep+lw > innerW {
			break
		}
		meta = append(meta, common.StyleFaint.Render(l))
		used += sep + lw
	}
	if len(meta) > 0 {
		metaLine = strings.Join(meta, "  ")
	}

	var dateLine string
	if m.sortMode == data.SortByUpdated && !issue.Updated.IsZero() {
		dateLine = common.StyleFaint.Render("Updated " + issue.Updated.Format("Jan 2 15:04"))
	} else if !issue.Created.IsZero() {
		dateLine = common.StyleFaint.Render("Created " + issue.Created.Format("Jan 2 15:04"))
	}

	content := line1 + "\n" + title + "\n" + metaLine + "\n" + dateLine
	return style.Render(content)
}

// findIssue looks up an issue by ID across all columns.
func (m Model) findIssue(id int) (data.Issue, bool) {
	for _, col := range m.columns {
		for _, iss := range col.issues {
			if iss.ID == id {
				return iss, true
			}
		}
	}
	return data.Issue{}, false
}

// overlayAt composites a small fg box on top of bg at position (x, y).
func (m Model) overlayAt(bg, fg string, x, y int) string {
	bgLines := strings.Split(bg, "\n")
	fgLines := strings.Split(fg, "\n")

	// Offset so the card appears below and to the right of the cursor
	startX := x + 1
	startY := y + 1

	// Measure fg width
	fgWidth := 0
	for _, line := range fgLines {
		if w := lipgloss.Width(line); w > fgWidth {
			fgWidth = w
		}
	}

	// Clamp to stay within bounds
	if startX+fgWidth > m.width {
		startX = m.width - fgWidth
	}
	if startX < 0 {
		startX = 0
	}
	if startY+len(fgLines) > len(bgLines) {
		startY = len(bgLines) - len(fgLines)
	}
	if startY < 0 {
		startY = 0
	}

	for i, fgLine := range fgLines {
		row := startY + i
		if row < 0 || row >= len(bgLines) {
			continue
		}
		bgLine := bgLines[row]

		// Left portion
		left := ansi.Truncate(bgLine, startX, "")
		leftW := lipgloss.Width(left)
		if leftW < startX {
			left += strings.Repeat(" ", startX-leftW)
		}

		// Right portion
		rightStart := startX + fgWidth
		right := ansi.TruncateLeft(bgLine, rightStart, "")

		bgLines[row] = left + fgLine + right
	}

	return strings.Join(bgLines, "\n")
}
