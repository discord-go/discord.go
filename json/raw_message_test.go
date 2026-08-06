package json

import (
	"reflect"
	"testing"
)

func TestRawMessage(t *testing.T) {
	var m RawMessage
	err := Unmarshal([]byte(`{"test":123}`), &m)
	if err != nil {
		t.Fatalf("Failed to unmarshal into RawMessage: %v", err)
	}

	want := RawMessage(`{"test":123}`)
	if !reflect.DeepEqual(m, want) {
		t.Errorf("RawMessage = %s, want %s", m, want)
	}

	b, err := Marshal(m)
	if err != nil {
		t.Fatalf("Failed to marshal RawMessage: %v", err)
	}
	if string(b) != `{"test":123}` {
		t.Errorf("Marshal(RawMessage) = %s, want %s", string(b), `{"test":123}`)
	}

	// Test MarshalJSON and UnmarshalJSON directly
	var r RawMessage
	if err := r.UnmarshalJSON([]byte(`"hello"`)); err != nil {
		t.Errorf("UnmarshalJSON error = %v", err)
	}

	b, err = r.MarshalJSON()
	if err != nil {
		t.Errorf("MarshalJSON error = %v", err)
	}
	if string(b) != `"hello"` {
		t.Errorf("MarshalJSON() = %s, want %s", string(b), `"hello"`)
	}
}
