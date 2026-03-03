package data

import (
	"sort"
	"strings"
	"time"
)

// Status represents an issue's workflow state.
type Status string

const (
	StatusBacklog    Status = "backlog"
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
	StatusCancelled  Status = "cancelled"
)

// AllStatuses defines the Kanban column order.
var AllStatuses = []Status{
	StatusBacklog,
	StatusTodo,
	StatusInProgress,
	StatusDone,
	StatusCancelled,
}

var statusLabels = map[Status]string{
	StatusBacklog:    "backlog",
	StatusTodo:       "todo",
	StatusInProgress: "in_progress",
	StatusDone:       "done",
	StatusCancelled:  "cancelled",
}

func (s Status) Label() string {
	if l, ok := statusLabels[s]; ok {
		return l
	}
	return string(s)
}

// Priority represents an issue's urgency level.
type Priority string

const (
	PriorityUrgent Priority = "urgent"
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

var priorityLabels = map[Priority]string{
	PriorityUrgent: "urgent",
	PriorityHigh:   "high",
	PriorityMedium: "medium",
	PriorityLow:    "low",
}

// AllPriorities defines the cycling order for the picker.
var AllPriorities = []Priority{
	PriorityLow,
	PriorityMedium,
	PriorityHigh,
	PriorityUrgent,
}

// PriorityOrder defines sort order (lower = higher priority).
var PriorityOrder = map[Priority]int{
	PriorityUrgent: 0,
	PriorityHigh:   1,
	PriorityMedium: 2,
	PriorityLow:    3,
}

func (p Priority) Label() string {
	if l, ok := priorityLabels[p]; ok {
		return l
	}
	return string(p)
}

// SortMode controls issue ordering in the TUI.
type SortMode int

const (
	SortByPriority SortMode = iota // Urgent → Low, tie-break by ID
	SortByUpdated                  // Most recently updated first, tie-break by ID
	SortByCreated                  // Most recently created first, tie-break by ID
	SortByID                       // Lowest ID first
	SortByTitle                    // Alphabetical by title
	SortByStatus                   // Status column order (backlog → cancelled)
	sortModeCount
)

// StatusOrder defines sort order for statuses (lower = earlier in workflow).
var StatusOrder = map[Status]int{
	StatusBacklog:    0,
	StatusTodo:       1,
	StatusInProgress: 2,
	StatusDone:       3,
	StatusCancelled:  4,
}

var sortModeLabels = map[SortMode]string{
	SortByPriority: "priority",
	SortByUpdated:  "updated",
	SortByCreated:  "created",
	SortByID:       "id",
	SortByTitle:    "title",
	SortByStatus:   "status",
}

func (s SortMode) Label() string {
	if l, ok := sortModeLabels[s]; ok {
		return l
	}
	return "unknown"
}

func (s SortMode) Next() SortMode {
	return (s + 1) % sortModeCount
}

// SortIssues sorts issues in place according to the given mode.
// When asc is true the natural order is reversed (e.g. oldest first, lowest priority first).
func SortIssues(issues []Issue, mode SortMode, asc bool) {
	sort.SliceStable(issues, func(i, j int) bool {
		if asc {
			i, j = j, i // flip comparison
		}
		switch mode {
		case SortByPriority:
			pi, pj := PriorityOrder[issues[i].Priority], PriorityOrder[issues[j].Priority]
			if pi != pj {
				return pi < pj
			}
			return issues[i].ID < issues[j].ID
		case SortByUpdated:
			if !issues[i].Updated.Equal(issues[j].Updated) {
				return issues[i].Updated.After(issues[j].Updated)
			}
			return issues[i].ID < issues[j].ID
		case SortByCreated:
			if !issues[i].Created.Equal(issues[j].Created) {
				return issues[i].Created.After(issues[j].Created)
			}
			return issues[i].ID < issues[j].ID
		case SortByID:
			return issues[i].ID < issues[j].ID
		case SortByTitle:
			ti := strings.ToLower(issues[i].Title)
			tj := strings.ToLower(issues[j].Title)
			if ti != tj {
				return ti < tj
			}
			return issues[i].ID < issues[j].ID
		case SortByStatus:
			si, sj := StatusOrder[issues[i].Status], StatusOrder[issues[j].Status]
			if si != sj {
				return si < sj
			}
			return issues[i].ID < issues[j].ID
		default:
			return issues[i].ID < issues[j].ID
		}
	})
}

type Comment struct {
	Date string `toml:"date"`
	Body string `toml:"body"`
}

type Issue struct {
	ID        int
	Title     string
	Status    Status
	Priority  Priority
	Labels    []string
	Parent    *int
	Children  []int
	BlockedBy []int
	Blocks    []int
	Created   time.Time
	Updated   time.Time
	Content   string
	Comments  []Comment
	SourceDir string // .grapes/ directory this issue was loaded from
	Worktree  string // worktree name (empty for main issues)
}
