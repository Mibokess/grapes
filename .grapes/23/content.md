## Goal
Add an interactive settings screen to the TUI with config persistence in `.grapes/config.toml`.

## Context
All colors, keybindings, and defaults are currently hardcoded. This adds:
- `internal/config/config.go` — Config struct, Load/Save
- `internal/tui/settings/settings.go` — Settings screen
- `ApplyTheme()` / `ApplyKeys()` functions in common package
- Integration into app.go and main.go

## Acceptance Criteria
- [ ] Settings screen accessible via `,` key and header tab
- [ ] Theme colors editable with live preview
- [ ] Keybindings editable
- [ ] Default view/sort configurable
- [ ] Config persists to `.grapes/config.toml`
- [ ] App works without config file (defaults)
- [ ] `go build ./...` and `go test ./...` pass
