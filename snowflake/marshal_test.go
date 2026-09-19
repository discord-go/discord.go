package snowflake

import (
	"encoding/json"
	"testing"
)

// TestIDMarshalJSONAsString pins the Discord wire format: snowflakes are
// JSON strings, never numbers. A numeric encoding is rejected by Discord
// with 50035 Invalid Form Body, most visibly in allowed_mentions.users.
func TestIDMarshalJSONAsString(t *testing.T) {
	id, err := Parse("1528182903698100354")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	data, err := json.Marshal(id)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(data) != `"1528182903698100354"` {
		t.Errorf("marshal = %s, want %q", data, `"1528182903698100354"`)
	}
}

// TestIDMarshalJSONInStruct covers struct fields both with and without the
// legacy quoted tag: the custom marshaler must win without double-quoting
// the tagged fields.
func TestIDMarshalJSONInStruct(t *testing.T) {
	id, err := Parse("1234567890123456789")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	payload := struct {
		Plain  ID  `json:"plain"`
		Quoted *ID `json:"quoted,string"`
	}{Plain: id, Quoted: &id}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := `{"plain":"1234567890123456789","quoted":"1234567890123456789"}`
	if string(data) != want {
		t.Errorf("marshal = %s, want %s", data, want)
	}
}

// TestIDRoundTripThroughQuotedTag keeps the decode side working: fields
// tagged ,string still decode a JSON string into the ID.
func TestIDRoundTripThroughQuotedTag(t *testing.T) {
	var payload struct {
		Quoted *ID `json:"quoted,string"`
	}
	if err := json.Unmarshal([]byte(`{"quoted":"1234567890123456789"}`), &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload.Quoted == nil || payload.Quoted.String() != "1234567890123456789" {
		t.Errorf("decoded = %v, want 1234567890123456789", payload.Quoted)
	}
}
