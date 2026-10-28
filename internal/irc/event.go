package irc

import "time"

type EventKind int

const (
	EvConnecting EventKind = iota
	EvConnected
	EvMessage
	EvDisconnected
	EvError
	EvLag
)

// Event is what a Client reports to the UI. Text carries a human-readable
// status or error message in the language active when it was created.
type Event struct {
	Network string
	Kind    EventKind
	Msg     Message
	Text    string
	Lag     time.Duration
	Time    time.Time
}
