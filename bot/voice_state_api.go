package bot

import (
	"github.com/discord-go/discord.go/events"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/voice"
)

// VoiceStateContext wraps a VOICE_STATE_UPDATE event with the decoded state.
type VoiceStateContext struct {
	BaseContext
	*events.VoiceStateUpdate
}

// State returns the decoded voice state.
func (v *VoiceStateContext) State() voice.VoiceState {
	if v.VoiceStateUpdate == nil {
		return voice.VoiceState{}
	}
	return v.VoiceStateUpdate.VoiceState
}

// VoiceStateUpdateHandler is called on VOICE_STATE_UPDATE events.
type VoiceStateUpdateHandler func(ctx *VoiceStateContext)

// OnVoiceStateUpdateTyped registers a typed VOICE_STATE_UPDATE handler. The
// returned function unsubscribes the handler and is safe to call more than
// once. The bot's built-in voice tracker is updated before handlers run, so
// VoiceChannelOf and VoiceStatesInChannel reflect the state that triggered
// the call.
func (b *Bot) OnVoiceStateUpdateTyped(handler VoiceStateUpdateHandler) func() {
	if handler == nil {
		return func() {}
	}
	return b.OnEvent("VOICE_STATE_UPDATE", func(ctx *EventContext) {
		var update events.VoiceStateUpdate
		if err := ctx.Decode(&update); err != nil {
			b.reportError(err)
			return
		}
		handler(&VoiceStateContext{BaseContext: ctx.BaseContext, VoiceStateUpdate: &update})
	})
}

// VoiceStatesInChannel returns the voice states of every user connected to
// the given channel across all guilds, in unspecified order. States come
// from the tracker fed by VOICE_STATE_UPDATE; it is empty until the first
// event for a user arrives and after resends during reconnects.
func (b *Bot) VoiceStatesInChannel(channelID snowflake.ID) []voice.VoiceState {
	if b.voice == nil {
		return nil
	}
	return b.voice.VoiceStatesInChannel(channelID)
}

// VoiceStateOf returns the current voice state for a user in a guild.
func (b *Bot) VoiceStateOf(guildID, userID snowflake.ID) (voice.VoiceState, bool) {
	if b.voice == nil {
		return voice.VoiceState{}, false
	}
	return b.voice.VoiceStateOf(guildID, userID)
}

// VoiceChannelOf returns the channel a user is connected to in a guild, or
// the zero ID when they are not in voice.
func (b *Bot) VoiceChannelOf(guildID, userID snowflake.ID) snowflake.ID {
	if b.voice == nil {
		return 0
	}
	return b.voice.VoiceChannelOf(guildID, userID)
}

// CountInChannel returns how many users the tracker believes are connected
// to a channel.
func (b *Bot) CountInChannel(channelID snowflake.ID) int {
	if b.voice == nil {
		return 0
	}
	return b.voice.CountInChannel(channelID)
}
