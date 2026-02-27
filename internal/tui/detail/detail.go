package detail

import (
	"fmt"
	"strings"

	"github.com/Mibokess/grapes/internal/data"
	"github.com/Mibokess/grapes/internal/tui/common"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	issue    data.Issue
	viewport viewport.Model
	ready    bool
	width    int
	height   int
}

func New(issue data.Issue, allIssues []data.Issue, width, height int) Model {
	vp := viewport.New(width, height)
	vp.SetContent(renderIssue(issue, allIssues, width))

	return Model{
		issue:    issue,
		viewport: vp,
		ready:    true,
		width:    width,
		height:   height,
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) SetSize(w, h int) Model {
	m.width = w
	m.height = h
	m.viewport.Width = w
	m.viewport.Height = h
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, common.DetailKeyMap.Back):
			return m, func() tea.Msg { return common.GoBackMsg{} }
		case key.Matches(msg, common.DetailKeyMap.ToBoard):
			return m, func() tea.Msg { return common.SwitchScreenMsg{Screen: common.ScreenBoard} }
		case key.Matches(msg, common.DetailKeyMap.ToList):
			return m, func() tea.Msg { return common.SwitchScreenMsg{Screen: common.ScreenList} }
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return m.viewport.View()
}

func renderIssue(issue data.Issue, allIssues []data.Issue, width int) string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#FAFAFA"}).
		Render(fmt.Sprintf("#%d  %s", issue.ID, issue.Title))
	b.WriteString(title + "\n\n")

	labelStyle := lipgloss.NewStyle().Bold(true)
	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#666666", Dark: "#999999"})

	statusStr := common.StatusStyle(issue.Status).Render(issue.Status.Label())
	prioStr := common.PriorityStyle(issue.Priority).Render(issue.Priority.Label())

	b.WriteString(labelStyle.Render("Status:   ") + statusStr + "\n")
	b.WriteString(labelStyle.Render("Priority: ") + prioStr + "\n")
	if issue.Assignee != "" {
		b.WriteString(labelStyle.Render("Assignee: ") + metaStyle.Render(issue.Assignee) + "\n")
	}
	if len(issue.Labels) > 0 {
		var labels []string
		for _, l := range issue.Labels {
			labels = append(labels, common.StyleLabel.Render(l))
		}
		b.WriteString(labelStyle.Render("Labels:   ") + strings.Join(labels, " ") + "\n")
	}
	if !issue.Created.IsZero() {
		b.WriteString(labelStyle.Render("Created:  ") + metaStyle.Render(issue.Created.Format("2006-01-02")) + "\n")
	}
	if !issue.Updated.IsZero() {
		b.WriteString(labelStyle.Render("Updated:  ") + metaStyle.Render(issue.Updated.Format("2006-01-02")) + "\n")
	}

	if issue.Parent != nil {
		parentTitle := ""
		for _, iss := range allIssues {
			if iss.ID == *issue.Parent {
				parentTitle = iss.Title
				break
			}
		}
		b.WriteString(labelStyle.Render("Parent:   ") + metaStyle.Render(fmt.Sprintf("#%d %s", *issue.Parent, parentTitle)) + "\n")
	}

	b.WriteString("\n")

	sectionStyle := lipgloss.NewStyle().Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "#6C40BF", Dark: "#B48EF7"})

	if issue.Content != "" {
		b.WriteString(sectionStyle.Render("── Description ──") + "\n\n")

		mdWidth := width - 4
		if mdWidth < 40 {
			mdWidth = 40
		}
		rendered := renderMarkdown(issue.Content, mdWidth)
		b.WriteString(rendered + "\n")
	}

	if len(issue.Children) > 0 {
		b.WriteString(sectionStyle.Render("── Sub-issues ──") + "\n\n")

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
			statusStr := common.StatusStyle(childStatus).Render(childStatus.Label())
			b.WriteString(fmt.Sprintf("  #%d  %s  [%s]\n", childID, childTitle, statusStr))
		}
		b.WriteString("\n")
	}

	if len(issue.Comments) > 0 {
		b.WriteString(sectionStyle.Render(fmt.Sprintf("── Comments (%d) ──", len(issue.Comments))) + "\n\n")

		commentHeaderStyle := lipgloss.NewStyle().Bold(true)

		mdWidth := width - 4
		if mdWidth < 40 {
			mdWidth = 40
		}

		for _, c := range issue.Comments {
			header := commentHeaderStyle.Render(fmt.Sprintf("%s — %s", c.Author, c.Date))
			b.WriteString(header + "\n")
			rendered := renderMarkdown(c.Body, mdWidth)
			b.WriteString(rendered + "\n")
		}
	}

	return b.String()
}

func renderMarkdown(content string, width int) string {
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
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
