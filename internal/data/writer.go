package data

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Mibokess/grapes/internal/fsutil"
	toml "github.com/pelletier/go-toml/v2"
)

// metaFileMode is the default permission used for newly created issue files.
const metaFileMode = 0o644

// serializedCommentsMarker identifies the generated comments section in an
// editor document. It must not be confused with ordinary Markdown content.
const serializedCommentsMarker = "<!-- grapes:comments -->\n## Comments\n"

func filePerm(path string) os.FileMode {
	if info, err := os.Stat(path); err == nil && info.Mode().IsRegular() {
		return info.Mode().Perm()
	}
	return metaFileMode
}

var issueWriteLocks sync.Map // map[string]*sync.Mutex

func issueWriteLock(issuesDir string, issueID int) *sync.Mutex {
	key, err := filepath.Abs(filepath.Join(issuesDir, strconv.Itoa(issueID)))
	if err != nil {
		key = filepath.Join(issuesDir, strconv.Itoa(issueID))
	}
	lock, _ := issueWriteLocks.LoadOrStore(key, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

func withIssueWriteLock(issuesDir string, issueID int, fn func() error) error {
	lock := issueWriteLock(issuesDir, issueID)
	lock.Lock()
	defer lock.Unlock()

	issueDir := filepath.Join(issuesDir, strconv.Itoa(issueID))
	fileLock, err := fsutil.OpenFileLock(filepath.Join(issueDir, ".lock"), 0o644)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("opening issue lock: %w", err)
		}
		// Preserve the normal missing-issue error from fn. StampTimestamps is
		// called after the CLI creates the directory, while update operations
		// should not create a missing issue as a side effect.
		return fn()
	}
	defer fileLock.Close()
	if err := fileLock.Lock(); err != nil {
		return fmt.Errorf("acquiring issue lock: %w", err)
	}
	defer fileLock.Unlock()
	return fn()
}

// metaPath returns the meta.toml path for an issue.
func metaPath(issuesDir string, issueID int) string {
	return filepath.Join(issuesDir, strconv.Itoa(issueID), "meta.toml")
}

// readMeta loads and parses an issue's meta.toml.
func readMeta(path string) (meta, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return meta{}, fmt.Errorf("read meta.toml: %w", err)
	}
	var m meta
	if err := toml.Unmarshal(raw, &m); err != nil {
		return meta{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return m, nil
}

// writeMeta marshals and atomically writes an issue's meta.toml.
func writeMeta(path string, m meta) error {
	out, err := toml.Marshal(&m)
	if err != nil {
		return fmt.Errorf("marshal meta.toml: %w", err)
	}
	return fsutil.WriteFile(path, out, filePerm(path))
}

// stamp sets `updated` to now, and `created` too when it is missing.
func stamp(m *meta, now time.Time) {
	if m.Created.IsZero() {
		m.Created = now
	}
	m.Updated = now
}

// UpdateField sets one meta.toml field and stamps the timestamps.
//
// The field name and the new value are both checked: writing a field grapes
// does not know, or a status/priority outside the accepted set, is an error
// rather than a write that appears to succeed and changes nothing.
func UpdateField(issuesDir string, issueID int, field, newValue string) error {
	return withIssueWriteLock(issuesDir, issueID, func() error {
		path := metaPath(issuesDir, issueID)
		m, err := readMeta(path)
		if err != nil {
			return err
		}

		switch field {
		case "title":
			if strings.TrimSpace(newValue) == "" {
				return fmt.Errorf("title must not be empty")
			}
			m.Title = newValue
		case "status":
			if !validStatuses[newValue] {
				return fmt.Errorf("%q is not a valid status", newValue)
			}
			m.Status = newValue
		case "priority":
			if !validPriorities[newValue] {
				return fmt.Errorf("%q is not a valid priority", newValue)
			}
			m.Priority = newValue
		default:
			return fmt.Errorf("unknown meta.toml field %q", field)
		}

		stamp(&m, time.Now().UTC().Truncate(time.Minute))
		return writeMeta(path, m)
	})
}

// UpdateLabels replaces the labels array in meta.toml.
func UpdateLabels(issuesDir string, issueID int, labels []string) error {
	return withIssueWriteLock(issuesDir, issueID, func() error {
		path := metaPath(issuesDir, issueID)
		m, err := readMeta(path)
		if err != nil {
			return err
		}
		m.Labels = append([]string(nil), labels...)
		stamp(&m, time.Now().UTC().Truncate(time.Minute))
		return writeMeta(path, m)
	})
}

// StampTimestamps reads meta.toml and sets `updated` to now. If `created` is
// zero/missing, it sets `created` to now as well. This is the canonical way to
// maintain timestamps — agents call `grapes issue <id>` which invokes this.
//
// A missing meta.toml is not an error: `grapes issue` stamps a directory it
// has just created, and the file is written from scratch.
func StampTimestamps(issuesDir string, issueID int) error {
	return withIssueWriteLock(issuesDir, issueID, func() error {
		path := metaPath(issuesDir, issueID)
		var m meta
		if _, err := os.Stat(path); err == nil {
			if m, err = readMeta(path); err != nil {
				return err
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("read meta.toml: %w", err)
		}

		stamp(&m, time.Now().UTC().Truncate(time.Minute))
		return writeMeta(path, m)
	})
}

// AppendComment appends a comment to an issue's comments.md using the standard
// grapes format:
//
//	### YYYY-MM-DDTHH:MM
//	comment body
//
// A blank line is prepended if the file already has content. Timestamps are
// UTC, matching the `created`/`updated` fields in meta.toml.
func AppendComment(issuesDir string, issueID int, body string) error {
	return withIssueWriteLock(issuesDir, issueID, func() error {
		issueDir := filepath.Join(issuesDir, strconv.Itoa(issueID))
		metaFile := filepath.Join(issueDir, "meta.toml")
		m, err := readMeta(metaFile)
		if err != nil {
			return fmt.Errorf("read meta.toml before appending comment: %w", err)
		}
		path := filepath.Join(issueDir, "comments.md")
		existing, err := os.ReadFile(path)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("read comments: %w", err)
		}

		now := time.Now().UTC().Truncate(time.Minute)
		var sb strings.Builder
		if len(existing) > 0 {
			sb.Write(existing)
			if existing[len(existing)-1] != '\n' {
				sb.WriteByte('\n')
			}
			sb.WriteByte('\n')
		}
		sb.WriteString("### " + now.Format("2006-01-02T15:04"))
		sb.WriteByte('\n')
		sb.WriteString(escapeCommentHeaders(body))
		sb.WriteByte('\n')

		stamp(&m, now)
		metaRaw, err := toml.Marshal(&m)
		if err != nil {
			return fmt.Errorf("marshal meta.toml: %w", err)
		}
		if err := fsutil.WriteFiles([]fsutil.AtomicFile{
			{Path: metaFile, Data: metaRaw, Perm: filePerm(metaFile)},
			{Path: path, Data: []byte(sb.String()), Perm: filePerm(path)},
		}); err != nil {
			return fmt.Errorf("appending comment: %w", err)
		}
		return nil
	})
}

// SerializeIssue renders a complete issue as an editable text document with
// TOML frontmatter, description body, and comments section.
func SerializeIssue(issue Issue) string {
	var sb strings.Builder

	// TOML frontmatter
	sb.WriteString("+++\n")
	sb.WriteString(fmt.Sprintf("title = %q\n", issue.Title))
	sb.WriteString(fmt.Sprintf("status = %q\n", string(issue.Status)))
	sb.WriteString(fmt.Sprintf("priority = %q\n", string(issue.Priority)))
	if len(issue.Labels) > 0 {
		quoted := make([]string, len(issue.Labels))
		for i, l := range issue.Labels {
			quoted[i] = fmt.Sprintf("%q", l)
		}
		sb.WriteString(fmt.Sprintf("labels = [%s]\n", strings.Join(quoted, ", ")))
	} else {
		sb.WriteString("labels = []\n")
	}
	if issue.Parent != nil {
		sb.WriteString(fmt.Sprintf("parent = %d\n", *issue.Parent))
	}
	if len(issue.BlockedBy) > 0 {
		parts := make([]string, len(issue.BlockedBy))
		for i, id := range issue.BlockedBy {
			parts[i] = strconv.Itoa(id)
		}
		sb.WriteString(fmt.Sprintf("blocked_by = [%s]\n", strings.Join(parts, ", ")))
	}
	sb.WriteString("+++\n")

	// Description
	if issue.Content != "" {
		sb.WriteString(issue.Content)
		if !strings.HasSuffix(issue.Content, "\n") {
			sb.WriteByte('\n')
		}
	}

	// Comments section. A comment with no date is text that preceded the first
	// "### " header in comments.md; it is written back without a header so the
	// edit round-trip preserves it.
	// The sentinel makes the generated comments section distinguishable from a
	// user-authored Markdown heading in the description.
	if len(issue.Comments) > 0 {
		sb.WriteString("\n" + serializedCommentsMarker)
		for _, c := range issue.Comments {
			if c.Date == "" {
				sb.WriteString("\n")
			} else {
				sb.WriteString(fmt.Sprintf("\n### %s\n", c.Date))
			}
			escapedBody := escapeCommentHeaders(c.Body)
			sb.WriteString(escapedBody)
			// Keep one structural newline after every body; ParseComments
			// removes that delimiter and preserves body whitespace.
			sb.WriteByte('\n')
		}
	}

	return sb.String()
}

func escapeCommentHeaders(body string) string {
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		slashes := 0
		for slashes < len(line) && line[slashes] == '\\' {
			slashes++
		}
		if commentHeader.MatchString(line[slashes:]) {
			// Prefix one escape marker. ParseComments removes exactly one,
			// preserving any literal backslashes already in the body.
			lines[i] = `\` + line
		}
	}
	return strings.Join(lines, "\n")
}

// EditValidationError is returned when the edited issue fails validation.
// The caller can use this to re-open the editor instead of discarding changes.
type EditValidationError struct {
	Message string
}

func (e *EditValidationError) Error() string {
	return "validation failed: " + e.Message
}

// editMeta is the frontmatter structure parsed back from the edited document.
type editMeta struct {
	Title     string   `toml:"title"`
	Status    string   `toml:"status"`
	Priority  string   `toml:"priority"`
	Labels    []string `toml:"labels"`
	Parent    *int     `toml:"parent,omitempty"`
	BlockedBy []int    `toml:"blocked_by,omitempty"`
}

// SaveIssueFromText parses an edited issue document and writes the changes
// back to meta.toml, content.md, and comments.md.
func SaveIssueFromText(issuesDir string, issueID int, text string) error {
	return withIssueWriteLock(issuesDir, issueID, func() error {
		frontmatter, body, err := splitEditorDocument(text)
		if err != nil {
			return err
		}
		var em editMeta
		if err := toml.Unmarshal([]byte(frontmatter), &em); err != nil {
			return fmt.Errorf("parsing frontmatter: %w", err)
		}
		if verrs := ValidateMeta(issueID, em.Title, em.Status, em.Priority); len(verrs) > 0 {
			return editValidationError(verrs)
		}

		content, commentsRaw := splitEditorBody(body)
		issueDir := filepath.Join(issuesDir, strconv.Itoa(issueID))
		metaFile := filepath.Join(issueDir, "meta.toml")
		contentPath := filepath.Join(issueDir, "content.md")
		commentsPath := filepath.Join(issueDir, "comments.md")
		existing, err := readMeta(metaFile)
		if err != nil {
			return err
		}
		relationships := readRelationshipGraph(issuesDir)
		relationships[issueID] = RelationshipMeta{Parent: em.Parent, BlockedBy: em.BlockedBy}
		if verrs := validateRelationshipsForEdit(relationships); len(verrs) > 0 {
			return editValidationError(verrs)
		}

		now := time.Now().UTC().Truncate(time.Minute)
		newMeta := meta{
			Title: em.Title, Status: em.Status, Priority: em.Priority,
			Labels: append([]string(nil), em.Labels...),
			Parent: em.Parent, BlockedBy: append([]int(nil), em.BlockedBy...),
			Comments: existing.Comments, Created: existing.Created,
		}
		stamp(&newMeta, now)
		metaRaw, err := toml.Marshal(&newMeta)
		if err != nil {
			return fmt.Errorf("marshal meta.toml: %w", err)
		}
		// Commit all three files from fully prepared temporary files. Keeping an
		// empty comments file is intentional: it avoids a non-atomic remove.
		if err := fsutil.WriteFiles([]fsutil.AtomicFile{
			{Path: metaFile, Data: metaRaw, Perm: filePerm(metaFile)},
			{Path: contentPath, Data: []byte(content), Perm: filePerm(contentPath)},
			{Path: commentsPath, Data: []byte(commentsRaw), Perm: filePerm(commentsPath)},
		}); err != nil {
			return fmt.Errorf("saving issue files: %w", err)
		}
		return nil
	})
}

func editValidationError(verrs []ValidationError) error {
	msgs := make([]string, len(verrs))
	for i, v := range verrs {
		msgs[i] = v.Field + ": " + v.Message
	}
	return &EditValidationError{Message: strings.Join(msgs, "; ")}
}

func splitEditorDocument(text string) (frontmatter, body string, err error) {
	text = strings.TrimPrefix(text, "\ufeff")
	text = strings.ReplaceAll(text, "\r\n", "\n")
	const opening = "+++\n"
	openAt := 0
	if !strings.HasPrefix(text, opening) {
		openAt = strings.Index(text, opening)
		if openAt < 0 || strings.TrimSpace(text[:openAt]) != "" {
			return "", "", fmt.Errorf("invalid format: missing TOML frontmatter delimiters")
		}
	}
	rest := text[openAt+len(opening):]
	closeAt := strings.Index(rest, "\n+++\n")
	if closeAt < 0 {
		return "", "", fmt.Errorf("invalid format: missing TOML frontmatter delimiters")
	}
	return rest[:closeAt], rest[closeAt+len("\n+++\n"):], nil
}

func splitEditorBody(body string) (content, comments string) {
	const generatedMarker = "\n\n" + serializedCommentsMarker
	if start := strings.Index(body, generatedMarker); start >= 0 {
		content = body[:start+1]
		comments = body[start+len(generatedMarker):]
		if strings.HasPrefix(comments, "\n") {
			comments = comments[1:]
		}
		return content, comments
	}
	const leadingGeneratedMarker = "\n" + serializedCommentsMarker
	if strings.HasPrefix(body, leadingGeneratedMarker) {
		comments = body[len(leadingGeneratedMarker):]
		if strings.HasPrefix(comments, "\n") {
			comments = comments[1:]
		}
		return "", comments
	}
	return body, ""
}
