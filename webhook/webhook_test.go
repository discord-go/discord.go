package webhook

import (
	"encoding/json"
	"testing"

	"github.com/discord-go/discord.go/snowflake"
)

func TestWebhook_UnmarshalJSON(t *testing.T) {
	data := []byte(`{
		"id": "1234567890",
		"type": 1,
		"guild_id": "9876543210",
		"channel_id": "1122334455",
		"user": {
			"id": "5566778899",
			"username": "testuser"
		},
		"name": "Test Webhook",
		"avatar": "avatar_hash",
		"token": "secure_token",
		"application_id": "9988776655",
		"source_guild": {
			"id": "9876543210",
			"name": "Test Guild"
		},
		"source_channel": {
			"id": "1122334455",
			"name": "Test Channel"
		},
		"url": "https://discord.com/api/webhooks/1234567890/secure_token"
	}`)

	var w Webhook
	err := json.Unmarshal(data, &w)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if w.ID != snowflake.ID(1234567890) {
		t.Errorf("Expected ID 1234567890, got %d", w.ID)
	}
	if w.Type != WebhookTypeIncoming {
		t.Errorf("Expected Type 1, got %d", w.Type)
	}
	if w.GuildID != snowflake.ID(9876543210) {
		t.Errorf("Expected GuildID 9876543210, got %d", w.GuildID)
	}
	if w.ChannelID != snowflake.ID(1122334455) {
		t.Errorf("Expected ChannelID 1122334455, got %d", w.ChannelID)
	}
	if w.User == nil || w.User.ID != snowflake.ID(5566778899) {
		t.Errorf("Expected User with ID 5566778899")
	}
	if w.Name != "Test Webhook" {
		t.Errorf("Expected Name 'Test Webhook', got '%s'", w.Name)
	}
	if w.Avatar != "avatar_hash" {
		t.Errorf("Expected Avatar 'avatar_hash', got '%s'", w.Avatar)
	}
	if w.Token != "secure_token" {
		t.Errorf("Expected Token 'secure_token', got '%s'", w.Token)
	}
	if w.ApplicationID != snowflake.ID(9988776655) {
		t.Errorf("Expected ApplicationID 9988776655, got %d", w.ApplicationID)
	}
	if w.SourceGuild == nil || w.SourceGuild.ID != snowflake.ID(9876543210) {
		t.Errorf("Expected SourceGuild with ID 9876543210")
	}
	if w.SourceChannel == nil || w.SourceChannel.ID != snowflake.ID(1122334455) {
		t.Errorf("Expected SourceChannel with ID 1122334455")
	}
	if w.URL != "https://discord.com/api/webhooks/1234567890/secure_token" {
		t.Errorf("Expected URL 'https://discord.com/api/webhooks/1234567890/secure_token', got '%s'", w.URL)
	}
}
