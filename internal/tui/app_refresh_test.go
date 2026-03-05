package tui

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/Mibokess/grapes/internal/config"
	"github.com/Mibokess/grapes/internal/data"
	"github.com/Mibokess/grapes/internal/tui/common"
	tea "charm.land/bubbletea/v2"
)

// createTestIssue writes a minimal meta.toml for an issue.
func createTestIssue(t *testing.T, grapesDir string, id int, title, status, priority string, extra string) {
	t.Helper()
	issueDir := filepath.Join(grapesDir, strconv.Itoa(id))
	if err := os.MkdirAll(issueDir, 0755); err != nil {
		t.Fatal(err)
	}
	meta := "title = '" + title + "'\n" +
		"status = '" + status + "'\n" +
		"priority = '" + priority + "'\n" +
		"labels = []\n" +
		"created = 2025-01-01T10:00:00Z\n" +
		"updated = 2025-01-01T10:00:00Z\n"
	if extra != "" {
		meta += extra + "\n"
	}
	if err := os.WriteFile(filepath.Join(issueDir, "meta.toml"), []byte(meta), 0644); err != nil {
		t.Fatal(err)
	}
}

// newTestModel creates a Model backed by a real temp directory.
// Returns the model and issues dir. The watcher is closed on test cleanup.
func newTestModel(t *testing.T, setup func(grapesDir string)) Model {
	t.Helper()
	dir := t.TempDir()
	grapesDir := filepath.Join(dir, ".grapes")
	if err := os.MkdirAll(grapesDir, 0755); err != nil {
		t.Fatal(err)
	}
	setup(grapesDir)
	issues, err := data.LoadAllIssues(grapesDir)
	if err != nil {
		t.Fatal(err)
	}
	m := NewModel(issues, grapesDir, config.Defaults(), "test")
	if m.watcher != nil {
		t.Cleanup(func() { m.watcher.Close() })
	}
	// Give it a size
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	return updated.(Model)
}

// execCmd runs a tea.Cmd and returns the resulting message (nil if cmd is nil).
func execCmd(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}
	return cmd()
}

// refreshModel sends a RefreshMsg to the model and returns the updated model.
func refreshModel(t *testing.T, m Model) Model {
	t.Helper()
	updated, _ := m.Update(common.RefreshMsg{})
	return updated.(Model)
}

// findIssue returns the issue with the given ID from the model's issues slice.
func findIssue(m Model, id int) *data.Issue {
	for i := range m.issues {
		if m.issues[i].ID == id {
			return &m.issues[i]
		}
	}
	return nil
}

// --- Tests ---

func TestApp_Refresh_MoveIssueMsg_ChangesStatus(t *testing.T) {
	m := newTestModel(t, func(dir string) {
		createTestIssue(t, dir, 1, "Test issue", "todo", "medium", "")
	})

	// Verify initial status
	iss := findIssue(m, 1)
	if iss == nil {
		t.Fatal("issue 1 not found")
	}
	if iss.Status != data.StatusTodo {
		t.Fatalf("expected initial status 'todo', got %q", iss.Status)
	}

	// Send MoveIssueMsg (simulates drag-drop on board)
	updated, cmd := m.Update(common.MoveIssueMsg{IssueID: 1, NewStatus: data.StatusInProgress})
	m = updated.(Model)

	// Execute the write command
	msg := execCmd(cmd)
	if msg != nil {
		t.Fatalf("expected nil (success), got %T: %v", msg, msg)
	}

	// Send RefreshMsg (simulates what fsnotify would do)
	m = refreshModel(t, m)

	// Verify the status changed
	iss = findIssue(m, 1)
	if iss == nil {
		t.Fatal("issue 1 not found after refresh")
	}
	if iss.Status != data.StatusInProgress {
		t.Errorf("expected status 'in_progress' after refresh, got %q", iss.Status)
	}
}

func TestApp_Refresh_PickerResult_ChangesStatus(t *testing.T) {
	m := newTestModel(t, func(dir string) {
		createTestIssue(t, dir, 1, "Test issue", "backlog", "low", "")
	})

	// Send PickerResultMsg for status change
	updated, cmd := m.Update(common.PickerResultMsg{IssueID: 1, Field: "status", Value: "done"})
	m = updated.(Model)

	msg := execCmd(cmd)
	if msg != nil {
		t.Fatalf("expected nil (success), got %T: %v", msg, msg)
	}

	m = refreshModel(t, m)

	iss := findIssue(m, 1)
	if iss == nil {
		t.Fatal("issue 1 not found after refresh")
	}
	if iss.Status != data.StatusDone {
		t.Errorf("expected status 'done' after refresh, got %q", iss.Status)
	}
}

func TestApp_Refresh_PickerResult_ChangesPriority(t *testing.T) {
	m := newTestModel(t, func(dir string) {
		createTestIssue(t, dir, 1, "Test issue", "todo", "low", "")
	})

	updated, cmd := m.Update(common.PickerResultMsg{IssueID: 1, Field: "priority", Value: "urgent"})
	m = updated.(Model)

	msg := execCmd(cmd)
	if msg != nil {
		t.Fatalf("expected nil (success), got %T: %v", msg, msg)
	}

	m = refreshModel(t, m)

	iss := findIssue(m, 1)
	if iss == nil {
		t.Fatal("issue 1 not found after refresh")
	}
	if iss.Priority != data.PriorityUrgent {
		t.Errorf("expected priority 'urgent' after refresh, got %q", iss.Priority)
	}
}

func TestApp_Refresh_SubissueStatusChange_VisibleInParent(t *testing.T) {
	m := newTestModel(t, func(dir string) {
		createTestIssue(t, dir, 1, "Parent issue", "in_progress", "high", "")
		createTestIssue(t, dir, 2, "Child issue", "todo", "high", "parent = 1")
	})

	// Verify child starts as todo
	child := findIssue(m, 2)
	if child == nil {
		t.Fatal("child issue 2 not found")
	}
	if child.Status != data.StatusTodo {
		t.Fatalf("expected child initial status 'todo', got %q", child.Status)
	}

	// Navigate to parent detail view
	updated, _ := m.Update(common.OpenDetailMsg{ID: 1})
	m = updated.(Model)
	if m.screen != common.ScreenDetail {
		t.Fatal("expected to be on detail screen")
	}

	// Change child status via picker
	updated, cmd := m.Update(common.PickerResultMsg{IssueID: 2, Field: "status", Value: "done"})
	m = updated.(Model)

	msg := execCmd(cmd)
	if msg != nil {
		t.Fatalf("expected nil (success), got %T: %v", msg, msg)
	}

	// Refresh
	m = refreshModel(t, m)

	// Verify child status updated in model
	child = findIssue(m, 2)
	if child == nil {
		t.Fatal("child issue 2 not found after refresh")
	}
	if child.Status != data.StatusDone {
		t.Errorf("expected child status 'done' after refresh, got %q", child.Status)
	}

	// Verify parent still on detail screen and detail was refreshed
	if m.screen != common.ScreenDetail {
		t.Error("should still be on detail screen after refresh")
	}
	if m.detail.IssueID() != 1 {
		t.Errorf("detail should show parent issue 1, got %d", m.detail.IssueID())
	}
}

func TestApp_Refresh_MultipleRapidChanges(t *testing.T) {
	m := newTestModel(t, func(dir string) {
		createTestIssue(t, dir, 1, "Issue A", "backlog", "low", "")
		createTestIssue(t, dir, 2, "Issue B", "todo", "medium", "")
		createTestIssue(t, dir, 3, "Issue C", "in_progress", "high", "")
	})

	// Change all three issues without refreshing between changes
	updated, cmd1 := m.Update(common.MoveIssueMsg{IssueID: 1, NewStatus: data.StatusTodo})
	m = updated.(Model)
	msg := execCmd(cmd1)
	if msg != nil {
		t.Fatalf("write 1 failed: %v", msg)
	}

	updated, cmd2 := m.Update(common.PickerResultMsg{IssueID: 2, Field: "status", Value: "done"})
	m = updated.(Model)
	msg = execCmd(cmd2)
	if msg != nil {
		t.Fatalf("write 2 failed: %v", msg)
	}

	updated, cmd3 := m.Update(common.PickerResultMsg{IssueID: 3, Field: "priority", Value: "low"})
	m = updated.(Model)
	msg = execCmd(cmd3)
	if msg != nil {
		t.Fatalf("write 3 failed: %v", msg)
	}

	// Single RefreshMsg after all writes (simulates debounced fsnotify)
	m = refreshModel(t, m)

	// Verify all changes are reflected
	iss1 := findIssue(m, 1)
	if iss1 == nil || iss1.Status != data.StatusTodo {
		t.Errorf("issue 1: expected status 'todo', got %q", iss1.Status)
	}
	iss2 := findIssue(m, 2)
	if iss2 == nil || iss2.Status != data.StatusDone {
		t.Errorf("issue 2: expected status 'done', got %q", iss2.Status)
	}
	iss3 := findIssue(m, 3)
	if iss3 == nil || iss3.Priority != data.PriorityLow {
		t.Errorf("issue 3: expected priority 'low', got %q", iss3.Priority)
	}
}

func TestApp_Refresh_ExternalFileChange(t *testing.T) {
	m := newTestModel(t, func(dir string) {
		createTestIssue(t, dir, 1, "Original title", "todo", "medium", "")
	})

	// Verify initial state
	iss := findIssue(m, 1)
	if iss == nil || iss.Title != "Original title" {
		t.Fatal("initial issue not loaded correctly")
	}

	// Simulate an external tool modifying the file directly (not through TUI)
	metaPath := filepath.Join(m.issuesDir, "1", "meta.toml")
	newMeta := "title = 'Modified externally'\n" +
		"status = 'in_progress'\n" +
		"priority = 'urgent'\n" +
		"labels = []\n" +
		"created = 2025-01-01T10:00:00Z\n" +
		"updated = 2025-01-02T10:00:00Z\n"
	if err := os.WriteFile(metaPath, []byte(newMeta), 0644); err != nil {
		t.Fatal(err)
	}

	m = refreshModel(t, m)

	iss = findIssue(m, 1)
	if iss == nil {
		t.Fatal("issue 1 not found after refresh")
	}
	if iss.Title != "Modified externally" {
		t.Errorf("expected title 'Modified externally', got %q", iss.Title)
	}
	if iss.Status != data.StatusInProgress {
		t.Errorf("expected status 'in_progress', got %q", iss.Status)
	}
	if iss.Priority != data.PriorityUrgent {
		t.Errorf("expected priority 'urgent', got %q", iss.Priority)
	}
}

func TestApp_Refresh_NewIssuePickedUp(t *testing.T) {
	m := newTestModel(t, func(dir string) {
		createTestIssue(t, dir, 1, "Existing issue", "todo", "medium", "")
	})

	if len(m.issues) != 1 {
		t.Fatalf("expected 1 issue initially, got %d", len(m.issues))
	}

	// Create a new issue on disk (simulates agent creating an issue)
	createTestIssue(t, m.issuesDir, 2, "New issue", "backlog", "low", "")

	m = refreshModel(t, m)

	if len(m.issues) != 2 {
		t.Fatalf("expected 2 issues after refresh, got %d", len(m.issues))
	}
	iss := findIssue(m, 2)
	if iss == nil {
		t.Fatal("new issue 2 not found after refresh")
	}
	if iss.Title != "New issue" {
		t.Errorf("expected title 'New issue', got %q", iss.Title)
	}
}

func TestApp_Refresh_BoardReflectsChanges(t *testing.T) {
	m := newTestModel(t, func(dir string) {
		createTestIssue(t, dir, 1, "Board test", "todo", "medium", "")
	})

	// Ensure we're on the board screen
	if m.screen != common.ScreenBoard {
		t.Fatal("expected to start on board screen")
	}

	// Change status
	updated, cmd := m.Update(common.MoveIssueMsg{IssueID: 1, NewStatus: data.StatusDone})
	m = updated.(Model)
	execCmd(cmd)

	m = refreshModel(t, m)

	// Verify the board's issues reflect the change
	boardView := m.board.View()
	if boardView == "" {
		t.Error("board view should not be empty")
	}
	// The issue should now be in the "done" column, not "todo"
	iss := findIssue(m, 1)
	if iss == nil || iss.Status != data.StatusDone {
		t.Errorf("board model should reflect status 'done', got %q", iss.Status)
	}
}

func TestApp_Refresh_DetailViewRecreatedOnRefresh(t *testing.T) {
	m := newTestModel(t, func(dir string) {
		createTestIssue(t, dir, 1, "Detail refresh test", "todo", "medium", "")
	})

	// Navigate to detail view
	updated, _ := m.Update(common.OpenDetailMsg{ID: 1})
	m = updated.(Model)
	if m.screen != common.ScreenDetail {
		t.Fatal("expected detail screen")
	}

	// Change status while viewing detail
	updated, cmd := m.Update(common.PickerResultMsg{IssueID: 1, Field: "status", Value: "in_progress"})
	m = updated.(Model)
	execCmd(cmd)

	m = refreshModel(t, m)

	// Verify detail view was recreated and shows updated data
	if m.screen != common.ScreenDetail {
		t.Error("should still be on detail screen")
	}
	if m.detail.IssueID() != 1 {
		t.Errorf("detail should show issue 1, got %d", m.detail.IssueID())
	}
	iss := findIssue(m, 1)
	if iss == nil || iss.Status != data.StatusInProgress {
		t.Errorf("expected status 'in_progress' after detail refresh, got %q", iss.Status)
	}
}

func TestApp_Refresh_WriteError_ShowsStatus(t *testing.T) {
	m := newTestModel(t, func(dir string) {
		createTestIssue(t, dir, 1, "Error test", "todo", "medium", "")
	})

	// Try to change status of non-existent issue
	// issueSourceDir will return m.issuesDir, but UpdateField will fail
	// because issue 999 doesn't exist
	updated, cmd := m.Update(common.MoveIssueMsg{IssueID: 999, NewStatus: data.StatusDone})
	m = updated.(Model)
	msg := execCmd(cmd)

	if msg == nil {
		t.Fatal("expected WriteErrMsg for non-existent issue")
	}
	if _, ok := msg.(common.WriteErrMsg); !ok {
		t.Fatalf("expected WriteErrMsg, got %T", msg)
	}

	// Process the error message
	updated, _ = m.Update(msg)
	m = updated.(Model)

	if m.statusMsg == "" {
		t.Error("expected status bar to show error message")
	}
}
