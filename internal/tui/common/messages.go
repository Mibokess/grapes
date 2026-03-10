package common

import (
	"github.com/Mibokess/grapes/internal/config"
	"github.com/Mibokess/grapes/internal/data"
)

// Screen identifies which view is active.
type Screen int

const (
	ScreenBoard Screen = iota
	ScreenList
	ScreenDetail
	ScreenSettings
)

// Messages for screen routing (sent by views, handled by app).
type OpenDetailMsg struct{ ID int }
type GoBackMsg struct{}
type SwitchScreenMsg struct{ Screen Screen }
type RefreshMsg struct{}

// Messages for write operations.
type ShowPickerMsg struct {
	IssueID int
	Field   string // "status" or "priority"
}
type PickerResultMsg struct {
	IssueID int
	Field   string
	Value   string
}
type PickerCancelMsg struct{}
type ShowLabelPickerMsg struct{ IssueID int }
type LabelPickerResultMsg struct {
	IssueID int
	Labels  []string
}
type LabelPickerCancelMsg struct{}
type LaunchEditorMsg struct{ ID int }
type EditorFinishedMsg struct{ Err error }
type LaunchEditMsg struct{ ID int }
type EditFinishedMsg struct{ Err error }
type WriteErrMsg struct{ Err error }
type CycleSortMsg struct{}
type ReverseSortMsg struct{}
type ToggleEmptyColumnsMsg struct{}
type ColumnSortMsg struct{ Mode data.SortMode }
type MoveIssueMsg struct {
	IssueID   int
	NewStatus data.Status
}

// Filter overlay messages.
type ShowFilterMenuMsg struct{}
type FilterMenuSelectMsg struct{ Field string }
type FilterPickerResultMsg struct {
	Field    string
	Selected []string
}
type FilterToggleChildrenMsg struct{}
type FilterToggleTopLevelMsg struct{}
type FilterCancelMsg struct{}
type ClearAllFiltersMsg struct{}

// Multi-source worktree messages.
type SwitchSourceMsg struct {
	IssueID   int
	SourceIdx int
}

// Settings screen messages.
type OpenSettingsMsg struct{}
type ConfigSavedMsg struct{ Config config.Config }
