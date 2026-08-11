package data

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// newIssueDir creates .grapes/<id>/meta.toml with the given contents and
// returns the .grapes/ path.
func newIssueDir(t *testing.T, id int, metaTOML string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), ".grapes")
	issueDir := filepath.Join(dir, strconv.Itoa(id))
	if err := os.MkdirAll(issueDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if metaTOML != "" {
		if err := os.WriteFile(filepath.Join(issueDir, "meta.toml"), []byte(metaTOML), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

const validMeta = `title = "example"
status = "todo"
priority = "medium"
labels = ["bug"]
created = 2026-01-01T00:00:00Z
updated = 2026-01-01T00:00:00Z
`

func TestUpdateField_ChangesStatus(t *testing.T) {
	dir := newIssueDir(t, 1, validMeta)

	if err := UpdateField(dir, 1, "status", "in_progress"); err != nil {
		t.Fatalf("UpdateField: %v", err)
	}

	iss, err := loadIssueMeta(dir, 1)
	if err != nil {
		t.Fatalf("loadIssueMeta: %v", err)
	}
	if iss.Status != StatusInProgress {
		t.Errorf("status = %q, want in_progress", iss.Status)
	}
	if iss.Title != "example" || len(iss.Labels) != 1 {
		t.Errorf("unrelated fields lost: %+v", iss)
	}
	if iss.Created.IsZero() {
		t.Error("created date was not preserved")
	}
	if !iss.Updated.After(iss.Created) {
		t.Error("updated should have been stamped to now")
	}
}

// A meta.toml missing the field being written used to be a silent no-op: sed
// matched nothing and the write reported success.
func TestUpdateField_MissingFieldStillWrites(t *testing.T) {
	dir := newIssueDir(t, 1, "title = \"no status here\"\npriority = \"low\"\n")

	if err := UpdateField(dir, 1, "status", "done"); err != nil {
		t.Fatalf("UpdateField: %v", err)
	}

	iss, err := loadIssueMeta(dir, 1)
	if err != nil {
		t.Fatalf("loadIssueMeta: %v", err)
	}
	if iss.Status != StatusDone {
		t.Errorf("status = %q, want done", iss.Status)
	}
}

func TestUpdateField_RejectsBadInput(t *testing.T) {
	dir := newIssueDir(t, 1, validMeta)

	tests := []struct {
		name, field, value string
	}{
		{"unknown field", "colour", "red"},
		{"invalid status", "status", "in-progress"},
		{"invalid priority", "priority", "critical"},
		{"empty title", "title", "  "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := UpdateField(dir, 1, tt.field, tt.value); err == nil {
				t.Fatal("expected an error, got nil")
			}
			iss, err := loadIssueMeta(dir, 1)
			if err != nil {
				t.Fatalf("loadIssueMeta: %v", err)
			}
			if iss.Status != StatusTodo || iss.Priority != PriorityMedium || iss.Title != "example" {
				t.Errorf("rejected write still changed the issue: %+v", iss)
			}
		})
	}
}

func TestUpdateField_MissingIssueIsAnError(t *testing.T) {
	dir := newIssueDir(t, 1, validMeta)
	if err := UpdateField(dir, 99, "status", "done"); err == nil {
		t.Error("writing to a nonexistent issue should fail")
	}
}

func TestUpdateLabels_ReplacesAndStamps(t *testing.T) {
	dir := newIssueDir(t, 1, validMeta)

	if err := UpdateLabels(dir, 1, []string{"docs", "ui"}); err != nil {
		t.Fatalf("UpdateLabels: %v", err)
	}
	iss, err := loadIssueMeta(dir, 1)
	if err != nil {
		t.Fatalf("loadIssueMeta: %v", err)
	}
	if strings.Join(iss.Labels, ",") != "docs,ui" {
		t.Errorf("labels = %v, want [docs ui]", iss.Labels)
	}
	if iss.Status != StatusTodo {
		t.Errorf("status changed to %q", iss.Status)
	}
}

func TestStampTimestamps_CreatesMissingMeta(t *testing.T) {
	dir := newIssueDir(t, 1, "")

	if err := StampTimestamps(dir, 1); err != nil {
		t.Fatalf("StampTimestamps: %v", err)
	}
	m, err := readMeta(metaPath(dir, 1))
	if err != nil {
		t.Fatalf("readMeta: %v", err)
	}
	if m.Created.IsZero() || m.Updated.IsZero() {
		t.Errorf("timestamps not set: %+v", m)
	}
	if !m.Created.Equal(m.Updated) {
		t.Error("a fresh issue should have created == updated")
	}
}

func TestStampTimestamps_PreservesCreated(t *testing.T) {
	dir := newIssueDir(t, 1, validMeta)
	want := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	if err := StampTimestamps(dir, 1); err != nil {
		t.Fatalf("StampTimestamps: %v", err)
	}
	iss, err := loadIssueMeta(dir, 1)
	if err != nil {
		t.Fatalf("loadIssueMeta: %v", err)
	}
	if !iss.Created.Equal(want) {
		t.Errorf("created = %v, want %v", iss.Created, want)
	}
}

func TestAppendComment_FormatAndSeparator(t *testing.T) {
	dir := newIssueDir(t, 1, validMeta)

	if err := AppendComment(dir, 1, "first"); err != nil {
		t.Fatalf("AppendComment: %v", err)
	}
	if err := AppendComment(dir, 1, "second"); err != nil {
		t.Fatalf("AppendComment: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, "1", "comments.md"))
	if err != nil {
		t.Fatal(err)
	}
	comments := ParseComments(string(raw))
	if len(comments) != 2 {
		t.Fatalf("parsed %d comments, want 2:\n%s", len(comments), raw)
	}
	if comments[0].Body != "first" || comments[1].Body != "second" {
		t.Errorf("bodies = %q, %q", comments[0].Body, comments[1].Body)
	}
	// Timestamps are UTC, like the meta.toml dates.
	wantPrefix := time.Now().UTC().Format("2006-01-02")
	if !strings.HasPrefix(comments[1].Date, wantPrefix) {
		t.Errorf("comment date %q is not today in UTC (%s)", comments[1].Date, wantPrefix)
	}
}

func TestSaveIssueFromText_RoundTrip(t *testing.T) {
	dir := newIssueDir(t, 1, validMeta)
	if err := os.WriteFile(filepath.Join(dir, "1", "content.md"), []byte("A description.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := AppendComment(dir, 1, "a comment"); err != nil {
		t.Fatal(err)
	}

	issues, problems, err := LoadAllIssues(dir)
	if err != nil || len(problems) > 0 {
		t.Fatalf("LoadAllIssues: %v %v", err, problems)
	}
	before := issues[0]

	if err := SaveIssueFromText(dir, 1, SerializeIssue(before)); err != nil {
		t.Fatalf("SaveIssueFromText: %v", err)
	}

	issues, _, err = LoadAllIssues(dir)
	if err != nil {
		t.Fatal(err)
	}
	after := issues[0]

	if after.Title != before.Title || after.Status != before.Status || after.Priority != before.Priority {
		t.Errorf("metadata changed:\n before=%+v\n after =%+v", before, after)
	}
	if strings.TrimSpace(after.Content) != strings.TrimSpace(before.Content) {
		t.Errorf("content changed: %q → %q", before.Content, after.Content)
	}
	if len(after.Comments) != len(before.Comments) {
		t.Fatalf("comment count changed: %d → %d", len(before.Comments), len(after.Comments))
	}
	if after.Comments[0] != before.Comments[0] {
		t.Errorf("comment changed: %+v → %+v", before.Comments[0], after.Comments[0])
	}
	if !after.Created.Equal(before.Created) {
		t.Errorf("created changed: %v → %v", before.Created, after.Created)
	}
}

// Text above the first "### " header is content a user typed by hand. It must
// survive being serialized into the editor and saved back.
func TestSaveIssueFromText_PreservesCommentPreamble(t *testing.T) {
	dir := newIssueDir(t, 1, validMeta)
	raw := "Notes that predate the log.\n\n### 2026-01-02T10:00\nreal comment\n"
	if err := os.WriteFile(filepath.Join(dir, "1", "comments.md"), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}

	issues, _, err := LoadAllIssues(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(issues[0].Comments); got != 2 {
		t.Fatalf("parsed %d comments, want 2 (preamble + one)", got)
	}
	if issues[0].Comments[0].Date != "" || issues[0].Comments[0].Body != "Notes that predate the log." {
		t.Fatalf("preamble not captured: %+v", issues[0].Comments[0])
	}

	if err := SaveIssueFromText(dir, 1, SerializeIssue(issues[0])); err != nil {
		t.Fatalf("SaveIssueFromText: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(dir, "1", "comments.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "Notes that predate the log.") {
		t.Errorf("preamble lost on save:\n%s", out)
	}
	if !strings.Contains(string(out), "real comment") {
		t.Errorf("comment lost on save:\n%s", out)
	}
}

func TestCommentHeadersInBodiesSurviveRoundTrip(t *testing.T) {
	dir := newIssueDir(t, 1, validMeta)
	body := "  first line  \n### 2027-03-04T05:06\n  last line  "
	if err := AppendComment(dir, 1, body); err != nil {
		t.Fatalf("AppendComment: %v", err)
	}

	issue, err := LoadIssue(dir, 1)
	if err != nil {
		t.Fatalf("LoadIssue: %v", err)
	}
	if len(issue.Comments) != 1 || issue.Comments[0].Body != body {
		t.Fatalf("comment body was split:\n%+v", issue.Comments)
	}

	if err := SaveIssueFromText(dir, 1, SerializeIssue(issue)); err != nil {
		t.Fatalf("SaveIssueFromText: %v", err)
	}
	issue, err = LoadIssue(dir, 1)
	if err != nil {
		t.Fatalf("LoadIssue after save: %v", err)
	}
	if len(issue.Comments) != 1 || issue.Comments[0].Body != body {
		t.Fatalf("comment body changed after editor round-trip:\n%+v", issue.Comments)
	}
}

func TestCommentTrailingWhitespaceAndEscapedSlashesSurviveRoundTrip(t *testing.T) {
	dir := newIssueDir(t, 1, validMeta)
	body := "\\### 2027-03-04T05:06\ntrailing\n"
	if err := AppendComment(dir, 1, body); err != nil {
		t.Fatalf("AppendComment: %v", err)
	}
	issue, err := LoadIssue(dir, 1)
	if err != nil {
		t.Fatalf("LoadIssue: %v", err)
	}
	if len(issue.Comments) != 1 || issue.Comments[0].Body != body {
		t.Fatalf("comment body changed on load: %+v", issue.Comments)
	}
	if err := SaveIssueFromText(dir, 1, SerializeIssue(issue)); err != nil {
		t.Fatalf("SaveIssueFromText: %v", err)
	}
	issue, err = LoadIssue(dir, 1)
	if err != nil {
		t.Fatalf("LoadIssue after save: %v", err)
	}
	if len(issue.Comments) != 1 || issue.Comments[0].Body != body {
		t.Fatalf("comment body changed after editor round-trip: %+v", issue.Comments)
	}
}

func TestSaveIssueFromText_AllowsCrossSourceRelationship(t *testing.T) {
	dir := newIssueDir(t, 1, validMeta)
	text := "+++\ntitle = \"example\"\nstatus = \"todo\"\npriority = \"medium\"\nparent = 99\n+++\nbody\n"
	if err := SaveIssueFromText(dir, 1, text); err != nil {
		t.Fatalf("SaveIssueFromText: %v", err)
	}
	issue, err := loadIssueMeta(dir, 1)
	if err != nil {
		t.Fatalf("loadIssueMeta: %v", err)
	}
	if issue.Parent == nil || *issue.Parent != 99 {
		t.Fatalf("parent relationship was not preserved: %+v", issue)
	}
}

func TestSaveIssueFromText_RejectsInvalid(t *testing.T) {
	dir := newIssueDir(t, 1, validMeta)

	tests := []struct {
		name string
		text string
	}{
		{"no frontmatter", "just a body\n"},
		{"text before frontmatter", "oops\n+++\ntitle = \"x\"\nstatus = \"todo\"\npriority = \"low\"\n+++\n"},
		{"bad status", "+++\ntitle = \"x\"\nstatus = \"nope\"\npriority = \"low\"\n+++\n"},
		{"empty title", "+++\ntitle = \"\"\nstatus = \"todo\"\npriority = \"low\"\n+++\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SaveIssueFromText(dir, 1, tt.text); err == nil {
				t.Fatal("expected an error, got nil")
			}
			iss, err := loadIssueMeta(dir, 1)
			if err != nil {
				t.Fatalf("loadIssueMeta: %v", err)
			}
			if iss.Title != "example" {
				t.Errorf("rejected save still wrote: %+v", iss)
			}
		})
	}
}

func TestSaveIssueFromText_AcceptsBOMAndCRLF(t *testing.T) {
	dir := newIssueDir(t, 1, validMeta)
	text := "\ufeff+++\r\ntitle = \"edited\"\r\nstatus = \"todo\"\r\npriority = \"low\"\r\n+++\r\nbody\r\n"
	if err := SaveIssueFromText(dir, 1, text); err != nil {
		t.Fatalf("SaveIssueFromText: %v", err)
	}
	issue, err := LoadIssue(dir, 1)
	if err != nil {
		t.Fatalf("LoadIssue: %v", err)
	}
	if issue.Title != "edited" || issue.Content != "body\n" {
		t.Fatalf("saved issue = title %q content %q", issue.Title, issue.Content)
	}
}

func TestSaveIssueFromText_PreservesExistingFileModes(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not expose Unix permission bits")
	}
	dir := newIssueDir(t, 1, validMeta)
	paths := []string{
		filepath.Join(dir, "1", "meta.toml"),
		filepath.Join(dir, "1", "content.md"),
		filepath.Join(dir, "1", "comments.md"),
	}
	if err := os.Chmod(paths[0], 0o600); err != nil {
		t.Fatal(err)
	}
	for _, path := range paths[1:] {
		if err := os.WriteFile(path, []byte("old\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	text := "+++\ntitle = \"edited\"\nstatus = \"todo\"\npriority = \"low\"\n+++\nbody\n"
	if err := SaveIssueFromText(dir, 1, text); err != nil {
		t.Fatalf("SaveIssueFromText: %v", err)
	}
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode().Perm(); got != 0o600 {
			t.Errorf("%s mode = %o, want 600", path, got)
		}
	}
}

// meta.toml may carry comments from an older layout. The editable document does
// not show them, so saving must not drop them.
func TestSaveIssueFromText_PreservesLegacyMetaComments(t *testing.T) {
	dir := newIssueDir(t, 1, validMeta+"\n[[comments]]\ndate = \"2026-01-01\"\nbody = \"legacy\"\n")

	text := "+++\ntitle = \"example\"\nstatus = \"todo\"\npriority = \"low\"\nlabels = []\n+++\nbody\n"
	if err := SaveIssueFromText(dir, 1, text); err != nil {
		t.Fatalf("SaveIssueFromText: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, "1", "meta.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "legacy") {
		t.Errorf("legacy meta comments dropped:\n%s", raw)
	}
}

func TestConcurrentIssueWritesPreserveFieldsAndComments(t *testing.T) {
	dir := newIssueDir(t, 1, validMeta)
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		if err := UpdateField(dir, 1, "status", "done"); err != nil {
			t.Errorf("UpdateField: %v", err)
		}
	}()
	go func() {
		defer wg.Done()
		if err := UpdateLabels(dir, 1, []string{"concurrent"}); err != nil {
			t.Errorf("UpdateLabels: %v", err)
		}
	}()
	go func() {
		defer wg.Done()
		if err := AppendComment(dir, 1, "concurrent comment"); err != nil {
			t.Errorf("AppendComment: %v", err)
		}
	}()
	wg.Wait()

	issue, err := loadIssueMeta(dir, 1)
	if err != nil {
		t.Fatal(err)
	}
	if issue.Status != StatusDone || len(issue.Labels) != 1 || issue.Labels[0] != "concurrent" {
		t.Fatalf("concurrent writes lost fields: %+v", issue)
	}
	comments, err := os.ReadFile(filepath.Join(dir, "1", "comments.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(comments), "concurrent comment") {
		t.Fatalf("comment was lost: %s", comments)
	}
	if issue.Updated.IsZero() {
		t.Fatal("comment/field writes did not stamp updated")
	}
}

func TestSaveIssueFromText_PreservesDescriptionHeadingAndWhitespace(t *testing.T) {
	dir := newIssueDir(t, 1, validMeta)
	text := "+++\ntitle = \"example\"\nstatus = \"todo\"\npriority = \"medium\"\nlabels = []\n+++\n  leading\n## Comments\ntrailing content  \n"
	if err := SaveIssueFromText(dir, 1, text); err != nil {
		t.Fatalf("SaveIssueFromText: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(dir, "1", "content.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "  leading\n## Comments\ntrailing content  \n" {
		t.Errorf("description whitespace/heading changed: %q", content)
	}
}

func TestSaveIssueFromText_PreservesDescriptionCommentLikeSection(t *testing.T) {
	dir := newIssueDir(t, 1, validMeta)
	text := "+++\ntitle = \"example\"\nstatus = \"todo\"\npriority = \"medium\"\nlabels = []\n+++\nDescription\n\n## Comments\n\n### 2025-01-01T00:00\nThis is still description text.\n"
	if err := SaveIssueFromText(dir, 1, text); err != nil {
		t.Fatalf("SaveIssueFromText: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(dir, "1", "content.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "Description\n\n## Comments\n\n### 2025-01-01T00:00\nThis is still description text.\n" {
		t.Errorf("description section was split: %q", content)
	}
	comments, err := os.ReadFile(filepath.Join(dir, "1", "comments.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(comments) != 0 {
		t.Errorf("description text was moved to comments: %q", comments)
	}
}
