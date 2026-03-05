### 2026-03-05T12:00
[FINDINGS] Investigation of current label state in TUI:

**Display (working)**:
- Detail view: colored pills (`detail.go:252-258`)
- Board view: colored text on cards (`board.go:525-547`)
- List view: column text (`list.go:509-513`)
- Filter: multi-select picker via filter menu (`filter/picker.go`)
- Search: labels included in text matching (`data/search.go:51-55`)

**Missing**:
- No label picker/editor component
- No keybinding to trigger label editing
- No click zones for label pills
- No `UpdateField` support for array fields
- No `*_label` key config entries

**Reference patterns**:
- Single-select picker: `picker/picker.go` (status/priority)
- Multi-select picker: `filter/picker.go` (filter labels)
- Message flow: `ShowPickerMsg` → `PickerResultMsg` in `app.go:503-536`
- All labels collection: `app.go:1017-1031` (`collectAllLabels()`)
