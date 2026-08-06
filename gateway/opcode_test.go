package gateway

import (
	"testing"
)

func TestOpcode(t *testing.T) {
	if OpcodeDispatch != 0 {
		t.Error("OpcodeDispatch should be 0")
	}
	if OpcodeHeartbeatACK != 11 {
		t.Error("OpcodeHeartbeatACK should be 11")
	}
}
