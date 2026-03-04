## Goal

Add adaptive light/dark theme support so the TUI is readable on both dark and light terminal backgrounds.

## Context

All colors in `internal/tui/common/theme.go` are hardcoded for dark terminals (GitHub dark palette). On light terminals (e.g., VS Code with a light theme), the terminal background shows through and text is unreadable.

Bubble Tea v2 provides `BackgroundColorMsg` for detecting terminal background. The approach: create a `Theme` struct, detect background in `Init()`, and propagate the theme to all views.

## Acceptance Criteria

- [ ] `Theme` struct in `common/theme.go` with dark and light palettes
- [ ] App detects terminal background via `tea.RequestBackgroundColor` / `BackgroundColorMsg`
- [ ] Theme propagated to board, list, detail, picker, and filter views
- [ ] All hardcoded colors in picker, filter, list, app consolidated into theme
- [ ] Glamour markdown rendering uses appropriate style ("dark"/"light")
- [ ] All tests pass

## Verify

```bash
go build ./...
go test ./...
```

## Pass Criteria

Build succeeds, all tests pass. TUI renders correctly on both dark and light terminal backgrounds.
