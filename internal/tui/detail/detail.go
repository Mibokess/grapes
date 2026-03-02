package detail

import (
	"fmt"
	"strings"

	"github.com/Mibokess/grapes/internal/data"
	"github.com/Mibokess/grapes/internal/tui/common"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/glamour"
	"github.com/muesli/termenv"
)

// clickZone represents a clickable rectangular region in the rendered content.
type clickZone struct {
	line   int    // content line number
	xStart int    // start X position (inclusive, screen coordinates)
	xEnd   int    // end X position (exclusive, screen coordinates)
	field  string // "status" or "priority"
}

type Model struct {
	issue      data.Issue
	viewport   viewport.Model
	ready      bool
	width      int
	height     int
	clickLines map[int]int  // content line number → issue ID for clickable links
	clickZones []clickZone  // rectangular click zones for metadata fields
	topOffset  int          // screen lines above this view's content (app header + filter bar)
}

func New(issue data.Issue, allIssues []data.Issue, width, height int) Model {
	content, clickLines, clickZones := renderIssue(issue, allIssues, width)
	vp := viewport.New(viewport.WithWidth(width), viewport.WithHeight(height))
	vp.SetContent(content)

	return Model{
		issue:      issue,
		viewport:   vp,
		ready:      true,
		width:      width,
		height:     height,
		clickLines: clickLines,
		clickZones: clickZones,
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) SetTopOffset(n int) Model {
	m.topOffset = n
	return m
}

func (m Model) SetSize(w, h int) Model {
	m.width = w
	m.height = h
	m.viewport.SetWidth(w)
	m.viewport.SetHeight(h)
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, common.DetailKeyMap.Back):
			return m, func() tea.Msg { return common.GoBackMsg{} }
		case key.Matches(msg, common.DetailKeyMap.ToBoard):
			return m, func() tea.Msg { return common.SwitchScreenMsg{Screen: common.ScreenBoard} }
		case key.Matches(msg, common.DetailKeyMap.ToList):
			return m, func() tea.Msg { return common.SwitchScreenMsg{Screen: common.ScreenList} }
		case key.Matches(msg, common.DetailKeyMap.CycleStatus):
			return m, func() tea.Msg {
				return common.ShowPickerMsg{IssueID: m.issue.ID, Field: "status"}
			}
		case key.Matches(msg, common.DetailKeyMap.CyclePriority):
			return m, func() tea.Msg {
				return common.ShowPickerMsg{IssueID: m.issue.ID, Field: "priority"}
			}
		case key.Matches(msg, common.DetailKeyMap.EditIssue):
			return m, func() tea.Msg {
				return common.LaunchEditMsg{ID: m.issue.ID}
			}
		case key.Matches(msg, common.DetailKeyMap.AddComment):
			return m, func() tea.Msg {
				return common.LaunchEditorMsg{ID: m.issue.ID}
			}
		}
	case tea.MouseClickMsg:
		mouse := msg.Mouse()
		if msg.Button == tea.MouseBackward {
			return m, func() tea.Msg { return common.GoBackMsg{} }
		}
		if msg.Button == tea.MouseLeft {
			viewportY := mouse.Y - m.topOffset
			if viewportY >= 0 && viewportY < m.viewport.Height() {
				contentLine := m.viewport.YOffset() + viewportY
				// Check click zones first (status/priority pickers)
				for _, zone := range m.clickZones {
					if contentLine == zone.line && mouse.X >= zone.xStart && mouse.X < zone.xEnd {
						field := zone.field
						return m, func() tea.Msg {
							return common.ShowPickerMsg{IssueID: m.issue.ID, Field: field}
						}
					}
				}
				// Fall through to line-based click links (issue navigation)
				if id, ok := m.clickLines[contentLine]; ok {
					return m, func() tea.Msg { return common.OpenDetailMsg{ID: id} }
				}
			}
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// IssueID returns the ID of the issue being displayed.
func (m Model) IssueID() int { return m.issue.ID }

func (m Model) View() string {
	return m.viewport.View()
}

func renderIssue(issue data.Issue, allIssues []data.Issue, width int) (string, map[int]int, []clickZone) {
	clickLines := make(map[int]int)
	var b strings.Builder

	// Title header
	idStr := common.StyleFaint.Render(fmt.Sprintf("#%d", issue.ID))
	title := common.StyleTitle.Render(issue.Title)
	b.WriteString(" " + idStr + "\n")
	b.WriteString(" " + title + "\n\n")

	// Metadata box: status pill + priority + labels + dates
	metaBoxW := width - 4
	if metaBoxW < 30 {
		metaBoxW = 30
	}
	var metaLines []string

	// Row 1: status pill + priority
	statusPill := common.StatusPillStyle(issue.Status).
		Render(common.StatusIcon(issue.Status) + " " + issue.Status.Label())
	prioStr := common.PriorityStyle(issue.Priority).
		Render(strings.TrimSpace(common.PriorityIcon(issue.Priority)) + " " + issue.Priority.Label())
	statusPillWidth := lipgloss.Width(statusPill)
	prioStrWidth := lipgloss.Width(prioStr)
	metaRow := statusPill + "  " + prioStr
	metaLines = append(metaLines, metaRow)

	// Row 2: labels
	if len(issue.Labels) > 0 {
		var labelStrs []string
		for _, l := range issue.Labels {
			labelStrs = append(labelStrs, common.RenderLabelPill(l))
		}
		metaLines = append(metaLines, strings.Join(labelStrs, " "))
	}

	// Row 3: dates
	var dateParts []string
	if !issue.Created.IsZero() {
		dateParts = append(dateParts, "Created "+issue.Created.Format("2006-01-02 15:04"))
	}
	if !issue.Updated.IsZero() {
		dateParts = append(dateParts, "Updated "+issue.Updated.Format("2006-01-02 15:04"))
	}
	if len(dateParts) > 0 {
		metaLines = append(metaLines, common.StyleFaint.Render(strings.Join(dateParts, "  ·  ")))
	}

	// Track clickable lines within meta box (lineIdx → issueID)
	type metaClick struct {
		lineIdx int
		issueID int
	}
	var metaClicks []metaClick

	// Row 4: parent link
	if issue.Parent != nil {
		parentTitle := ""
		for _, iss := range allIssues {
			if iss.ID == *issue.Parent {
				parentTitle = iss.Title
				break
			}
		}
		parentLink := common.StyleSectionHeader.Render("↑") +
			common.StyleFaint.Render(" Parent: ") +
			common.StyleSectionHeader.Render(fmt.Sprintf("#%d", *issue.Parent)) +
			"  " + common.StyleSubtitle.Render(parentTitle)
		metaClicks = append(metaClicks, metaClick{lineIdx: len(metaLines), issueID: *issue.Parent})
		metaLines = append(metaLines, parentLink)
	}

	// Row 5: blocked by
	for _, blockerID := range issue.BlockedBy {
		blockerTitle := ""
		for _, iss := range allIssues {
			if iss.ID == blockerID {
				blockerTitle = iss.Title
				break
			}
		}
		link := common.StyleSectionHeader.Render("⊘") +
			common.StyleFaint.Render(" Blocked by: ") +
			common.StyleSectionHeader.Render(fmt.Sprintf("#%d", blockerID)) +
			"  " + common.StyleSubtitle.Render(blockerTitle)
		metaClicks = append(metaClicks, metaClick{lineIdx: len(metaLines), issueID: blockerID})
		metaLines = append(metaLines, link)
	}

	// Row 6: blocks
	for _, blockedID := range issue.Blocks {
		blockedTitle := ""
		for _, iss := range allIssues {
			if iss.ID == blockedID {
				blockedTitle = iss.Title
				break
			}
		}
		link := common.StyleSectionHeader.Render("▸") +
			common.StyleFaint.Render(" Blocks: ") +
			common.StyleSectionHeader.Render(fmt.Sprintf("#%d", blockedID)) +
			"  " + common.StyleSubtitle.Render(blockedTitle)
		metaClicks = append(metaClicks, metaClick{lineIdx: len(metaLines), issueID: blockedID})
		metaLines = append(metaLines, link)
	}

	metaBoxStartLine := strings.Count(b.String(), "\n")
	metaContent := strings.Join(metaLines, "\n")
	metaBox := common.StyleMetaBox.Width(metaBoxW).Render(metaContent)
	b.WriteString(metaBox + "\n")

	// Register click lines for all links inside the meta box
	for _, mc := range metaClicks {
		clickLines[metaBoxStartLine+1+mc.lineIdx] = mc.issueID
	}

	// Register click zones for status pill and priority text
	// Meta box content X offset: MarginLeft(1) + Border(1) + Padding(1) = 3
	const metaContentXOffset = 3
	statusPrioLine := metaBoxStartLine + 1
	var zones []clickZone
	zones = append(zones, clickZone{
		line:   statusPrioLine,
		xStart: metaContentXOffset,
		xEnd:   metaContentXOffset + statusPillWidth,
		field:  "status",
	})
	zones = append(zones, clickZone{
		line:   statusPrioLine,
		xStart: metaContentXOffset + statusPillWidth + 2, // +2 for "  " separator
		xEnd:   metaContentXOffset + statusPillWidth + 2 + prioStrWidth,
		field:  "priority",
	})

	b.WriteString("\n")

	mdWidth := width - 4
	if mdWidth < 40 {
		mdWidth = 40
	}

	sectionUnderline := common.StyleSectionHeader.Render(strings.Repeat("━", 2))

	if issue.Content != "" {
		b.WriteString(" " + common.StyleSectionHeader.Render("Description") + " " + sectionUnderline + "\n\n")
		rendered := renderMarkdown(issue.Content, mdWidth)
		b.WriteString(rendered + "\n")
	}

	if len(issue.Children) > 0 {
		b.WriteString(" " + common.StyleSectionHeader.Render("Sub-issues") + " " + sectionUnderline + "\n\n")
		for _, childID := range issue.Children {
			childTitle := ""
			childStatus := data.Status("")
			for _, iss := range allIssues {
				if iss.ID == childID {
					childTitle = iss.Title
					childStatus = iss.Status
					break
				}
			}
			icon := common.StatusStyle(childStatus).Render(common.StatusIcon(childStatus))
			lineNum := strings.Count(b.String(), "\n")
			clickLines[lineNum] = childID
			b.WriteString(fmt.Sprintf("  %s  #%d  %s\n", icon, childID, common.StyleSubtitle.Render(childTitle)))
		}
		b.WriteString("\n")
	}

	if len(issue.Comments) > 0 {
		b.WriteString(" " + common.StyleSectionHeader.Render(fmt.Sprintf("Comments (%d)", len(issue.Comments))) + " " + sectionUnderline + "\n\n")

		commentW := width - 6
		if commentW < 30 {
			commentW = 30
		}
		for _, c := range issue.Comments {
			commentBox := common.StyleCommentBox.Width(commentW).
				Render(
					common.StyleFaint.Render(c.Date) + "\n" +
						renderMarkdown(c.Body, commentW-4),
				)
			b.WriteString(commentBox + "\n\n")
		}
	}

	return b.String(), clickLines, zones
}

func renderMarkdown(content string, width int) string {
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(width),
		glamour.WithColorProfile(termenv.TrueColor),
	)
	if err != nil {
		return content
	}
	out, err := r.Render(content)
	if err != nil {
		return content
	}
	return strings.TrimRight(out, "\n")
}
