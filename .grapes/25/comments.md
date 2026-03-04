### 2026-03-04T12:00
[STARTED] Settings screen ignores all mouse events. Adding MouseClickMsg and MouseWheelMsg handling to settings.Update().

### 2026-03-04T12:15
[FINDINGS] Root cause: two issues found:
1. `settings.Update()` only handled `tea.KeyPressMsg` — no mouse events at all.
2. Settings model had no `topOffset` field (unlike board, list, detail). Mouse Y coords are absolute, so without subtracting the app header height, clicks landed on the wrong row. This is the "offset" bug visible on Theme and Keys categories.

### 2026-03-04T12:30
[DONE] Fixed in 3 commits on `25/settings-mouse-navigation`:
- Added `MouseClickMsg` handling (click categories, click fields, click-to-edit)
- Added `MouseWheelMsg` handling (scroll through long field lists)
- Added `topOffset` field + `SetTopOffset()` method to settings Model
- Updated `app.go` to pass `topOffset` to settings (3 call sites)
- Added 13 tests covering mouse clicks, scrolling, offset, editing guard, keyboard nav
