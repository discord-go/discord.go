package voice

import "github.com/discord-go/discord.go/snowflake"

// VoiceServerUpdate represents the payload for the VOICE_SERVER_UPDATE
// gateway event. It is sent when a guild's voice server is updated,
// indicating the client should connect to a new voice endpoint.
type VoiceServerUpdate struct {
	// Token is the voice connection token.
	Token string `json:"token"`
	// GuildID is the guild this voice server update is for.
	GuildID snowflake.ID `json:"guild_id,string"`
	// Endpoint is the voice server host (may be nil if the voice server
	// has been deallocated).
	Endpoint *string `json:"endpoint"`
}
