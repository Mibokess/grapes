package filter

import (
	"github.com/Mibokess/grapes/internal/data"
)

// FilterSet holds the active structured filters. Categories are AND'd together;
// values within a category are OR'd.
type FilterSet struct {
	Statuses     []data.Status
	Priorities   []data.Priority
	Labels       []string
	Sources      []string // worktree names; "main" for main issues
	HasChildren  *bool
	TopLevelOnly bool // when true, only show issues without a parent
	TextQuery    string
}

// Matches returns true if the issue passes all active filters.
func (f FilterSet) Matches(issue data.Issue) bool {
	if len(f.Statuses) > 0 && !containsStatus(f.Statuses, issue.Status) {
		return false
	}
	if len(f.Priorities) > 0 && !containsPriority(f.Priorities, issue.Priority) {
		return false
	}
	if len(f.Labels) > 0 && !hasAnyLabel(f.Labels, issue.Labels) {
		return false
	}
	if len(f.Sources) > 0 {
		src := issue.Worktree
		if src == "" {
			src = "main"
		}
		if !containsString(f.Sources, src) {
			return false
		}
	}
	if f.HasChildren != nil {
		has := len(issue.Children) > 0
		if has != *f.HasChildren {
			return false
		}
	}
	if f.TopLevelOnly && issue.Parent != nil {
		return false
	}
	if f.TextQuery != "" {
		if !data.MatchesQuery(issue, f.TextQuery) {
			return false
		}
	}
	return true
}

// IsEmpty returns true when no filters are active.
func (f FilterSet) IsEmpty() bool {
	return len(f.Statuses) == 0 &&
		len(f.Priorities) == 0 &&
		len(f.Labels) == 0 &&
		len(f.Sources) == 0 &&
		f.HasChildren == nil &&
		!f.TopLevelOnly &&
		f.TextQuery == ""
}

// ActiveCount returns the number of active filter categories (excluding text query).
func (f FilterSet) ActiveCount() int {
	n := 0
	if len(f.Statuses) > 0 {
		n++
	}
	if len(f.Priorities) > 0 {
		n++
	}
	if len(f.Labels) > 0 {
		n++
	}
	if len(f.Sources) > 0 {
		n++
	}
	if f.HasChildren != nil {
		n++
	}
	if f.TopLevelOnly {
		n++
	}
	return n
}

// Clear resets all filters.
func (f *FilterSet) Clear() {
	f.Statuses = nil
	f.Priorities = nil
	f.Labels = nil
	f.Sources = nil
	f.HasChildren = nil
	f.TopLevelOnly = false
	f.TextQuery = ""
}

// ToggleStatus adds or removes a status from the filter.
func (f *FilterSet) ToggleStatus(s data.Status) {
	for i, v := range f.Statuses {
		if v == s {
			f.Statuses = append(f.Statuses[:i], f.Statuses[i+1:]...)
			return
		}
	}
	f.Statuses = append(f.Statuses, s)
}

// TogglePriority adds or removes a priority from the filter.
func (f *FilterSet) TogglePriority(p data.Priority) {
	for i, v := range f.Priorities {
		if v == p {
			f.Priorities = append(f.Priorities[:i], f.Priorities[i+1:]...)
			return
		}
	}
	f.Priorities = append(f.Priorities, p)
}

// ToggleLabel adds or removes a label from the filter.
func (f *FilterSet) ToggleLabel(l string) {
	for i, v := range f.Labels {
		if v == l {
			f.Labels = append(f.Labels[:i], f.Labels[i+1:]...)
			return
		}
	}
	f.Labels = append(f.Labels, l)
}

// Default returns a FilterSet with the default filters applied (top-level only).
func Default() FilterSet {
	return FilterSet{TopLevelOnly: true}
}

// ToggleTopLevelOnly toggles the top-level-only filter.
func (f *FilterSet) ToggleTopLevelOnly() {
	f.TopLevelOnly = !f.TopLevelOnly
}

// ToggleHasChildren cycles nil → true → false → nil.
func (f *FilterSet) ToggleHasChildren() {
	if f.HasChildren == nil {
		t := true
		f.HasChildren = &t
	} else if *f.HasChildren {
		fa := false
		f.HasChildren = &fa
	} else {
		f.HasChildren = nil
	}
}

// SetStatuses replaces the status filter with the given values.
func (f *FilterSet) SetStatuses(statuses []string) {
	f.Statuses = nil
	for _, s := range statuses {
		f.Statuses = append(f.Statuses, data.Status(s))
	}
}

// SetPriorities replaces the priority filter with the given values.
func (f *FilterSet) SetPriorities(priorities []string) {
	f.Priorities = nil
	for _, p := range priorities {
		f.Priorities = append(f.Priorities, data.Priority(p))
	}
}

// SetLabels replaces the label filter with the given values.
func (f *FilterSet) SetLabels(labels []string) {
	f.Labels = labels
}

// SetSources replaces the source filter with the given values.
func (f *FilterSet) SetSources(sources []string) {
	f.Sources = sources
}

func containsStatus(ss []data.Status, s data.Status) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func containsPriority(ps []data.Priority, p data.Priority) bool {
	for _, v := range ps {
		if v == p {
			return true
		}
	}
	return false
}

func containsString(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func hasAnyLabel(filter, issueLabels []string) bool {
	for _, fl := range filter {
		for _, il := range issueLabels {
			if fl == il {
				return true
			}
		}
	}
	return false
}
