package data

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Mibokess/grapes/internal/fsutil"
	toml "github.com/pelletier/go-toml/v2"
)

// metaFileMode is the permission used for every file grapes writes.
const metaFileMode = 0o644

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
	return fsutil.WriteFile(path, out, metaFileMode)
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
}

// UpdateLabels replaces the labels array in meta.toml.
func UpdateLabels(issuesDir string, issueID int, labels []string) error {
	path := metaPath(issuesDir, issueID)
	m, err := readMeta(path)
	if err != nil {
		return err
	}
	m.Labels = labels
	stamp(&m, time.Now().UTC().Truncate(time.Minute))
	return writeMeta(path, m)
}

// StampTimestamps reads meta.toml and sets `updated` to now. If `created` is
// zero/missing, it sets `created` to now as well. This is the canonical way to
// maintain timestamps — agents call `grapes issue <id>` which invokes this.
//
// A missing meta.toml is not an error: `grapes issue` stamps a directory it
// has just created, and the file is written from scratch.
func StampTimestamps(issuesDir string, issueID int) error {
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
	path := filepath.Join(issuesDir, strconv.Itoa(issueID), "comments.md")

	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read comments: %w", err)
	}

	now := time.Now().UTC().Format("2006-01-02T15:04")

	var sb strings.Builder
	if len(existing) > 0 {
		sb.Write(existing)
		// Ensure existing content ends with newline
		if existing[len(existing)-1] != '\n' {
			sb.WriteByte('\n')
		}
		// Blank line separator before new comment
		sb.WriteByte('\n')
	}
	sb.WriteString("### " + now)
	sb.WriteByte('\n')
	sb.WriteString(body)
	sb.WriteByte('\n')

	return fsutil.WriteFile(path, []byte(sb.String()), metaFileMode)
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
	if len(issue.Comments) > 0 {
		sb.WriteString("\n## Comments\n")
		for _, c := range issue.Comments {
			if c.Date == "" {
				sb.WriteString("\n")
			} else {
				sb.WriteString(fmt.Sprintf("\n### %s\n", c.Date))
			}
			sb.WriteString(c.Body)
			if !strings.HasSuffix(c.Body, "\n") {
				sb.WriteByte('\n')
			}
		}
	}

	return sb.String()
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
	// Split frontmatter from body
	parts := strings.SplitN(text, "+++\n", 3)
	if len(parts) < 3 {
		return fmt.Errorf("invalid format: missing TOML frontmatter delimiters")
	}
	if strings.TrimSpace(parts[0]) != "" {
		return fmt.Errorf("invalid format: text before the opening +++ delimiter")
	}
	frontmatter := parts[1]
	body := parts[2]

	// Parse frontmatter
	var em editMeta
	if err := toml.Unmarshal([]byte(frontmatter), &em); err != nil {
		return fmt.Errorf("parsing frontmatter: %w", err)
	}

	// Validate before writing anything
	if verrs := ValidateMeta(issueID, em.Title, em.Status, em.Priority); len(verrs) > 0 {
		msgs := make([]string, len(verrs))
		for i, v := range verrs {
			msgs[i] = v.Field + ": " + v.Message
		}
		return &EditValidationError{Message: strings.Join(msgs, "; ")}
	}

	// Split body into content and comments at "## Comments" marker
	content := body
	var commentsRaw string
	if idx := strings.Index(body, "\n## Comments\n"); idx >= 0 {
		content = body[:idx]
		commentsRaw = body[idx+len("\n## Comments\n"):]
	} else if strings.HasPrefix(body, "## Comments\n") {
		content = ""
		commentsRaw = body[len("## Comments\n"):]
	}
	content = strings.TrimSpace(content)
	commentsRaw = strings.TrimSpace(commentsRaw)

	issueDir := filepath.Join(issuesDir, strconv.Itoa(issueID))

	// Read existing meta to preserve fields the editable document does not
	// carry: the created date, and any legacy comments stored in meta.toml.
	existing, err := readMeta(filepath.Join(issueDir, "meta.toml"))
	if err != nil {
		return err
	}

	newMeta := meta{
		Title:     em.Title,
		Status:    em.Status,
		Priority:  em.Priority,
		Labels:    em.Labels,
		Parent:    em.Parent,
		BlockedBy: em.BlockedBy,
		Comments:  existing.Comments,
		Created:   existing.Created,
	}
	stamp(&newMeta, time.Now().UTC().Truncate(time.Minute))
	if err := writeMeta(filepath.Join(issueDir, "meta.toml"), newMeta); err != nil {
		return err
	}

	// Write content.md
	if content != "" {
		content += "\n"
	}
	if err := fsutil.WriteFile(filepath.Join(issueDir, "content.md"), []byte(content), metaFileMode); err != nil {
		return fmt.Errorf("writing content.md: %w", err)
	}

	// Write comments.md only when there are comments; remove the file otherwise.
	commentsPath := filepath.Join(issueDir, "comments.md")
	if commentsRaw != "" {
		if err := fsutil.WriteFile(commentsPath, []byte(commentsRaw+"\n"), metaFileMode); err != nil {
			return fmt.Errorf("writing comments.md: %w", err)
		}
	} else {
		// Remove stale comments.md; ignore error if it doesn't exist.
		if err := os.Remove(commentsPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("removing comments.md: %w", err)
		}
	}

	return nil
}
