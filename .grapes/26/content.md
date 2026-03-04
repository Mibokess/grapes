## Goal
Fix `NewThemeFromConfig` to respect terminal dark/light detection so Glamour markdown rendering and base colors adapt correctly on light terminals.

## Context
- `internal/tui/common/theme.go:156` — `NewThemeFromConfig` always calls `NewTheme(true)` (dark), ignoring actual terminal background.
- `internal/tui/app.go:206-216` — `BackgroundColorMsg` handler has `msg.IsDark()` but doesn't pass it to `NewThemeFromConfig`.
- `GlamourStyle` is always `"dark"` when config overrides exist, causing inline code (backtick text like `meta.yaml`) to render with dark backgrounds on light terminals.
- The saved `.grapes/config.toml` `[theme]` section contains hardcoded dark-mode hex colors from the settings screen, which override light-mode colors.

## Root Cause
1. `NewThemeFromConfig` signature takes only `config.ThemeConfig` — no `isDark` parameter.
2. It starts from `NewTheme(true)` so `GlamourStyle` is always `"dark"`.
3. The `BackgroundColorMsg` handler doesn't pass `isDark` through when config overrides exist.

## Fix
1. Change `NewThemeFromConfig` to accept `isDark bool` parameter: `NewThemeFromConfig(cfg config.ThemeConfig, isDark bool)`.
2. Start from `NewTheme(isDark)` instead of `NewTheme(true)` so the correct base palette and `GlamourStyle` are set before applying overrides.
3. Update the call site in `app.go` to pass `msg.IsDark()`.

## Acceptance Criteria
- [ ] `NewThemeFromConfig` accepts `isDark bool` and starts from `NewTheme(isDark)`.
- [ ] `BackgroundColorMsg` handler passes `msg.IsDark()` to `NewThemeFromConfig`.
- [ ] On a light terminal with theme config, `GlamourStyle` is `"light"` and base colors are light-mode.
- [ ] On a dark terminal with theme config, behavior is unchanged (dark base + overrides).
- [ ] All existing tests pass.

## Verify
```bash
cd /projects/mboss/dev/grapes && go build ./... && go test ./...
```

## Pass Criteria
Build succeeds. All tests pass. `NewThemeFromConfig` uses `NewTheme(isDark)` not `NewTheme(true)`.
