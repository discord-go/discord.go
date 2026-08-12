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
	if w.Type != TypeIncoming {
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

func TestWebhookHelpers(t *testing.T) {
	w := Webhook{
		ID:    snowflake.ID(1234567890),
		Type:  TypeIncoming,
		Name:  "alerts",
		Token: "secure_token",
	}

	if !w.IsIncoming() {
		t.Error("expected IsIncoming true")
	}
	if w.IsChannelFollower() {
		t.Error("expected IsChannelFollower false")
	}
	if w.IsApplication() {
		t.Error("expected IsApplication false")
	}
	if w.IsZero() {
		t.Error("expected IsZero false")
	}
	if !w.HasToken() {
		t.Error("expected HasToken true")
	}

	execURL := w.ExecutionURL()
	expected := "https://discord.com/api/webhooks/1234567890/secure_token"
	if execURL != expected {
		t.Errorf("ExecutionURL = %q, want %q", execURL, expected)
	}

	// No token -> empty URL
	w.Token = ""
	if w.ExecutionURL() != "" {
		t.Errorf("expected empty ExecutionURL without token, got %q", w.ExecutionURL())
	}

	// Zero webhook
	var zero Webhook
	if !zero.IsZero() {
		t.Error("expected IsZero true for zero-value Webhook")
	}
	if zero.HasToken() {
		t.Error("expected HasToken false for zero-value Webhook")
	}
}

func TestWebhookAvatarURL(t *testing.T) {
	// With avatar hash
	w := Webhook{
		ID:     snowflake.ID(1234567890),
		Avatar: "abc123",
	}
	url := w.AvatarURL(AvatarURLOptions{})
	if url != "https://cdn.discordapp.com/avatars/1234567890/abc123.png" {
		t.Errorf("AvatarURL = %q", url)
	}

	// With gif extension
	url = w.AvatarURL(AvatarURLOptions{Extension: "gif"})
	if url != "https://cdn.discordapp.com/avatars/1234567890/abc123.gif" {
		t.Errorf("AvatarURL gif = %q", url)
	}

	// Without avatar -> default avatar
	w.Avatar = ""
	url = w.AvatarURL(AvatarURLOptions{})
	if url == "" {
		t.Error("expected default avatar URL, got empty")
	}
}

func TestTypeString(t *testing.T) {
	tests := []struct {
		typ  Type
		want string
	}{
		{TypeIncoming, "Incoming"},
		{TypeChannelFollower, "ChannelFollower"},
		{TypeApplication, "Application"},
		{Type(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.typ.String(); got != tt.want {
			t.Errorf("Type(%d).String() = %q, want %q", tt.typ, got, tt.want)
		}
	}
}
