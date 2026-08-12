package tmux

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Session describes a Grapes-managed tmux session and one representative pane.
type Session struct {
	Name     string // actual tmux session name (may have been renamed)
	Target   string // session:window.pane target for the representative pane
	IssueID  int
	Worktree string
	Agent    string
	Path     string
	Pane     string
	Attached bool
}

// These hooks keep command interaction small and make the package's behavior
// testable without requiring tmux to be installed or a server to be running.
var tmuxLookPath = exec.LookPath
var tmuxRunCommand = defaultRunCommand

type commandRunner func(path string, args ...string) (stdout, stderr []byte, err error)

func defaultRunCommand(path string, args ...string) ([]byte, []byte, error) {
	cmd := exec.Command(path, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.Output()
	return stdout, stderr.Bytes(), err
}

type commandError struct {
	err    error
	stderr string
}

func (e *commandError) Error() string {
	if e.stderr == "" {
		return e.err.Error()
	}
	return fmt.Sprintf("%v: %s", e.err, strings.TrimSpace(e.stderr))
}

func (e *commandError) Unwrap() error { return e.err }

const listPanesFormat = "#{session_name}\t#{session_attached}\t#{@grapes_project}\t#{@grapes_issue}\t#{@grapes_worktree}\t#{@grapes_agent}\t#{window_index}\t#{pane_index}\t#{pane_current_path}\t#{pane_current_command}"

// List returns Grapes-managed sessions associated with projectRoot. Sessions
// are identified by their metadata, rather than by their names, so renaming a
// managed tmux session does not lose the association.
func List(projectRoot string) ([]Session, error) {
	path, err := tmuxLookPath("tmux")
	if err != nil {
		// tmux is optional. A normal TUI startup must not fail just because it
		// is unavailable.
		return []Session{}, nil
	}

	stdout, stderr, err := tmuxRunCommand(path, "list-panes", "-a", "-F", listPanesFormat)
	if err != nil {
		if isNoServerError(err, stderr) {
			return []Session{}, nil
		}
		if len(stderr) != 0 {
			return nil, &commandError{err: err, stderr: string(stderr)}
		}
		return nil, err
	}
	return parseSessions(string(stdout), normalizeRoot(projectRoot)), nil
}

// Ensure finds the managed session for an issue, or creates a detached shell
// session in cwd and records its Grapes metadata.
func Ensure(projectRoot string, issueID int, worktree, cwd string) (Session, error) {
	root := normalizeRoot(projectRoot)
	sessions, err := List(root)
	if err != nil {
		return Session{}, err
	}
	for _, session := range sessions {
		if session.IssueID == issueID && session.Worktree == worktree {
			return session, nil
		}
	}

	path, err := tmuxLookPath("tmux")
	if err != nil {
		return Session{}, fmt.Errorf("tmux is unavailable: %w", err)
	}
	if cwd == "" {
		cwd = root
	}
	if cwd == "" {
		cwd = "."
	}
	name := sessionName(root, issueID, worktree)
	if _, stderr, err := tmuxRunCommand(path, "new-session", "-d", "-s", name, "-c", cwd); err != nil {
		return Session{}, commandFailure(err, stderr)
	}
	for _, option := range [][2]string{
		{"@grapes_project", root},
		{"@grapes_issue", strconv.Itoa(issueID)},
		{"@grapes_worktree", worktree},
	} {
		if _, stderr, err := tmuxRunCommand(path, "set-option", "-t", name, option[0], option[1]); err != nil {
			return Session{}, commandFailure(err, stderr)
		}
	}

	// Query tmux again so callers receive the actual target and pane details.
	sessions, err = List(root)
	if err != nil {
		return Session{}, err
	}
	for _, session := range sessions {
		if session.IssueID == issueID && session.Worktree == worktree {
			return session, nil
		}
	}

	// A newly-created tmux session always starts with window and pane 0. Keep
	// a useful result if a compatible tmux implementation does not expose it
	// through list-panes immediately.
	return Session{
		Name:     name,
		Target:   name + ":0.0",
		IssueID:  issueID,
		Worktree: worktree,
		Path:     cwd,
	}, nil
}

// AttachCommand builds the terminal handoff command without invoking a shell.
func AttachCommand(target string) *exec.Cmd {
	return exec.Command("tmux", "attach-session", "-t", target)
}

func commandFailure(err error, stderr []byte) error {
	if len(stderr) == 0 {
		return err
	}
	return &commandError{err: err, stderr: string(stderr)}
}

func isNoServerError(err error, stderr []byte) bool {
	message := strings.ToLower(string(stderr) + " " + err.Error())
	for _, phrase := range []string{
		"no server running",
		"no sessions",
		"failed to connect to server",
		"error connecting to",
		"can't find socket",
		"cannot find socket",
		"no such file or directory",
	} {
		if strings.Contains(message, phrase) {
			return true
		}
	}
	return false
}

func parseSessions(output, projectRoot string) []Session {
	byName := make(map[string]Session)
	for _, line := range strings.Split(output, "\n") {
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 10 || fields[0] == "" || fields[2] == "" || fields[3] == "" {
			continue
		}
		if fields[2] != projectRoot {
			continue
		}
		issueID, err := strconv.Atoi(fields[3])
		if err != nil || issueID <= 0 {
			continue
		}
		window, err := strconv.Atoi(fields[6])
		if err != nil || window < 0 {
			continue
		}
		pane, err := strconv.Atoi(fields[7])
		if err != nil || pane < 0 {
			continue
		}
		session := Session{
			Name:     fields[0],
			Target:   fmt.Sprintf("%s:%d.%d", fields[0], window, pane),
			IssueID:  issueID,
			Worktree: fields[4],
			Agent:    fields[5],
			Path:     fields[8],
			Pane:     fields[9],
			Attached: fields[1] == "1" || strings.EqualFold(fields[1], "true"),
		}
		if previous, ok := byName[session.Name]; ok {
			if targetLess(session, previous) {
				byName[session.Name] = session
			}
			continue
		}
		byName[session.Name] = session
	}

	sessions := make([]Session, 0, len(byName))
	for _, session := range byName {
		sessions = append(sessions, session)
	}
	sort.Slice(sessions, func(i, j int) bool { return sessions[i].Name < sessions[j].Name })
	return sessions
}

func targetLess(a, b Session) bool {
	return a.Target < b.Target
}

func sessionName(projectRoot string, issueID int, worktree string) string {
	return fmt.Sprintf("grapes-%s-issue-%d-%s", projectIdentity(projectRoot), issueID, slug(worktree, "main"))
}

func projectIdentity(projectRoot string) string {
	root := normalizeRoot(projectRoot)
	base := "project"
	if root != "" {
		base = slug(filepath.Base(root), "project")
	}
	sum := sha256.Sum256([]byte(root))
	return base + "-" + hex.EncodeToString(sum[:4])
}

func normalizeRoot(root string) string {
	if root == "" {
		return ""
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return filepath.Clean(root)
	}
	return filepath.Clean(absolute)
}

func slug(value, fallback string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-' {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	result := strings.Trim(b.String(), "-")
	if result == "" {
		return fallback
	}
	return result
}
