package data

import (
	"fmt"
	"testing"
)

func BenchmarkSortIssuesByTitle(b *testing.B) {
	issues := make([]Issue, 2048)
	for i := range issues {
		issues[i] = Issue{ID: i + 1, Title: fmt.Sprintf("Issue title %04d", (i*37)%issuesLen)}
	}
	b.ResetTimer()
	for range b.N {
		SortIssues(issues, SortByTitle, false)
	}
}

const issuesLen = 2048

func TestSortIssues_TitleReverseAndUnknownValues(t *testing.T) {
	issues := []Issue{
		{ID: 1, Title: "same", Priority: PriorityUrgent, Status: StatusTodo},
		{ID: 2, Title: "same", Priority: PriorityUrgent, Status: StatusTodo},
		{ID: 3, Title: "other", Priority: Priority("future"), Status: Status("future")},
	}

	SortIssues(issues, SortByTitle, true)
	if issues[0].ID != 2 || issues[1].ID != 1 || issues[2].ID != 3 {
		t.Fatalf("reverse title sort = %v, want [2 1 3]", []int{issues[0].ID, issues[1].ID, issues[2].ID})
	}
	SortIssues(issues, SortByPriority, false)
	if issues[len(issues)-1].ID != 3 {
		t.Fatalf("unknown priority did not sort last: %v", issues)
	}
	SortIssues(issues, SortByStatus, false)
	if issues[len(issues)-1].ID != 3 {
		t.Fatalf("unknown status did not sort last: %v", issues)
	}
}
