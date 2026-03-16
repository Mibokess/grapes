### 2026-03-16T14:10
[STARTED] Fix detail view not re-rendering on terminal resize. The fix is to call `renderIssue()` inside `SetSize()`, guarded by `m.ready`.

### 2026-03-16T14:11
[DONE] Added `renderIssue()` call in `SetSize()`. All tests pass. Updated pre-existing stale golden file for narrow test.
