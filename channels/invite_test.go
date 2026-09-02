package channels

import (
	"encoding/json"
	"testing"
	"time"
)

// TestInviteUnmarshalGatewayPayload verifies that INVITE_CREATE and
// INVITE_DELETE gateway payloads decode the metadata fields invite trackers
// depend on: uses, max_uses, max_age, temporary, created_at, and the flat
// guild_id/channel_id/inviter_id snowflakes.
func TestInviteUnmarshalGatewayPayload(t *testing.T) {
	data := []byte(`{
		"channel_id": "555",
		"code": "aBcDeF",
		"created_at": "2026-09-02T12:00:00+00:00",
		"guild_id": "123",
		"inviter": {"id": "42", "username": "ada", "discriminator": "0"},
		"max_age": 604800,
		"max_uses": 25,
		"temporary": false,
		"uses": 7
	}`)

	var inv Invite
	if err := json.Unmarshal(data, &inv); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if inv.Code != "aBcDeF" {
		t.Errorf("Code = %q, want aBcDeF", inv.Code)
	}
	if inv.GuildID.String() != "123" {
		t.Errorf("GuildID = %s, want 123", inv.GuildID)
	}
	if inv.ChannelID.String() != "555" {
		t.Errorf("ChannelID = %s, want 555", inv.ChannelID)
	}
	if inv.Uses != 7 {
		t.Errorf("Uses = %d, want 7", inv.Uses)
	}
	if inv.MaxUses != 25 {
		t.Errorf("MaxUses = %d, want 25", inv.MaxUses)
	}
	if inv.MaxAge != 604800 {
		t.Errorf("MaxAge = %d, want 604800", inv.MaxAge)
	}
	if inv.Temporary {
		t.Errorf("Temporary = true, want false")
	}
	if inv.Inviter == nil || inv.Inviter.ID.String() != "42" {
		t.Errorf("Inviter = %+v, want user 42", inv.Inviter)
	}
	if inv.CreatedAt == nil {
		t.Fatalf("CreatedAt = nil, want parsed time")
	}
	if inv.CreatedAt.Year() != 2026 {
		t.Errorf("CreatedAt = %v, want 2026", inv.CreatedAt)
	}
}

// TestInviteUnmarshalExpiresAt verifies the existing ExpiresAt field still
// decodes alongside the new metadata.
func TestInviteUnmarshalExpiresAt(t *testing.T) {
	data := []byte(`{"code":"x","expires_at":"2030-01-01T00:00:00+00:00"}`)
	var inv Invite
	if err := json.Unmarshal(data, &inv); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if inv.ExpiresAt == nil {
		t.Fatalf("ExpiresAt = nil, want parsed time")
	}
	if inv.ExpiresAt.Before(time.Now()) {
		t.Errorf("ExpiresAt = %v, want future date", inv.ExpiresAt)
	}
}
