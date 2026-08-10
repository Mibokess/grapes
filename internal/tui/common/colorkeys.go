package common

import "github.com/Mibokess/grapes/internal/config"

// ColorKeys lists every themeable color, in the order the settings screen shows
// them. The key strings are also the TOML field names in config.toml.
var ColorKeys = []string{
	"accent",
	"accent_bg",
	"border",
	"text",
	"muted",
	"faint",
	"surface",
	"color_backlog",
	"color_todo",
	"color_in_progress",
	"color_done",
	"color_cancelled",
	"color_urgent",
	"color_high",
	"color_medium",
	"color_low",
}

// ColorLabels maps each color key to its settings-screen label.
var ColorLabels = map[string]string{
	"accent":            "Accent",
	"accent_bg":         "Accent BG",
	"border":            "Border",
	"text":              "Text",
	"muted":             "Muted",
	"faint":             "Faint",
	"surface":           "Surface",
	"color_backlog":     "Backlog",
	"color_todo":        "Todo",
	"color_in_progress": "In Progress",
	"color_done":        "Done",
	"color_cancelled":   "Cancelled",
	"color_urgent":      "Urgent",
	"color_high":        "High",
	"color_medium":      "Medium",
	"color_low":         "Low",
}

// colorField returns a pointer to the field named by key. Returning the pointer
// keeps get and set from drifting apart, which is what two parallel switch
// statements over sixteen keys eventually do.
func colorField(c *config.ColorSetConfig, key string) *string {
	switch key {
	case "accent":
		return &c.Accent
	case "accent_bg":
		return &c.AccentBg
	case "border":
		return &c.Border
	case "text":
		return &c.Text
	case "muted":
		return &c.Muted
	case "faint":
		return &c.Faint
	case "surface":
		return &c.Surface
	case "color_backlog":
		return &c.ColorBacklog
	case "color_todo":
		return &c.ColorTodo
	case "color_in_progress":
		return &c.ColorInProgress
	case "color_done":
		return &c.ColorDone
	case "color_cancelled":
		return &c.ColorCancelled
	case "color_urgent":
		return &c.ColorUrgent
	case "color_high":
		return &c.ColorHigh
	case "color_medium":
		return &c.ColorMedium
	case "color_low":
		return &c.ColorLow
	}
	return nil
}

// IsColorKey reports whether key names a themeable color.
func IsColorKey(key string) bool {
	return colorField(&config.ColorSetConfig{}, key) != nil
}

// GetColor returns the value of the named color, or "" for an unknown key.
func GetColor(c config.ColorSetConfig, key string) string {
	if p := colorField(&c, key); p != nil {
		return *p
	}
	return ""
}

// SetColor writes the named color. Unknown keys are ignored.
func SetColor(c *config.ColorSetConfig, key, val string) {
	if p := colorField(c, key); p != nil {
		*p = val
	}
}
