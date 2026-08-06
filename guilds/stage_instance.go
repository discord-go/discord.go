package guilds

import "github.com/discord-go/discord.go/snowflake"

// StageInstance represents a live Stage channel instance.
type StageInstance struct {
	ID                    snowflake.ID  `json:"id,string"`
	GuildID               snowflake.ID  `json:"guild_id,string"`
	ChannelID             snowflake.ID  `json:"channel_id,string"`
	Topic                 string        `json:"topic"`
	PrivacyLevel          int           `json:"privacy_level"`
	DiscoverableDisabled  bool          `json:"discoverable_disabled"`
	GuildScheduledEventID *snowflake.ID `json:"guild_scheduled_event_id,string,omitempty"`
}
