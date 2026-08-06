package voice

import (
	"testing"
)

func TestConnectionState_String(t *testing.T) {
	tests := []struct {
		state ConnectionState
		want  string
	}{
		{StateIdle, "Idle"},
		{StateConnecting, "Connecting"},
		{StateConnected, "Connected"},
		{StateDisconnecting, "Disconnecting"},
		{StateReconnecting, "Reconnecting"},
		{ConnectionState(99), "ConnectionState(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.state.String(); got != tt.want {
				t.Errorf("ConnectionState.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConnectionState_IsActive(t *testing.T) {
	tests := []struct {
		state ConnectionState
		want  bool
	}{
		{StateIdle, false},
		{StateConnecting, true},
		{StateConnected, true},
		{StateDisconnecting, true},
		{StateReconnecting, true},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			if got := tt.state.IsActive(); got != tt.want {
				t.Errorf("ConnectionState.IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}
