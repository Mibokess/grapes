package settings

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Mibokess/grapes/internal/config"
	"github.com/Mibokess/grapes/internal/tui/common"
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
	cfgKey  string // key used to get/set on Config
	kind    fieldKind
	options []string // for enum fields
}

// keyBindings lists every rebindable action once: the settings row it renders
// and the config field it reads and writes. One table means a new binding
// cannot show up in the UI without being editable, or vice versa.
var keyBindings = []struct {
	label  string
	cfgKey string
	field  func(*config.KeysConfig) *string
}{
	{"Quit", "quit", func(k *config.KeysConfig) *string { return &k.Quit }},
	{"Board: Up", "board_up", func(k *config.KeysConfig) *string { return &k.BoardUp }},
	{"Board: Down", "board_down", func(k *config.KeysConfig) *string { return &k.BoardDown }},
	{"Board: Left", "board_left", func(k *config.KeysConfig) *string { return &k.BoardLeft }},
	{"Board: Right", "board_right", func(k *config.KeysConfig) *string { return &k.BoardRight }},
	{"Board: Open", "board_open", func(k *config.KeysConfig) *string { return &k.BoardOpen }},
	{"Board: Edit", "board_edit", func(k *config.KeysConfig) *string { return &k.BoardEdit }},
	{"Board: To list", "board_to_list", func(k *config.KeysConfig) *string { return &k.BoardToList }},
	{"Board: Search", "board_search", func(k *config.KeysConfig) *string { return &k.BoardSearch }},
	{"Board: Filter", "board_filter", func(k *config.KeysConfig) *string { return &k.BoardFilter }},
	{"Board: Status", "board_status", func(k *config.KeysConfig) *string { return &k.BoardStatus }},
	{"Board: Priority", "board_priority", func(k *config.KeysConfig) *string { return &k.BoardPriority }},
	{"Board: Label", "board_label", func(k *config.KeysConfig) *string { return &k.BoardLabel }},
	{"Board: Sort", "board_sort", func(k *config.KeysConfig) *string { return &k.BoardSort }},
	{"Board: Reverse", "board_reverse", func(k *config.KeysConfig) *string { return &k.BoardReverse }},
	{"Board: Empty cols", "board_empty", func(k *config.KeysConfig) *string { return &k.BoardEmpty }},
	{"List: Up", "list_up", func(k *config.KeysConfig) *string { return &k.ListUp }},
	{"List: Down", "list_down", func(k *config.KeysConfig) *string { return &k.ListDown }},
	{"List: Open", "list_open", func(k *config.KeysConfig) *string { return &k.ListOpen }},
	{"List: Edit", "list_edit", func(k *config.KeysConfig) *string { return &k.ListEdit }},
	{"List: To board", "list_to_board", func(k *config.KeysConfig) *string { return &k.ListToBoard }},
	{"List: Search", "list_search", func(k *config.KeysConfig) *string { return &k.ListSearch }},
	{"List: Filter", "list_filter", func(k *config.KeysConfig) *string { return &k.ListFilter }},
	{"List: Status", "list_status", func(k *config.KeysConfig) *string { return &k.ListStatus }},
	{"List: Priority", "list_priority", func(k *config.KeysConfig) *string { return &k.ListPriority }},
	{"List: Label", "list_label", func(k *config.KeysConfig) *string { return &k.ListLabel }},
	{"List: Sort", "list_sort", func(k *config.KeysConfig) *string { return &k.ListSort }},
	{"List: Reverse", "list_reverse", func(k *config.KeysConfig) *string { return &k.ListReverse }},
	{"Detail: Back", "detail_back", func(k *config.KeysConfig) *string { return &k.DetailBack }},
	{"Detail: To board", "detail_to_board", func(k *config.KeysConfig) *string { return &k.DetailToBoard }},
	{"Detail: To list", "detail_to_list", func(k *config.KeysConfig) *string { return &k.DetailToList }},
	{"Detail: Status", "detail_status", func(k *config.KeysConfig) *string { return &k.DetailStatus }},
	{"Detail: Priority", "detail_priority", func(k *config.KeysConfig) *string { return &k.DetailPriority }},
	{"Detail: Label", "detail_label", func(k *config.KeysConfig) *string { return &k.DetailLabel }},
	{"Detail: Comment", "detail_comment", func(k *config.KeysConfig) *string { return &k.DetailComment }},
	{"Detail: Edit", "detail_edit", func(k *config.KeysConfig) *string { return &k.DetailEdit }},
}

// keyBindingField returns a pointer to the configured key for cfgKey, or nil.
func keyBindingField(k *config.KeysConfig, cfgKey string) *string {
	for _, kb := range keyBindings {
		if kb.cfgKey == cfgKey {
			return kb.field(k)
		}
	}
	return nil
}

// themeFields builds the Theme category rows from the shared color registry.
func themeFields() []field {
	fields := []field{
		{label: "Theme", cfgKey: "theme_preset", kind: fieldEnum, options: common.CuratedPresets},
	}
	for _, key := range common.ColorKeys {
		fields = append(fields, field{label: common.ColorLabels[key], cfgKey: key, kind: fieldColor})
	}
	return fields
}

// keysFields builds the Keys category rows from the binding table.
func keysFields() []field {
	fields := make([]field, 0, len(keyBindings))
	for _, kb := range keyBindings {
		fields = append(fields, field{label: kb.label, cfgKey: kb.cfgKey, kind: fieldKey})
	}
	return fields
}

type category struct {
	name   string
	fields []field
}

// Model is the settings screen model.
type Model struct {
	cfg        config.Config
	original   config.Config // snapshot for cancel
	issuesDir  string
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
		{name: "Theme", fields: themeFields()},
		{name: "Keys", fields: keysFields()},
		{
			name:   "Sources",
			fields: []field{}, // populated dynamically by effectiveFields()
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

	f, _ := m.currentField()
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
				return m.activateField(fields[idx])
			}
			m.fieldIdx = idx
			m.focus = paneFields
		}
	}
	return m, nil
}

func (m Model) updateEditing(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	f, ok := m.currentField()
	if !ok {
		m.editing = false
		return m, nil
	}

	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		val := m.input.Value()

		// Adding a worktree directory
		if f.cfgKey == "add_source_pattern" {
			val = strings.TrimSpace(val)
			if val != "" {
				m.cfg.Sources.Dirs = append(m.cfg.Sources.Dirs, val)
			}
			m.editing = false
			m.input.CharLimit = 32
			m.statusMsg = ""
			return m, nil
		}

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
		m.input.CharLimit = 32
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
		f, ok := m.currentField()
		if !ok {
			return m, nil
		}
		return m.activateField(f)
	}

	return m, nil
}

// activateField performs the "Enter" action for a field: run it, cycle it,
// open a picker for it, or start editing it. Keyboard and mouse both come
// through here so the two cannot drift apart.
func (m Model) activateField(f field) (Model, tea.Cmd) {
	switch f.kind {
	case fieldAction:
		return m.activateAction(f)

	case fieldEnum:
		// Long option lists get a picker; short ones cycle in place.
		if len(f.options) > 3 {
			m.openPicker(f)
			return m, nil
		}
		if len(f.options) == 0 {
			return m, nil
		}
		next := f.options[0]
		cur := m.getFieldValue(f.cfgKey)
		for i, opt := range f.options {
			if opt == cur {
				next = f.options[(i+1)%len(f.options)]
				break
			}
		}
		return m.applyEnum(f.cfgKey, next)

	default:
		// Start inline editing
		m.editing = true
		m.input.SetValue(m.getFieldValue(f.cfgKey))
		m.input.Focus()
		m.input.CursorEnd()
		return m, nil
	}
}

// applyEnum stores an enum choice and rebuilds the theme when the choice was a
// theme preset, so the whole app repaints immediately.
func (m Model) applyEnum(cfgKey, val string) (Model, tea.Cmd) {
	m.setFieldValue(cfgKey, val)
	if cfgKey != "theme_preset" {
		return m, nil
	}
	newTheme := common.NewThemeFromConfig(m.cfg.Theme, m.termIsDark)
	m.theme = newTheme
	return m, func() tea.Msg { return common.ThemeMsg{Theme: newTheme} }
}

// currentField returns the selected field. The cursor can outlive the row it
// pointed at, since the field list changes with the config.
func (m Model) currentField() (field, bool) {
	fields := m.effectiveFields()
	if m.fieldIdx < 0 || m.fieldIdx >= len(fields) {
		return field{}, false
	}
	return fields[m.fieldIdx], true
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
				if f.cfgKey == "default_source_dir" {
					valStr = "default"
				} else if strings.HasPrefix(f.cfgKey, "source_dir_") {
					valStr = "×"
				} else {
					valStr = ""
				}
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

// isColorOverridden reports whether the user has set this color explicitly.
// An empty stored value means "whatever the active theme says", so overridden
// is simply "non-empty" — and it stays meaningful under a preset, where the
// theme's own colors are not the built-in ones.
func (m Model) isColorOverridden(cfgKey string) bool {
	return common.GetColor(m.cfg.Theme.ColorsFor(m.effectiveIsDark()), cfgKey) != ""
}

// hasAnyColorOverride returns true if any theme color has been overridden.
func (m Model) hasAnyColorOverride() bool {
	colors := m.cfg.Theme.ColorsFor(m.effectiveIsDark())
	for _, key := range common.ColorKeys {
		if common.GetColor(colors, key) != "" {
			return true
		}
	}
	return false
}

// effectiveFields returns the fields for the current category, dynamically
// appending a "Reset colors" action when theme color overrides exist,
// or building the Sources list from config.
func (m Model) effectiveFields() []field {
	catName := m.categories[m.catIdx].name
	fields := m.categories[m.catIdx].fields

	if catName == "Theme" && m.hasAnyColorOverride() {
		result := make([]field, len(fields), len(fields)+1)
		copy(result, fields)
		result = append(result, field{
			label:  "Reset colors",
			cfgKey: "reset_colors",
			kind:   fieldAction,
		})
		return result
	}

	if catName == "Sources" {
		result := []field{
			{
				label:  ".grapes",
				cfgKey: "default_source_dir",
				kind:   fieldAction,
			},
		}
		for i, dir := range m.cfg.Sources.Dirs {
			result = append(result, field{
				label:  dir,
				cfgKey: fmt.Sprintf("source_dir_%d", i),
				kind:   fieldAction,
			})
		}
		result = append(result, field{
			label:  "+ Add pattern",
			cfgKey: "add_source_pattern",
			kind:   fieldAction,
		})
		return result
	}

	return fields
}

// activateAction handles activation of action-type fields.
func (m Model) activateAction(f field) (Model, tea.Cmd) {
	switch {
	case f.cfgKey == "reset_colors":
		// Clearing the overrides hands the colors back to the active theme,
		// which is the built-in palette or the selected preset.
		m.cfg.Theme.SetColorsFor(m.effectiveIsDark(), config.ColorSetConfig{})
		newTheme := common.NewThemeFromConfig(m.cfg.Theme, m.termIsDark)
		m.theme = newTheme
		// Clamp fieldIdx since reset row may disappear
		fields := m.effectiveFields()
		if m.fieldIdx >= len(fields) {
			m.fieldIdx = len(fields) - 1
		}
		return m, func() tea.Msg { return common.ThemeMsg{Theme: newTheme} }

	case f.cfgKey == "add_source_pattern":
		m.editing = true
		m.input.CharLimit = 256
		m.input.SetValue("")
		m.input.Focus()
		return m, nil

	case strings.HasPrefix(f.cfgKey, "source_dir_"):
		idxStr := strings.TrimPrefix(f.cfgKey, "source_dir_")
		idx, err := strconv.Atoi(idxStr)
		if err == nil && idx >= 0 && idx < len(m.cfg.Sources.Dirs) {
			m.cfg.Sources.Dirs = append(
				m.cfg.Sources.Dirs[:idx],
				m.cfg.Sources.Dirs[idx+1:]...,
			)
		}
		fields := m.effectiveFields()
		if m.fieldIdx >= len(fields) {
			m.fieldIdx = len(fields) - 1
		}
		return m, nil
	}
	return m, nil
}

// getFieldValue reads the current value for a config key.
//
// Colors report the *effective* value — the one the active theme renders with,
// preset included — not the stored override, so the swatch always matches what
// is on screen.
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
	}
	if common.IsColorKey(cfgKey) {
		return common.GetColor(m.theme.Colors, cfgKey)
	}
	keys := m.cfg.Keys
	if p := keyBindingField(&keys, cfgKey); p != nil {
		return *p
	}
	return ""
}

// setFieldValue writes a value to the config for a given key.
func (m *Model) setFieldValue(cfgKey, val string) {
	switch cfgKey {
	case "default_screen":
		m.cfg.View.DefaultScreen = val
		return
	case "default_sort":
		m.cfg.View.DefaultSort = val
		return
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
		return
	case "auto_close_subs":
		m.cfg.View.AutoCloseSubs = val == "on"
		return
	case "hide_empty_columns":
		b := val == "on"
		m.cfg.View.HideEmptyColumns = &b
		return
	}
	if common.IsColorKey(cfgKey) {
		isDark := m.effectiveIsDark()
		colors := m.cfg.Theme.ColorsFor(isDark)
		common.SetColor(&colors, cfgKey, val)
		m.cfg.Theme.SetColorsFor(isDark, colors)
		return
	}
	if p := keyBindingField(&m.cfg.Keys, cfgKey); p != nil {
		*p = val
	}
}
