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
	fieldAction
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

	picking     bool
	pickCfgKey  string
	pickOptions []string
	pickCursor  int
	pickCurrent int
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
			{label: "Hide empty columns", cfgKey: "hide_empty_columns", kind: fieldEnum, options: []string{"off", "on"}},
			},
		},
		{
			name: "Theme",
			fields: []field{
				{label: "Theme", cfgKey: "theme_preset", kind: fieldEnum, options: common.CuratedPresets},
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
				{label: "Board: Label", cfgKey: "board_label", kind: fieldKey},
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
				{label: "List: Label", cfgKey: "list_label", kind: fieldKey},
				{label: "List: Sort", cfgKey: "list_sort", kind: fieldKey},
				{label: "List: Reverse", cfgKey: "list_reverse", kind: fieldKey},
				{label: "Detail: Back", cfgKey: "detail_back", kind: fieldKey},
				{label: "Detail: To board", cfgKey: "detail_to_board", kind: fieldKey},
				{label: "Detail: To list", cfgKey: "detail_to_list", kind: fieldKey},
				{label: "Detail: Status", cfgKey: "detail_status", kind: fieldKey},
				{label: "Detail: Priority", cfgKey: "detail_priority", kind: fieldKey},
				{label: "Detail: Label", cfgKey: "detail_label", kind: fieldKey},
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

// PickerActive returns whether an enum picker overlay is open.
func (m Model) PickerActive() bool { return m.picking }

// PickerView renders the enum picker overlay box.
func (m Model) PickerView() string {
	if !m.picking {
		return ""
	}

	cursorStyle := lipgloss.NewStyle().Foreground(m.theme.ColorAccent).Bold(true)
	checkStyle := lipgloss.NewStyle().Foreground(m.theme.ColorDone)
	rowActiveStyle := lipgloss.NewStyle().Background(m.theme.ColorAccentBg)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(m.theme.ColorAccent)

	maxVis := m.pickMaxVisible()
	scrollOff := m.pickScrollOffset()

	rowWidth := 24
	for _, opt := range m.pickOptions {
		if w := lipgloss.Width(opt) + 4; w > rowWidth {
			rowWidth = w
		}
	}

	var rows []string
	for i := scrollOff; i < scrollOff+maxVis && i < len(m.pickOptions); i++ {
		opt := m.pickOptions[i]
		isCursor := i == m.pickCursor
		isCurrent := i == m.pickCurrent

		var prefix string
		switch {
		case isCursor:
			prefix = cursorStyle.Render("›") + " "
		case isCurrent:
			prefix = checkStyle.Render("✓") + " "
		default:
			prefix = "  "
		}

		row := prefix + opt
		visible := lipgloss.Width(row)
		if visible < rowWidth {
			row += strings.Repeat(" ", rowWidth-visible)
		}

		if isCursor {
			row = rowActiveStyle.Render(row)
		}

		rows = append(rows, row)
	}

	content := strings.Join(rows, "\n")

	f := m.currentField()
	title := " " + titleStyle.Render(f.label) + " "

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.ColorAccent).
		Padding(1, 2).
		Render(content)

	// Insert title in top border
	lines := strings.Split(box, "\n")
	if len(lines) > 0 {
		topBorder := lines[0]
		if len(topBorder) > 4 {
			runeTop := []rune(topBorder)
			titleRunes := []rune(title)
			insertAt := 2
			end := insertAt + len(titleRunes)
			if end < len(runeTop) {
				result := make([]rune, 0, len(runeTop))
				result = append(result, runeTop[:insertAt]...)
				result = append(result, titleRunes...)
				result = append(result, runeTop[end:]...)
				lines[0] = string(result)
			}
		}
		box = strings.Join(lines, "\n")
	}

	return box
}

func (m Model) pickMaxVisible() int {
	mv := m.height - 6
	if mv < 3 {
		mv = 3
	}
	if mv > len(m.pickOptions) {
		mv = len(m.pickOptions)
	}
	return mv
}

func (m Model) pickScrollOffset() int {
	mv := m.pickMaxVisible()
	if len(m.pickOptions) <= mv {
		return 0
	}
	if m.pickCursor < mv {
		return 0
	}
	return m.pickCursor - mv + 1
}

func (m *Model) openPicker(f field) {
	cur := m.getFieldValue(f.cfgKey)
	m.picking = true
	m.pickCfgKey = f.cfgKey
	m.pickOptions = f.options
	m.pickCurrent = 0
	for i, opt := range f.options {
		if opt == cur {
			m.pickCurrent = i
			break
		}
	}
	m.pickCursor = m.pickCurrent
}

func (m Model) pickerScreenPos() (x, y, boxW, boxH int) {
	maxVis := m.pickMaxVisible()

	rowWidth := 24
	for _, opt := range m.pickOptions {
		if w := lipgloss.Width(opt) + 4; w > rowWidth {
			rowWidth = w
		}
	}

	// border(1) + padding(1) + rows + padding(1) + border(1)
	boxH = maxVis + 4
	// border(1) + padding(2) + content + padding(2) + border(1)
	boxW = rowWidth + 6

	x = (m.width - boxW) / 2
	y = m.topOffset + (m.height-boxH)/2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	return
}

func (m Model) updatePicking(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up"))):
		if m.pickCursor > 0 {
			m.pickCursor--
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("j", "down"))):
		if m.pickCursor < len(m.pickOptions)-1 {
			m.pickCursor++
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		val := m.pickOptions[m.pickCursor]
		m.picking = false
		m.setFieldValue(m.pickCfgKey, val)
		if m.pickCfgKey == "theme_preset" {
			newTheme := common.NewThemeFromConfig(m.cfg.Theme, m.termIsDark)
			m.theme = newTheme
			return m, func() tea.Msg { return common.ThemeMsg{Theme: newTheme} }
		}
		return m, nil
	case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
		m.picking = false
		return m, nil
	}
	return m, nil
}

func (m Model) handlePickerClick(mouse tea.Mouse) (Model, tea.Cmd) {
	sx, sy, boxW, boxH := m.pickerScreenPos()
	relY := mouse.Y - sy - 2 // border + padding
	inBox := mouse.Y >= sy && mouse.Y < sy+boxH &&
		mouse.X >= sx && mouse.X < sx+boxW
	scrollOff := m.pickScrollOffset()
	maxVis := m.pickMaxVisible()

	if inBox && relY >= 0 && relY < maxVis {
		idx := scrollOff + relY
		if idx < len(m.pickOptions) {
			m.pickCursor = idx
			val := m.pickOptions[idx]
			m.picking = false
			m.setFieldValue(m.pickCfgKey, val)
			if m.pickCfgKey == "theme_preset" {
				newTheme := common.NewThemeFromConfig(m.cfg.Theme, m.termIsDark)
				m.theme = newTheme
				return m, func() tea.Msg { return common.ThemeMsg{Theme: newTheme} }
			}
			return m, nil
		}
	}
	if !inBox {
		m.picking = false
	}
	return m, nil
}

func (m Model) handlePickerMotion(mouse tea.Mouse) (Model, tea.Cmd) {
	_, sy, _, _ := m.pickerScreenPos()
	relY := mouse.Y - sy - 2
	scrollOff := m.pickScrollOffset()
	maxVis := m.pickMaxVisible()
	if relY >= 0 && relY < maxVis {
		idx := scrollOff + relY
		if idx < len(m.pickOptions) {
			m.pickCursor = idx
		}
	}
	return m, nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.picking {
			return m.updatePicking(msg)
		}
		if m.editing {
			return m.updateEditing(msg)
		}
		return m.updateNavigating(msg)

	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft {
			if m.picking {
				return m.handlePickerClick(msg.Mouse())
			}
			return m.handleMouseClick(msg.Mouse())
		}

	case tea.MouseMotionMsg:
		if m.picking {
			return m.handlePickerMotion(msg.Mouse())
		}

	case tea.MouseWheelMsg:
		if m.picking {
			if msg.Button == tea.MouseWheelUp && m.pickCursor > 0 {
				m.pickCursor--
			} else if msg.Button == tea.MouseWheelDown && m.pickCursor < len(m.pickOptions)-1 {
				m.pickCursor++
			}
			return m, nil
		}
		if msg.Button == tea.MouseWheelUp {
			if m.fieldIdx > 0 {
				m.fieldIdx--
			}
		} else if msg.Button == tea.MouseWheelDown {
			fields := m.effectiveFields()
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
		fields := m.effectiveFields()

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
				switch f.kind {
				case fieldAction:
					return m.activateAction(f)
				case fieldEnum:
					if len(f.options) > 3 {
						m.openPicker(f)
						return m, nil
					}
					cur := m.getFieldValue(f.cfgKey)
					for i, opt := range f.options {
						if opt == cur {
							next := f.options[(i+1)%len(f.options)]
							m.setFieldValue(f.cfgKey, next)
							if f.cfgKey == "theme_preset" {
								newTheme := common.NewThemeFromConfig(m.cfg.Theme, m.termIsDark)
								m.theme = newTheme
								return m, func() tea.Msg { return common.ThemeMsg{Theme: newTheme} }
							}
							return m, nil
						}
					}
					m.setFieldValue(f.cfgKey, f.options[0])
					if f.cfgKey == "theme_preset" {
						newTheme := common.NewThemeFromConfig(m.cfg.Theme, m.termIsDark)
						m.theme = newTheme
						return m, func() tea.Msg { return common.ThemeMsg{Theme: newTheme} }
					}
				default:
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
			fields := m.effectiveFields()
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
		switch f.kind {
		case fieldAction:
			return m.activateAction(f)
		case fieldEnum:
			if len(f.options) > 3 {
				m.openPicker(f)
				return m, nil
			}
			// Cycle through options (for small lists)
			cur := m.getFieldValue(f.cfgKey)
			for i, opt := range f.options {
				if opt == cur {
					next := f.options[(i+1)%len(f.options)]
					m.setFieldValue(f.cfgKey, next)
					if f.cfgKey == "theme_preset" {
						newTheme := common.NewThemeFromConfig(m.cfg.Theme, m.termIsDark)
						m.theme = newTheme
						return m, func() tea.Msg { return common.ThemeMsg{Theme: newTheme} }
					}
					return m, nil
				}
			}
			m.setFieldValue(f.cfgKey, f.options[0])
			if f.cfgKey == "theme_preset" {
				newTheme := common.NewThemeFromConfig(m.cfg.Theme, m.termIsDark)
				m.theme = newTheme
				return m, func() tea.Msg { return common.ThemeMsg{Theme: newTheme} }
			}
			return m, nil
		default:
			// Start inline editing
			m.editing = true
			m.input.SetValue(m.getFieldValue(f.cfgKey))
			m.input.Focus()
			m.input.CursorEnd()
			return m, nil
		}
	}

	return m, nil
}

func (m Model) currentField() field {
	return m.effectiveFields()[m.fieldIdx]
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

	fields := m.effectiveFields()
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

	overrideBoldStyle := lipgloss.NewStyle().Bold(true)

	var rightLines []string
	for i := scrollOffset; i < len(fields) && i < scrollOffset+visibleH; i++ {
		f := fields[i]
		val := m.getFieldValue(f.cfgKey)
		label := fmt.Sprintf(" %-*s", labelW+2, f.label)
		isOverridden := f.kind == fieldColor && m.isColorOverridden(f.cfgKey)

		var valStr string
		if m.editing && m.focus == paneFields && i == m.fieldIdx {
			valStr = m.input.View()
		} else {
			switch f.kind {
			case fieldColor:
				swatch := lipgloss.NewStyle().Foreground(lipgloss.Color(val)).Render("██")
				hexStr := val
				if isOverridden {
					hexStr = overrideBoldStyle.Render(val)
				}
				valStr = swatch + " " + hexStr
			case fieldEnum:
				valStr = val
			case fieldKey:
				valStr = val
			case fieldAction:
				valStr = ""
			}
		}

		if i == m.fieldIdx && m.focus == paneFields {
			if f.kind == fieldAction {
				label = m.theme.StyleSectionHeader.Render(label)
			} else {
				label = m.theme.StyleTitle.Render(label)
				if !m.editing {
					valStr = m.theme.StyleSectionHeader.Render(valStr)
				}
			}
		} else {
			if f.kind == fieldAction {
				label = m.theme.StyleFaint.Render(label)
			} else {
				label = m.theme.StyleSubtitle.Render(label)
				valStr = m.theme.StyleFaint.Render(valStr)
			}
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

// defaultColorForKey returns the default color value for a theme color key
// based on the current effective mode.
func (m Model) defaultColorForKey(cfgKey string) string {
	defaults := config.Defaults()
	isDark := m.effectiveIsDark()
	return getColorFromSet(defaults.Theme.ColorsFor(isDark), cfgKey)
}

// isColorOverridden returns true if a color field differs from the theme default.
func (m Model) isColorOverridden(cfgKey string) bool {
	current := getColorFromSet(m.cfg.Theme.ColorsFor(m.effectiveIsDark()), cfgKey)
	return current != m.defaultColorForKey(cfgKey)
}

// hasAnyColorOverride returns true if any theme color differs from defaults.
func (m Model) hasAnyColorOverride() bool {
	for _, f := range m.categories[1].fields { // Theme category
		if f.kind == fieldColor && m.isColorOverridden(f.cfgKey) {
			return true
		}
	}
	return false
}

// effectiveFields returns the fields for the current category, dynamically
// appending a "Reset colors" action when theme color overrides exist.
func (m Model) effectiveFields() []field {
	fields := m.categories[m.catIdx].fields
	if m.catIdx == 1 && m.hasAnyColorOverride() { // Theme category
		result := make([]field, len(fields), len(fields)+1)
		copy(result, fields)
		result = append(result, field{
			label:  "Reset colors",
			cfgKey: "reset_colors",
			kind:   fieldAction,
		})
		return result
	}
	return fields
}

// activateAction handles activation of action-type fields.
func (m Model) activateAction(f field) (Model, tea.Cmd) {
	switch f.cfgKey {
	case "reset_colors":
		defaults := config.Defaults()
		isDark := m.effectiveIsDark()
		m.cfg.Theme.SetColorsFor(isDark, defaults.Theme.ColorsFor(isDark))
		newTheme := common.NewThemeFromConfig(m.cfg.Theme, m.termIsDark)
		m.theme = newTheme
		// Clamp fieldIdx since reset row may disappear
		fields := m.effectiveFields()
		if m.fieldIdx >= len(fields) {
			m.fieldIdx = len(fields) - 1
		}
		return m, func() tea.Msg { return common.ThemeMsg{Theme: newTheme} }
	}
	return m, nil
}

// getFieldValue reads the current value for a config key.
func (m Model) getFieldValue(cfgKey string) string {
	switch cfgKey {
	case "default_screen":
		return m.cfg.View.DefaultScreen
	case "default_sort":
		return m.cfg.View.DefaultSort
	case "theme_preset":
		p := m.cfg.Theme.Preset
		if p == "" || p == "default" {
			switch m.cfg.Theme.Mode {
			case "light":
				return "Light"
			case "dark":
				return "Dark"
			default:
				return "Auto"
			}
		}
		return p
	case "auto_close_subs":
		if m.cfg.View.AutoCloseSubs {
			return "on"
		}
		return "off"
	case "hide_empty_columns":
		if m.cfg.View.HideEmpty() {
			return "on"
		}
		return "off"
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
	case "board_label":
		return m.cfg.Keys.BoardLabel
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
	case "list_label":
		return m.cfg.Keys.ListLabel
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
	case "detail_label":
		return m.cfg.Keys.DetailLabel
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
		switch val {
		case "Auto":
			m.cfg.Theme.Preset = ""
			m.cfg.Theme.Mode = "auto"
		case "Light":
			m.cfg.Theme.Preset = ""
			m.cfg.Theme.Mode = "light"
		case "Dark":
			m.cfg.Theme.Preset = ""
			m.cfg.Theme.Mode = "dark"
		default:
			m.cfg.Theme.Preset = val
			m.cfg.Theme.Mode = "auto"
		}
	case "auto_close_subs":
		m.cfg.View.AutoCloseSubs = val == "on"
	case "hide_empty_columns":
		b := val == "on"
		m.cfg.View.HideEmptyColumns = &b
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
	case "board_label":
		m.cfg.Keys.BoardLabel = val
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
	case "list_label":
		m.cfg.Keys.ListLabel = val
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
	case "detail_label":
		m.cfg.Keys.DetailLabel = val
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
