package bot

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/voice"
)

func mustSnowflake(t *testing.T, s string) snowflake.ID {
	t.Helper()
	id, err := snowflake.Parse(s)
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return id
}

func TestVoiceTrackerApplyAndQuery(t *testing.T) {
	tr := newVoiceTracker()
	guild := mustSnowflake(t, "123")
	user := mustSnowflake(t, "42")
	channel := mustSnowflake(t, "777")

	tr.apply(voice.VoiceState{GuildID: guild, UserID: user, ChannelID: channel, SessionID: "s1"})

	if got := tr.CountInChannel(channel); got != 1 {
		t.Fatalf("CountInChannel = %d, want 1", got)
	}
	states := tr.VoiceStatesInChannel(channel)
	if len(states) != 1 || states[0].UserID != user {
		t.Fatalf("VoiceStatesInChannel = %+v, want one state for user 42", states)
	}
	if got := tr.VoiceChannelOf(guild, user); got != channel {
		t.Fatalf("VoiceChannelOf = %s, want %s", got, channel)
	}
	state, ok := tr.VoiceStateOf(guild, user)
	if !ok || state.SessionID != "s1" {
		t.Fatalf("VoiceStateOf = (%+v, %v), want session s1", state, ok)
	}
}

// TestVoiceTrackerMoveAndLeave verifies that moving between channels updates
// the entry and that a nil/zero channel removes it.
func TestVoiceTrackerMoveAndLeave(t *testing.T) {
	tr := newVoiceTracker()
	guild := mustSnowflake(t, "123")
	user := mustSnowflake(t, "42")
	first := mustSnowflake(t, "777")
	second := mustSnowflake(t, "888")

	tr.apply(voice.VoiceState{GuildID: guild, UserID: user, ChannelID: first})
	tr.apply(voice.VoiceState{GuildID: guild, UserID: user, ChannelID: second})

	if got := tr.CountInChannel(first); got != 0 {
		t.Fatalf("old channel count = %d, want 0 after move", got)
	}
	if got := tr.CountInChannel(second); got != 1 {
		t.Fatalf("new channel count = %d, want 1 after move", got)
	}

	tr.apply(voice.VoiceState{GuildID: guild, UserID: user, ChannelID: snowflake.ID(0)})
	if got := tr.VoiceChannelOf(guild, user); !got.IsZero() {
		t.Fatalf("VoiceChannelOf after disconnect = %s, want zero", got)
	}

	tr.apply(voice.VoiceState{GuildID: guild, UserID: user, ChannelID: snowflake.ID(0)})
	if _, ok := tr.VoiceStateOf(guild, user); ok {
		t.Fatalf("zero channel should remove the entry")
	}
}

// TestVoiceTrackerDropGuild verifies GUILD_DELETE cleanup.
func TestVoiceTrackerDropGuild(t *testing.T) {
	tr := newVoiceTracker()
	guild := mustSnowflake(t, "123")
	user := mustSnowflake(t, "42")
	channel := mustSnowflake(t, "777")
	tr.apply(voice.VoiceState{GuildID: guild, UserID: user, ChannelID: channel})

	tr.dropGuild(guild)
	if got := tr.CountInChannel(channel); got != 0 {
		t.Fatalf("count after dropGuild = %d, want 0", got)
	}
}

// TestBotVoiceStateDispatch verifies the dispatch path feeds the tracker and
// that the typed handler receives the decoded state.
func TestBotVoiceStateDispatch(t *testing.T) {
	b := New("test-token")
	b.voice = newVoiceTracker()

	var wg sync.WaitGroup
	var gotSession string
	wg.Add(1)
	unsub := b.OnVoiceStateUpdateTyped(func(ctx *VoiceStateContext) {
		defer wg.Done()
		gotSession = ctx.State().SessionID
	})
	defer unsub()

	payload := []byte(`{"t":"VOICE_STATE_UPDATE","s":1,"op":0,"d":{"guild_id":"123","user_id":"42","session_id":"s9","channel_id":"777","deaf":false,"mute":false,"self_deaf":false,"self_mute":false}}`)
	b.handleRawDispatch(payload)
	wg.Wait()

	if gotSession != "s9" {
		t.Fatalf("typed handler session = %q, want s9", gotSession)
	}
	if got := b.VoiceChannelOf(mustSnowflake(t, "123"), mustSnowflake(t, "42")); got.String() != "777" {
		t.Fatalf("VoiceChannelOf = %s, want 777", got)
	}
	if got := b.CountInChannel(mustSnowflake(t, "777")); got != 1 {
		t.Fatalf("CountInChannel = %d, want 1", got)
	}
}

// TestBotVoiceStateDispatchDisconnect verifies that a disconnect payload
// (channel_id null) removes the user from the tracker.
func TestBotVoiceStateDispatchDisconnect(t *testing.T) {
	b := New("test-token")
	b.voice = newVoiceTracker()

	guild := mustSnowflake(t, "123")
	user := mustSnowflake(t, "42")
	channel := mustSnowflake(t, "777")

	var raw voice.VoiceState
	join := []byte(`{"guild_id":"123","user_id":"42","session_id":"s1","channel_id":"777"}`)
	if err := json.Unmarshal(join, &raw); err != nil {
		t.Fatalf("unmarshal join: %v", err)
	}
	b.voice.apply(raw)
	if b.CountInChannel(channel) != 1 {
		t.Fatalf("precondition failed: user not tracked")
	}

	var wg sync.WaitGroup
	wg.Add(1)
	unsub := b.OnVoiceStateUpdateTyped(func(ctx *VoiceStateContext) { wg.Done() })
	defer unsub()

	b.handleRawDispatch([]byte(`{"t":"VOICE_STATE_UPDATE","s":2,"op":0,"d":{"guild_id":"123","user_id":"42","session_id":"s1","channel_id":null}}`))
	wg.Wait()

	if got := b.VoiceChannelOf(guild, user); !got.IsZero() {
		t.Fatalf("VoiceChannelOf after disconnect = %s, want zero", got)
	}
}
