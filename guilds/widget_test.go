package guilds

import (
	"encoding/json"
	"testing"
)

func TestWidgetUnmarshalJSON(t *testing.T) {
	// Test Widget
	data := []byte(`{
		"id": "123456789012345678",
		"name": "Widget Name",
		"instant_invite": null,
		"channels": [
			{
				"id": "987654321098765432",
				"name": "Channel 1",
				"position": 0
			}
		],
		"members": [],
		"presence_count": 0
	}`)

	var w Widget
	if err := json.Unmarshal(data, &w); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.ID.String() != "123456789012345678" {
		t.Errorf("expected 123456789012345678, got %v", w.ID)
	}
	if len(w.Channels) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(w.Channels))
	}
	if w.Channels[0].ID.String() != "987654321098765432" {
		t.Errorf("expected 987654321098765432, got %v", w.Channels[0].ID)
	}

	// Test invalid JSON
	if err := w.UnmarshalJSON([]byte(`{`)); err == nil {
		t.Error("expected error for invalid JSON")
	}
	if err := json.Unmarshal([]byte(`{"id": "invalid"}`), &w); err == nil {
		t.Error("expected error for invalid id")
	}

	// Test WidgetChannel separately
	var wc WidgetChannel
	if err := wc.UnmarshalJSON([]byte(`{`)); err == nil {
		t.Error("expected error for invalid JSON")
	}
	if err := json.Unmarshal([]byte(`{"id": "invalid"}`), &wc); err == nil {
		t.Error("expected error for invalid id")
	}

	// Test WidgetSettings
	wsData := []byte(`{
		"enabled": true,
		"channel_id": "111111111111111111"
	}`)
	var ws WidgetSettings
	if err := json.Unmarshal(wsData, &ws); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ws.Enabled {
		t.Error("expected enabled to be true")
	}
	if ws.ChannelID == nil || ws.ChannelID.String() != "111111111111111111" {
		t.Errorf("expected 111111111111111111, got %v", ws.ChannelID)
	}

	// Test invalid WidgetSettings
	if err := ws.UnmarshalJSON([]byte(`{`)); err == nil {
		t.Error("expected error for invalid JSON")
	}
	if err := json.Unmarshal([]byte(`{"channel_id": "invalid"}`), &ws); err == nil {
		t.Error("expected error for invalid channel_id")
	}
}
