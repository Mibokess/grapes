package list

import (
	"fmt"
	"strings"
	"time"

	"github.com/Mibokess/grapes/internal/data"
	"github.com/Mibokess/grapes/internal/tui/common"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// columnSortModes maps table column indices to their sort modes.
// Column 6 (Labels) has no sort mode (-1).
var columnSortModes = []data.SortMode{
	data.SortByID,
	data.SortByTitle,
	data.SortByStatus,
	data.SortByPriority,
	data.SortByCreated,
	data.SortByUpdated,
	-1, // Labels — not sortable
}

// stickyWidth is the rendered width of the sticky ID column (content + padding).
const stickyWidth = 6 // ID column width (4) + 2 padding

type Model struct {
	allIssues    []data.Issue
	table        table.Model
	filter       textinput.Model
	filtering    bool
	width        int
	height       int
	visibleStart int // first visible row index, mirrors table's internal start
	scrollX      int // horizontal scroll offset for columns after the sticky ID
	sortMode     data.SortMode
	sortAsc      bool
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

func (m Model) Filtering() bool    { return m.filtering }
func (m Model) HScrollActive() bool { return m.needsHScroll() }

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
	m.clampScrollX()
	return m
}

func (m Model) SetIssues(issues []data.Issue) Model {
	m.allIssues = issues
	m.table = m.buildTable(m.filteredIssues(), m.width, m.height-3)
	return m
}

func (m Model) SetSortState(mode data.SortMode, asc bool) Model {
	m.sortMode = mode
	m.sortAsc = asc
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
		case key.Matches(msg, common.ListKeyMap.ScrollLeft):
			if m.needsHScroll() {
				m.scrollX -= 8
				if m.scrollX < 0 {
					m.scrollX = 0
				}
				return m, nil
			}
		case key.Matches(msg, common.ListKeyMap.ScrollRight):
			if m.needsHScroll() {
				m.scrollX += 8
				m.clampScrollX()
				return m, nil
			}
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
			// Header row is right after app header (+filter line if shown).
			headerY := common.AppHeaderHeight
			if m.filter.Value() != "" {
				headerY = common.AppHeaderHeight + 1
			}
			// 2-line app header + 2-line table header = 4; +1 if filter line shown.
			tableTopY := common.AppHeaderHeight + 2
			if m.filter.Value() != "" {
				tableTopY = common.AppHeaderHeight + 3
			}
			if mouse.Y == headerY {
				// Click on column header → sort by that column
				col := m.columnAtX(mouse.X)
				if col >= 0 && col < len(columnSortModes) {
					mode := columnSortModes[col]
					if mode >= 0 {
						return m, func() tea.Msg { return common.ColumnSortMsg{Mode: mode} }
					}
				}
			} else if mouse.Y >= tableTopY {
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

	if m.needsHScroll() {
		tableView = m.applyHScroll(tableView)
	}

	if filterLine != "" {
		return lipgloss.JoinVertical(lipgloss.Left, filterLine, tableView)
	}
	return tableView
}

// applyHScroll post-processes the table output to create a sticky ID column
// and a horizontally scrollable area for the remaining columns.
// A thin vertical line separates the frozen ID pane from the scrollable area.
func (m Model) applyHScroll(view string) string {
	lines := strings.Split(view, "\n")
	sep := common.StyleFaint.Render("│")
	sepW := lipgloss.Width(sep)
	avail := m.width - stickyWidth - sepW
	if avail < 1 {
		avail = 1
	}

	scrollX := m.scrollX
	if max := m.maxScrollX(); scrollX > max {
		scrollX = max
	}

	var out []string
	for _, line := range lines {
		if lipgloss.Width(line) == 0 {
			out = append(out, line)
			continue
		}

		// Extract sticky ID column
		sticky := ansi.Truncate(line, stickyWidth, "")
		if w := lipgloss.Width(sticky); w < stickyWidth {
			sticky += strings.Repeat(" ", stickyWidth-w)
		}

		// Get the scrollable rest, applying horizontal offset
		rest := ansi.TruncateLeft(line, stickyWidth, "")
		if scrollX > 0 {
			rest = ansi.TruncateLeft(rest, scrollX, "")
		}
		rest = ansi.Truncate(rest, avail, "")

		out = append(out, sticky+sep+rest)
	}
	return strings.Join(out, "\n")
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

// columnAtX returns the column index for a given X coordinate in the rendered view.
// Accounts for horizontal scroll offset when scrolling is active.
func (m Model) columnAtX(x int) int {
	cw := m.colWidths()
	// When scrolling, map viewport X to full-table X
	if m.needsHScroll() && x >= stickyWidth {
		x = x + m.scrollX
	}
	cumX := 0
	for i, w := range cw {
		cumX += w + 2
		if x < cumX {
			return i
		}
	}
	return len(cw) - 1
}

// colWidths returns the column widths. The title column gets a minimum of 30
// so content isn't excessively truncated even on narrow terminals (horizontal
// scrolling takes over instead).
func (m Model) colWidths() []int {
	w := m.width
	if w < 40 {
		w = 40
	}
	// Fixed widths: 4+13+9+10+10+15 = 61; cell padding: 7 cols × 2 = 14; total overhead = 75.
	titleW := w - 75
	if titleW < 20 {
		titleW = 20
	}
	return []int{4, titleW, 13, 9, 10, 10, 15}
}

// tableFullWidth returns the full rendered width of all columns with padding.
func (m Model) tableFullWidth() int {
	total := 0
	for _, w := range m.colWidths() {
		total += w + 2
	}
	return total
}

// needsHScroll returns true when the full table is wider than the terminal.
func (m Model) needsHScroll() bool {
	return m.tableFullWidth() > m.width
}

// maxScrollX returns the maximum horizontal scroll offset.
func (m Model) maxScrollX() int {
	max := m.tableFullWidth() - m.width
	if max < 0 {
		return 0
	}
	return max
}

// clampScrollX ensures scrollX is within valid bounds.
func (m *Model) clampScrollX() {
	if max := m.maxScrollX(); m.scrollX > max {
		m.scrollX = max
	}
	if m.scrollX < 0 {
		m.scrollX = 0
	}
}

func (m Model) sortIndicator(col data.SortMode) string {
	if m.sortMode != col {
		return ""
	}
	if m.sortAsc {
		return "▲"
	}
	return "▼"
}

func (m Model) buildTable(issues []data.Issue, width, height int) table.Model {
	cw := m.colWidths()

	cols := []table.Column{
		{Title: "ID" + m.sortIndicator(data.SortByID), Width: cw[0]},
		{Title: "Title" + m.sortIndicator(data.SortByTitle), Width: cw[1]},
		{Title: "Status" + m.sortIndicator(data.SortByStatus), Width: cw[2]},
		{Title: "Priority" + m.sortIndicator(data.SortByPriority), Width: cw[3]},
		{Title: "Created" + m.sortIndicator(data.SortByCreated), Width: cw[4]},
		{Title: "Updated" + m.sortIndicator(data.SortByUpdated), Width: cw[5]},
		{Title: "Labels", Width: cw[6]},
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
		createdCell := formatDate(iss.Created)
		updatedCell := formatDate(iss.Updated)
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", iss.ID),
			iss.Title,
			statusCell,
			prioCell,
			createdCell,
			updatedCell,
			labels,
		})
	}

	// Use the wider of terminal width or natural table width so
	// the bubbles/table doesn't truncate columns — we handle the
	// viewport ourselves via applyHScroll.
	tableW := m.tableFullWidth()
	if width > tableW {
		tableW = width
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithWidth(tableW),
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

// formatDate renders a time as a compact date string.
// Zero times render as "—".
func formatDate(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	return t.Format("Jan 02 '06")
}
