package tui

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Mibokess/grapes/internal/data"
	"github.com/Mibokess/grapes/internal/tui/board"
	"github.com/Mibokess/grapes/internal/tui/common"
	"github.com/Mibokess/grapes/internal/tui/detail"
	"github.com/Mibokess/grapes/internal/tui/list"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fsnotify/fsnotify"
)

type Model struct {
	issues     []data.Issue
	issuesDir  string
	width      int
	height     int
	screen     common.Screen
	prevScreen common.Screen
	watcher    *fsnotify.Watcher

	board  board.Model
	list   list.Model
	detail detail.Model
}

func NewModel(issues []data.Issue, issuesDir string) Model {
	w, _ := fsnotify.NewWatcher()
	if w != nil {
		addWatchDirs(w, issuesDir)
	}

	return Model{
		issues:    issues,
		issuesDir: issuesDir,
		screen:    common.ScreenBoard,
		board:     board.New(issues),
		list:      list.New(issues),
		watcher:   w,
	}
}

// addWatchDirs watches the issues directory and all numeric subdirectories.
func addWatchDirs(w *fsnotify.Watcher, issuesDir string) {
	w.Add(issuesDir)
	entries, err := os.ReadDir(issuesDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := strconv.Atoi(e.Name()); err != nil {
			continue
		}
		w.Add(filepath.Join(issuesDir, e.Name()))
	}
}

// watchCmd blocks on the fsnotify watcher and returns a RefreshMsg when files change.
// It debounces rapid events by draining for 100ms after the first event.
func (m Model) watchCmd() tea.Cmd {
	if m.watcher == nil {
		return nil
	}
	w := m.watcher
	return func() tea.Msg {
		for {
			select {
			case _, ok := <-w.Events:
				if !ok {
					return nil
				}
				// Debounce: drain events for 100ms
				timer := time.NewTimer(100 * time.Millisecond)
			drain:
				for {
					select {
					case _, ok := <-w.Events:
						if !ok {
							timer.Stop()
							return nil
						}
					case <-timer.C:
						break drain
					}
				}
				return common.RefreshMsg{}
			case _, ok := <-w.Errors:
				if !ok {
					return nil
				}
			}
		}
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.board.Init(), m.list.Init(), m.watchCmd())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		const overhead = 3 // app header (2 lines) + status bar (1 line)
		contentHeight := m.height - overhead
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

	case tea.MouseMsg:
		if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress && msg.Y == 0 {
			// Detect clicks on header tabs (Board / List)
			boardTabW := lipgloss.Width(common.StyleTabInactive.Render("Board"))
			listTabW := lipgloss.Width(common.StyleTabInactive.Render("List"))
			tabsStart := m.width - boardTabW - listTabW
			if msg.X >= tabsStart && msg.X < tabsStart+boardTabW {
				m.screen = common.ScreenBoard
				return m, nil
			}
			if msg.X >= tabsStart+boardTabW && msg.X < m.width {
				m.screen = common.ScreenList
				return m, nil
			}
		}
		// Non-tab clicks fall through to active screen delegation

	case common.OpenDetailMsg:
		var iss *data.Issue
		for i := range m.issues {
			if m.issues[i].ID == msg.ID {
				iss = &m.issues[i]
				break
			}
		}
		if iss != nil {
			// Only update prevScreen when coming from a non-detail screen,
			// so Esc from nested detail views returns to the original screen.
			if m.screen != common.ScreenDetail {
				m.prevScreen = m.screen
			}
			m.screen = common.ScreenDetail
			m.detail = detail.New(*iss, m.issues, m.width, m.height-3)
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
			return m, m.watchCmd()
		}
		m.issues = issues
		m.board = m.board.SetIssues(issues)
		m.list = m.list.SetIssues(issues)
		// Re-sync watched dirs (picks up new issue folders) and keep watching
		if m.watcher != nil {
			addWatchDirs(m.watcher, m.issuesDir)
		}
		return m, m.watchCmd()
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

func (m Model) renderHeader() string {
	title := common.StyleAppTitle.Render("grapes")

	// Active tab follows the current screen; detail inherits from previous screen.
	activeScreen := m.screen
	if activeScreen == common.ScreenDetail {
		activeScreen = m.prevScreen
	}

	var boardTab, listTab string
	if activeScreen == common.ScreenBoard {
		boardTab = common.StyleTabActive.Render("Board")
	} else {
		boardTab = common.StyleTabInactive.Render("Board")
	}
	if activeScreen == common.ScreenList {
		listTab = common.StyleTabActive.Render("List")
	} else {
		listTab = common.StyleTabInactive.Render("List")
	}

	tabs := lipgloss.JoinHorizontal(lipgloss.Top, boardTab, " ", listTab)
	spacerW := m.width - lipgloss.Width(title) - lipgloss.Width(tabs)
	if spacerW < 0 {
		spacerW = 0
	}
	row := title + strings.Repeat(" ", spacerW) + tabs
	sep := common.StyleSeparator.Render(strings.Repeat("━", m.width))
	return lipgloss.JoinVertical(lipgloss.Left, row, sep)
}

func (m Model) View() string {
	header := m.renderHeader()

	var content string
	var helpParts []string
	dot := common.StyleStatusSep.Render(" · ")

	switch m.screen {
	case common.ScreenBoard:
		content = m.board.View()
		helpParts = []string{
			common.FormatKeyHint("hjkl", "navigate"),
			common.FormatKeyHint("enter", "open"),
			common.FormatKeyHint("L", "list"),
			common.FormatKeyHint("r", "refresh"),
			common.FormatKeyHint("q", "quit"),
		}
	case common.ScreenList:
		content = m.list.View()
		helpParts = []string{
			common.FormatKeyHint("jk", "navigate"),
			common.FormatKeyHint("enter", "open"),
			common.FormatKeyHint("/", "filter"),
			common.FormatKeyHint("esc", "clear"),
			common.FormatKeyHint("b", "board"),
			common.FormatKeyHint("q", "quit"),
		}
	case common.ScreenDetail:
		content = m.detail.View()
		helpParts = []string{
			common.FormatKeyHint("jk", "scroll"),
			common.FormatKeyHint("esc", "back"),
			common.FormatKeyHint("b", "board"),
			common.FormatKeyHint("l", "list"),
			common.FormatKeyHint("q", "quit"),
		}
	}

	helpText := "  " + strings.Join(helpParts, dot)
	bar := common.StyleStatusBar.Width(m.width).Render(helpText)
	return lipgloss.JoinVertical(lipgloss.Left, header, content, bar)
}
