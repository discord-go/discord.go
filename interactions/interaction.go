package interactions

import (
	"encoding/json"

	"github.com/discord-go/discord.go/channels"
	"github.com/discord-go/discord.go/guilds"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

// InteractionType is the type of an Interaction
type InteractionType int

const (
	InteractionTypePing                           InteractionType = 1
	InteractionTypeApplicationCommand             InteractionType = 2
	InteractionTypeMessageComponent               InteractionType = 3
	InteractionTypeApplicationCommandAutocomplete InteractionType = 4
	InteractionTypeModalSubmit                    InteractionType = 5
)

// Interaction represents a Discord Interaction
type Interaction struct {
	ID            snowflake.ID    `json:"id,string"`
	ApplicationID snowflake.ID    `json:"application_id,string"`
	Type          InteractionType `json:"type"`
	Data          json.RawMessage `json:"data,omitempty"`
	Guild         *guilds.Guild   `json:"guild,omitempty"`
	// GuildID is the guild the interaction was used in; zero for DMs.
	GuildID snowflake.ID      `json:"guild_id,string,omitempty"`
	Channel *channels.Channel `json:"channel,omitempty"`
	// ChannelID is the channel the interaction was used in; zero when
	// absent.
	ChannelID      snowflake.ID      `json:"channel_id,string,omitempty"`
	Member         *users.Member     `json:"member,omitempty"`
	User           *users.User       `json:"user,omitempty"`
	Token          string            `json:"token"`
	Version        int               `json:"version"`
	Message        *messages.Message `json:"message,omitempty"`
	AppPermissions string            `json:"app_permissions,omitempty"`
	Locale         string            `json:"locale,omitempty"`
	GuildLocale    string            `json:"guild_locale,omitempty"`
}
