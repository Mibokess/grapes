package config

import (
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"
)

// ViewConfig controls startup defaults.
type ViewConfig struct {
	DefaultScreen string `toml:"default_screen"`
	DefaultSort   string `toml:"default_sort"`
}

// ThemeConfig holds all UI color values as hex strings.
type ThemeConfig struct {
	Accent   string `toml:"accent"`
	AccentBg string `toml:"accent_bg"`
	Border   string `toml:"border"`
	Text     string `toml:"text"`
	Muted    string `toml:"muted"`
	Faint    string `toml:"faint"`
	Surface  string `toml:"surface"`

	ColorBacklog    string `toml:"color_backlog"`
	ColorTodo       string `toml:"color_todo"`
	ColorInProgress string `toml:"color_in_progress"`
	ColorDone       string `toml:"color_done"`
	ColorCancelled  string `toml:"color_cancelled"`

	ColorUrgent string `toml:"color_urgent"`
	ColorHigh   string `toml:"color_high"`
	ColorMedium string `toml:"color_medium"`
	ColorLow    string `toml:"color_low"`
}

// KeysConfig holds customizable keybinding strings.
type KeysConfig struct {
	Quit string `toml:"quit"`

	BoardUp       string `toml:"board_up"`
	BoardDown     string `toml:"board_down"`
	BoardLeft     string `toml:"board_left"`
	BoardRight    string `toml:"board_right"`
	BoardOpen     string `toml:"board_open"`
	BoardEdit     string `toml:"board_edit"`
	BoardToList   string `toml:"board_to_list"`
	BoardFilter   string `toml:"board_filter"`
	BoardStatus   string `toml:"board_status"`
	BoardPriority string `toml:"board_priority"`
	BoardSort     string `toml:"board_sort"`
	BoardReverse  string `toml:"board_reverse"`

	ListUp        string `toml:"list_up"`
	ListDown      string `toml:"list_down"`
	ListOpen      string `toml:"list_open"`
	ListEdit      string `toml:"list_edit"`
	ListToBoard   string `toml:"list_to_board"`
	ListSearch    string `toml:"list_search"`
	ListFilter    string `toml:"list_filter"`
	ListStatus    string `toml:"list_status"`
	ListPriority  string `toml:"list_priority"`
	ListSort      string `toml:"list_sort"`
	ListReverse   string `toml:"list_reverse"`

	DetailBack     string `toml:"detail_back"`
	DetailToBoard  string `toml:"detail_to_board"`
	DetailToList   string `toml:"detail_to_list"`
	DetailStatus   string `toml:"detail_status"`
	DetailPriority string `toml:"detail_priority"`
	DetailComment  string `toml:"detail_comment"`
	DetailEdit     string `toml:"detail_edit"`
}

// Config is the full application configuration.
type Config struct {
	View  ViewConfig  `toml:"view"`
	Theme ThemeConfig `toml:"theme"`
	Keys  KeysConfig  `toml:"keys"`
}

// Defaults returns the default configuration matching the hardcoded values.
func Defaults() Config {
	return Config{
		View: ViewConfig{
			DefaultScreen: "board",
			DefaultSort:   "priority",
		},
		Theme: ThemeConfig{
			Accent:          "#a371f7",
			AccentBg:        "#2d1b69",
			Border:          "#30363d",
			Text:            "#e6edf3",
			Muted:           "#8b949e",
			Faint:           "#484f58",
			Surface:         "#161b22",
			ColorBacklog:    "#8b949e",
			ColorTodo:       "#388bfd",
			ColorInProgress: "#d29922",
			ColorDone:       "#3fb950",
			ColorCancelled:  "#6e7681",
			ColorUrgent:     "#f85149",
			ColorHigh:       "#d29922",
			ColorMedium:     "#388bfd",
			ColorLow:        "#6e7681",
		},
		Keys: KeysConfig{
			Quit:          "q",
			BoardUp:       "k",
			BoardDown:     "j",
			BoardLeft:     "h",
			BoardRight:    "l",
			BoardOpen:     "enter",
			BoardEdit:     "e",
			BoardToList:   "L",
			BoardFilter:   "f",
			BoardStatus:   "s",
			BoardPriority: "p",
			BoardSort:     "o",
			BoardReverse:  "O",
			ListUp:        "k",
			ListDown:      "j",
			ListOpen:      "enter",
			ListEdit:      "e",
			ListToBoard:   "B",
			ListSearch:    "/",
			ListFilter:    "f",
			ListStatus:    "s",
			ListPriority:  "p",
			ListSort:      "o",
			ListReverse:   "O",
			DetailBack:     "esc",
			DetailToBoard:  "B",
			DetailToList:   "l",
			DetailStatus:   "s",
			DetailPriority: "p",
			DetailComment:  "c",
			DetailEdit:     "e",
		},
	}
}

// Load reads config from .grapes/config.toml, falling back to defaults.
func Load(issuesDir string) Config {
	cfg := Defaults()
	path := filepath.Join(issuesDir, "config.toml")
	raw, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}
	_ = toml.Unmarshal(raw, &cfg)
	return cfg
}

// Save writes the config to .grapes/config.toml.
func Save(issuesDir string, cfg Config) error {
	path := filepath.Join(issuesDir, "config.toml")
	raw, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0644)
}
