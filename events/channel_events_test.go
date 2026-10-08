package events

import (
	"encoding/json"
	"testing"

	"github.com/discord-go/discord.go/snowflake"
)

func TestChannelInfo(t *testing.T) {
	raw := `{"guild_id":"613425648685547541","channels":[
		{"id":"100","status":"standup","voice_start_time":1756000000},
		{"id":"200","status":null,"voice_start_time":null}
	]}`
	var info ChannelInfo
	if err := json.Unmarshal([]byte(raw), &info); err != nil {
		t.Fatal(err)
	}
	if info.GuildID != snowflake.ID(613425648685547541) {
		t.Errorf("guild_id = %s", info.GuildID)
	}
	if len(info.Channels) != 2 {
		t.Fatalf("len = %d", len(info.Channels))
	}
	first := info.Channels[0]
	if first.ID != 100 || first.Status == nil || *first.Status != "standup" {
		t.Errorf("first = %+v", first)
	}
	if first.VoiceStartTime == nil || *first.VoiceStartTime != 1756000000 {
		t.Errorf("voice_start_time = %v", first.VoiceStartTime)
	}
	second := info.Channels[1]
	if second.Status != nil || second.VoiceStartTime != nil {
		t.Errorf("second = %+v", second)
	}
}

func TestVoiceChannelStatusUpdate(t *testing.T) {
	set := VoiceChannelStatusUpdate{}
	if err := json.Unmarshal([]byte(`{"id":"100","guild_id":"200","status":"in a meeting"}`), &set); err != nil {
		t.Fatal(err)
	}
	if set.ID != 100 || set.GuildID != 200 {
		t.Errorf("ids = %+v", set)
	}
	if set.Status == nil || *set.Status != "in a meeting" {
		t.Errorf("status = %v", set.Status)
	}

	var cleared VoiceChannelStatusUpdate
	if err := json.Unmarshal([]byte(`{"id":"100","guild_id":"200","status":null}`), &cleared); err != nil {
		t.Fatal(err)
	}
	if cleared.Status != nil {
		t.Errorf("status = %v, want nil", cleared.Status)
	}
}

func TestVoiceChannelStartTimeUpdate(t *testing.T) {
	var started VoiceChannelStartTimeUpdate
	if err := json.Unmarshal([]byte(`{"id":"100","guild_id":"200","voice_start_time":1756000000}`), &started); err != nil {
		t.Fatal(err)
	}
	if started.ID != 100 || started.GuildID != 200 {
		t.Errorf("ids = %+v", started)
	}
	if started.VoiceStartTime == nil || *started.VoiceStartTime != 1756000000 {
		t.Errorf("voice_start_time = %v", started.VoiceStartTime)
	}

	var cleared VoiceChannelStartTimeUpdate
	if err := json.Unmarshal([]byte(`{"id":"100","guild_id":"200","voice_start_time":null}`), &cleared); err != nil {
		t.Fatal(err)
	}
	if cleared.VoiceStartTime != nil {
		t.Errorf("voice_start_time = %v, want nil", cleared.VoiceStartTime)
	}
}
