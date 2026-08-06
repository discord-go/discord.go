package voice

import "testing"

func TestOpcodes(t *testing.T) {
	if Identify != 0 || Ready != 2 || Speaking != 5 || ClientDisconnect != 13 {
		t.Errorf("Opcodes are incorrect")
	}
}
