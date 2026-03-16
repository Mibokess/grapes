package detail

import (
	"fmt"
	"image/color"
	"strconv"
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
	field  string // "status", "priority", or "source:N"
}

type Model struct {
	issue         data.Issue
	allIssues     []data.Issue // all issues for rendering relationships
	viewport      viewport.Model
	ready         bool
	width         int
	height        int
	clickLines    map[int]int  // content line number → issue ID for clickable links
	clickZones    []clickZone  // rectangular click zones for metadata fields
	topOffset     int          // screen lines above this view's content (app header + filter bar)
	worktreeNames []string     // sorted worktree names for consistent color assignment
	theme         common.Theme
}

// SetWorktreeNames sets the sorted worktree names for color assignment.
// Re-renders the view if the issue has multiple sources (to show colored pills).
func (m Model) SetWorktreeNames(names []string) Model {
	m.worktreeNames = names
	if len(m.issue.Sources) > 1 {
		content, clickLines, clickZones := renderIssue(m.issue, m.allIssues, m.width, m.theme, names)
		m.viewport.SetContent(content)
		m.clickLines = clickLines
		m.clickZones = clickZones
	}
	return m
}

func New(issue data.Issue, allIssues []data.Issue, width, height int, theme common.Theme) Model {
	content, clickLines, clickZones := renderIssue(issue, allIssues, width, theme, nil)
	vp := viewport.New(viewport.WithWidth(width), viewport.WithHeight(height))
	vp.SetContent(content)

	return Model{
		issue:      issue,
		allIssues:  allIssues,
		viewport:   vp,
		ready:      true,
		width:      width,
		height:     height,
		clickLines: clickLines,
		clickZones: clickZones,
		theme:      theme,
	}
}

// UpdateIssue re-renders the detail view with updated data while preserving
// the current scroll position.
func (m Model) UpdateIssue(issue data.Issue, allIssues []data.Issue) Model {
	m.issue = issue
	m.allIssues = allIssues
	content, clickLines, clickZones := renderIssue(issue, allIssues, m.width, m.theme, m.worktreeNames)
	m.viewport.SetContent(content)
	m.clickLines = clickLines
	m.clickZones = clickZones
	return m
}

func (m Model) SetTheme(t common.Theme) Model {
	m.theme = t
	return m
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
	if m.ready {
		content, clickLines, clickZones := renderIssue(m.issue, m.allIssues, w, m.theme, m.worktreeNames)
		m.viewport.SetContent(content)
		m.clickLines = clickLines
		m.clickZones = clickZones
	}
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
		case key.Matches(msg, common.DetailKeyMap.Labels):
			return m, func() tea.Msg {
				return common.ShowLabelPickerMsg{IssueID: m.issue.ID}
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
				// Check click zones first (status/priority pickers, source switching)
				for _, zone := range m.clickZones {
					if contentLine == zone.line && mouse.X >= zone.xStart && mouse.X < zone.xEnd {
						field := zone.field
						if strings.HasPrefix(field, "source:") {
							idx, _ := strconv.Atoi(strings.TrimPrefix(field, "source:"))
							issueID := m.issue.ID
							return m, func() tea.Msg {
								return common.SwitchSourceMsg{IssueID: issueID, SourceIdx: idx}
							}
						}
						if field == "labels" {
							return m, func() tea.Msg {
								return common.ShowLabelPickerMsg{IssueID: m.issue.ID}
							}
						}
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

func renderIssue(issue data.Issue, allIssues []data.Issue, width int, theme common.Theme, wtNames []string) (string, map[int]int, []clickZone) {
	clickLines := make(map[int]int)
	var zones []clickZone
	var b strings.Builder

	// Title header
	idStr := theme.StyleFaint.Render(fmt.Sprintf("#%d", issue.ID))
	title := theme.StyleTitle.Render(issue.Title)
	b.WriteString(" " + idStr + "\n")
	b.WriteString(" " + title + "\n")

	// Source pills: show when issue exists in multiple sources
	if len(issue.Sources) > 1 {
		sourceLine := strings.Count(b.String(), "\n")
		xPos := 1 // leading space
		var pills []string
		for i, s := range issue.Sources {
			var label string
			if s.Name == "" {
				label = common.MainIcon() + " main"
			} else {
				label = common.WorktreeIcon() + " " + s.Name
			}

			var pill string
			if i == issue.ActiveSource {
				// Active source: bold, colored background
				var bg color.Color
				if s.Name == "" {
					bg = theme.ColorAccent
				} else {
					bg = theme.WorktreeColorFor(s.Name, wtNames)
				}
				pill = lipgloss.NewStyle().
					Foreground(theme.ColorContrast).
					Background(bg).
					Padding(0, 1).
					Bold(true).
					Render(label)
			} else {
				// Inactive source: muted
				var fg color.Color
				if s.Name == "" {
					fg = theme.ColorMuted
				} else {
					fg = theme.WorktreeColorFor(s.Name, wtNames)
				}
				pill = lipgloss.NewStyle().
					Foreground(fg).
					Padding(0, 1).
					Render(label)
			}

			pillW := lipgloss.Width(pill)
			zones = append(zones, clickZone{
				line:   sourceLine,
				xStart: xPos,
				xEnd:   xPos + pillW,
				field:  fmt.Sprintf("source:%d", i),
			})
			xPos += pillW + 1 // +1 for space between pills
			pills = append(pills, pill)
		}
		b.WriteString(" " + strings.Join(pills, " ") + "\n")
	} else if issue.Worktree != "" {
		wtBadge := theme.StyleWorktreeBadge.Render(common.WorktreeIcon() + " " + issue.Worktree)
		b.WriteString(" " + wtBadge + "\n")
	}
	b.WriteString("\n")

	// Metadata box: status pill + priority + labels + dates
	metaBoxW := width - 4
	if metaBoxW < 30 {
		metaBoxW = 30
	}
	var metaLines []string

	// Row 1: status pill + priority
	statusPill := theme.StatusPillStyle(issue.Status).
		Render(common.StatusIcon(issue.Status) + " " + issue.Status.Label())
	prioStr := theme.PriorityStyle(issue.Priority).
		Render(strings.TrimSpace(common.PriorityIcon(issue.Priority)) + " " + issue.Priority.Label())
	statusPillWidth := lipgloss.Width(statusPill)
	prioStrWidth := lipgloss.Width(prioStr)
	metaRow := statusPill + "  " + prioStr
	metaLines = append(metaLines, metaRow)

	// Row 2: labels (clickable)
	labelsLineIdx := -1
	if len(issue.Labels) > 0 {
		labelsLineIdx = len(metaLines)
		var labelStrs []string
		for _, l := range issue.Labels {
			labelStrs = append(labelStrs, theme.RenderLabelPill(l))
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
		metaLines = append(metaLines, theme.StyleFaint.Render(strings.Join(dateParts, "  ·  ")))
	}

	// Track clickable lines within meta box (lineIdx → issueID)
	type metaClick struct {
		lineIdx int
		issueID int
	}
	var metaClicks []metaClick

	// Row 4: parent link
	if issue.Parent != nil {
		parentTitle, _, _ := findRelatedIssue(allIssues, *issue.Parent, issue.Worktree)
		parentLink := theme.StyleSectionHeader.Render("↑") +
			theme.StyleFaint.Render(" Parent: ") +
			theme.StyleSectionHeader.Render(fmt.Sprintf("#%d", *issue.Parent)) +
			"  " + theme.StyleSubtitle.Render(parentTitle)
		metaClicks = append(metaClicks, metaClick{lineIdx: len(metaLines), issueID: *issue.Parent})
		metaLines = append(metaLines, parentLink)
	}

	// Row 5: blocked by
	for _, blockerID := range issue.BlockedBy {
		blockerTitle, _, _ := findRelatedIssue(allIssues, blockerID, issue.Worktree)
		link := theme.StyleSectionHeader.Render("⊘") +
			theme.StyleFaint.Render(" Blocked by: ") +
			theme.StyleSectionHeader.Render(fmt.Sprintf("#%d", blockerID)) +
			"  " + theme.StyleSubtitle.Render(blockerTitle)
		metaClicks = append(metaClicks, metaClick{lineIdx: len(metaLines), issueID: blockerID})
		metaLines = append(metaLines, link)
	}

	// Row 6: blocks
	for _, blockedID := range issue.Blocks {
		blockedTitle, _, _ := findRelatedIssue(allIssues, blockedID, issue.Worktree)
		link := theme.StyleSectionHeader.Render("▸") +
			theme.StyleFaint.Render(" Blocks: ") +
			theme.StyleSectionHeader.Render(fmt.Sprintf("#%d", blockedID)) +
			"  " + theme.StyleSubtitle.Render(blockedTitle)
		metaClicks = append(metaClicks, metaClick{lineIdx: len(metaLines), issueID: blockedID})
		metaLines = append(metaLines, link)
	}

	metaBoxStartLine := strings.Count(b.String(), "\n")
	metaContent := strings.Join(metaLines, "\n")
	metaBox := theme.StyleMetaBox.Width(metaBoxW).Render(metaContent)
	b.WriteString(metaBox + "\n")

	// Register click lines for all links inside the meta box
	for _, mc := range metaClicks {
		clickLines[metaBoxStartLine+1+mc.lineIdx] = mc.issueID
	}

	// Register click zones for status pill and priority text
	// Meta box content X offset: MarginLeft(1) + Border(1) + Padding(1) = 3
	const metaContentXOffset = 3
	statusPrioLine := metaBoxStartLine + 1
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

	// Register click zone for labels row
	if labelsLineIdx >= 0 {
		labelsLine := metaBoxStartLine + 1 + labelsLineIdx
		zones = append(zones, clickZone{
			line:   labelsLine,
			xStart: metaContentXOffset,
			xEnd:   metaContentXOffset + metaBoxW, // entire row is clickable
			field:  "labels",
		})
	}

	b.WriteString("\n")

	mdWidth := width - 4
	if mdWidth < 40 {
		mdWidth = 40
	}

	sectionUnderline := theme.StyleSectionHeader.Render(strings.Repeat("━", 2))

	if issue.Content != "" {
		b.WriteString(" " + theme.StyleSectionHeader.Render("Description") + " " + sectionUnderline + "\n\n")
		rendered := renderMarkdown(issue.Content, mdWidth, theme.GlamourStyle)
		b.WriteString(rendered + "\n")
	}

	if len(issue.Children) > 0 {
		b.WriteString(" " + theme.StyleSectionHeader.Render("Sub-issues") + " " + sectionUnderline + "\n\n")
		for _, childID := range issue.Children {
			childTitle, childStatus, childLabels := findRelatedIssue(allIssues, childID, issue.Worktree)
			icon := theme.StatusStyle(childStatus).Render(common.StatusIcon(childStatus) + " " + childStatus.Label())
			lineNum := strings.Count(b.String(), "\n")
			clickLines[lineNum] = childID
			labelStr := ""
			for _, l := range childLabels {
				labelStr += "  " + theme.RenderLabelPill(l)
			}
			b.WriteString(fmt.Sprintf("  %s  #%d  %s%s\n", icon, childID, theme.StyleSubtitle.Render(childTitle), labelStr))
		}
		b.WriteString("\n")
	}

	if len(issue.Comments) > 0 {
		b.WriteString(" " + theme.StyleSectionHeader.Render(fmt.Sprintf("Comments (%d)", len(issue.Comments))) + " " + sectionUnderline + "\n\n")

		commentW := width - 6
		if commentW < 30 {
			commentW = 30
		}
		for _, c := range issue.Comments {
			commentBox := theme.StyleCommentBox.Width(commentW).
				Render(
					theme.StyleFaint.Render(c.Date) + "\n" +
						renderMarkdown(c.Body, commentW-4, theme.GlamourStyle),
				)
			b.WriteString(commentBox + "\n\n")
		}
	}

	return b.String(), clickLines, zones
}

// findRelatedIssue returns title, status, and labels for a related issue,
// preferring the source that matches the viewing issue's worktree name.
func findRelatedIssue(allIssues []data.Issue, id int, worktree string) (string, data.Status, []string) {
	for i := range allIssues {
		if allIssues[i].ID != id {
			continue
		}
		for _, s := range allIssues[i].Sources {
			if s.Name == worktree {
				return s.Title, s.Status, s.Labels
			}
		}
		return allIssues[i].Title, allIssues[i].Status, allIssues[i].Labels
	}
	return "", "", nil
}

func renderMarkdown(content string, width int, glamourStyle string) string {
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(glamourStyle),
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
