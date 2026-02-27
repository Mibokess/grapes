package common

import "github.com/charmbracelet/bubbles/key"

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
	Up      key.Binding
	Down    key.Binding
	Left    key.Binding
	Right   key.Binding
	Open    key.Binding
	ToList  key.Binding
	Refresh key.Binding
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
	ToList: key.NewBinding(
		key.WithKeys("L"),
		key.WithHelp("L", "list view"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
}

// ListKeys are specific to the list screen.
type ListKeys struct {
	Up      key.Binding
	Down    key.Binding
	Open    key.Binding
	ToBoard key.Binding
	Filter  key.Binding
	Clear   key.Binding
	Refresh key.Binding
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
	ToBoard: key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "board view"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	Clear: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "clear filter"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
}

// DetailKeys are specific to the detail screen.
type DetailKeys struct {
	Back    key.Binding
	ToBoard key.Binding
	ToList  key.Binding
}

var DetailKeyMap = DetailKeys{
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	ToBoard: key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "board view"),
	),
	ToList: key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("l", "list view"),
	),
}
