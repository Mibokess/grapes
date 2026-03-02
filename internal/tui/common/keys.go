package common

import "charm.land/bubbles/v2/key"

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
