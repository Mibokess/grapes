package tui

import (
	"github.com/Mibokess/grapes/internal/data"
	"github.com/Mibokess/grapes/internal/tui/board"
	"github.com/Mibokess/grapes/internal/tui/common"
	"github.com/Mibokess/grapes/internal/tui/detail"
	"github.com/Mibokess/grapes/internal/tui/list"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	issues     []data.Issue
	issuesDir  string
	width      int
	height     int
	screen     common.Screen
	prevScreen common.Screen

	board  board.Model
	list   list.Model
	detail detail.Model
}

func NewModel(issues []data.Issue, issuesDir string) Model {
	return Model{
		issues:    issues,
		issuesDir: issuesDir,
		screen:    common.ScreenBoard,
		board:     board.New(issues),
		list:      list.New(issues),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.board.Init(), m.list.Init())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		helpHeight := 1
		contentHeight := m.height - helpHeight
		m.board = m.board.SetSize(m.width, contentHeight)
		m.list = m.list.SetSize(m.width, contentHeight)
		m.detail = m.detail.SetSize(m.width, contentHeight)
		return m, nil

	case tea.KeyMsg:
		// Global quit — but not when filtering in list view
		if m.screen == common.ScreenList && m.list.Filtering() {
			break // fall through to screen-specific handler
		}
		if key.Matches(msg, common.GlobalKeyMap.Quit) {
			return m, tea.Quit
		}

	case common.OpenDetailMsg:
		var iss *data.Issue
		for i := range m.issues {
			if m.issues[i].ID == msg.ID {
				iss = &m.issues[i]
				break
			}
		}
		if iss != nil {
			m.prevScreen = m.screen
			m.screen = common.ScreenDetail
			m.detail = detail.New(*iss, m.issues, m.width, m.height-1)
			return m, m.detail.Init()
		}
		return m, nil

	case common.GoBackMsg:
		m.screen = m.prevScreen
		return m, nil

	case common.SwitchScreenMsg:
		m.screen = msg.Screen
		return m, nil

	case common.RefreshMsg:
		issues, err := data.LoadAllIssues(m.issuesDir)
		if err != nil {
			return m, nil
		}
		m.issues = issues
		m.board = m.board.SetIssues(issues)
		m.list = m.list.SetIssues(issues)
		return m, nil
	}

	// Delegate to active screen
	var cmd tea.Cmd
	switch m.screen {
	case common.ScreenBoard:
		m.board, cmd = m.board.Update(msg)
	case common.ScreenList:
		m.list, cmd = m.list.Update(msg)
	case common.ScreenDetail:
		m.detail, cmd = m.detail.Update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	var content string
	var helpText string

	switch m.screen {
	case common.ScreenBoard:
		content = m.board.View()
		helpText = "  hjkl/arrows: navigate • enter: open • L: list • r: refresh • q: quit"
	case common.ScreenList:
		content = m.list.View()
		helpText = "  j/k: navigate • enter: open • /: filter • esc: clear • b: board • r: refresh • q: quit"
	case common.ScreenDetail:
		content = m.detail.View()
		helpText = "  j/k: scroll • esc: back • b: board • l: list • q: quit"
	}

	bar := common.StyleStatusBar.Width(m.width).Render(helpText)
	return lipgloss.JoinVertical(lipgloss.Left, content, bar)
}
