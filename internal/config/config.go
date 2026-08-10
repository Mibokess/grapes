package config

import (
	"os"
	"path/filepath"

	"github.com/Mibokess/grapes/internal/fsutil"
	toml "github.com/pelletier/go-toml/v2"
)

// ViewConfig controls startup defaults.
type ViewConfig struct {
	DefaultScreen    string `toml:"default_screen"`
	DefaultSort      string `toml:"default_sort"`
	AutoCloseSubs    bool   `toml:"auto_close_subs"`
	HideEmptyColumns *bool  `toml:"hide_empty_columns,omitempty"`
}

// HideEmpty returns whether empty board columns should be hidden (default true).
func (v ViewConfig) HideEmpty() bool {
	if v.HideEmptyColumns == nil {
		return true
	}
	return *v.HideEmptyColumns
}

// ColorSetConfig holds color overrides for one theme mode.
type ColorSetConfig struct {
	Accent          string `toml:"accent,omitempty"`
	AccentBg        string `toml:"accent_bg,omitempty"`
	Border          string `toml:"border,omitempty"`
	Text            string `toml:"text,omitempty"`
	Muted           string `toml:"muted,omitempty"`
	Faint           string `toml:"faint,omitempty"`
	Surface         string `toml:"surface,omitempty"`
	ColorBacklog    string `toml:"color_backlog,omitempty"`
	ColorTodo       string `toml:"color_todo,omitempty"`
	ColorInProgress string `toml:"color_in_progress,omitempty"`
	ColorDone       string `toml:"color_done,omitempty"`
	ColorCancelled  string `toml:"color_cancelled,omitempty"`
	ColorUrgent     string `toml:"color_urgent,omitempty"`
	ColorHigh       string `toml:"color_high,omitempty"`
	ColorMedium     string `toml:"color_medium,omitempty"`
	ColorLow        string `toml:"color_low,omitempty"`
}

// ThemeConfig holds theme mode and color values for dark and light modes.
type ThemeConfig struct {
	Mode   string `toml:"mode"`   // "auto", "light", "dark" (empty = auto)
	Preset string `toml:"preset"` // external theme name; empty/"default" = built-in

	// Dark-mode colors (top-level for backward compatibility).
	Accent          string `toml:"accent"`
	AccentBg        string `toml:"accent_bg"`
	Border          string `toml:"border"`
	Text            string `toml:"text"`
	Muted           string `toml:"muted"`
	Faint           string `toml:"faint"`
	Surface         string `toml:"surface"`
	ColorBacklog    string `toml:"color_backlog"`
	ColorTodo       string `toml:"color_todo"`
	ColorInProgress string `toml:"color_in_progress"`
	ColorDone       string `toml:"color_done"`
	ColorCancelled  string `toml:"color_cancelled"`
	ColorUrgent     string `toml:"color_urgent"`
	ColorHigh       string `toml:"color_high"`
	ColorMedium     string `toml:"color_medium"`
	ColorLow        string `toml:"color_low"`

	// Light-mode colors.
	Light ColorSetConfig `toml:"light"`
}

// EffectiveIsDark resolves the mode to a boolean. termIsDark is the terminal-detected value.
func (tc ThemeConfig) EffectiveIsDark(termIsDark bool) bool {
	switch tc.Mode {
	case "light":
		return false
	case "dark":
		return true
	default:
		return termIsDark
	}
}

// ColorsFor returns the color overrides for the given mode.
func (tc ThemeConfig) ColorsFor(isDark bool) ColorSetConfig {
	if isDark {
		return ColorSetConfig{
			Accent: tc.Accent, AccentBg: tc.AccentBg, Border: tc.Border,
			Text: tc.Text, Muted: tc.Muted, Faint: tc.Faint, Surface: tc.Surface,
			ColorBacklog: tc.ColorBacklog, ColorTodo: tc.ColorTodo,
			ColorInProgress: tc.ColorInProgress, ColorDone: tc.ColorDone,
			ColorCancelled: tc.ColorCancelled, ColorUrgent: tc.ColorUrgent,
			ColorHigh: tc.ColorHigh, ColorMedium: tc.ColorMedium, ColorLow: tc.ColorLow,
		}
	}
	return tc.Light
}

// SetColorsFor writes a ColorSetConfig back to the appropriate location.
func (tc *ThemeConfig) SetColorsFor(isDark bool, c ColorSetConfig) {
	if isDark {
		tc.Accent = c.Accent
		tc.AccentBg = c.AccentBg
		tc.Border = c.Border
		tc.Text = c.Text
		tc.Muted = c.Muted
		tc.Faint = c.Faint
		tc.Surface = c.Surface
		tc.ColorBacklog = c.ColorBacklog
		tc.ColorTodo = c.ColorTodo
		tc.ColorInProgress = c.ColorInProgress
		tc.ColorDone = c.ColorDone
		tc.ColorCancelled = c.ColorCancelled
		tc.ColorUrgent = c.ColorUrgent
		tc.ColorHigh = c.ColorHigh
		tc.ColorMedium = c.ColorMedium
		tc.ColorLow = c.ColorLow
	} else {
		tc.Light = c
	}
}

// KeysConfig holds customizable keybinding strings.
type KeysConfig struct {
	Quit     string `toml:"quit"`
	Settings string `toml:"settings"`

	BoardUp       string `toml:"board_up"`
	BoardDown     string `toml:"board_down"`
	BoardLeft     string `toml:"board_left"`
	BoardRight    string `toml:"board_right"`
	BoardOpen     string `toml:"board_open"`
	BoardEdit     string `toml:"board_edit"`
	BoardToList   string `toml:"board_to_list"`
	BoardSearch   string `toml:"board_search"`
	BoardFilter   string `toml:"board_filter"`
	BoardStatus   string `toml:"board_status"`
	BoardPriority string `toml:"board_priority"`
	BoardLabel    string `toml:"board_label"`
	BoardSort     string `toml:"board_sort"`
	BoardReverse  string `toml:"board_reverse"`
	BoardEmpty    string `toml:"board_empty"`

	ListUp       string `toml:"list_up"`
	ListDown     string `toml:"list_down"`
	ListOpen     string `toml:"list_open"`
	ListEdit     string `toml:"list_edit"`
	ListToBoard  string `toml:"list_to_board"`
	ListSearch   string `toml:"list_search"`
	ListFilter   string `toml:"list_filter"`
	ListStatus   string `toml:"list_status"`
	ListPriority string `toml:"list_priority"`
	ListLabel    string `toml:"list_label"`
	ListSort     string `toml:"list_sort"`
	ListReverse  string `toml:"list_reverse"`

	DetailBack     string `toml:"detail_back"`
	DetailToBoard  string `toml:"detail_to_board"`
	DetailToList   string `toml:"detail_to_list"`
	DetailStatus   string `toml:"detail_status"`
	DetailPriority string `toml:"detail_priority"`
	DetailLabel    string `toml:"detail_label"`
	DetailComment  string `toml:"detail_comment"`
	DetailEdit     string `toml:"detail_edit"`
}

// SourcesConfig controls where grapes looks for additional issue directories.
type SourcesConfig struct {
	// Dirs lists glob patterns that resolve to issue directories.
	// Patterns can be absolute or relative to the project root.
	//
	// This repo's own git worktrees are discovered automatically (via
	// git worktree list), so they do not need to be listed here. Use Dirs only
	// for additional non-worktree sources, such as a separate repository with a
	// custom issue-folder name, e.g. "../other-project/.potatoes".
	Dirs []string `toml:"dirs"`
}

// Config is the full application configuration.
type Config struct {
	View    ViewConfig    `toml:"view"`
	Sources SourcesConfig `toml:"sources"`
	Theme   ThemeConfig   `toml:"theme"`
	Keys    KeysConfig    `toml:"keys"`
}

// Defaults returns the default configuration.
//
// Theme colors are deliberately empty: they are *overrides*, and the built-in
// palette lives in the theme package, which is the one place that knows what a
// color means. An empty value therefore says "use the active theme's color",
// which is what makes an override on top of a preset possible.
func Defaults() Config {
	return Config{
		Sources: SourcesConfig{
			Dirs: []string{},
		},
		View: ViewConfig{
			DefaultScreen: "board",
			DefaultSort:   "priority",
		},
		Theme: ThemeConfig{
			Mode: "auto",
		},
		Keys: KeysConfig{
			Quit:           "q",
			Settings:       "C",
			BoardUp:        "k",
			BoardDown:      "j",
			BoardLeft:      "h",
			BoardRight:     "l",
			BoardOpen:      "enter",
			BoardEdit:      "e",
			BoardToList:    "L",
			BoardSearch:    "/",
			BoardFilter:    "f",
			BoardStatus:    "s",
			BoardPriority:  "p",
			BoardLabel:     "t",
			BoardSort:      "o",
			BoardReverse:   "O",
			BoardEmpty:     "E",
			ListUp:         "k",
			ListDown:       "j",
			ListOpen:       "enter",
			ListEdit:       "e",
			ListToBoard:    "B",
			ListSearch:     "/",
			ListFilter:     "f",
			ListStatus:     "s",
			ListPriority:   "p",
			ListLabel:      "t",
			ListSort:       "o",
			ListReverse:    "O",
			DetailBack:     "esc",
			DetailToBoard:  "B",
			DetailToList:   "l",
			DetailStatus:   "s",
			DetailPriority: "p",
			DetailLabel:    "t",
			DetailComment:  "c",
			DetailEdit:     "e",
		},
	}
}

// Load reads config from .grapes/config.toml, falling back to defaults.
//
// A missing file yields defaults and no error. A malformed file yields
// defaults and the parse error, so the caller can report it. The parse
// targets a scratch copy so a partial parse never leaks into the result:
// toml.Unmarshal populates fields as it reads, and a config that silently
// half-applies is harder to diagnose than one that plainly falls back.
func Load(issuesDir string) (Config, error) {
	path := filepath.Join(issuesDir, "config.toml")
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Defaults(), nil
		}
		return Defaults(), err
	}
	parsed := Defaults()
	if err := toml.Unmarshal(raw, &parsed); err != nil {
		return Defaults(), err
	}
	return parsed, nil
}

// Save writes the config to .grapes/config.toml.
func Save(issuesDir string, cfg Config) error {
	path := filepath.Join(issuesDir, "config.toml")
	raw, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return fsutil.WriteFile(path, raw, 0o644)
}
