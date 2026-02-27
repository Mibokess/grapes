package data

import "time"

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
	StatusBacklog:    "Backlog",
	StatusTodo:       "Todo",
	StatusInProgress: "In Progress",
	StatusDone:       "Done",
	StatusCancelled:  "Cancelled",
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
	PriorityUrgent: "Urgent",
	PriorityHigh:   "High",
	PriorityMedium: "Medium",
	PriorityLow:    "Low",
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

type Comment struct {
	Author string
	Date   string
	Body   string
}

type Issue struct {
	ID       int
	Title    string
	Status   Status
	Priority Priority
	Assignee string
	Labels   []string
	Parent   *int
	Children []int
	Created  time.Time
	Updated  time.Time
	Content  string
	Comments []Comment
}
