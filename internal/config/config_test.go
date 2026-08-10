package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// writeConfig drops a config.toml into a fresh temp dir and returns the dir.
func writeConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
	return dir
}

func TestLoad_MissingFileYieldsDefaults(t *testing.T) {
	cfg, err := Load(t.TempDir())
	if err != nil {
		t.Fatalf("a missing config.toml should not be an error, got %v", err)
	}
	if !reflect.DeepEqual(cfg, Defaults()) {
		t.Error("missing config.toml should yield exactly Defaults()")
	}
}

func TestLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	want := Defaults()
	want.Theme.Preset = "Dracula"
	want.Theme.Mode = "light"
	want.View.DefaultScreen = "list"
	want.View.AutoCloseSubs = true
	want.Keys.BoardOpen = "x"
	want.Sources.Dirs = []string{"../other/.grapes"}

	if err := Save(dir, want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("round-trip mismatch:\n want=%+v\n got =%+v", want, got)
	}
}

func TestLoad_PartialConfigKeepsDefaults(t *testing.T) {
	dir := writeConfig(t, "[theme]\npreset = \"Nord\"\n")

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Theme.Preset != "Nord" {
		t.Errorf("preset = %q, want %q", cfg.Theme.Preset, "Nord")
	}
	// Keys were not mentioned in the file, so they must survive intact.
	if cfg.Keys.BoardOpen != Defaults().Keys.BoardOpen {
		t.Errorf("unset key was clobbered: BoardOpen = %q, want %q",
			cfg.Keys.BoardOpen, Defaults().Keys.BoardOpen)
	}
}

// A malformed file must yield clean defaults, never a half-applied config.
// go-toml populates fields as it parses, so a document that is valid up to a
// syntax error would otherwise leave the leading sections applied and silently
// drop the rest.
func TestLoad_MalformedYieldsCleanDefaults(t *testing.T) {
	dir := writeConfig(t, "[theme]\npreset = \"Nord\"\n\n[view]\ndefault_screen = \"list\"\n\n[keys]\nboard_open = = broken\n")

	cfg, err := Load(dir)
	if err == nil {
		t.Fatal("a malformed config.toml should report a parse error")
	}
	if !reflect.DeepEqual(cfg, Defaults()) {
		t.Errorf("malformed config should yield exactly Defaults(), got preset=%q default_screen=%q",
			cfg.Theme.Preset, cfg.View.DefaultScreen)
	}
}

func TestLoad_UnreadableFileReportsError(t *testing.T) {
	dir := t.TempDir()
	// A directory where config.toml is expected makes the read fail with
	// something other than "not exist".
	if err := os.Mkdir(filepath.Join(dir, "config.toml"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cfg, err := Load(dir)
	if err == nil {
		t.Fatal("an unreadable config.toml should report an error")
	}
	if !reflect.DeepEqual(cfg, Defaults()) {
		t.Error("an unreadable config.toml should still yield Defaults()")
	}
}

func TestHideEmpty(t *testing.T) {
	if !(ViewConfig{}).HideEmpty() {
		t.Error("HideEmpty should default to true when unset")
	}
	no, yes := false, true
	if (ViewConfig{HideEmptyColumns: &no}).HideEmpty() {
		t.Error("HideEmpty should honour an explicit false")
	}
	if !(ViewConfig{HideEmptyColumns: &yes}).HideEmpty() {
		t.Error("HideEmpty should honour an explicit true")
	}
}
