## Goal
Fix the misaligned `│` separator on the settings screen when the left category pane has fewer rows than the right fields pane.

## Context
- File: `internal/tui/settings/settings.go`, `View()` method (line ~414)
- Category labels are formatted with `fmt.Sprintf("  %-*s", catW-4, cat.name)` producing strings of width `catW - 2` (16 chars when `catW=18`).
- Padding lines for the left column use `strings.Repeat(" ", catW)` (18 chars).
- This 2-character mismatch causes the `│` separator to shift right on rows below the last category.
- Visible on the Theme category page where 17 fields exceed the 3 categories.
- The golden test file `testdata/TestSettingsView_ThemeCategory.golden` was generated with the bug, so it asserts the wrong output.

## Acceptance Criteria
- [ ] Padding lines in the left column are `catW - 2` chars wide, matching the category label width.
- [ ] The `│` separator is vertically aligned on all rows of the settings screen.
- [ ] Golden file `TestSettingsView_ThemeCategory.golden` reflects the corrected alignment.
- [ ] All settings tests pass.

## Verify
```bash
go test ./internal/tui/settings/ -v
```

## Pass Criteria
All tests pass. The golden file shows the `│` separator at the same column on every line.
