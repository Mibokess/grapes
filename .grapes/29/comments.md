### 2026-03-04T12:00
[STARTED] Created issue from discussion about adding theme presets. Key design decisions:
- Use `go.withmatt.com/themes` for 450+ presets via ANSI color mapping
- Keep current GitHub Primer palettes as the default (backward compatible)
- Preset + per-color overrides layering
- Derived colors (faint, border, accent_bg) computed from palette
