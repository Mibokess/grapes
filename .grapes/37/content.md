## Goal
Add text search (/) to the board view, matching the existing search in the list view.

## Context
- List view has text search via `/` key using a `textinput.Model` that filters issues with `data.MatchesQuery`
- Board view (`internal/tui/board/board.go`) has no text search — only structured filters via `f`
- The board groups issues by status columns; search should filter which issues appear in each column

## Acceptance Criteria
- [ ] Pressing `/` on the board view activates a text input for search
- [ ] Typing filters issues across all columns using `data.MatchesQuery`
- [ ] Enter confirms search (leaves filter active, exits input mode)
- [ ] Esc clears search and exits input mode
- [ ] Filter line shown above board when search is active
- [ ] `q` does not quit while search input is focused
- [ ] `BoardSearch` config key added for customization
- [ ] Help bar shows `/` hint on board view
- [ ] Golden files updated

## Verify
```bash
go build ./... && go test ./internal/tui/...
```
