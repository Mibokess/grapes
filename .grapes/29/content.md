## Goal

Replace the hardcoded dark/light GitHub Primer palettes with a preset-based theme system backed by `go.withmatt.com/themes` (450+ iTerm2-Color-Schemes). Users pick a preset name (e.g., "Nord", "Dracula") and get a full theme derived from its ANSI palette via a fixed mapping. Individual colors can still be overridden on top of a preset.

## Context

### Current system

- `internal/config/config.go`: `ThemeConfig` has `Mode` (auto/light/dark) and 16 hex color fields per mode (dark at top level, light under `[theme.light]`).
- `internal/tui/common/theme.go`: `Theme` struct has ~20 raw `color.Color` fields + pre-built lipgloss styles. `NewTheme(isDark)` calls `setDarkColors()`/`setLightColors()` which hardcode GitHub Primer hex values. `NewThemeFromConfig()` applies per-field overrides from config, then rebuilds styles.
- `internal/tui/settings/settings.go`: Settings UI exposes `theme_mode` (enum) + 16 color fields for the current mode.

### External dependency

`go.withmatt.com/themes` — MIT, zero runtime I/O, 450+ themes compiled into binary.

```go
type Theme struct {
    Name                                           string
    Foreground, Background, Cursor                 string // hex
    Black, Red, Green, Yellow, Blue, Magenta, Cyan, White string // ANSI 0-7
    BrightBlack, BrightRed, BrightGreen, BrightYellow,
    BrightBlue, BrightMagenta, BrightCyan, BrightWhite string // ANSI 8-15
}

func GetTheme(name string) (*Theme, error) // case-insensitive
func ListThemes() []string
```

### ANSI → semantic color mapping

| themes field | → grapes role | Rationale |
|---|---|---|
| `Foreground` | text | Primary text |
| `Background` | surface | Background |
| `BrightBlack` | muted | Gray/secondary text |
| derived (midpoint muted↔surface) | faint | Very dim elements |
| derived (midpoint muted↔surface) | border | Border color |
| `Magenta` | accent | Primary accent |
| derived (darken/lighten accent) | accent_bg | Accent background |
| `Red` | color_urgent, error | Danger/urgent |
| `Yellow` | color_high, color_in_progress | Warning/active |
| `Blue` | color_medium, color_todo | Info |
| `BrightBlack` | color_low, color_backlog, color_cancelled | Dim/inactive |
| `Green` | color_done | Success |

Colors marked "derived" need to be computed (e.g., blend two hex colors). A helper like `blendHex(a, b string, t float64) string` would handle this.

### Config change

```toml
[theme]
mode = "auto"
preset = "Nord"          # NEW — name from go.withmatt.com/themes, or "default"
accent = "#ff00ff"       # per-color override still works, applied on top of preset
```

Precedence: **ANSI mapping from preset → per-color overrides**. When `preset` is empty or `"default"`, fall back to the current hardcoded GitHub Primer palettes (preserving backward compatibility).

### Default presets

- Dark default: current GitHub Primer dark palette (shipped as `"default"` or empty preset)
- Light default: current GitHub Primer light palette
- All 450+ `go.withmatt.com/themes` presets available by name

### Detecting dark vs light for external presets

When `mode = "auto"`, the app detects terminal background. For external presets, the theme's `Background` color luminance determines whether to use the "dark" or "light" code path for derived colors (contrast, pill backgrounds, label palette, glamour style, worktree colors).

### Settings UI change

Add a `preset` field to the Theme category (above the individual color fields). Could be a text input with the preset name, or an enum-style selector. When changed, the live preview updates immediately. Individual color fields show the *resolved* value (from preset + overrides) and editing one creates an override.

## Acceptance Criteria

- [ ] `go.withmatt.com/themes` added as a dependency
- [ ] New `Preset` field on `ThemeConfig` in config, serialized as `preset` in TOML
- [ ] `NewThemeFromConfig` resolves preset → ANSI mapping → semantic colors → per-color overrides → build styles
- [ ] When preset is empty or `"default"`, behavior is identical to current (backward compatible)
- [ ] When a valid preset name is set, all theme colors derive from its ANSI palette
- [ ] Per-color overrides in config take precedence over preset-derived colors
- [ ] Settings UI has a `preset` field that live-previews theme changes
- [ ] Derived colors (faint, border, accent_bg) are computed from the preset palette, not hardcoded
- [ ] Dark/light detection works correctly: `mode` setting is respected, and for external presets the background luminance drives derived-color logic
- [ ] Invalid preset names fall back to default palette (no crash)

## Verify

```bash
cd /projects/mboss/dev/grapes && go build ./...
```

```bash
cd /projects/mboss/dev/grapes && go vet ./...
```

Manual verification:
1. Run the TUI with no config changes — should look identical to current
2. Set `preset = "Nord"` in config.toml — TUI should render with Nord colors
3. Set `preset = "Dracula"` — TUI should render with Dracula colors
4. Set `preset = "Nord"` + `accent = "#ff0000"` — accent should be red, rest from Nord
5. Open settings, change preset — live preview updates
6. Set `preset = "nonexistent"` — should fall back to default, no crash

## Pass Criteria

- Build and vet pass with zero errors
- Manual tests 1-6 above all behave as described
