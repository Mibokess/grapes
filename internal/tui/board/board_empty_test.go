package board_test

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Mibokess/grapes/internal/data"
	"github.com/Mibokess/grapes/internal/tui/board"
	"github.com/Mibokess/grapes/internal/tui/common"
	"github.com/Mibokess/grapes/internal/tui/testutil"
)

// emptyBoard returns a board that is hiding empty columns and has no issues —
// the state the app reaches whenever a filter matches nothing.
func emptyBoard() board.Model {
	return board.New(nil).SetHideEmpty(true).SetSize(100, 30).SetIssues(nil)
}

// Pressing down on a board with no visible columns used to index into an empty
// slice and take the whole TUI with it.
func TestBoard_NoColumns_KeysDoNotPanic(t *testing.T) {
	keys := []string{"j", "k", "h", "l", "e", "s", "p", "t", "o", "O", "E"}
	for _, k := range keys {
		t.Run(k, func(t *testing.T) {
			m := emptyBoard()
			m, _ = m.Update(keyMsg(k))
			m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
			_ = m.View()
		})
	}
}

func TestBoard_NoColumns_MouseDoesNotPanic(t *testing.T) {
	m := emptyBoard()
	m, _ = m.Update(tea.MouseClickMsg{X: 10, Y: 10, Button: tea.MouseLeft})
	m, _ = m.Update(tea.MouseMotionMsg{X: 40, Y: 12, Button: tea.MouseLeft})
	m, _ = m.Update(tea.MouseReleaseMsg{X: 40, Y: 12, Button: tea.MouseLeft})
	m, _ = m.Update(tea.MouseWheelMsg{X: 10, Y: 10, Button: tea.MouseWheelDown})
	_ = m.View()
}

// A filtered-empty board can become non-empty on refresh. Its previous
// SetSize call may have computed zero visible columns, so rebuilding must
// restore a usable column count before rendering.
func TestBoard_EmptyThenNonEmpty_Renders(t *testing.T) {
	m := emptyBoard()
	m = m.SetIssues([]data.Issue{{ID: 7, Title: "appeared", Status: data.StatusTodo, Priority: data.PriorityLow}})
	view := testutil.StripANSI(m.View())
	if !strings.Contains(view, "appeared") {
		t.Fatalf("refreshed board did not render issue:\n%s", view)
	}
}

// SetHideEmpty has to regroup: otherwise the board keeps showing the columns it
// built before the option was applied.
func TestBoard_SetHideEmpty_RegroupsColumns(t *testing.T) {
	issues := []data.Issue{
		{ID: 1, Title: "only todo", Status: data.StatusTodo, Priority: data.PriorityLow, Created: time.Now()},
	}
	shown := board.New(issues).SetSize(120, 40)

	withEmpty := shown.SetHideEmpty(false).View()
	if !strings.Contains(withEmpty, "BACKLOG") {
		t.Error("with hide-empty off, the empty backlog column should be visible")
	}

	hidden := shown.SetHideEmpty(true).View()
	if strings.Contains(hidden, "BACKLOG") {
		t.Errorf("with hide-empty on, empty columns should be gone:\n%s", hidden)
	}
	if !strings.Contains(hidden, "TODO") {
		t.Error("the non-empty column should still be there")
	}
}

// Priority is a string type, so an unguarded `<=` compares alphabetically and
// urgent ("urgent" > "high") loses its marker.
func TestBoard_UrgentCardShowsPriorityIcon(t *testing.T) {
	tests := []struct {
		priority data.Priority
		wantIcon bool
	}{
		{data.PriorityUrgent, true},
		{data.PriorityHigh, true},
		{data.PriorityMedium, false},
		{data.PriorityLow, false},
	}
	for _, tt := range tests {
		t.Run(string(tt.priority), func(t *testing.T) {
			issues := []data.Issue{{ID: 1, Title: "card", Status: data.StatusTodo, Priority: tt.priority}}
			view := testutil.StripANSI(board.New(issues).SetSize(120, 40).View())

			icon := strings.TrimSpace(common.PriorityIcon(tt.priority))
			has := icon != "" && strings.Contains(view, "#1 "+icon)
			if has != tt.wantIcon {
				t.Errorf("priority %s: icon shown = %v, want %v\n%s", tt.priority, has, tt.wantIcon, view)
			}
		})
	}
}

// Every worktree holds a copy of every issue, so a badge that merely meant "a
// copy exists elsewhere" would appear on nearly every card. It marks ownership:
// a worktree working on this issue ahead of the main checkout.
func TestBoard_SourceBadgeMarksOwnershipOnly(t *testing.T) {
	// Owned: a worktree has the newest version of this issue.
	owned := data.Issue{
		ID: 1, Title: "Owned by a worktree", Status: data.StatusTodo,
		Priority: data.PriorityMedium, Worktree: "worker",
		Sources: []data.IssueSource{{Name: ""}, {Name: "worker"}},
	}
	// Diverged: a worktree also has a version, but the main checkout's is newer.
	diverged := data.Issue{
		ID: 2, Title: "Main is newer", Status: data.StatusTodo,
		Priority: data.PriorityMedium, Worktree: "",
		Sources: []data.IssueSource{{Name: ""}, {Name: "worker"}},
	}

	render := func(issues ...data.Issue) string {
		m := board.New(issues).SetWorktreeNames([]string{"worker"}).SetSize(120, 40)
		return testutil.StripANSI(m.View())
	}

	if got := render(owned); !strings.Contains(got, common.WorktreeIcon()) {
		t.Errorf("an issue owned by a worktree should be marked, got:\n%s", got)
	}
	if got := render(diverged); strings.Contains(got, common.WorktreeIcon()) {
		t.Errorf("an issue the main checkout owns should carry no badge, got:\n%s", got)
	}
}
