package bot

import (
	"errors"

	"github.com/discord-go/discord.go/snowflake"
)

// JoinVoiceChannel requests that the bot join a guild voice channel. The
// gateway voice-server update is then available through OnEvent("VOICE_SERVER_UPDATE", ...)
// for applications that create a full voice.Client session.
func (b *Bot) JoinVoiceChannel(guildID, channelID snowflake.ID, selfMute, selfDeaf bool) error {
	b.stateMu.RLock()
	client := b.gwClient
	manager := b.shardManager
	b.stateMu.RUnlock()
	if manager != nil {
		return manager.JoinVoiceChannel(guildID, channelID, selfMute, selfDeaf)
	}
	if client == nil {
		return errors.New("bot: gateway is not running")
	}
	return client.JoinVoiceChannel(guildID, channelID, selfMute, selfDeaf)
}

// LeaveVoiceChannel leaves the guild voice channel.
func (b *Bot) LeaveVoiceChannel(guildID snowflake.ID) error {
	return b.JoinVoiceChannel(guildID, 0, false, false)
}
