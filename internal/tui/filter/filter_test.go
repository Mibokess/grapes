package filter

import (
	"testing"

	"github.com/Mibokess/grapes/internal/data"
)

func boolPtr(b bool) *bool { return &b }

func TestFilterSet_Matches(t *testing.T) {
	base := data.Issue{
		ID:       1,
		Title:    "Fix login bug",
		Status:   data.StatusInProgress,
		Priority: data.PriorityHigh,
		Labels:   []string{"bug", "auth"},
		Children: []int{2, 3},
		Content:  "The login page crashes on submit",
	}

	tests := []struct {
		name   string
		filter FilterSet
		issue  data.Issue
		want   bool
	}{
		{
			name:   "empty filter matches everything",
			filter: FilterSet{},
			issue:  base,
			want:   true,
		},
		{
			name:   "status match",
			filter: FilterSet{Statuses: []data.Status{data.StatusInProgress}},
			issue:  base,
			want:   true,
		},
		{
			name:   "status no match",
			filter: FilterSet{Statuses: []data.Status{data.StatusDone}},
			issue:  base,
			want:   false,
		},
		{
			name:   "status OR within category",
			filter: FilterSet{Statuses: []data.Status{data.StatusDone, data.StatusInProgress}},
			issue:  base,
			want:   true,
		},
		{
			name:   "priority match",
			filter: FilterSet{Priorities: []data.Priority{data.PriorityHigh}},
			issue:  base,
			want:   true,
		},
		{
			name:   "priority no match",
			filter: FilterSet{Priorities: []data.Priority{data.PriorityLow}},
			issue:  base,
			want:   false,
		},
		{
			name:   "label match one of many",
			filter: FilterSet{Labels: []string{"bug"}},
			issue:  base,
			want:   true,
		},
		{
			name:   "label no match",
			filter: FilterSet{Labels: []string{"docs"}},
			issue:  base,
			want:   false,
		},
		{
			name:   "label OR - any filter label matches any issue label",
			filter: FilterSet{Labels: []string{"docs", "auth"}},
			issue:  base,
			want:   true,
		},
		{
			name:   "has children true - issue has children",
			filter: FilterSet{HasChildren: boolPtr(true)},
			issue:  base,
			want:   true,
		},
		{
			name:   "has children false - issue has children",
			filter: FilterSet{HasChildren: boolPtr(false)},
			issue:  base,
			want:   false,
		},
		{
			name:   "has children true - issue has no children",
			filter: FilterSet{HasChildren: boolPtr(true)},
			issue:  data.Issue{ID: 2, Status: data.StatusTodo, Priority: data.PriorityLow},
			want:   false,
		},
		{
			name:   "has children false - issue has no children",
			filter: FilterSet{HasChildren: boolPtr(false)},
			issue:  data.Issue{ID: 2, Status: data.StatusTodo, Priority: data.PriorityLow},
			want:   true,
		},
		{
			name:   "text query matches title",
			filter: FilterSet{TextQuery: "login"},
			issue:  base,
			want:   true,
		},
		{
			name:   "text query matches content",
			filter: FilterSet{TextQuery: "crashes"},
			issue:  base,
			want:   true,
		},
		{
			name:   "text query case insensitive",
			filter: FilterSet{TextQuery: "LOGIN"},
			issue:  base,
			want:   true,
		},
		{
			name:   "text query no match",
			filter: FilterSet{TextQuery: "dashboard"},
			issue:  base,
			want:   false,
		},
		{
			name: "AND across categories - all match",
			filter: FilterSet{
				Statuses:   []data.Status{data.StatusInProgress},
				Priorities: []data.Priority{data.PriorityHigh},
				Labels:     []string{"bug"},
			},
			issue: base,
			want:  true,
		},
		{
			name: "AND across categories - one fails",
			filter: FilterSet{
				Statuses:   []data.Status{data.StatusInProgress},
				Priorities: []data.Priority{data.PriorityLow},
				Labels:     []string{"bug"},
			},
			issue: base,
			want:  false,
		},
		{
			name:   "issue with no labels - label filter active",
			filter: FilterSet{Labels: []string{"bug"}},
			issue:  data.Issue{ID: 2, Status: data.StatusTodo, Priority: data.PriorityLow},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.Matches(tt.issue)
			if got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterSet_IsEmpty(t *testing.T) {
	if !(FilterSet{}).IsEmpty() {
		t.Error("empty FilterSet should be empty")
	}
	if (FilterSet{Statuses: []data.Status{data.StatusDone}}).IsEmpty() {
		t.Error("FilterSet with statuses should not be empty")
	}
	if (FilterSet{TextQuery: "foo"}).IsEmpty() {
		t.Error("FilterSet with text query should not be empty")
	}
}

func TestFilterSet_Toggle(t *testing.T) {
	var f FilterSet
	f.ToggleStatus(data.StatusDone)
	if len(f.Statuses) != 1 || f.Statuses[0] != data.StatusDone {
		t.Error("ToggleStatus should add")
	}
	f.ToggleStatus(data.StatusDone)
	if len(f.Statuses) != 0 {
		t.Error("ToggleStatus should remove")
	}

	f.TogglePriority(data.PriorityHigh)
	if len(f.Priorities) != 1 {
		t.Error("TogglePriority should add")
	}
	f.TogglePriority(data.PriorityHigh)
	if len(f.Priorities) != 0 {
		t.Error("TogglePriority should remove")
	}

	f.ToggleLabel("bug")
	if len(f.Labels) != 1 {
		t.Error("ToggleLabel should add")
	}
	f.ToggleLabel("bug")
	if len(f.Labels) != 0 {
		t.Error("ToggleLabel should remove")
	}
}

func TestFilterSet_ToggleHasChildren(t *testing.T) {
	var f FilterSet
	if f.HasChildren != nil {
		t.Error("should start nil")
	}
	f.ToggleHasChildren()
	if f.HasChildren == nil || !*f.HasChildren {
		t.Error("first toggle should be true")
	}
	f.ToggleHasChildren()
	if f.HasChildren == nil || *f.HasChildren {
		t.Error("second toggle should be false")
	}
	f.ToggleHasChildren()
	if f.HasChildren != nil {
		t.Error("third toggle should be nil")
	}
}

func TestFilterSet_Clear(t *testing.T) {
	f := FilterSet{
		Statuses:    []data.Status{data.StatusDone},
		Priorities:  []data.Priority{data.PriorityHigh},
		Labels:      []string{"bug"},
		HasChildren: boolPtr(true),
		TextQuery:   "foo",
	}
	f.Clear()
	if !f.IsEmpty() {
		t.Error("Clear should make FilterSet empty")
	}
}

func TestFilterSet_ActiveCount(t *testing.T) {
	f := FilterSet{
		Statuses:   []data.Status{data.StatusDone},
		Priorities: []data.Priority{data.PriorityHigh},
		TextQuery:  "foo", // not counted
	}
	if f.ActiveCount() != 2 {
		t.Errorf("ActiveCount() = %d, want 2", f.ActiveCount())
	}
}
