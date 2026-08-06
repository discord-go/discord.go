package voice

import (
	"encoding/json"
	"testing"

	"github.com/discord-go/discord.go/snowflake"
)

func TestVoiceServerUpdate_JSON(t *testing.T) {
	raw := `{
		"token": "my_token",
		"guild_id": "41771983423143937",
		"endpoint": "smart.loyal.discord.gg"
	}`

	var update VoiceServerUpdate
	if err := json.Unmarshal([]byte(raw), &update); err != nil {
		t.Fatalf("failed to unmarshal VoiceServerUpdate: %v", err)
	}

	if update.Token != "my_token" {
		t.Errorf("expected Token %q, got %q", "my_token", update.Token)
	}
	if update.GuildID != snowflake.ID(41771983423143937) {
		t.Errorf("expected GuildID %v, got %v", snowflake.ID(41771983423143937), update.GuildID)
	}
	if update.Endpoint == nil || *update.Endpoint != "smart.loyal.discord.gg" {
		t.Errorf("unexpected Endpoint: %v", update.Endpoint)
	}
}

func TestVoiceServerUpdate_NilEndpoint(t *testing.T) {
	raw := `{
		"token": "a_token",
		"guild_id": "123456789",
		"endpoint": null
	}`

	var update VoiceServerUpdate
	if err := json.Unmarshal([]byte(raw), &update); err != nil {
		t.Fatalf("failed to unmarshal VoiceServerUpdate: %v", err)
	}

	if update.Endpoint != nil {
		t.Errorf("expected Endpoint to be nil, got %v", *update.Endpoint)
	}
}

func TestVoiceServerUpdate_Marshal(t *testing.T) {
	endpoint := "us-west.discord.gg"
	update := VoiceServerUpdate{
		Token:    "token123",
		GuildID:  snowflake.ID(999),
		Endpoint: &endpoint,
	}

	data, err := json.Marshal(update)
	if err != nil {
		t.Fatalf("failed to marshal VoiceServerUpdate: %v", err)
	}

	var decoded VoiceServerUpdate
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal marshaled VoiceServerUpdate: %v", err)
	}

	if decoded.Token != update.Token {
		t.Errorf("expected Token %q, got %q", update.Token, decoded.Token)
	}
	if decoded.GuildID != update.GuildID {
		t.Errorf("expected GuildID %v, got %v", update.GuildID, decoded.GuildID)
	}
	if decoded.Endpoint == nil || *decoded.Endpoint != endpoint {
		t.Errorf("expected Endpoint %q, got %v", endpoint, decoded.Endpoint)
	}
}
