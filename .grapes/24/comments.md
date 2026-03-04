## Comments

### 2026-03-04

[STARTED] Implementing adaptive theme support with Theme struct + BackgroundColorMsg approach.

[DONE] All compilation errors fixed, build and tests pass. Implementation:
- `Theme` struct in `common/theme.go` with dark and light palettes via `NewTheme(isDark)`
- Global `T` variable + `ApplyTheme(cfg)` for settings live preview
- App detects background via `tea.RequestBackgroundColor` / `BackgroundColorMsg` in `Init()`/`Update()`
- Theme propagated to board, list, detail, picker, and filter views via constructors + `SetTheme()`
- Glamour markdown uses `theme.GlamourStyle` ("dark"/"light")
- Config overrides via `NewThemeFromConfig(cfg)` + settings screen
- All tests pass
