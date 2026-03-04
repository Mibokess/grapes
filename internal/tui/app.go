package tui

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"sort"

	"github.com/Mibokess/grapes/internal/config"
	"github.com/Mibokess/grapes/internal/data"
	"github.com/Mibokess/grapes/internal/tui/board"
	"github.com/Mibokess/grapes/internal/tui/common"
	"github.com/Mibokess/grapes/internal/tui/detail"
	"github.com/Mibokess/grapes/internal/tui/filter"
	"github.com/Mibokess/grapes/internal/tui/list"
	"github.com/Mibokess/grapes/internal/tui/picker"
	"github.com/Mibokess/grapes/internal/tui/settings"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/fsnotify/fsnotify"
)

// clearStatusMsg is sent after a delay to clear transient status bar messages.
type clearStatusMsg struct{}

// navEntry captures one frame in the navigation history.
type navEntry struct {
	screen common.Screen
	detail detail.Model // only meaningful when screen == ScreenDetail
}

type Model struct {
	issues      []data.Issue
	issuesDir   string
	projectRoot string
	width       int
	height      int
	screen     common.Screen
	navStack   []navEntry
	watcher    *fsnotify.Watcher
	sortMode data.SortMode
	sortAsc  bool // ascending order (reversed from default)
	theme    common.Theme

	cfg    config.Config
	board  board.Model
	list   list.Model
	detail detail.Model
	settings settings.Model

	picker       *picker.Model       // non-nil when picker overlay is active
	filterMenu   *filter.Menu        // non-nil when filter menu is open
	filterPicker *filter.MultiPicker // non-nil when filter multi-picker is open
	filters      filter.FilterSet    // structured filter state

	statusMsg      string // transient error/info message for status bar
	editingIssueID int    // issue ID for in-progress editor session
	editingTmpPath string // temp file path for editor
	editingMode    string // "comment" or "edit"
}

func NewModel(issues []data.Issue, issuesDir string, cfg config.Config) Model {
	projectRoot := data.ProjectRoot(issuesDir)

	w, _ := fsnotify.NewWatcher()
	if w != nil {
		addWatchDirs(w, issuesDir)
		for _, dir := range data.FindWorktreeIssuesDirs(projectRoot) {
			addWatchDirs(w, dir)
		}
	}

	sortMode := data.SortByPriority
	data.SortIssues(issues, sortMode, false)

	filters := filter.Default()
	var filtered []data.Issue
	for _, iss := range issues {
		if filters.Matches(iss) {
			filtered = append(filtered, iss)
		}
	}

	l := list.New(filtered)
	l = l.SetSortState(sortMode, false)

	theme := common.NewThemeFromConfig(cfg.Theme)

	// Apply configured default screen
	screen := common.ScreenBoard
	if cfg.View.DefaultScreen == "list" {
		screen = common.ScreenList
	}

	// Apply configured keybindings
	common.ApplyKeys(cfg.Keys)

	return Model{
		issues:      issues,
		issuesDir:   issuesDir,
		projectRoot: projectRoot,
		screen:      screen,
		sortMode:    sortMode,
		filters:     filters,
		cfg:         cfg,
		theme:       theme,
		board:       board.New(filtered).SetTheme(theme),
		list:        l.SetTheme(theme),
		watcher:     w,
	}
}

// issueSourceDir returns the .grapes/ directory for the given issue ID.
func (m Model) issueSourceDir(issueID int) string {
	for _, iss := range m.issues {
		if iss.ID == issueID {
			if iss.SourceDir != "" {
				return iss.SourceDir
			}
			return m.issuesDir
		}
	}
	return m.issuesDir
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
	return tea.Batch(m.board.Init(), m.list.Init(), m.watchCmd(), tea.RequestBackgroundColor)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		contentHeight := m.contentHeight()
		off := m.topOffset()
		m.board = m.board.SetTopOffset(off).SetSize(m.width, contentHeight)
		m.list = m.list.SetTopOffset(off).SetSize(m.width, contentHeight)
		m.detail = m.detail.SetTopOffset(off).SetSize(m.width, contentHeight)
		m.settings = m.settings.SetSize(m.width, contentHeight)
		return m, nil

	case tea.BackgroundColorMsg:
		// Detect dark/light and apply config overrides on top
		if m.cfg.Theme != (config.ThemeConfig{}) {
			m.theme = common.NewThemeFromConfig(m.cfg.Theme)
		} else {
			m.theme = common.NewTheme(msg.IsDark())
		}
		m.board = m.board.SetTheme(m.theme)
		m.list = m.list.SetTheme(m.theme)
		m.detail = m.detail.SetTheme(m.theme)
		return m, nil

	case tea.KeyPressMsg:
		// When filter overlays are active, route all input to them
		if m.filterPicker != nil {
			var cmd tea.Cmd
			fp := *m.filterPicker
			fp, cmd = fp.Update(msg)
			m.filterPicker = &fp
			return m, cmd
		}
		if m.filterMenu != nil {
			var cmd tea.Cmd
			fm := *m.filterMenu
			fm, cmd = fm.Update(msg)
			m.filterMenu = &fm
			return m, cmd
		}
		// When picker is active, route all input to it
		if m.picker != nil {
			var cmd tea.Cmd
			p := *m.picker
			p, cmd = p.Update(msg)
			m.picker = &p
			return m, cmd
		}
		// Global quit — but not when filtering in list view
		if m.screen == common.ScreenList && m.list.Filtering() {
			break // fall through to screen-specific handler
		}
		if key.Matches(msg, common.GlobalKeyMap.Quit) {
			return m, tea.Quit
		}
		// Open settings with ","
		if msg.String() == "," && m.screen != common.ScreenSettings {
			m.settings = settings.New(m.cfg, m.issuesDir, m.width, m.contentHeight(), m.theme)
			m.navStack = append(m.navStack, navEntry{screen: m.screen})
			m.screen = common.ScreenSettings
			return m, nil
		}

	case tea.MouseClickMsg, tea.MouseMotionMsg:
		// When picker is active, route all mouse events to it
		if m.picker != nil {
			m.updatePickerPosition()
			var cmd tea.Cmd
			p := *m.picker
			p, cmd = p.Update(msg)
			m.picker = &p
			return m, cmd
		}

		if click, ok := msg.(tea.MouseClickMsg); ok {
			mouse := click.Mouse()
			if click.Button == tea.MouseLeft && mouse.Y == 0 {
				// Detect clicks on header tabs (Board / List / Settings)
				boardTabW := lipgloss.Width(m.theme.StyleTabInactive.Render("Board"))
				listTabW := lipgloss.Width(m.theme.StyleTabInactive.Render("List"))
				settingsTabW := lipgloss.Width(m.theme.StyleTabInactive.Render(",Settings"))
				totalTabsW := boardTabW + 1 + listTabW + 1 + settingsTabW // +1 for spaces
				tabsStart := m.width - totalTabsW
				x := mouse.X
				if x >= tabsStart && x < tabsStart+boardTabW {
					m.navStack = nil
					m.screen = common.ScreenBoard
					return m, nil
				}
				listStart := tabsStart + boardTabW + 1
				if x >= listStart && x < listStart+listTabW {
					m.navStack = nil
					m.screen = common.ScreenList
					return m, nil
				}
				settingsStart := listStart + listTabW + 1
				if x >= settingsStart && x < settingsStart+settingsTabW {
					m.settings = settings.New(m.cfg, m.issuesDir, m.width, m.contentHeight(), m.theme)
					m.navStack = append(m.navStack, navEntry{screen: m.screen})
					m.screen = common.ScreenSettings
					return m, nil
				}
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
			m.navStack = append(m.navStack, navEntry{screen: m.screen, detail: m.detail})
			m.screen = common.ScreenDetail
			m.detail = detail.New(*iss, m.issues, m.width, m.contentHeight(), m.theme).SetTopOffset(m.topOffset())
			return m, m.detail.Init()
		}
		return m, nil

	case common.GoBackMsg:
		if len(m.navStack) == 0 {
			m.screen = common.ScreenBoard
			return m, nil
		}
		top := m.navStack[len(m.navStack)-1]
		m.navStack = m.navStack[:len(m.navStack)-1]
		m.screen = top.screen
		if top.screen == common.ScreenDetail {
			m.detail = top.detail
			m.detail = m.detail.SetTopOffset(m.topOffset()).SetSize(m.width, m.contentHeight())
		}
		return m, nil

	case common.SwitchScreenMsg:
		m.navStack = nil
		m.screen = msg.Screen
		return m, nil

	case common.ConfigSavedMsg:
		m.cfg = msg.Config
		m.theme = common.NewThemeFromConfig(msg.Config.Theme)
		m.board = m.board.SetTheme(m.theme)
		m.list = m.list.SetTheme(m.theme)
		m.detail = m.detail.SetTheme(m.theme)
		common.ApplyKeys(msg.Config.Keys)
		// Go back to previous screen
		if len(m.navStack) > 0 {
			top := m.navStack[len(m.navStack)-1]
			m.navStack = m.navStack[:len(m.navStack)-1]
			m.screen = top.screen
		} else {
			m.screen = common.ScreenBoard
		}
		return m, nil

	case common.ThemeMsg:
		m.theme = msg.Theme
		m.board = m.board.SetTheme(m.theme)
		m.list = m.list.SetTheme(m.theme)
		m.detail = m.detail.SetTheme(m.theme)
		m.settings = m.settings.SetTheme(m.theme)
		return m, nil

	case common.CycleSortMsg:
		m.sortMode = m.sortMode.Next()
		m.sortAsc = false // reset direction when changing sort field
		data.SortIssues(m.issues, m.sortMode, m.sortAsc)
		filtered := m.filteredIssues()
		m.board = m.board.SetSortMode(m.sortMode).SetIssues(filtered)
		m.list = m.list.SetSortState(m.sortMode, m.sortAsc).SetIssues(filtered)
		return m, nil

	case common.ReverseSortMsg:
		m.sortAsc = !m.sortAsc
		data.SortIssues(m.issues, m.sortMode, m.sortAsc)
		filtered := m.filteredIssues()
		m.board = m.board.SetIssues(filtered)
		m.list = m.list.SetSortState(m.sortMode, m.sortAsc).SetIssues(filtered)
		return m, nil

	case common.ColumnSortMsg:
		if m.sortMode == msg.Mode {
			m.sortAsc = !m.sortAsc
		} else {
			m.sortMode = msg.Mode
			m.sortAsc = false
		}
		data.SortIssues(m.issues, m.sortMode, m.sortAsc)
		filtered := m.filteredIssues()
		m.board = m.board.SetSortMode(m.sortMode).SetIssues(filtered)
		m.list = m.list.SetSortState(m.sortMode, m.sortAsc).SetIssues(filtered)
		return m, nil

	case common.RefreshMsg:
		issues, err := data.LoadAllIssues(m.issuesDir)
		if err != nil {
			return m, m.watchCmd()
		}
		// Merge worktree-only issues
		mainIDs := make(map[int]bool)
		for _, iss := range issues {
			mainIDs[iss.ID] = true
		}
		wtIssues, _ := data.LoadWorktreeIssues(m.projectRoot, mainIDs)
		if len(wtIssues) > 0 {
			issues = append(issues, wtIssues...)
			data.RewireRelationships(issues)
		}
		data.SortIssues(issues, m.sortMode, m.sortAsc)
		m.issues = issues
		filtered := m.filteredIssues()
		m.board = m.board.SetIssues(filtered)
		m.list = m.list.SetIssues(filtered)
		// Re-create detail view if it's showing, so changes are visible
		if m.screen == common.ScreenDetail {
			for _, iss := range issues {
				if iss.ID == m.detail.IssueID() {
					m.detail = detail.New(iss, issues, m.width, m.contentHeight(), m.theme).SetTopOffset(m.topOffset())
					break
				}
			}
		}
		// Re-sync watched dirs (picks up new issue folders) and keep watching
		if m.watcher != nil {
			addWatchDirs(m.watcher, m.issuesDir)
			for _, dir := range data.FindWorktreeIssuesDirs(m.projectRoot) {
				addWatchDirs(m.watcher, dir)
			}
		}
		return m, m.watchCmd()

	case common.ShowPickerMsg:
		p := m.buildPicker(msg.IssueID, msg.Field)
		m.picker = &p
		return m, nil

	case common.MoveIssueMsg:
		srcDir := m.issueSourceDir(msg.IssueID)
		return m, func() tea.Msg {
			if err := data.UpdateField(srcDir, msg.IssueID, "status", string(msg.NewStatus)); err != nil {
				return common.WriteErrMsg{Err: err}
			}
			return nil // fsnotify will trigger refresh
		}

	case common.PickerResultMsg:
		m.picker = nil
		srcDir := m.issueSourceDir(msg.IssueID)
		return m, func() tea.Msg {
			if err := data.UpdateField(srcDir, msg.IssueID, msg.Field, msg.Value); err != nil {
				return common.WriteErrMsg{Err: err}
			}
			return nil // fsnotify will trigger refresh
		}

	case common.PickerCancelMsg:
		m.picker = nil
		return m, nil

	case common.ShowFilterMenuMsg:
		menu := filter.NewMenu(m.filters, len(m.collectAllLabels()), m.theme)
		m.filterMenu = &menu
		return m, nil

	case common.FilterMenuSelectMsg:
		m.filterMenu = nil
		picker := m.buildFilterPicker(msg.Field)
		m.filterPicker = &picker
		return m, nil

	case common.FilterToggleChildrenMsg:
		m.filterMenu = nil
		m.filters.ToggleHasChildren()
		m.propagateFilters()
		return m, nil

	case common.FilterToggleTopLevelMsg:
		m.filterMenu = nil
		m.filters.ToggleTopLevelOnly()
		m.propagateFilters()
		return m, nil

	case common.FilterPickerResultMsg:
		m.filterPicker = nil
		m.applyFilterSelection(msg.Field, msg.Selected)
		m.propagateFilters()
		return m, nil

	case common.FilterCancelMsg:
		m.filterMenu = nil
		m.filterPicker = nil
		return m, nil

	case common.ClearAllFiltersMsg:
		m.filterMenu = nil
		m.filters.Clear()
		m.propagateFilters()
		return m, nil

	case common.LaunchEditorMsg:
		tmpFile, err := os.CreateTemp("", "grapes-comment-*.md")
		if err != nil {
			m.statusMsg = "Error: " + err.Error()
			return m, m.clearStatusAfter(3 * time.Second)
		}
		tmpFile.Close()
		m.editingIssueID = msg.ID
		m.editingTmpPath = tmpFile.Name()
		m.editingMode = "comment"

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}
		c := exec.Command(editor, m.editingTmpPath)
		return m, tea.ExecProcess(c, func(err error) tea.Msg {
			return common.EditorFinishedMsg{Err: err}
		})

	case common.LaunchEditMsg:
		// Find the issue to serialize
		var issue *data.Issue
		for i := range m.issues {
			if m.issues[i].ID == msg.ID {
				issue = &m.issues[i]
				break
			}
		}
		if issue == nil {
			return m, nil
		}

		tmpFile, err := os.CreateTemp("", "grapes-edit-*.md")
		if err != nil {
			m.statusMsg = "Error: " + err.Error()
			return m, m.clearStatusAfter(3 * time.Second)
		}
		if _, err := tmpFile.WriteString(data.SerializeIssue(*issue)); err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			m.statusMsg = "Error: " + err.Error()
			return m, m.clearStatusAfter(3 * time.Second)
		}
		tmpFile.Close()

		m.editingIssueID = msg.ID
		m.editingTmpPath = tmpFile.Name()
		m.editingMode = "edit"

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}
		c := exec.Command(editor, m.editingTmpPath)
		return m, tea.ExecProcess(c, func(err error) tea.Msg {
			return common.EditFinishedMsg{Err: err}
		})

	case common.EditFinishedMsg:
		if msg.Err != nil {
			m.statusMsg = "Editor error: " + msg.Err.Error()
			os.Remove(m.editingTmpPath)
			return m, m.clearStatusAfter(3 * time.Second)
		}
		body, err := os.ReadFile(m.editingTmpPath)
		if err != nil {
			os.Remove(m.editingTmpPath)
			m.statusMsg = "Error reading file: " + err.Error()
			return m, m.clearStatusAfter(3 * time.Second)
		}
		text := string(body)
		// Strip error banner (from previous validation retry) before checking emptiness
		cleaned := stripErrorBanner(text)
		if strings.TrimSpace(cleaned) == "" {
			os.Remove(m.editingTmpPath)
			m.statusMsg = "Edit cancelled."
			return m, m.clearStatusAfter(3 * time.Second)
		}

		issueID := m.editingIssueID
		tmpPath := m.editingTmpPath
		srcDir := m.issueSourceDir(issueID)

		saveErr := data.SaveIssueFromText(srcDir, issueID, cleaned)
		if saveErr == nil {
			os.Remove(tmpPath)
			return m, nil // fsnotify will trigger refresh
		}

		// On validation error, prepend the error to the file and re-open the editor
		var valErr *data.EditValidationError
		if errors.As(saveErr, &valErr) {
			// Strip any previous error banner before prepending a fresh one
			cleaned := stripErrorBanner(text)
			banner := "# ERROR: " + valErr.Message + "\n# Fix the issue above, then save and quit. Empty file to cancel.\n\n"
			os.WriteFile(tmpPath, []byte(banner+cleaned), 0644)

			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vi"
			}
			c := exec.Command(editor, tmpPath)
			return m, tea.ExecProcess(c, func(err error) tea.Msg {
				return common.EditFinishedMsg{Err: err}
			})
		}

		// Non-validation error — clean up and show
		os.Remove(tmpPath)
		m.statusMsg = "Write error: " + saveErr.Error()
		return m, m.clearStatusAfter(3 * time.Second)

	case common.EditorFinishedMsg:
		if msg.Err != nil {
			m.statusMsg = "Editor error: " + msg.Err.Error()
			os.Remove(m.editingTmpPath)
			return m, m.clearStatusAfter(3 * time.Second)
		}
		body, err := os.ReadFile(m.editingTmpPath)
		os.Remove(m.editingTmpPath)
		if err != nil {
			m.statusMsg = "Error reading comment: " + err.Error()
			return m, m.clearStatusAfter(3 * time.Second)
		}
		trimmed := strings.TrimSpace(string(body))
		if trimmed == "" {
			return m, nil // empty comment, no-op
		}
		issueID := m.editingIssueID
		srcDir := m.issueSourceDir(issueID)
		return m, func() tea.Msg {
			if err := data.AppendComment(srcDir, issueID, trimmed); err != nil {
				return common.WriteErrMsg{Err: err}
			}
			return nil // fsnotify will trigger refresh
		}

	case common.WriteErrMsg:
		m.statusMsg = "Write error: " + msg.Err.Error()
		return m, m.clearStatusAfter(3 * time.Second)

	case clearStatusMsg:
		m.statusMsg = ""
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
	case common.ScreenSettings:
		m.settings, cmd = m.settings.Update(msg)
	}
	return m, cmd
}

func (m Model) renderHeader() string {
	title := m.theme.StyleAppTitle.Render("grapes")

	// Active tab follows the current screen; detail inherits from origin screen.
	activeScreen := m.screen
	if activeScreen == common.ScreenDetail {
		activeScreen = m.originScreen()
	}

	renderTab := func(label string, screen common.Screen) string {
		if activeScreen == screen {
			return m.theme.StyleTabActive.Render(label)
		}
		return m.theme.StyleTabInactive.Render(label)
	}

	boardTab := renderTab("Board", common.ScreenBoard)
	listTab := renderTab("List", common.ScreenList)
	settingsTab := renderTab(",Settings", common.ScreenSettings)

	tabs := lipgloss.JoinHorizontal(lipgloss.Top, boardTab, " ", listTab, " ", settingsTab)
	spacerW := m.width - lipgloss.Width(title) - lipgloss.Width(tabs)
	if spacerW < 0 {
		spacerW = 0
	}
	row := title + strings.Repeat(" ", spacerW) + tabs
	sep := m.theme.StyleSeparator.Render(strings.Repeat("━", m.width))
	return lipgloss.JoinVertical(lipgloss.Left, row, sep)
}

// originScreen returns the non-detail screen that was active before
// entering the detail view chain. Used for tab highlighting.
func (m Model) originScreen() common.Screen {
	for i := len(m.navStack) - 1; i >= 0; i-- {
		if m.navStack[i].screen != common.ScreenDetail {
			return m.navStack[i].screen
		}
	}
	return common.ScreenBoard
}

func (m Model) View() tea.View {
	header := m.renderHeader()

	var content string
	var helpParts []string
	dot := m.theme.StyleStatusSep.Render(" · ")

	contentHeight := m.contentHeight()

	sortArrow := "▼"
	if m.sortAsc {
		sortArrow = "▲"
	}
	sortLabel := m.sortMode.Label() + " " + sortArrow

	switch m.screen {
	case common.ScreenBoard:
		content = m.board.View()
		helpParts = []string{
			m.theme.FormatKeyHint("hjkl", "navigate"),
			m.theme.FormatKeyHint("enter", "open"),
			m.theme.FormatKeyHint("e", "edit"),
			m.theme.FormatKeyHint("s", "status"),
			m.theme.FormatKeyHint("p", "priority"),
			m.theme.FormatKeyHint("drag", "move"),
			m.theme.FormatKeyHint("f", "filter"),
			m.theme.FormatKeyHint("o/O", sortLabel),
			m.theme.FormatKeyHint("L", "list"),
			m.theme.FormatKeyHint("q", "quit"),
		}
	case common.ScreenList:
		content = m.list.View()
		navHint := "jk"
		if m.list.HScrollActive() {
			navHint = "hjkl"
		}
		helpParts = []string{
			m.theme.FormatKeyHint(navHint, "navigate"),
			m.theme.FormatKeyHint("enter", "open"),
			m.theme.FormatKeyHint("e", "edit"),
			m.theme.FormatKeyHint("s", "status"),
			m.theme.FormatKeyHint("p", "priority"),
			m.theme.FormatKeyHint("o/O", sortLabel),
			m.theme.FormatKeyHint("f", "filter"),
			m.theme.FormatKeyHint("/", "search"),
			m.theme.FormatKeyHint("B", "board"),
			m.theme.FormatKeyHint("q", "quit"),
		}
	case common.ScreenDetail:
		content = m.detail.View()
		helpParts = []string{
			m.theme.FormatKeyHint("jk", "scroll"),
			m.theme.FormatKeyHint("e", "edit"),
			m.theme.FormatKeyHint("s", "status"),
			m.theme.FormatKeyHint("p", "priority"),
			m.theme.FormatKeyHint("c", "comment"),
			m.theme.FormatKeyHint("esc", "back"),
			m.theme.FormatKeyHint("q", "quit"),
		}
	case common.ScreenSettings:
		content = m.settings.View()
		helpParts = []string{
			m.theme.FormatKeyHint("jk", "navigate"),
			m.theme.FormatKeyHint("tab", "pane"),
			m.theme.FormatKeyHint("enter", "edit"),
			m.theme.FormatKeyHint("ctrl+s", "save"),
			m.theme.FormatKeyHint("esc", "back"),
		}
	}

	// Pad content to fill the content area
	contentLines := strings.Count(content, "\n") + 1
	if contentLines < contentHeight {
		content += strings.Repeat("\n", contentHeight-contentLines)
	}

	// Picker overlay: composite the picker box on top of the real content
	if m.picker != nil {
		content = overlayCenter(content, m.picker.View(), m.width, contentHeight)
		helpParts = []string{
			m.theme.FormatKeyHint("jk", "navigate"),
			m.theme.FormatKeyHint("enter", "select"),
			m.theme.FormatKeyHint("esc", "cancel"),
		}
	}

	// Filter overlays
	if m.filterPicker != nil {
		content = overlayCenter(content, m.filterPicker.View(), m.width, contentHeight)
		helpParts = []string{
			m.theme.FormatKeyHint("jk", "navigate"),
			m.theme.FormatKeyHint("space", "toggle"),
			m.theme.FormatKeyHint("enter", "apply"),
			m.theme.FormatKeyHint("esc", "cancel"),
		}
	} else if m.filterMenu != nil {
		content = overlayCenter(content, m.filterMenu.View(), m.width, contentHeight)
		helpParts = []string{
			m.theme.FormatKeyHint("jk", "navigate"),
			m.theme.FormatKeyHint("enter", "select"),
			m.theme.FormatKeyHint("esc", "cancel"),
		}
	}

	var helpText string
	if m.statusMsg != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#f85149"))
		helpText = "  " + errStyle.Render(m.statusMsg)
	} else {
		helpText = "  " + strings.Join(helpParts, dot)
	}
	bar := m.theme.StyleStatusBar.Width(m.width).Render(helpText)

	// Render filter bar between header and content when filters are active
	filterBar := filter.RenderBar(m.filters, m.width, m.theme)
	var full string
	if filterBar != "" {
		full = lipgloss.JoinVertical(lipgloss.Left, header, filterBar, content, bar)
	} else {
		full = lipgloss.JoinVertical(lipgloss.Left, header, content, bar)
	}

	v := tea.NewView(full)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

// clearStatusAfter returns a command that clears the status message after a delay.
func (m Model) clearStatusAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

// buildPicker creates a picker model for the given issue field.
func (m Model) buildPicker(issueID int, field string) picker.Model {
	var issue *data.Issue
	for i := range m.issues {
		if m.issues[i].ID == issueID {
			issue = &m.issues[i]
			break
		}
	}

	switch field {
	case "status":
		var opts []picker.Option
		current := 0
		for i, s := range data.AllStatuses {
			if issue != nil && issue.Status == s {
				current = i
			}
			opts = append(opts, picker.Option{
				Value: string(s),
				Label: s.Label(),
				Icon:  common.StatusIcon(s),
				Style: m.theme.StatusStyle(s),
			})
		}
		return picker.New("Status", opts, current, issueID, field, m.theme)

	case "priority":
		var opts []picker.Option
		current := 0
		for i, p := range data.AllPriorities {
			if issue != nil && issue.Priority == p {
				current = i
			}
			opts = append(opts, picker.Option{
				Value: string(p),
				Label: p.Label(),
				Icon:  strings.TrimSpace(common.PriorityIcon(p)),
				Style: m.theme.PriorityStyle(p),
			})
		}
		return picker.New("Priority", opts, current, issueID, field, m.theme)
	}

	// Fallback (shouldn't happen)
	return picker.New(field, nil, 0, issueID, field, m.theme)
}

// updatePickerPosition computes the centered screen position of the picker
// overlay and stores it on the picker model for mouse hit-testing.
func (m *Model) updatePickerPosition() {
	if m.picker == nil {
		return
	}
	pickerView := m.picker.View()
	pickerLines := strings.Split(pickerView, "\n")
	fgH := len(pickerLines)
	fgW := 0
	for _, l := range pickerLines {
		if w := lipgloss.Width(l); w > fgW {
			fgW = w
		}
	}
	contentH := m.contentHeight()
	offsetY := common.AppHeaderHeight + filter.BarHeight(m.filters)
	m.picker.ScreenX = (m.width - fgW) / 2
	m.picker.ScreenY = offsetY + (contentH-fgH)/2
}

// topOffset returns the number of screen lines above the view content
// (app header + optional structured filter bar).
func (m Model) topOffset() int {
	return common.AppHeaderHeight + filter.BarHeight(m.filters)
}

// contentHeight returns the available height for view content, accounting for
// the app header, status bar, and optional filter bar.
func (m Model) contentHeight() int {
	h := m.height - 3 // header(2) + status bar(1)
	h -= filter.BarHeight(m.filters)
	if h < 0 {
		h = 0
	}
	return h
}

// filteredIssues returns issues matching the current structured filters.
func (m Model) filteredIssues() []data.Issue {
	if m.filters.IsEmpty() {
		return m.issues
	}
	var out []data.Issue
	for _, iss := range m.issues {
		if m.filters.Matches(iss) {
			out = append(out, iss)
		}
	}
	return out
}

// collectAllLabels extracts unique labels from all loaded issues (unfiltered).
func (m Model) collectAllLabels() []string {
	seen := make(map[string]bool)
	var labels []string
	for _, iss := range m.issues {
		for _, l := range iss.Labels {
			if !seen[l] {
				seen[l] = true
				labels = append(labels, l)
			}
		}
	}
	sort.Strings(labels)
	return labels
}

// buildFilterPicker creates a MultiPicker for the given filter field.
// Only values that exist in the current issue set are shown.
func (m Model) buildFilterPicker(field string) filter.MultiPicker {
	switch field {
	case "status":
		present := make(map[data.Status]bool)
		for _, iss := range m.issues {
			present[iss.Status] = true
		}
		var opts []filter.PickerOption
		var preSelected []string
		for _, s := range data.AllStatuses {
			if !present[s] {
				continue
			}
			opts = append(opts, filter.PickerOption{
				Value: string(s),
				Label: s.Label(),
				Icon:  common.StatusIcon(s),
				Style: m.theme.StatusStyle(s),
			})
		}
		for _, s := range m.filters.Statuses {
			preSelected = append(preSelected, string(s))
		}
		return filter.NewMultiPicker("Status", "status", opts, preSelected, m.theme)

	case "priority":
		present := make(map[data.Priority]bool)
		for _, iss := range m.issues {
			present[iss.Priority] = true
		}
		var opts []filter.PickerOption
		var preSelected []string
		for _, p := range data.AllPriorities {
			if !present[p] {
				continue
			}
			opts = append(opts, filter.PickerOption{
				Value: string(p),
				Label: p.Label(),
				Icon:  strings.TrimSpace(common.PriorityIcon(p)),
				Style: m.theme.PriorityStyle(p),
			})
		}
		for _, p := range m.filters.Priorities {
			preSelected = append(preSelected, string(p))
		}
		return filter.NewMultiPicker("Priority", "priority", opts, preSelected, m.theme)

	case "labels":
		var opts []filter.PickerOption
		for _, l := range m.collectAllLabels() {
			opts = append(opts, filter.PickerOption{
				Value: l,
				Label: l,
				Style: m.theme.StatusStyle(data.StatusTodo), // neutral color
			})
		}
		return filter.NewMultiPicker("Label", "labels", opts, m.filters.Labels, m.theme)

	case "source":
		sourceSet := make(map[string]bool)
		for _, iss := range m.issues {
			if iss.Worktree == "" {
				sourceSet["main"] = true
			} else {
				sourceSet[iss.Worktree] = true
			}
		}
		var opts []filter.PickerOption
		if sourceSet["main"] {
			opts = append(opts, filter.PickerOption{
				Value: "main",
				Label: "main",
				Style: m.theme.StyleSubtitle,
			})
		}
		for name := range sourceSet {
			if name != "main" && name != "" {
				opts = append(opts, filter.PickerOption{
					Value: name,
					Label: common.WorktreeIcon() + " " + name,
					Style: m.theme.StyleWorktreeLabel,
				})
			}
		}
		return filter.NewMultiPicker("Source", "source", opts, m.filters.Sources, m.theme)
	}

	return filter.NewMultiPicker(field, field, nil, nil, m.theme)
}

// applyFilterSelection updates the filter set from a multi-picker result.
func (m *Model) applyFilterSelection(field string, selected []string) {
	switch field {
	case "status":
		m.filters.SetStatuses(selected)
	case "priority":
		m.filters.SetPriorities(selected)
	case "labels":
		m.filters.SetLabels(selected)
	case "source":
		m.filters.SetSources(selected)
	}
}

// propagateFilters sends filtered issues to both views and adjusts sizes.
func (m *Model) propagateFilters() {
	filtered := m.filteredIssues()
	m.board = m.board.SetStatusFilter(m.filters.Statuses).SetIssues(filtered)
	m.list = m.list.SetIssues(filtered)
	contentHeight := m.contentHeight()
	off := m.topOffset()
	m.board = m.board.SetTopOffset(off).SetSize(m.width, contentHeight)
	m.list = m.list.SetTopOffset(off).SetSize(m.width, contentHeight)
}

// stripErrorBanner removes a leading "# ERROR: ..." banner that was prepended by
// a previous validation failure, so it doesn't accumulate on repeated retries.
func stripErrorBanner(text string) string {
	lines := strings.Split(text, "\n")
	i := 0
	for i < len(lines) && strings.HasPrefix(lines[i], "# ") {
		i++
	}
	// Skip blank lines after the banner
	for i < len(lines) && lines[i] == "" {
		i++
	}
	if i == 0 {
		return text
	}
	return strings.Join(lines[i:], "\n")
}

// overlayCenter composites fg (a small box) centered on top of bg (the full content).
// Uses ANSI-aware truncation to preserve the background content on both sides
// of the overlay box, so board columns / list rows stay visible around the picker.
func overlayCenter(bg, fg string, bgWidth, bgHeight int) string {
	bgLines := strings.Split(bg, "\n")
	fgLines := strings.Split(fg, "\n")

	// Ensure bg has enough lines
	for len(bgLines) < bgHeight {
		bgLines = append(bgLines, "")
	}

	// Measure fg box width
	fgWidth := 0
	for _, line := range fgLines {
		if w := lipgloss.Width(line); w > fgWidth {
			fgWidth = w
		}
	}

	// Calculate centering offsets
	startY := (bgHeight - len(fgLines)) / 2
	startX := (bgWidth - fgWidth) / 2
	if startY < 0 {
		startY = 0
	}
	if startX < 0 {
		startX = 0
	}

	// Splice fg lines into bg lines, preserving left and right bg content
	for i, fgLine := range fgLines {
		y := startY + i
		if y >= len(bgLines) {
			break
		}
		bgLine := bgLines[y]

		// Left portion: first startX visible chars of the bg line
		left := ansi.Truncate(bgLine, startX, "")
		// Pad left if the bg line is shorter than startX
		leftW := lipgloss.Width(left)
		if leftW < startX {
			left += strings.Repeat(" ", startX-leftW)
		}

		// Right portion: bg content after the fg box ends
		rightStart := startX + fgWidth
		right := ansi.TruncateLeft(bgLine, rightStart, "")

		bgLines[y] = left + fgLine + right
	}

	return strings.Join(bgLines[:bgHeight], "\n")
}
