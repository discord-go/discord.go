package voice

import (
	"time"

	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

// VoiceState represents a user's voice connection status, matching the
// Discord API voice state object.
type VoiceState struct {
	// GuildID is the guild this voice state is for, if applicable.
	GuildID *snowflake.ID `json:"guild_id,string,omitempty"`
	// ChannelID is the channel the user is connected to.
	ChannelID *snowflake.ID `json:"channel_id,string,omitempty"`
	// UserID is the ID of the user this voice state is for.
	UserID snowflake.ID `json:"user_id,string"`
	// Member is the guild member this voice state is for, if applicable.
	Member *users.Member `json:"member,omitempty"`
	// SessionID is the session ID for this voice state.
	SessionID string `json:"session_id"`
	// Deaf indicates whether the user is deafened by the server.
	Deaf bool `json:"deaf"`
	// Mute indicates whether the user is muted by the server.
	Mute bool `json:"mute"`
	// SelfDeaf indicates whether the user is locally deafened.
	SelfDeaf bool `json:"self_deaf"`
	// SelfMute indicates whether the user is locally muted.
	SelfMute bool `json:"self_mute"`
	// SelfStream indicates whether the user is streaming using "Go Live".
	SelfStream bool `json:"self_stream,omitempty"`
	// SelfVideo indicates whether the user's camera is enabled.
	SelfVideo bool `json:"self_video"`
	// Suppress indicates whether the user's permission to speak is denied.
	Suppress bool `json:"suppress"`
	// RequestToSpeakTimestamp is the time at which the user requested to speak.
	RequestToSpeakTimestamp *time.Time `json:"request_to_speak_timestamp,omitempty"`
}
