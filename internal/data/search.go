package data

import (
	"strings"
)

// MatchesQuery returns true if the issue matches all space-separated words
// in the query. Each word must appear somewhere in the issue's combined
// searchable text (title, status, priority, labels, content, comments).
// Matching is case-insensitive.
func MatchesQuery(issue Issue, query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return true
	}
	words := strings.Fields(q)
	text := issueSearchText(issue)
	for _, w := range words {
		if !strings.Contains(text, w) {
			return false
		}
	}
	return true
}

// MatchSnippet returns a short context snippet showing why an issue matched
// the query. It only returns a snippet for matches outside the title (since
// the title is already visible). Returns "" if all matches are in the title.
func MatchSnippet(issue Issue, query string) string {
	words := strings.Fields(strings.ToLower(strings.TrimSpace(query)))
	if len(words) == 0 {
		return ""
	}
	titleLower := strings.ToLower(issue.Title)

	for _, w := range words {
		if strings.Contains(titleLower, w) {
			continue // visible in title already
		}

		// Search content
		if idx := strings.Index(strings.ToLower(issue.Content), w); idx >= 0 {
			return snippetAround(issue.Content, idx, len(w), 40)
		}
		// Search comments
		for _, c := range issue.Comments {
			if idx := strings.Index(strings.ToLower(c.Body), w); idx >= 0 {
				return snippetAround(c.Body, idx, len(w), 40)
			}
		}
		// Search labels
		for _, l := range issue.Labels {
			if strings.Contains(strings.ToLower(l), w) {
				return l
			}
		}
		if strings.Contains(strings.ToLower(string(issue.Status)), w) {
			return string(issue.Status)
		}
		if strings.Contains(strings.ToLower(string(issue.Priority)), w) {
			return string(issue.Priority)
		}
	}
	return ""
}

// snippetAround extracts a short excerpt from text centered on the match
// at position idx with the given match length. maxLen controls the total
// snippet length. Newlines are replaced with spaces.
func snippetAround(text string, idx, matchLen, maxLen int) string {
	// Flatten newlines
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", "")

	// Recalculate idx after flattening (same positions since we only replace single chars)
	if idx > len(text) {
		idx = len(text)
	}

	half := (maxLen - matchLen) / 2
	start := idx - half
	end := idx + matchLen + half

	prefix := ""
	suffix := ""
	if start < 0 {
		start = 0
	} else {
		prefix = "..."
	}
	if end > len(text) {
		end = len(text)
	} else {
		suffix = "..."
	}

	return prefix + strings.TrimSpace(text[start:end]) + suffix
}

// issueSearchText builds a single lowercase string from all searchable
// fields of an issue. Called once per issue per query evaluation.
func issueSearchText(iss Issue) string {
	var b strings.Builder
	b.WriteString(strings.ToLower(iss.Title))
	b.WriteByte(' ')
	b.WriteString(strings.ToLower(string(iss.Status)))
	b.WriteByte(' ')
	b.WriteString(strings.ToLower(string(iss.Priority)))
	for _, l := range iss.Labels {
		b.WriteByte(' ')
		b.WriteString(strings.ToLower(l))
	}
	if iss.Content != "" {
		b.WriteByte(' ')
		b.WriteString(strings.ToLower(iss.Content))
	}
	for _, c := range iss.Comments {
		b.WriteByte(' ')
		b.WriteString(strings.ToLower(c.Body))
	}
	if iss.Worktree != "" {
		b.WriteByte(' ')
		b.WriteString(strings.ToLower(iss.Worktree))
	}
	return b.String()
}
