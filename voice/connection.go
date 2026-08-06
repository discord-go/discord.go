package voice

import "fmt"

// ConnectionState represents the state of a voice connection.
type ConnectionState int

const (
	// StateIdle indicates no active voice connection.
	StateIdle ConnectionState = iota
	// StateConnecting indicates the voice connection is being established.
	StateConnecting
	// StateConnected indicates the voice connection is active.
	StateConnected
	// StateDisconnecting indicates the voice connection is being torn down.
	StateDisconnecting
	// StateReconnecting indicates the voice connection is reconnecting.
	StateReconnecting
)

// String returns the string representation of a ConnectionState.
func (s ConnectionState) String() string {
	switch s {
	case StateIdle:
		return "Idle"
	case StateConnecting:
		return "Connecting"
	case StateConnected:
		return "Connected"
	case StateDisconnecting:
		return "Disconnecting"
	case StateReconnecting:
		return "Reconnecting"
	default:
		return fmt.Sprintf("ConnectionState(%d)", int(s))
	}
}

// IsActive returns true if the connection state represents an active or
// transitioning connection (anything other than Idle).
func (s ConnectionState) IsActive() bool {
	return s != StateIdle
}
