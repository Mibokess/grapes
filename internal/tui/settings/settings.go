package settings

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Mibokess/grapes/internal/config"
	"github.com/Mibokess/grapes/internal/tui/common"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	themes "go.withmatt.com/themes"
)

var hexColorRe = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

type pane int

const (
	paneCategories pane = iota
	paneFields
)

type fieldKind int

const (
	fieldEnum fieldKind = iota
	fieldColor
	fieldKey
)

type field struct {
	label   string
	cfgKey  string   // key used to get/set on Config
	kind    fieldKind
	options []string // for enum fields
}

type category struct {
	name   string
	fields []field
}

// Model is the settings screen model.
type Model struct {
	cfg       config.Config
	original  config.Config // snapshot for cancel
	issuesDir string
	theme      common.Theme
	termIsDark bool

	categories []category
	catIdx     int
	fieldIdx   int
	focus      pane
	editing    bool
	input      textinput.Model
	statusMsg  string

	width     int
	height    int
	topOffset int // screen lines above this view's content (app header + filter bar)
}

// New creates a new settings screen model.
func New(cfg config.Config, issuesDir string, w, h int, theme common.Theme) Model {
	ti := textinput.New()
	ti.CharLimit = 32

	cats := []category{
		{
			name: "View",
			fields: []field{
				{label: "Default screen", cfgKey: "default_screen", kind: fieldEnum, options: []string{"board", "list"}},
				{label: "Default sort", cfgKey: "default_sort", kind: fieldEnum, options: []string{"priority", "updated", "created", "id", "title", "status"}},
				{label: "Auto-close sub-issues", cfgKey: "auto_close_subs", kind: fieldEnum, options: []string{"off", "on"}},
			},
		},
		{
			name: "Theme",
			fields: []field{
				{label: "Preset", cfgKey: "theme_preset", kind: fieldEnum, options: common.CuratedPresets},
				{label: "Mode", cfgKey: "theme_mode", kind: fieldEnum, options: []string{"auto", "light", "dark"}},
				{label: "Accent", cfgKey: "accent", kind: fieldColor},
				{label: "Accent BG", cfgKey: "accent_bg", kind: fieldColor},
				{label: "Border", cfgKey: "border", kind: fieldColor},
				{label: "Text", cfgKey: "text", kind: fieldColor},
				{label: "Muted", cfgKey: "muted", kind: fieldColor},
				{label: "Faint", cfgKey: "faint", kind: fieldColor},
				{label: "Surface", cfgKey: "surface", kind: fieldColor},
				{label: "Backlog", cfgKey: "color_backlog", kind: fieldColor},
				{label: "Todo", cfgKey: "color_todo", kind: fieldColor},
				{label: "In Progress", cfgKey: "color_in_progress", kind: fieldColor},
				{label: "Done", cfgKey: "color_done", kind: fieldColor},
				{label: "Cancelled", cfgKey: "color_cancelled", kind: fieldColor},
				{label: "Urgent", cfgKey: "color_urgent", kind: fieldColor},
				{label: "High", cfgKey: "color_high", kind: fieldColor},
				{label: "Medium", cfgKey: "color_medium", kind: fieldColor},
				{label: "Low", cfgKey: "color_low", kind: fieldColor},
			},
		},
		{
			name: "Keys",
			fields: []field{
				{label: "Quit", cfgKey: "quit", kind: fieldKey},
				{label: "Board: Up", cfgKey: "board_up", kind: fieldKey},
				{label: "Board: Down", cfgKey: "board_down", kind: fieldKey},
				{label: "Board: Left", cfgKey: "board_left", kind: fieldKey},
				{label: "Board: Right", cfgKey: "board_right", kind: fieldKey},
				{label: "Board: Open", cfgKey: "board_open", kind: fieldKey},
				{label: "Board: Edit", cfgKey: "board_edit", kind: fieldKey},
				{label: "Board: To list", cfgKey: "board_to_list", kind: fieldKey},
				{label: "Board: Filter", cfgKey: "board_filter", kind: fieldKey},
				{label: "Board: Status", cfgKey: "board_status", kind: fieldKey},
				{label: "Board: Priority", cfgKey: "board_priority", kind: fieldKey},
				{label: "Board: Sort", cfgKey: "board_sort", kind: fieldKey},
				{label: "Board: Reverse", cfgKey: "board_reverse", kind: fieldKey},
				{label: "List: Up", cfgKey: "list_up", kind: fieldKey},
				{label: "List: Down", cfgKey: "list_down", kind: fieldKey},
				{label: "List: Open", cfgKey: "list_open", kind: fieldKey},
				{label: "List: Edit", cfgKey: "list_edit", kind: fieldKey},
				{label: "List: To board", cfgKey: "list_to_board", kind: fieldKey},
				{label: "List: Search", cfgKey: "list_search", kind: fieldKey},
				{label: "List: Filter", cfgKey: "list_filter", kind: fieldKey},
				{label: "List: Status", cfgKey: "list_status", kind: fieldKey},
				{label: "List: Priority", cfgKey: "list_priority", kind: fieldKey},
				{label: "List: Sort", cfgKey: "list_sort", kind: fieldKey},
				{label: "List: Reverse", cfgKey: "list_reverse", kind: fieldKey},
				{label: "Detail: Back", cfgKey: "detail_back", kind: fieldKey},
				{label: "Detail: To board", cfgKey: "detail_to_board", kind: fieldKey},
				{label: "Detail: To list", cfgKey: "detail_to_list", kind: fieldKey},
				{label: "Detail: Status", cfgKey: "detail_status", kind: fieldKey},
				{label: "Detail: Priority", cfgKey: "detail_priority", kind: fieldKey},
				{label: "Detail: Comment", cfgKey: "detail_comment", kind: fieldKey},
				{label: "Detail: Edit", cfgKey: "detail_edit", kind: fieldKey},
			},
		},
	}

	return Model{
		cfg:        cfg,
		original:   cfg,
		issuesDir:  issuesDir,
		theme:      theme,
		categories: cats,
		input:      ti,
		width:      w,
		height:     h,
	}
}

func (m Model) Init() tea.Cmd { return nil }

// SetSize updates the viewport dimensions.
func (m Model) SetSize(w, h int) Model {
	m.width = w
	m.height = h
	return m
}

// SetTopOffset sets the number of screen lines above this view's content.
func (m Model) SetTopOffset(n int) Model {
	m.topOffset = n
	return m
}

// SetTheme updates the theme used for rendering.
func (m Model) SetTheme(t common.Theme) Model {
	m.theme = t
	return m
}

// SetDark updates the terminal dark/light detection flag.
func (m Model) SetDark(isDark bool) Model {
	m.termIsDark = isDark
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.editing {
			return m.updateEditing(msg)
		}
		return m.updateNavigating(msg)

	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft {
			return m.handleMouseClick(msg.Mouse())
		}

	case tea.MouseWheelMsg:
		if msg.Button == tea.MouseWheelUp {
			if m.fieldIdx > 0 {
				m.fieldIdx--
			}
		} else if msg.Button == tea.MouseWheelDown {
			fields := m.categories[m.catIdx].fields
			if m.fieldIdx < len(fields)-1 {
				m.fieldIdx++
			}
		}
		return m, nil
	}
	return m, nil
}

func (m Model) handleMouseClick(mouse tea.Mouse) (Model, tea.Cmd) {
	if m.editing {
		return m, nil
	}

	const catW = 18
	// Mouse Y is absolute (0 = top of terminal). Subtract topOffset for
	// the app header/filter bar, then 1 more for the view's own top padding line.
	row := mouse.Y - m.topOffset - 1

	if row < 0 {
		return m, nil
	}

	sepX := catW // approximate separator x position

	if mouse.X < sepX {
		// Clicked in categories pane
		if row >= 0 && row < len(m.categories) {
			m.catIdx = row
			m.fieldIdx = 0
			m.focus = paneCategories
		}
	} else {
		// Clicked in fields pane
		fields := m.categories[m.catIdx].fields

		// Account for scroll offset (same logic as View)
		visibleH := m.height - 2
		if visibleH < 1 {
			visibleH = 1
		}
		scrollOffset := 0
		if len(fields) > visibleH && m.fieldIdx >= visibleH {
			scrollOffset = m.fieldIdx - visibleH + 1
		}

		idx := row + scrollOffset
		if idx >= 0 && idx < len(fields) {
			if m.focus == paneFields && m.fieldIdx == idx {
				// Already selected — activate (same as pressing Enter)
				f := fields[idx]
				if f.kind == fieldEnum {
					cur := m.getFieldValue(f.cfgKey)
					for i, opt := range f.options {
						if opt == cur {
							next := f.options[(i+1)%len(f.options)]
							m.setFieldValue(f.cfgKey, next)
							if f.cfgKey == "theme_mode" || f.cfgKey == "theme_preset" {
								newTheme := common.NewThemeFromConfig(m.cfg.Theme, m.termIsDark)
								m.theme = newTheme
								return m, func() tea.Msg { return common.ThemeMsg{Theme: newTheme} }
							}
							return m, nil
						}
					}
					m.setFieldValue(f.cfgKey, f.options[0])
					if f.cfgKey == "theme_mode" || f.cfgKey == "theme_preset" {
						newTheme := common.NewThemeFromConfig(m.cfg.Theme, m.termIsDark)
						m.theme = newTheme
						return m, func() tea.Msg { return common.ThemeMsg{Theme: newTheme} }
					}
				} else {
					m.editing = true
					m.input.SetValue(m.getFieldValue(f.cfgKey))
					m.input.Focus()
					m.input.CursorEnd()
				}
			} else {
				m.fieldIdx = idx
				m.focus = paneFields
			}
		}
	}
	return m, nil
}

func (m Model) updateEditing(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	f := m.currentField()

	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		val := m.input.Value()

		// Validate color fields
		if f.kind == fieldColor && !hexColorRe.MatchString(val) {
			m.statusMsg = "Invalid hex color (use #RRGGBB)"
			return m, nil
		}

		m.setFieldValue(f.cfgKey, val)
		m.editing = false
		m.statusMsg = ""

		// Live preview for theme colors — rebuild theme and send to app
		if f.kind == fieldColor {
			newTheme := common.NewThemeFromConfig(m.cfg.Theme, m.termIsDark)
			m.theme = newTheme
			return m, func() tea.Msg {
				return common.ThemeMsg{Theme: newTheme}
			}
		}
		return m, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
		m.editing = false
		m.statusMsg = ""
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) updateNavigating(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, common.SettingsKeyMap.Save):
		// Save and go back
		if err := config.Save(m.issuesDir, m.cfg); err != nil {
			m.statusMsg = "Save error: " + err.Error()
			return m, nil
		}
		common.ApplyKeys(m.cfg.Keys)
		return m, func() tea.Msg {
			return common.ConfigSavedMsg{Config: m.cfg}
		}

	case key.Matches(msg, common.SettingsKeyMap.Back):
		if m.focus == paneFields {
			// Step back to categories pane first
			m.focus = paneCategories
			return m, nil
		}
		// Cancel — restore original theme and go back
		origTheme := common.NewThemeFromConfig(m.original.Theme, m.termIsDark)
		return m, tea.Batch(
			func() tea.Msg { return common.ThemeMsg{Theme: origTheme} },
			func() tea.Msg { return common.GoBackMsg{} },
		)

	case key.Matches(msg, common.SettingsKeyMap.Right):
		if m.focus == paneCategories {
			m.focus = paneFields
			m.fieldIdx = 0
		}
		return m, nil

	case key.Matches(msg, common.SettingsKeyMap.Left):
		if m.focus == paneFields {
			m.focus = paneCategories
		}
		return m, nil

	case key.Matches(msg, common.SettingsKeyMap.Tab):
		if m.focus == paneCategories {
			m.focus = paneFields
		} else {
			m.focus = paneCategories
		}
		m.fieldIdx = 0
		return m, nil

	case key.Matches(msg, common.SettingsKeyMap.Up):
		if m.focus == paneCategories {
			if m.catIdx > 0 {
				m.catIdx--
				m.fieldIdx = 0
			}
		} else {
			if m.fieldIdx > 0 {
				m.fieldIdx--
			}
		}
		return m, nil

	case key.Matches(msg, common.SettingsKeyMap.Down):
		if m.focus == paneCategories {
			if m.catIdx < len(m.categories)-1 {
				m.catIdx++
				m.fieldIdx = 0
			}
		} else {
			fields := m.categories[m.catIdx].fields
			if m.fieldIdx < len(fields)-1 {
				m.fieldIdx++
			}
		}
		return m, nil

	case key.Matches(msg, common.SettingsKeyMap.Enter):
		if m.focus == paneCategories {
			m.focus = paneFields
			m.fieldIdx = 0
			return m, nil
		}
		f := m.currentField()
		if f.kind == fieldEnum {
			// Cycle through options
			cur := m.getFieldValue(f.cfgKey)
			for i, opt := range f.options {
				if opt == cur {
					next := f.options[(i+1)%len(f.options)]
					m.setFieldValue(f.cfgKey, next)
					if f.cfgKey == "theme_mode" || f.cfgKey == "theme_preset" {
						newTheme := common.NewThemeFromConfig(m.cfg.Theme, m.termIsDark)
						m.theme = newTheme
						return m, func() tea.Msg { return common.ThemeMsg{Theme: newTheme} }
					}
					return m, nil
				}
			}
			m.setFieldValue(f.cfgKey, f.options[0])
			if f.cfgKey == "theme_mode" || f.cfgKey == "theme_preset" {
				newTheme := common.NewThemeFromConfig(m.cfg.Theme, m.termIsDark)
				m.theme = newTheme
				return m, func() tea.Msg { return common.ThemeMsg{Theme: newTheme} }
			}
			return m, nil
		}
		// Start inline editing
		m.editing = true
		m.input.SetValue(m.getFieldValue(f.cfgKey))
		m.input.Focus()
		m.input.CursorEnd()
		return m, nil
	}

	return m, nil
}

func (m Model) currentField() field {
	return m.categories[m.catIdx].fields[m.fieldIdx]
}

func (m Model) View() string {
	catW := 18
	sep := m.theme.StyleFaint.Render("│")

	var leftLines []string
	for i, cat := range m.categories {
		line := fmt.Sprintf("  %-*s", catW-4, cat.name)
		if i == m.catIdx {
			if m.focus == paneCategories {
				line = m.theme.StyleSectionHeader.Render(line)
			} else {
				line = m.theme.StyleTitle.Render(line)
			}
		} else {
			line = m.theme.StyleSubtitle.Render(line)
		}
		leftLines = append(leftLines, line)
	}

	fields := m.categories[m.catIdx].fields
	labelW := 0
	for _, f := range fields {
		if len(f.label) > labelW {
			labelW = len(f.label)
		}
	}

	// Scrolling: if there are more fields than fit in the visible area,
	// scroll to keep the selected field visible
	visibleH := m.height - 2 // leave room for padding
	if visibleH < 1 {
		visibleH = 1
	}
	scrollOffset := 0
	if len(fields) > visibleH {
		if m.fieldIdx >= visibleH {
			scrollOffset = m.fieldIdx - visibleH + 1
		}
	}

	var rightLines []string
	for i := scrollOffset; i < len(fields) && i < scrollOffset+visibleH; i++ {
		f := fields[i]
		val := m.getFieldValue(f.cfgKey)
		label := fmt.Sprintf(" %-*s", labelW+2, f.label)

		var valStr string
		if m.editing && m.focus == paneFields && i == m.fieldIdx {
			valStr = m.input.View()
		} else {
			switch f.kind {
			case fieldColor:
				swatch := lipgloss.NewStyle().Foreground(lipgloss.Color(val)).Render("██")
				valStr = swatch + " " + val
			case fieldEnum:
				valStr = val
			case fieldKey:
				valStr = val
			}
		}

		if i == m.fieldIdx && m.focus == paneFields {
			label = m.theme.StyleTitle.Render(label)
			if !m.editing {
				valStr = m.theme.StyleSectionHeader.Render(valStr)
			}
		} else {
			label = m.theme.StyleSubtitle.Render(label)
			valStr = m.theme.StyleFaint.Render(valStr)
		}

		rightLines = append(rightLines, label+"  "+valStr)
	}

	// Pad shorter column
	maxLines := len(leftLines)
	if len(rightLines) > maxLines {
		maxLines = len(rightLines)
	}
	for len(leftLines) < maxLines {
		leftLines = append(leftLines, strings.Repeat(" ", catW-2))
	}
	for len(rightLines) < maxLines {
		rightLines = append(rightLines, "")
	}

	var lines []string
	lines = append(lines, "") // top padding
	for i := 0; i < maxLines; i++ {
		left := leftLines[i]
		right := ""
		if i < len(rightLines) {
			right = rightLines[i]
		}
		lines = append(lines, left+" "+sep+" "+right)
	}

	// Status message at bottom
	if m.statusMsg != "" {
		errStyle := lipgloss.NewStyle().Foreground(m.theme.ColorError)
		lines = append(lines, "")
		lines = append(lines, "  "+errStyle.Render(m.statusMsg))
	}

	content := strings.Join(lines, "\n")

	// Pad to fill height
	contentLines := strings.Count(content, "\n") + 1
	if contentLines < m.height {
		content += strings.Repeat("\n", m.height-contentLines)
	}

	return content
}

func (m Model) effectiveIsDark() bool {
	if p := m.cfg.Theme.Preset; p != "" && p != "default" {
		if ext, err := themes.GetTheme(p); err == nil {
			isDark := common.PresetIsDark(ext)
			switch m.cfg.Theme.Mode {
			case "light":
				return false
			case "dark":
				return true
			}
			return isDark
		}
	}
	return m.cfg.Theme.EffectiveIsDark(m.termIsDark)
}

// getFieldValue reads the current value for a config key.
func (m Model) getFieldValue(cfgKey string) string {
	switch cfgKey {
	case "default_screen":
		return m.cfg.View.DefaultScreen
	case "default_sort":
		return m.cfg.View.DefaultSort
	case "theme_preset":
		if m.cfg.Theme.Preset == "" {
			return "default"
		}
		return m.cfg.Theme.Preset
	case "auto_close_subs":
		if m.cfg.View.AutoCloseSubs {
			return "on"
		}
		return "off"
	case "theme_mode":
		if m.cfg.Theme.Mode == "" {
			return "auto"
		}
		return m.cfg.Theme.Mode
	case "accent", "accent_bg", "border", "text", "muted", "faint", "surface",
		"color_backlog", "color_todo", "color_in_progress", "color_done", "color_cancelled",
		"color_urgent", "color_high", "color_medium", "color_low":
		return getColorFromSet(m.cfg.Theme.ColorsFor(m.effectiveIsDark()), cfgKey)
	case "quit":
		return m.cfg.Keys.Quit
	case "board_up":
		return m.cfg.Keys.BoardUp
	case "board_down":
		return m.cfg.Keys.BoardDown
	case "board_left":
		return m.cfg.Keys.BoardLeft
	case "board_right":
		return m.cfg.Keys.BoardRight
	case "board_open":
		return m.cfg.Keys.BoardOpen
	case "board_edit":
		return m.cfg.Keys.BoardEdit
	case "board_to_list":
		return m.cfg.Keys.BoardToList
	case "board_filter":
		return m.cfg.Keys.BoardFilter
	case "board_status":
		return m.cfg.Keys.BoardStatus
	case "board_priority":
		return m.cfg.Keys.BoardPriority
	case "board_sort":
		return m.cfg.Keys.BoardSort
	case "board_reverse":
		return m.cfg.Keys.BoardReverse
	case "list_up":
		return m.cfg.Keys.ListUp
	case "list_down":
		return m.cfg.Keys.ListDown
	case "list_open":
		return m.cfg.Keys.ListOpen
	case "list_edit":
		return m.cfg.Keys.ListEdit
	case "list_to_board":
		return m.cfg.Keys.ListToBoard
	case "list_search":
		return m.cfg.Keys.ListSearch
	case "list_filter":
		return m.cfg.Keys.ListFilter
	case "list_status":
		return m.cfg.Keys.ListStatus
	case "list_priority":
		return m.cfg.Keys.ListPriority
	case "list_sort":
		return m.cfg.Keys.ListSort
	case "list_reverse":
		return m.cfg.Keys.ListReverse
	case "detail_back":
		return m.cfg.Keys.DetailBack
	case "detail_to_board":
		return m.cfg.Keys.DetailToBoard
	case "detail_to_list":
		return m.cfg.Keys.DetailToList
	case "detail_status":
		return m.cfg.Keys.DetailStatus
	case "detail_priority":
		return m.cfg.Keys.DetailPriority
	case "detail_comment":
		return m.cfg.Keys.DetailComment
	case "detail_edit":
		return m.cfg.Keys.DetailEdit
	}
	return ""
}

// setFieldValue writes a value to the config for a given key.
func (m *Model) setFieldValue(cfgKey, val string) {
	switch cfgKey {
	case "default_screen":
		m.cfg.View.DefaultScreen = val
	case "default_sort":
		m.cfg.View.DefaultSort = val
	case "theme_preset":
		if val == "default" {
			m.cfg.Theme.Preset = ""
		} else {
			m.cfg.Theme.Preset = val
		}
	case "auto_close_subs":
		m.cfg.View.AutoCloseSubs = val == "on"
	case "theme_mode":
		m.cfg.Theme.Mode = val
	case "accent", "accent_bg", "border", "text", "muted", "faint", "surface",
		"color_backlog", "color_todo", "color_in_progress", "color_done", "color_cancelled",
		"color_urgent", "color_high", "color_medium", "color_low":
		isDark := m.effectiveIsDark()
		colors := m.cfg.Theme.ColorsFor(isDark)
		setColorOnSet(&colors, cfgKey, val)
		m.cfg.Theme.SetColorsFor(isDark, colors)
	case "quit":
		m.cfg.Keys.Quit = val
	case "board_up":
		m.cfg.Keys.BoardUp = val
	case "board_down":
		m.cfg.Keys.BoardDown = val
	case "board_left":
		m.cfg.Keys.BoardLeft = val
	case "board_right":
		m.cfg.Keys.BoardRight = val
	case "board_open":
		m.cfg.Keys.BoardOpen = val
	case "board_edit":
		m.cfg.Keys.BoardEdit = val
	case "board_to_list":
		m.cfg.Keys.BoardToList = val
	case "board_filter":
		m.cfg.Keys.BoardFilter = val
	case "board_status":
		m.cfg.Keys.BoardStatus = val
	case "board_priority":
		m.cfg.Keys.BoardPriority = val
	case "board_sort":
		m.cfg.Keys.BoardSort = val
	case "board_reverse":
		m.cfg.Keys.BoardReverse = val
	case "list_up":
		m.cfg.Keys.ListUp = val
	case "list_down":
		m.cfg.Keys.ListDown = val
	case "list_open":
		m.cfg.Keys.ListOpen = val
	case "list_edit":
		m.cfg.Keys.ListEdit = val
	case "list_to_board":
		m.cfg.Keys.ListToBoard = val
	case "list_search":
		m.cfg.Keys.ListSearch = val
	case "list_filter":
		m.cfg.Keys.ListFilter = val
	case "list_status":
		m.cfg.Keys.ListStatus = val
	case "list_priority":
		m.cfg.Keys.ListPriority = val
	case "list_sort":
		m.cfg.Keys.ListSort = val
	case "list_reverse":
		m.cfg.Keys.ListReverse = val
	case "detail_back":
		m.cfg.Keys.DetailBack = val
	case "detail_to_board":
		m.cfg.Keys.DetailToBoard = val
	case "detail_to_list":
		m.cfg.Keys.DetailToList = val
	case "detail_status":
		m.cfg.Keys.DetailStatus = val
	case "detail_priority":
		m.cfg.Keys.DetailPriority = val
	case "detail_comment":
		m.cfg.Keys.DetailComment = val
	case "detail_edit":
		m.cfg.Keys.DetailEdit = val
	}
}

func getColorFromSet(c config.ColorSetConfig, key string) string {
	switch key {
	case "accent":
		return c.Accent
	case "accent_bg":
		return c.AccentBg
	case "border":
		return c.Border
	case "text":
		return c.Text
	case "muted":
		return c.Muted
	case "faint":
		return c.Faint
	case "surface":
		return c.Surface
	case "color_backlog":
		return c.ColorBacklog
	case "color_todo":
		return c.ColorTodo
	case "color_in_progress":
		return c.ColorInProgress
	case "color_done":
		return c.ColorDone
	case "color_cancelled":
		return c.ColorCancelled
	case "color_urgent":
		return c.ColorUrgent
	case "color_high":
		return c.ColorHigh
	case "color_medium":
		return c.ColorMedium
	case "color_low":
		return c.ColorLow
	}
	return ""
}

func setColorOnSet(c *config.ColorSetConfig, key, val string) {
	switch key {
	case "accent":
		c.Accent = val
	case "accent_bg":
		c.AccentBg = val
	case "border":
		c.Border = val
	case "text":
		c.Text = val
	case "muted":
		c.Muted = val
	case "faint":
		c.Faint = val
	case "surface":
		c.Surface = val
	case "color_backlog":
		c.ColorBacklog = val
	case "color_todo":
		c.ColorTodo = val
	case "color_in_progress":
		c.ColorInProgress = val
	case "color_done":
		c.ColorDone = val
	case "color_cancelled":
		c.ColorCancelled = val
	case "color_urgent":
		c.ColorUrgent = val
	case "color_high":
		c.ColorHigh = val
	case "color_medium":
		c.ColorMedium = val
	case "color_low":
		c.ColorLow = val
	}
}
