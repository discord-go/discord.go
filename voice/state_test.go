package voice

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/discord-go/discord.go/snowflake"
)

func TestVoiceState_JSON(t *testing.T) {
	raw := `{
		"guild_id": "41771983423143937",
		"channel_id": "157733188964188161",
		"user_id": "80351110224678912",
		"session_id": "90326bd25d71d39b9ef95b299e3872ff",
		"deaf": false,
		"mute": false,
		"self_deaf": false,
		"self_mute": true,
		"self_stream": true,
		"self_video": false,
		"suppress": false,
		"request_to_speak_timestamp": "2021-03-31T18:45:31.297561+00:00"
	}`

	var state VoiceState
	if err := json.Unmarshal([]byte(raw), &state); err != nil {
		t.Fatalf("failed to unmarshal VoiceState: %v", err)
	}

	if state.GuildID != snowflake.ID(41771983423143937) {
		t.Errorf("unexpected GuildID: %v", state.GuildID)
	}
	if state.ChannelID != snowflake.ID(157733188964188161) {
		t.Errorf("unexpected ChannelID: %v", state.ChannelID)
	}
	if state.UserID != snowflake.ID(80351110224678912) {
		t.Errorf("unexpected UserID: %v", state.UserID)
	}
	if state.SessionID != "90326bd25d71d39b9ef95b299e3872ff" {
		t.Errorf("unexpected SessionID: %q", state.SessionID)
	}
	if state.Deaf {
		t.Error("expected Deaf to be false")
	}
	if state.Mute {
		t.Error("expected Mute to be false")
	}
	if state.SelfDeaf {
		t.Error("expected SelfDeaf to be false")
	}
	if !state.SelfMute {
		t.Error("expected SelfMute to be true")
	}
	if !state.SelfStream {
		t.Error("expected SelfStream to be true")
	}
	if state.SelfVideo {
		t.Error("expected SelfVideo to be false")
	}
	if state.Suppress {
		t.Error("expected Suppress to be false")
	}
	if state.RequestToSpeakTimestamp == nil {
		t.Fatal("expected RequestToSpeakTimestamp to be non-nil")
	}
	if state.RequestToSpeakTimestamp.Year() != 2021 {
		t.Errorf("unexpected RequestToSpeakTimestamp year: %d", state.RequestToSpeakTimestamp.Year())
	}
}

func TestVoiceState_NilOptionalFields(t *testing.T) {
	raw := `{
		"user_id": "80351110224678912",
		"session_id": "abc123",
		"deaf": true,
		"mute": true,
		"self_deaf": true,
		"self_mute": false,
		"self_video": true,
		"suppress": true
	}`

	var state VoiceState
	if err := json.Unmarshal([]byte(raw), &state); err != nil {
		t.Fatalf("failed to unmarshal VoiceState: %v", err)
	}

	if !state.GuildID.IsZero() {
		t.Errorf("expected GuildID to be zero, got %v", state.GuildID)
	}
	if !state.ChannelID.IsZero() {
		t.Errorf("expected ChannelID to be zero, got %v", state.ChannelID)
	}
	if state.Member != nil {
		t.Error("expected Member to be nil")
	}
	if state.RequestToSpeakTimestamp != nil {
		t.Error("expected RequestToSpeakTimestamp to be nil")
	}
	if !state.Deaf {
		t.Error("expected Deaf to be true")
	}
	if !state.Mute {
		t.Error("expected Mute to be true")
	}
	if !state.SelfDeaf {
		t.Error("expected SelfDeaf to be true")
	}
	if state.SelfMute {
		t.Error("expected SelfMute to be false")
	}
	if !state.SelfVideo {
		t.Error("expected SelfVideo to be true")
	}
	if !state.Suppress {
		t.Error("expected Suppress to be true")
	}
}

func TestVoiceState_Marshal(t *testing.T) {
	ts := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	state := VoiceState{
		GuildID:                 snowflake.ID(123),
		ChannelID:               snowflake.ID(456),
		UserID:                  snowflake.ID(789),
		SessionID:               "session-abc",
		Deaf:                    true,
		Mute:                    true,
		SelfDeaf:                false,
		SelfMute:                false,
		SelfStream:              true,
		SelfVideo:               true,
		Suppress:                false,
		RequestToSpeakTimestamp: &ts,
	}

	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("failed to marshal VoiceState: %v", err)
	}

	var decoded VoiceState
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal marshaled VoiceState: %v", err)
	}

	if decoded.UserID != state.UserID {
		t.Errorf("expected UserID %v, got %v", state.UserID, decoded.UserID)
	}
	if decoded.SessionID != state.SessionID {
		t.Errorf("expected SessionID %q, got %q", state.SessionID, decoded.SessionID)
	}
}
