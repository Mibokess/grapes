package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Mibokess/grapes/internal/config"
	"github.com/Mibokess/grapes/internal/data"
	"github.com/Mibokess/grapes/internal/tui/common"
	"github.com/Mibokess/grapes/internal/tui/testutil"
)

// Reloading used to happen inside Update, which froze the event loop for the
// whole load — the reason grapes felt unresponsive once a project had many
// worktrees. Update must now only schedule the work.
func TestRefresh_DoesNotLoadInUpdate(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".grapes")
	if err := os.MkdirAll(filepath.Join(dir, "1"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "1", "meta.toml"),
		[]byte("title = \"first\"\nstatus = \"todo\"\npriority = \"low\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	m := NewModel(data.Workspace{}, nil, dir, config.Defaults(), "test")
	t.Cleanup(func() {
		if m.watcher != nil {
			m.watcher.Close()
		}
	})

	updated, cmd := m.Update(common.RefreshMsg{})
	m = updated.(Model)

	if len(m.issues) != 0 {
		t.Error("Update read the workspace itself; the load belongs in a command")
	}
	if !m.loading {
		t.Error("Update should have marked a load as in flight")
	}
	if cmd == nil {
		t.Fatal("Update should have returned a command to do the loading")
	}

	// The command carries the work, and its result is what updates the model.
	updated, _ = m.Update(m.loadWorkspaceCmd()())
	m = updated.(Model)
	if len(m.issues) != 1 {
		t.Errorf("loaded %d issues after the command ran, want 1", len(m.issues))
	}
	if m.loading {
		t.Error("the load result should clear the in-flight flag")
	}
}

// A second refresh arriving while a load is in flight must not start another.
func TestRefresh_DoesNotStackConcurrentLoads(t *testing.T) {
	m := NewModel(data.Workspace{}, nil, t.TempDir(), config.Defaults(), "test")
	t.Cleanup(func() {
		if m.watcher != nil {
			m.watcher.Close()
		}
	})

	updated, _ := m.Update(common.RefreshMsg{})
	m = updated.(Model)
	before := m.loading

	updated, _ = m.Update(common.RefreshMsg{})
	m = updated.(Model)

	if !before || !m.loading {
		t.Fatal("both refreshes should leave a load in flight")
	}
}

// The source filter lists worktrees that are working on something. Idle
// worktrees are the reason the old list ran to thirty or forty entries.
func TestSourceFilter_ListsOnlyWorkingWorktrees(t *testing.T) {
	m := NewModel(data.Workspace{
		Issues: testutil.SampleIssues(),
		Worktrees: []data.WorktreeInfo{
			{Name: "busy", Touched: []int{1, 2}, Owned: []int{1}},
		},
	}, nil, t.TempDir(), config.Defaults(), "test")
	t.Cleanup(func() {
		if m.watcher != nil {
			m.watcher.Close()
		}
	})

	view := testutil.StripANSI(m.buildFilterPicker("source").View())
	if !strings.Contains(view, "busy (2)") {
		t.Errorf("source filter should name the working worktree and its issue count, got:\n%s", view)
	}
}

// Outside git there are no worktrees, and the filter should say why rather than
// looking identical to a project where every worktree happens to be idle.
func TestSourceFilter_ExplainsMissingAttribution(t *testing.T) {
	m := NewModel(data.Workspace{
		Issues:         testutil.SampleIssues(),
		AttributionErr: os.ErrNotExist,
	}, nil, t.TempDir(), config.Defaults(), "test")
	t.Cleanup(func() {
		if m.watcher != nil {
			m.watcher.Close()
		}
	})

	if !strings.Contains(m.sourcePickerTitle(), "unavailable") {
		t.Errorf("title = %q, want it to explain why worktrees are missing", m.sourcePickerTitle())
	}
}
