package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Mibokess/grapes/internal/config"
	"github.com/Mibokess/grapes/internal/data"
	"github.com/Mibokess/grapes/internal/tui/common"
	"github.com/Mibokess/grapes/internal/tui/testutil"
)

// view renders the model at a fixed size, with ANSI stripped.
func view(t *testing.T, m Model) string {
	t.Helper()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	return testutil.StripANSI(updated.(Model).View().Content)
}

// view.default_sort was accepted by the config and the settings screen, but the
// app started on priority regardless.
func TestNewModel_AppliesConfiguredDefaultSort(t *testing.T) {
	cfg := config.Defaults()
	cfg.View.DefaultSort = "title"

	m := NewModel(data.Workspace{Issues: testutil.SampleIssues()}, nil, t.TempDir(), cfg, "test")
	t.Cleanup(func() {
		if m.watcher != nil {
			m.watcher.Close()
		}
	})

	if m.sortMode != data.SortByTitle {
		t.Errorf("sortMode = %v, want title", m.sortMode.Label())
	}
	if !strings.Contains(view(t, m), "title ▼") {
		t.Error("status bar should show the configured sort")
	}
}

func TestNewModel_UnknownDefaultSortIsReported(t *testing.T) {
	cfg := config.Defaults()
	cfg.View.DefaultSort = "nonsense"

	m := NewModel(data.Workspace{Issues: testutil.SampleIssues()}, nil, t.TempDir(), cfg, "test")
	t.Cleanup(func() {
		if m.watcher != nil {
			m.watcher.Close()
		}
	})

	if m.sortMode != data.SortByPriority {
		t.Errorf("sortMode = %v, want the priority fallback", m.sortMode.Label())
	}
	if !strings.Contains(m.statusMsg, "nonsense") {
		t.Errorf("a bad default_sort should be reported, got %q", m.statusMsg)
	}
}

// Status-bar hints are built from the live keymaps, so a rebound key has to
// show up in the help text.
func TestView_HelpHintsFollowConfiguredKeys(t *testing.T) {
	t.Cleanup(func() { common.ApplyKeys(config.Defaults().Keys) })

	cfg := config.Defaults()
	cfg.Keys.BoardEdit = "x"
	cfg.Keys.Quit = "Q"

	m := NewModel(data.Workspace{Issues: testutil.SampleIssues()}, nil, t.TempDir(), cfg, "test")
	t.Cleanup(func() {
		if m.watcher != nil {
			m.watcher.Close()
		}
	})

	got := view(t, m)
	if !strings.Contains(got, "x edit") {
		t.Errorf("help bar should advertise the rebound edit key:\n%s", got)
	}
	if !strings.Contains(got, "Q quit") {
		t.Errorf("help bar should advertise the rebound quit key:\n%s", got)
	}
}

// An empty key in config.toml must not bind an action to nothing.
func TestApplyKeys_EmptyKeyKeepsDefault(t *testing.T) {
	t.Cleanup(func() { common.ApplyKeys(config.Defaults().Keys) })

	keys := config.Defaults().Keys
	keys.BoardOpen = ""
	common.ApplyKeys(keys)

	if got := common.KeyLabel(common.BoardKeyMap.Open); got != "enter" {
		t.Errorf("BoardKeyMap.Open = %q, want the default \"enter\"", got)
	}
}

// A malformed issue is skipped rather than failing the reload, but the skip has
// to reach the status bar.
func TestRefresh_ReportsSkippedIssues(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".grapes")
	if err := os.MkdirAll(filepath.Join(dir, "1"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "1", "meta.toml"),
		[]byte("title = \"ok\"\nstatus = \"todo\"\npriority = \"low\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	m := NewModel(data.Workspace{}, nil, dir, config.Defaults(), "test")
	t.Cleanup(func() {
		if m.watcher != nil {
			m.watcher.Close()
		}
	})

	// Now break a second issue and reload.
	if err := os.MkdirAll(filepath.Join(dir, "2"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "2", "meta.toml"), []byte("= not toml ="), 0o644); err != nil {
		t.Fatal(err)
	}

	// Update only schedules the load; the command carries it out.
	updated, _ := m.Update(common.RefreshMsg{})
	m = updated.(Model)
	updated, _ = m.Update(m.loadWorkspaceCmd()())
	got := updated.(Model)

	if len(got.issues) != 1 {
		t.Errorf("loaded %d issues, want the one good issue", len(got.issues))
	}
	if !strings.Contains(got.statusMsg, "#2") {
		t.Errorf("status bar should name the skipped issue, got %q", got.statusMsg)
	}
}
