package webhook

import (
	"github.com/discord-go/discord.go/channels"
	"github.com/discord-go/discord.go/guilds"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

// Webhook represents a Discord webhook.
type Webhook struct {
	ID            snowflake.ID      `json:"id,string"`
	Type          Type              `json:"type"`
	GuildID       snowflake.ID      `json:"guild_id,string,omitempty"`
	ChannelID     snowflake.ID      `json:"channel_id,string,omitempty"`
	User          *users.User       `json:"user,omitempty"`
	Name          string            `json:"name,omitempty"`
	Avatar        string            `json:"avatar,omitempty"`
	Token         string            `json:"token,omitempty"`
	ApplicationID snowflake.ID      `json:"application_id,string,omitempty"`
	SourceGuild   *guilds.Guild     `json:"source_guild,omitempty"`
	SourceChannel *channels.Channel `json:"source_channel,omitempty"`
	URL           string            `json:"url,omitempty"`
}
