package data

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
)

// ValidationError represents a single validation problem.
type ValidationError struct {
	IssueID int
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("#%d %s: %s", e.IssueID, e.Field, e.Message)
}

// validStatuses is the set of accepted status values.
var validStatuses = map[string]bool{
	string(StatusBacklog):    true,
	string(StatusTodo):       true,
	string(StatusInProgress): true,
	string(StatusDone):       true,
	string(StatusCancelled):  true,
}

// validPriorities is the set of accepted priority values.
var validPriorities = map[string]bool{
	string(PriorityUrgent): true,
	string(PriorityHigh):   true,
	string(PriorityMedium): true,
	string(PriorityLow):    true,
}

// ValidateIssue checks a single issue directory for correctness.
// It reads meta.toml and comments.md from disk and returns all problems found.
func ValidateIssue(issuesDir string, issueID int) []ValidationError {
	dir := filepath.Join(issuesDir, strconv.Itoa(issueID))
	var errs []ValidationError

	// Check meta.toml exists and is valid
	metaPath := filepath.Join(dir, "meta.toml")
	raw, err := os.ReadFile(metaPath)
	if err != nil {
		return []ValidationError{{IssueID: issueID, Field: "meta.toml", Message: "cannot read file"}}
	}

	var m meta
	if err := toml.Unmarshal(raw, &m); err != nil {
		return []ValidationError{{IssueID: issueID, Field: "meta.toml", Message: "invalid TOML: " + err.Error()}}
	}

	errs = append(errs, validateMeta(issueID, m)...)

	// Check comment headers
	commentsPath := filepath.Join(dir, "comments.md")
	commentsRaw, err := os.ReadFile(commentsPath)
	if err == nil {
		errs = append(errs, validateComments(issueID, string(commentsRaw))...)
	}

	return errs
}

// ValidateMeta checks the parsed metadata fields for correctness.
// Exported so the edit flow can validate before writing.
func ValidateMeta(issueID int, title, status, priority string) []ValidationError {
	m := meta{Title: title, Status: status, Priority: priority}
	return validateMeta(issueID, m)
}

func validateMeta(issueID int, m meta) []ValidationError {
	var errs []ValidationError

	if strings.TrimSpace(m.Title) == "" {
		errs = append(errs, ValidationError{IssueID: issueID, Field: "title", Message: "must not be empty"})
	}

	if !validStatuses[m.Status] {
		valid := make([]string, 0, len(validStatuses))
		for _, s := range AllStatuses {
			valid = append(valid, string(s))
		}
		errs = append(errs, ValidationError{
			IssueID: issueID, Field: "status",
			Message: fmt.Sprintf("%q is not valid (use: %s)", m.Status, strings.Join(valid, ", ")),
		})
	}

	if !validPriorities[m.Priority] {
		valid := make([]string, 0, len(validPriorities))
		for _, p := range AllPriorities {
			valid = append(valid, string(p))
		}
		errs = append(errs, ValidationError{
			IssueID: issueID, Field: "priority",
			Message: fmt.Sprintf("%q is not valid (use: %s)", m.Priority, strings.Join(valid, ", ")),
		})
	}

	if m.Created != "" && parseDate(m.Created).IsZero() {
		errs = append(errs, ValidationError{IssueID: issueID, Field: "created", Message: fmt.Sprintf("%q is not a valid date", m.Created)})
	}
	if m.Updated != "" && parseDate(m.Updated).IsZero() {
		errs = append(errs, ValidationError{IssueID: issueID, Field: "updated", Message: fmt.Sprintf("%q is not a valid date", m.Updated)})
	}

	return errs
}

func validateComments(issueID int, raw string) []ValidationError {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	var errs []ValidationError
	lineNum := 0
	for _, line := range strings.Split(raw, "\n") {
		lineNum++
		if strings.HasPrefix(line, "### ") {
			if m := commentHeader.FindStringSubmatch(line); m == nil {
				errs = append(errs, ValidationError{
					IssueID: issueID, Field: "comments.md",
					Message: fmt.Sprintf("line %d: invalid comment header %q (expected ### YYYY-MM-DD or ### YYYY-MM-DDTHH:MM)", lineNum, line),
				})
			}
		}
	}
	return errs
}

// ValidateAll checks every issue in the issues directory.
// It also verifies parent references point to existing issues.
func ValidateAll(issuesDir string) []ValidationError {
	entries, err := os.ReadDir(issuesDir)
	if err != nil {
		return []ValidationError{{Field: "directory", Message: "cannot read " + issuesDir}}
	}

	// Collect existing IDs
	existingIDs := make(map[int]bool)
	var issueIDs []int
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		existingIDs[id] = true
		issueIDs = append(issueIDs, id)
	}

	var errs []ValidationError
	for _, id := range issueIDs {
		errs = append(errs, ValidateIssue(issuesDir, id)...)
	}

	// Check parent and blocked_by references
	for _, id := range issueIDs {
		metaPath := filepath.Join(issuesDir, strconv.Itoa(id), "meta.toml")
		raw, err := os.ReadFile(metaPath)
		if err != nil {
			continue
		}
		var m meta
		if err := toml.Unmarshal(raw, &m); err != nil {
			continue
		}
		if m.Parent != nil && !existingIDs[*m.Parent] {
			errs = append(errs, ValidationError{
				IssueID: id, Field: "parent",
				Message: fmt.Sprintf("references #%d which does not exist", *m.Parent),
			})
		}
		for _, blockerID := range m.BlockedBy {
			if blockerID == id {
				errs = append(errs, ValidationError{
					IssueID: id, Field: "blocked_by",
					Message: "cannot be blocked by itself",
				})
			} else if !existingIDs[blockerID] {
				errs = append(errs, ValidationError{
					IssueID: id, Field: "blocked_by",
					Message: fmt.Sprintf("references #%d which does not exist", blockerID),
				})
			}
		}
	}

	return errs
}
