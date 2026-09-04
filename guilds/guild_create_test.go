package guilds

import (
	"encoding/json"
	"testing"
)

// TestGuildUnmarshalGuildCreatePayload verifies that the extra arrays Discord
// sends only in the GUILD_CREATE gateway event (voice_states, presences,
// threads, channels, members) decode into the Guild struct. REST responses
// omit these keys, so the fields must simply stay nil there.
func TestGuildUnmarshalGuildCreatePayload(t *testing.T) {
	data := []byte(`{
		"id": "123456789012345678",
		"name": "Test Guild",
		"owner_id": "111111111111111111",
		"voice_states": [
			{"user_id": "42", "session_id": "abc", "channel_id": "999"}
		],
		"presences": [
			{"user": {"id": "42", "username": "ada"}, "status": "online"}
		],
		"channels": [
			{"id": "777", "type": 2, "name": "general"}
		],
		"threads": [
			{"id": "888", "type": 11, "name": "thread"}
		],
		"members": [
			{"user": {"id": "42", "username": "ada"}, "roles": []}
		]
	}`)

	var g Guild
	if err := json.Unmarshal(data, &g); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if g.Name != "Test Guild" {
		t.Errorf("Name = %q, want %q", g.Name, "Test Guild")
	}
	if g.OwnerID.String() != "111111111111111111" {
		t.Errorf("OwnerID = %s, want 111111111111111111", g.OwnerID)
	}

	if len(g.VoiceStates) != 1 {
		t.Fatalf("VoiceStates length = %d, want 1", len(g.VoiceStates))
	}
	vs := g.VoiceStates[0]
	if vs.UserID.String() != "42" {
		t.Errorf("VoiceStates[0].UserID = %s, want 42", vs.UserID)
	}
	if vs.SessionID != "abc" {
		t.Errorf("VoiceStates[0].SessionID = %q, want %q", vs.SessionID, "abc")
	}
	if vs.ChannelID.String() != "999" {
		t.Errorf("VoiceStates[0].ChannelID = %v, want 999", vs.ChannelID)
	}

	if len(g.Presences) != 1 {
		t.Fatalf("Presences length = %d, want 1", len(g.Presences))
	}
	if g.Presences[0].Status != "online" {
		t.Errorf("Presences[0].Status = %q, want %q", g.Presences[0].Status, "online")
	}
	if g.Presences[0].User.ID.String() != "42" {
		t.Errorf("Presences[0].User.ID = %s, want 42", g.Presences[0].User.ID)
	}

	if len(g.GuildChannels) != 1 {
		t.Fatalf("GuildChannels length = %d, want 1", len(g.GuildChannels))
	}
	if g.GuildChannels[0].ID.String() != "777" {
		t.Errorf("GuildChannels[0].ID = %s, want 777", g.GuildChannels[0].ID)
	}

	if len(g.Threads) != 1 {
		t.Fatalf("Threads length = %d, want 1", len(g.Threads))
	}
	if g.Threads[0].ID.String() != "888" {
		t.Errorf("Threads[0].ID = %s, want 888", g.Threads[0].ID)
	}

	if len(g.Members) != 1 {
		t.Fatalf("Members length = %d, want 1", len(g.Members))
	}
	if g.Members[0].User == nil || g.Members[0].User.Username != "ada" {
		t.Errorf("Members[0].User = %+v, want username ada", g.Members[0].User)
	}
}

// TestGuildUnmarshalRESTPayload verifies that a REST-shaped guild (no
// gateway-only arrays) unmarshals with the new fields left nil.
func TestGuildUnmarshalRESTPayload(t *testing.T) {
	data := []byte(`{"id": "123456789012345678", "name": "REST Guild"}`)
	var g Guild
	if err := json.Unmarshal(data, &g); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if g.VoiceStates != nil {
		t.Errorf("VoiceStates = %v, want nil", g.VoiceStates)
	}
	if g.Members != nil {
		t.Errorf("Members = %v, want nil", g.Members)
	}
	if g.GuildChannels != nil {
		t.Errorf("GuildChannels = %v, want nil", g.GuildChannels)
	}
}
