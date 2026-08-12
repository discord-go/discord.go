package webhook

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

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

// AvatarURLOptions controls the CDN representation returned by AvatarURL.
type AvatarURLOptions struct {
	Extension string
	Size      int
}

// AvatarURL returns a Discord CDN URL for this webhook's avatar.
// If the webhook has no custom avatar, it returns the default Discord
// avatar derived from the webhook ID.
func (w Webhook) AvatarURL(options AvatarURLOptions) string {
	if w.Avatar == "" {
		index := (uint64(w.ID) >> 22) % 6
		return fmt.Sprintf("https://cdn.discordapp.com/embed/avatars/%d.png", index)
	}
	extension := strings.TrimPrefix(strings.ToLower(options.Extension), ".")
	if extension == "" {
		extension = "png"
	}
	if extension != "png" && extension != "jpg" && extension != "jpeg" && extension != "webp" && extension != "gif" {
		extension = "png"
	}
	result := fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.%s", w.ID, url.PathEscape(w.Avatar), extension)
	if options.Size > 0 {
		result += "?size=" + strconv.Itoa(options.Size)
	}
	return result
}

// IsIncoming returns true if the webhook is an incoming webhook (Type 1).
func (w Webhook) IsIncoming() bool { return w.Type == TypeIncoming }

// IsChannelFollower returns true if the webhook is a channel follower webhook (Type 2).
func (w Webhook) IsChannelFollower() bool { return w.Type == TypeChannelFollower }

// IsApplication returns true if the webhook is an application webhook (Type 3).
func (w Webhook) IsApplication() bool { return w.Type == TypeApplication }

// IsZero returns true if the webhook has no ID (uninitialized or zero value).
func (w Webhook) IsZero() bool { return w.ID == 0 }

// HasToken returns true if the webhook has a token, meaning it can be
// executed without bot authentication.
func (w Webhook) HasToken() bool { return w.Token != "" }

// ExecutionURL returns the webhook execution URL if the token is present.
// Returns empty string if the webhook ID or token is not set.
func (w Webhook) ExecutionURL() string {
	if w.ID == 0 || w.Token == "" {
		return ""
	}
	return fmt.Sprintf("https://discord.com/api/webhooks/%s/%s", w.ID, w.Token)
}

// String returns a human-readable representation of the webhook type.
func (t Type) String() string {
	switch t {
	case TypeIncoming:
		return "Incoming"
	case TypeChannelFollower:
		return "ChannelFollower"
	case TypeApplication:
		return "Application"
	default:
		return "Unknown"
	}
}
