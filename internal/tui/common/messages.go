package common

// Screen identifies which view is active.
type Screen int

const (
	ScreenBoard Screen = iota
	ScreenList
	ScreenDetail
)

// Messages for screen routing (sent by views, handled by app).
type OpenDetailMsg struct{ ID int }
type GoBackMsg struct{}
type SwitchScreenMsg struct{ Screen Screen }
type RefreshMsg struct{}
