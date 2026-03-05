## Goal
Make overridden theme colors visually distinct in the TUI settings page and provide a way to reset them to defaults.

## Context
- Settings page: `internal/tui/settings/settings.go`
- Config model: `internal/config/config.go` — `ColorSetConfig`, `ThemeConfig`, `Defaults()`
- Theme building: `internal/tui/common/theme.go` — `NewThemeFromConfig`, `applyColorOverrides`
- When a user edits a color in the Theme settings category, it overwrites the value in `cfg.Theme`. There is currently no visual indicator that a color differs from the theme's default, and no way to reset colors back to defaults.
- Default colors come from `config.Defaults()` for the current mode (dark/light).
- The Theme category has 16 color fields (accent, accent_bg, border, text, muted, faint, surface, plus status and priority colors).

## Acceptance Criteria
- [ ] Color fields whose value differs from `config.Defaults()` for the current mode are rendered with bold value text.
- [ ] When any color is overridden, a "Reset colors" action appears as the last item in the Theme category field list.
- [ ] Selecting "Reset colors" (Enter) resets all color fields for the current mode to `config.Defaults()` values, rebuilds the theme, and sends a `ThemeMsg` for live preview.
- [ ] When no colors are overridden, the "Reset colors" item does not appear.
- [ ] Navigation (up/down, scroll, click) works correctly with the dynamic reset row.
- [ ] Project compiles and existing tests pass.

## Verify
```bash
cd /projects/mboss/dev/grapes && go build ./...
cd /projects/mboss/dev/grapes && go test ./...
```

## Pass Criteria
- Build succeeds with no errors.
- All tests pass.
- Manual: open settings → Theme, edit a color → value turns bold, "Reset colors" appears at bottom. Activate reset → colors revert, bold goes away, reset row disappears.
