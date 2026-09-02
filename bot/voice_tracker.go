package bot

import (
	"sync"

	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/voice"
)

// voiceTracker maintains the guild voice-state table from VOICE_STATE_UPDATE
// dispatches. It answers "who is connected to channel X" without REST calls,
// which voice-driven applications (join-to-create, music bots, moderation)
// need on every event.
type voiceTracker struct {
	mu sync.RWMutex
	// states maps userID -> latest voice state per guild.
	states map[string]map[string]voice.VoiceState
}

func newVoiceTracker() *voiceTracker {
	return &voiceTracker{states: make(map[string]map[string]voice.VoiceState)}
}

// apply records a VOICE_STATE_UPDATE payload. A nil ChannelID or the zero ID
// means the user disconnected and removes their entry.
func (t *voiceTracker) apply(state voice.VoiceState) {
	if state.GuildID == nil || state.GuildID.IsZero() {
		return
	}
	guildKey := state.GuildID.String()
	userKey := state.UserID.String()

	t.mu.Lock()
	defer t.mu.Unlock()
	if t.states[guildKey] == nil {
		t.states[guildKey] = make(map[string]voice.VoiceState)
	}
	if state.ChannelID == nil || state.ChannelID.IsZero() {
		delete(t.states[guildKey], userKey)
		return
	}
	t.states[guildKey][userKey] = state
}

// dropGuild forgets every voice state for a guild (GUILD_DELETE).
func (t *voiceTracker) dropGuild(guildID snowflake.ID) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.states, guildID.String())
}

// VoiceStatesInChannel returns the voice states of every user connected to
// the given channel, in unspecified order.
func (t *voiceTracker) VoiceStatesInChannel(channelID snowflake.ID) []voice.VoiceState {
	key := channelID.String()
	t.mu.RLock()
	defer t.mu.RUnlock()
	var out []voice.VoiceState
	for _, guild := range t.states {
		for _, state := range guild {
			if state.ChannelID != nil && state.ChannelID.String() == key {
				out = append(out, state)
			}
		}
	}
	return out
}

// VoiceStateOf returns the current voice state for a user in a guild.
func (t *voiceTracker) VoiceStateOf(guildID, userID snowflake.ID) (voice.VoiceState, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	guild, ok := t.states[guildID.String()]
	if !ok {
		return voice.VoiceState{}, false
	}
	state, ok := guild[userID.String()]
	return state, ok
}

// VoiceChannelOf returns the channel a user is connected to in a guild, or
// the zero ID when they are not in voice.
func (t *voiceTracker) VoiceChannelOf(guildID, userID snowflake.ID) snowflake.ID {
	state, ok := t.VoiceStateOf(guildID, userID)
	if !ok || state.ChannelID == nil {
		return 0
	}
	return *state.ChannelID
}

// CountInChannel returns how many users are connected to a channel.
func (t *voiceTracker) CountInChannel(channelID snowflake.ID) int {
	return len(t.VoiceStatesInChannel(channelID))
}
