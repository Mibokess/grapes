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
