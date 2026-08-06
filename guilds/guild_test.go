package guilds

import (
	"encoding/json"
	"testing"
)

func TestGuildUnmarshalJSON(t *testing.T) {
	data := []byte(`{
		"id": "123456789012345678",
		"name": "Discord API",
		"owner_id": "987654321098765432",
		"afk_channel_id": "111111111111111111",
		"widget_channel_id": null,
		"application_id": "",
		"system_channel_id": "222222222222222222",
		"rules_channel_id": "333333333333333333",
		"public_updates_channel_id": "444444444444444444",
		"safety_alerts_channel_id": "555555555555555555"
	}`)

	var g Guild
	if err := json.Unmarshal(data, &g); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if g.ID.String() != "123456789012345678" {
		t.Errorf("expected 123456789012345678, got %v", g.ID)
	}
	if g.OwnerID.String() != "987654321098765432" {
		t.Errorf("expected 987654321098765432, got %v", g.OwnerID)
	}
	if g.AFKChannelID == nil || g.AFKChannelID.String() != "111111111111111111" {
		t.Errorf("expected 111111111111111111, got %v", g.AFKChannelID)
	}
	if g.WidgetChannelID != nil {
		t.Errorf("expected WidgetChannelID to be nil, got %v", g.WidgetChannelID)
	}
	if g.ApplicationID != nil {
		t.Errorf("expected ApplicationID to be nil, got %v", g.ApplicationID)
	}

	// Test invalid JSON syntax
	if err := g.UnmarshalJSON([]byte(`{`)); err == nil {
		t.Error("expected error for invalid JSON syntax")
	}

	invalidTests := []string{
		`{"id": "invalid"}`,
		`{"owner_id": "invalid"}`,
		`{"afk_channel_id": "invalid"}`,
		`{"widget_channel_id": "invalid"}`,
		`{"application_id": "invalid"}`,
		`{"system_channel_id": "invalid"}`,
		`{"rules_channel_id": "invalid"}`,
		`{"public_updates_channel_id": "invalid"}`,
		`{"safety_alerts_channel_id": "invalid"}`,
	}

	for _, tc := range invalidTests {
		if err := json.Unmarshal([]byte(tc), &g); err == nil {
			t.Errorf("expected error for invalid JSON: %s", tc)
		}
	}
}

func TestUnavailableGuild(t *testing.T) {
	data := []byte(`{
		"id": "123456789012345678",
		"unavailable": true
	}`)

	var ug UnavailableGuild
	if err := json.Unmarshal(data, &ug); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ug.ID.String() != "123456789012345678" {
		t.Errorf("expected 123456789012345678, got %v", ug.ID)
	}
	if !ug.Unavailable {
		t.Error("expected true")
	}
}

func TestWelcomeScreenChannelUnmarshalJSON(t *testing.T) {
	data := []byte(`{
		"channel_id": "123456789012345678",
		"description": "Welcome!",
		"emoji_id": "876543210987654321",
		"emoji_name": "wave"
	}`)

	var wsc WelcomeScreenChannel
	if err := json.Unmarshal(data, &wsc); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if wsc.ChannelID.String() != "123456789012345678" {
		t.Errorf("expected 123456789012345678, got %v", wsc.ChannelID)
	}
	if wsc.EmojiID == nil || wsc.EmojiID.String() != "876543210987654321" {
		t.Errorf("expected 876543210987654321, got %v", wsc.EmojiID)
	}

	// test null/empty
	data2 := []byte(`{
		"channel_id": "123456789012345678",
		"description": "Welcome 2!",
		"emoji_id": null
	}`)
	var wsc2 WelcomeScreenChannel
	if err := json.Unmarshal(data2, &wsc2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wsc2.EmojiID != nil {
		t.Errorf("expected nil emoji_id, got %v", wsc2.EmojiID)
	}

	// Test invalid JSON syntax
	if err := wsc.UnmarshalJSON([]byte(`{`)); err == nil {
		t.Error("expected error for invalid JSON syntax")
	}

	if err := json.Unmarshal([]byte(`{"channel_id": "invalid"}`), &wsc); err == nil {
		t.Error("expected error for invalid channel_id")
	}
	if err := json.Unmarshal([]byte(`{"emoji_id": "invalid"}`), &wsc); err == nil {
		t.Error("expected error for invalid emoji_id")
	}
}

func TestWelcomeScreen(t *testing.T) {
	data := []byte(`{
		"description": "Welcome to the server",
		"welcome_channels": [
			{
				"channel_id": "123456789012345678",
				"description": "General chatting",
				"emoji_id": null,
				"emoji_name": "speech_balloon"
			}
		]
	}`)

	var ws WelcomeScreen
	if err := json.Unmarshal(data, &ws); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ws.Description == nil || *ws.Description != "Welcome to the server" {
		t.Errorf("unexpected description: %v", ws.Description)
	}
	if len(ws.WelcomeChannels) != 1 {
		t.Fatalf("expected 1 channel, got %v", len(ws.WelcomeChannels))
	}
}
