package data

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	toml "github.com/pelletier/go-toml/v2"
)

// UpdateField runs sed to replace a field value in meta.toml and updates the
// "updated" field. This mirrors what an agent does when editing issue metadata.
//
//	sed -i "s/^field = .*/field = 'newValue'/" .grapes/<id>/meta.toml
//	sed -i "s/^updated = .*/updated = '2026-03-02T15:04'/" .grapes/<id>/meta.toml
func UpdateField(issuesDir string, issueID int, field, newValue string) error {
	path := filepath.Join(issuesDir, strconv.Itoa(issueID), "meta.toml")

	// Update the target field
	fieldPattern := fmt.Sprintf(`s/^%s = .*/%s = '%s'/`, field, field, newValue)
	if err := exec.Command("sed", "-i", fieldPattern, path).Run(); err != nil {
		return fmt.Errorf("sed %s: %w", field, err)
	}

	// Update the "updated" datetime
	now := time.Now().Format("2006-01-02T15:04")
	datePattern := fmt.Sprintf(`s/^updated = .*/updated = '%s'/`, now)
	if err := exec.Command("sed", "-i", datePattern, path).Run(); err != nil {
		return fmt.Errorf("sed updated: %w", err)
	}

	return nil
}

// AppendComment appends a comment to an issue's comments.md using the standard
// grapes format:
//
//	### YYYY-MM-DD
//	comment body
//
// A blank line is prepended if the file already has content.
func AppendComment(issuesDir string, issueID int, body string) error {
	path := filepath.Join(issuesDir, strconv.Itoa(issueID), "comments.md")

	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read comments: %w", err)
	}

	now := time.Now().Format("2006-01-02T15:04")
	header := fmt.Sprintf("### %s", now)

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
	sb.WriteString(header)
	sb.WriteByte('\n')
	sb.WriteString(body)
	sb.WriteByte('\n')

	return os.WriteFile(path, []byte(sb.String()), 0644)
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

	// Comments section
	if len(issue.Comments) > 0 {
		sb.WriteString("\n## Comments\n")
		for _, c := range issue.Comments {
			sb.WriteString(fmt.Sprintf("\n### %s\n", c.Date))
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

	// Write meta.toml
	issueDir := filepath.Join(issuesDir, strconv.Itoa(issueID))
	now := time.Now().Format("2006-01-02T15:04")

	// Read existing meta to preserve created date
	existingMeta, err := os.ReadFile(filepath.Join(issueDir, "meta.toml"))
	if err != nil {
		return fmt.Errorf("reading existing meta: %w", err)
	}
	var existing meta
	if err := toml.Unmarshal(existingMeta, &existing); err != nil {
		return fmt.Errorf("parsing existing meta: %w", err)
	}

	newMeta := meta{
		Title:     em.Title,
		Status:    em.Status,
		Priority:  em.Priority,
		Labels:    em.Labels,
		Parent:    em.Parent,
		BlockedBy: em.BlockedBy,
		Created:   existing.Created,
		Updated:   now,
	}
	metaBytes, err := toml.Marshal(&newMeta)
	if err != nil {
		return fmt.Errorf("marshaling meta: %w", err)
	}
	if err := os.WriteFile(filepath.Join(issueDir, "meta.toml"), metaBytes, 0644); err != nil {
		return fmt.Errorf("writing meta.toml: %w", err)
	}

	// Write content.md
	if content != "" {
		content += "\n"
	}
	if err := os.WriteFile(filepath.Join(issueDir, "content.md"), []byte(content), 0644); err != nil {
		return fmt.Errorf("writing content.md: %w", err)
	}

	// Write comments.md only when there are comments; remove the file otherwise.
	commentsPath := filepath.Join(issueDir, "comments.md")
	if commentsRaw != "" {
		if err := os.WriteFile(commentsPath, []byte(commentsRaw+"\n"), 0644); err != nil {
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
