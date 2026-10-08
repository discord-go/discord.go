package messages

import (
	"encoding/json"
	"testing"
)

func TestNewAttachmentParams(t *testing.T) {
	params := NewAttachmentParams(0, "report.pdf")
	if params.ID != 0 || params.Filename != "report.pdf" {
		t.Errorf("params = %+v", params)
	}
}

func TestAttachmentParamsRequestJSON(t *testing.T) {
	params := AttachmentParams{
		ID:           0,
		Filename:     "clip.mp4",
		Title:        "Demo",
		Description:  "A clip",
		DurationSecs: 12.5,
		Waveform:     "AAAA",
		IsSpoiler:    true,
	}
	raw, err := json.Marshal(params)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"id":            "0",
		"filename":      "clip.mp4",
		"title":         "Demo",
		"description":   "A clip",
		"duration_secs": 12.5,
		"waveform":      "AAAA",
		"is_spoiler":    true,
	}
	if len(decoded) != len(want) {
		t.Fatalf("keys = %v", decoded)
	}
	for key, value := range want {
		if decoded[key] != value {
			t.Errorf("%s = %v, want %v", key, decoded[key], value)
		}
	}
}

func TestAttachmentParamsExistingReferenceOmitsOptionalFields(t *testing.T) {
	raw, err := json.Marshal(AttachmentParams{ID: 123456789012345678})
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != `{"id":"123456789012345678"}` {
		t.Errorf("json = %s", raw)
	}
}

func TestAttachmentParamsUnmarshal(t *testing.T) {
	var params AttachmentParams
	if err := json.Unmarshal([]byte(`{"id":"42","filename":"a.png","is_spoiler":true}`), &params); err != nil {
		t.Fatal(err)
	}
	if params.ID != 42 || params.Filename != "a.png" || !params.IsSpoiler {
		t.Errorf("params = %+v", params)
	}
}
