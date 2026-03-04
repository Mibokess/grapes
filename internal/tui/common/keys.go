package common

import (
	"github.com/Mibokess/grapes/internal/config"
	"charm.land/bubbles/v2/key"
)

// GlobalKeys are available on every screen.
type GlobalKeys struct {
	Quit key.Binding
	Help key.Binding
}

var GlobalKeyMap = GlobalKeys{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}

// BoardKeys are specific to the Kanban board screen.
type BoardKeys struct {
	Up            key.Binding
	Down          key.Binding
	Left          key.Binding
	Right         key.Binding
	Open          key.Binding
	EditIssue     key.Binding
	ToList        key.Binding
	Filter        key.Binding
	Refresh       key.Binding
	CycleStatus   key.Binding
	CyclePriority key.Binding
	CycleSort     key.Binding
	ReverseSort   key.Binding
}

var BoardKeyMap = BoardKeys{
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k/up", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j/down", "down"),
	),
	Left: key.NewBinding(
		key.WithKeys("h", "left"),
		key.WithHelp("h/left", "left"),
	),
	Right: key.NewBinding(
		key.WithKeys("l", "right"),
		key.WithHelp("l/right", "right"),
	),
	Open: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open"),
	),
	EditIssue: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
	),
	ToList: key.NewBinding(
		key.WithKeys("L"),
		key.WithHelp("L", "list view"),
	),
	Filter: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "filter"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	CycleStatus: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "status"),
	),
	CyclePriority: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "priority"),
	),
	CycleSort: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "order"),
	),
	ReverseSort: key.NewBinding(
		key.WithKeys("O"),
		key.WithHelp("O", "reverse"),
	),
}

// ListKeys are specific to the list screen.
type ListKeys struct {
	Up               key.Binding
	Down             key.Binding
	Open             key.Binding
	EditIssue        key.Binding
	ToBoard          key.Binding
	Filter           key.Binding
	StructuredFilter key.Binding
	Clear            key.Binding
	Refresh          key.Binding
	CycleStatus      key.Binding
	CyclePriority    key.Binding
	CycleSort        key.Binding
	ReverseSort      key.Binding
	ScrollLeft       key.Binding
	ScrollRight      key.Binding
}

var ListKeyMap = ListKeys{
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k/up", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j/down", "down"),
	),
	Open: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open"),
	),
	EditIssue: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
	),
	ToBoard: key.NewBinding(
		key.WithKeys("B"),
		key.WithHelp("B", "board view"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	StructuredFilter: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "filter"),
	),
	Clear: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "clear filter"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	CycleStatus: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "status"),
	),
	CyclePriority: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "priority"),
	),
	CycleSort: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "order"),
	),
	ReverseSort: key.NewBinding(
		key.WithKeys("O"),
		key.WithHelp("O", "reverse"),
	),
	ScrollLeft: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "scroll left"),
	),
	ScrollRight: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "scroll right"),
	),
}

// DetailKeys are specific to the detail screen.
type DetailKeys struct {
	Back          key.Binding
	ToBoard       key.Binding
	ToList        key.Binding
	CycleStatus   key.Binding
	CyclePriority key.Binding
	AddComment    key.Binding
	EditIssue     key.Binding
}

var DetailKeyMap = DetailKeys{
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	ToBoard: key.NewBinding(
		key.WithKeys("B"),
		key.WithHelp("B", "board view"),
	),
	ToList: key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("l", "list view"),
	),
	CycleStatus: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "status"),
	),
	CyclePriority: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "priority"),
	),
	AddComment: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "comment"),
	),
	EditIssue: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
	),
}

// SettingsKeys are available on the settings screen.
type SettingsKeys struct {
	Up    key.Binding
	Down  key.Binding
	Tab   key.Binding
	Enter key.Binding
	Save  key.Binding
	Back  key.Binding
}

var SettingsKeyMap = SettingsKeys{
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k/up", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j/down", "down"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch pane"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "edit"),
	),
	Save: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
}

// ApplyKeys updates all keybinding vars from a KeysConfig.
func ApplyKeys(k config.KeysConfig) {
	GlobalKeyMap.Quit = key.NewBinding(key.WithKeys(k.Quit, "ctrl+c"), key.WithHelp(k.Quit, "quit"))

	BoardKeyMap.Up = key.NewBinding(key.WithKeys(k.BoardUp, "up"), key.WithHelp(k.BoardUp+"/up", "up"))
	BoardKeyMap.Down = key.NewBinding(key.WithKeys(k.BoardDown, "down"), key.WithHelp(k.BoardDown+"/down", "down"))
	BoardKeyMap.Left = key.NewBinding(key.WithKeys(k.BoardLeft, "left"), key.WithHelp(k.BoardLeft+"/left", "left"))
	BoardKeyMap.Right = key.NewBinding(key.WithKeys(k.BoardRight, "right"), key.WithHelp(k.BoardRight+"/right", "right"))
	BoardKeyMap.Open = key.NewBinding(key.WithKeys(k.BoardOpen), key.WithHelp(k.BoardOpen, "open"))
	BoardKeyMap.EditIssue = key.NewBinding(key.WithKeys(k.BoardEdit), key.WithHelp(k.BoardEdit, "edit"))
	BoardKeyMap.ToList = key.NewBinding(key.WithKeys(k.BoardToList), key.WithHelp(k.BoardToList, "list view"))
	BoardKeyMap.Filter = key.NewBinding(key.WithKeys(k.BoardFilter), key.WithHelp(k.BoardFilter, "filter"))
	BoardKeyMap.CycleStatus = key.NewBinding(key.WithKeys(k.BoardStatus), key.WithHelp(k.BoardStatus, "status"))
	BoardKeyMap.CyclePriority = key.NewBinding(key.WithKeys(k.BoardPriority), key.WithHelp(k.BoardPriority, "priority"))
	BoardKeyMap.CycleSort = key.NewBinding(key.WithKeys(k.BoardSort), key.WithHelp(k.BoardSort, "order"))
	BoardKeyMap.ReverseSort = key.NewBinding(key.WithKeys(k.BoardReverse), key.WithHelp(k.BoardReverse, "reverse"))

	ListKeyMap.Up = key.NewBinding(key.WithKeys(k.ListUp, "up"), key.WithHelp(k.ListUp+"/up", "up"))
	ListKeyMap.Down = key.NewBinding(key.WithKeys(k.ListDown, "down"), key.WithHelp(k.ListDown+"/down", "down"))
	ListKeyMap.Open = key.NewBinding(key.WithKeys(k.ListOpen), key.WithHelp(k.ListOpen, "open"))
	ListKeyMap.EditIssue = key.NewBinding(key.WithKeys(k.ListEdit), key.WithHelp(k.ListEdit, "edit"))
	ListKeyMap.ToBoard = key.NewBinding(key.WithKeys(k.ListToBoard), key.WithHelp(k.ListToBoard, "board view"))
	ListKeyMap.Filter = key.NewBinding(key.WithKeys(k.ListSearch), key.WithHelp(k.ListSearch, "search"))
	ListKeyMap.StructuredFilter = key.NewBinding(key.WithKeys(k.ListFilter), key.WithHelp(k.ListFilter, "filter"))
	ListKeyMap.CycleStatus = key.NewBinding(key.WithKeys(k.ListStatus), key.WithHelp(k.ListStatus, "status"))
	ListKeyMap.CyclePriority = key.NewBinding(key.WithKeys(k.ListPriority), key.WithHelp(k.ListPriority, "priority"))
	ListKeyMap.CycleSort = key.NewBinding(key.WithKeys(k.ListSort), key.WithHelp(k.ListSort, "order"))
	ListKeyMap.ReverseSort = key.NewBinding(key.WithKeys(k.ListReverse), key.WithHelp(k.ListReverse, "reverse"))

	DetailKeyMap.Back = key.NewBinding(key.WithKeys(k.DetailBack), key.WithHelp(k.DetailBack, "back"))
	DetailKeyMap.ToBoard = key.NewBinding(key.WithKeys(k.DetailToBoard), key.WithHelp(k.DetailToBoard, "board view"))
	DetailKeyMap.ToList = key.NewBinding(key.WithKeys(k.DetailToList), key.WithHelp(k.DetailToList, "list view"))
	DetailKeyMap.CycleStatus = key.NewBinding(key.WithKeys(k.DetailStatus), key.WithHelp(k.DetailStatus, "status"))
	DetailKeyMap.CyclePriority = key.NewBinding(key.WithKeys(k.DetailPriority), key.WithHelp(k.DetailPriority, "priority"))
	DetailKeyMap.AddComment = key.NewBinding(key.WithKeys(k.DetailComment), key.WithHelp(k.DetailComment, "comment"))
	DetailKeyMap.EditIssue = key.NewBinding(key.WithKeys(k.DetailEdit), key.WithHelp(k.DetailEdit, "edit"))
}
