package events

import (
	"github.com/discord-go/discord.go/channels"
	"github.com/discord-go/discord.go/snowflake"
)

// ChannelCreate represents the CHANNEL_CREATE event.
type ChannelCreate struct {
	channels.Channel
}

// ChannelUpdate represents the CHANNEL_UPDATE event.
type ChannelUpdate struct {
	channels.Channel
}

// ChannelInfo represents the "Channel Info" dispatch Discord sends in answer
// to a gateway Request Channel Info (opcode 43). It carries the ephemeral
// per-channel values Discord does not put on the channel object.
type ChannelInfo struct {
	GuildID  snowflake.ID       `json:"guild_id,string"`
	Channels []ChannelInfoEntry `json:"channels"`
}

// ChannelInfoEntry is one channel's ephemeral data inside a Channel Info
// event. Only the requested fields are present.
type ChannelInfoEntry struct {
	ID             snowflake.ID `json:"id,string"`
	Status         *string      `json:"status,omitempty"`
	VoiceStartTime *int         `json:"voice_start_time,omitempty"`
}

// VoiceChannelStatusUpdate represents the VOICE_CHANNEL_STATUS_UPDATE event,
// sent when a voice channel's status changes. Status is null when the status
// was cleared.
type VoiceChannelStatusUpdate struct {
	ID      snowflake.ID `json:"id,string"`
	GuildID snowflake.ID `json:"guild_id,string"`
	Status  *string      `json:"status"`
}

// VoiceChannelStartTimeUpdate represents the VOICE_CHANNEL_START_TIME_UPDATE
// event, sent when a voice channel's session start time changes.
type VoiceChannelStartTimeUpdate struct {
	ID             snowflake.ID `json:"id,string"`
	GuildID        snowflake.ID `json:"guild_id,string"`
	VoiceStartTime *int         `json:"voice_start_time"`
}
