package data

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

// RelationshipMeta contains the persisted directed relationship fields used
// when validating references and cycles.
type RelationshipMeta struct {
	Parent    *int
	BlockedBy []int
}

// ValidateRelationships checks references and rejects self-links and cycles.
// The map must contain every issue visible to the caller; missing IDs are
// reported as dangling references.
func ValidateRelationships(relationships map[int]RelationshipMeta) []ValidationError {
	return validateRelationships(relationships, true)
}

func validateRelationshipsForEdit(relationships map[int]RelationshipMeta) []ValidationError {
	// An editor only receives one source directory. A reference absent there
	// may live in another configured store, so existence checks belong to the
	// full workspace validator, not this per-source save path.
	return validateRelationships(relationships, false)
}

func validateRelationships(relationships map[int]RelationshipMeta, checkReferences bool) []ValidationError {
	ids := make(map[int]bool, len(relationships))
	for id := range relationships {
		ids[id] = true
	}
	var errs []ValidationError
	for id, rel := range relationships {
		if rel.Parent != nil {
			if *rel.Parent == id {
				errs = append(errs, ValidationError{IssueID: id, Field: "parent", Message: "cannot be its own parent"})
			} else if checkReferences && !ids[*rel.Parent] {
				errs = append(errs, ValidationError{IssueID: id, Field: "parent", Message: fmt.Sprintf("references #%d which does not exist", *rel.Parent)})
			}
		}
		for _, blockerID := range rel.BlockedBy {
			if blockerID == id {
				errs = append(errs, ValidationError{IssueID: id, Field: "blocked_by", Message: "cannot be blocked by itself"})
			} else if checkReferences && !ids[blockerID] {
				errs = append(errs, ValidationError{IssueID: id, Field: "blocked_by", Message: fmt.Sprintf("references #%d which does not exist", blockerID)})
			}
		}
	}
	errs = append(errs, validateRelationshipCycles(relationships, false)...)
	errs = append(errs, validateRelationshipCycles(relationships, true)...)
	sort.SliceStable(errs, func(i, j int) bool {
		if errs[i].IssueID != errs[j].IssueID {
			return errs[i].IssueID < errs[j].IssueID
		}
		return errs[i].Field < errs[j].Field
	})
	return errs
}

func validateRelationshipCycles(relationships map[int]RelationshipMeta, blockers bool) []ValidationError {
	edges := make(map[int][]int, len(relationships))
	for id, rel := range relationships {
		if blockers {
			edges[id] = rel.BlockedBy
		} else if rel.Parent != nil {
			edges[id] = []int{*rel.Parent}
		}
	}
	state := make(map[int]uint8, len(edges))
	cyclic := make(map[int]bool)
	var visit func(int, []int)
	visit = func(id int, stack []int) {
		if state[id] == 1 {
			cycleStart := -1
			for i := len(stack) - 2; i >= 0; i-- {
				if stack[i] == id {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				for _, cycleID := range stack[cycleStart : len(stack)-1] {
					cyclic[cycleID] = true
				}
			}
			return
		}
		if state[id] == 2 {
			return
		}
		state[id] = 1
		for _, next := range edges[id] {
			if next == id {
				continue
			}
			if _, ok := relationships[next]; ok {
				visit(next, append(stack, next))
			}
		}
		state[id] = 2
	}
	for id := range relationships {
		if state[id] == 0 {
			visit(id, []int{id})
		}
	}
	field := "parent"
	kind := "parent"
	if blockers {
		field, kind = "blocked_by", "blocked_by"
	}
	var errs []ValidationError
	for id := range cyclic {
		errs = append(errs, ValidationError{IssueID: id, Field: field, Message: fmt.Sprintf("%s relationship cycle detected", kind)})
	}
	return errs
}

func readRelationshipGraph(issuesDir string) map[int]RelationshipMeta {
	entries, err := os.ReadDir(issuesDir)
	if err != nil {
		return nil
	}
	graph := make(map[int]RelationshipMeta)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		id, err := strconv.Atoi(entry.Name())
		if err != nil || id <= 0 {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(issuesDir, entry.Name(), "meta.toml"))
		if err != nil {
			continue
		}
		var m meta
		if toml.Unmarshal(raw, &m) == nil {
			graph[id] = RelationshipMeta{Parent: m.Parent, BlockedBy: append([]int(nil), m.BlockedBy...)}
		}
	}
	return graph
}

// ValidateIssue checks a single issue directory for correctness.
// It reads meta.toml and comments.md from disk and returns all problems found.
func ValidateIssue(issuesDir string, issueID int) []ValidationError {
	if issueID <= 0 {
		return []ValidationError{{IssueID: issueID, Field: "id", Message: "must be positive"}}
	}
	dir := filepath.Join(issuesDir, strconv.Itoa(issueID))
	var errs []ValidationError

	metaPath := filepath.Join(dir, "meta.toml")
	raw, err := os.ReadFile(metaPath)
	if err != nil {
		return []ValidationError{{IssueID: issueID, Field: "meta.toml", Message: fmt.Sprintf("cannot read file: %v", err)}}
	}

	var m meta
	if err := toml.Unmarshal(raw, &m); err != nil {
		return []ValidationError{{IssueID: issueID, Field: "meta.toml", Message: "invalid TOML: " + err.Error()}}
	}
	errs = append(errs, validateMeta(issueID, m)...)

	commentsPath := filepath.Join(dir, "comments.md")
	commentsRaw, present, err := readSecondaryFile(commentsPath)
	if err != nil {
		errs = append(errs, ValidationError{IssueID: issueID, Field: "comments.md", Message: fmt.Sprintf("cannot read file: %v", err)})
	} else if present {
		errs = append(errs, validateComments(issueID, string(commentsRaw))...)
	}
	return errs
}

// readSecondaryFile distinguishes an absent optional file from an unreadable
// one. Callers may treat absent content as empty while still surfacing I/O
// failures such as permissions and wrong file types.
func readSecondaryFile(path string) ([]byte, bool, error) {
	raw, err := os.ReadFile(path)
	if err == nil {
		return raw, true, nil
	}
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	return nil, false, err
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

	// Collect existing positive issue IDs for deterministic validation order.
	var (
		issueIDs []int
		errs     []ValidationError
	)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		if id <= 0 {
			errs = append(errs, ValidationError{IssueID: id, Field: "id", Message: "must be positive"})
			continue
		}
		issueIDs = append(issueIDs, id)
	}
	sort.Ints(issueIDs)

	for _, id := range issueIDs {
		errs = append(errs, ValidateIssue(issuesDir, id)...)
	}

	// Relationship validation is performed after all individual metadata has
	// been loaded, so references and cycles are checked consistently.
	errs = append(errs, ValidateRelationships(readRelationshipGraph(issuesDir))...)

	return errs
}
