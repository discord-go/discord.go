package rest

import (
	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/snowflake"
)

type CreateWebhookParams struct {
	Name      string       `json:"name"`
	ChannelID snowflake.ID `json:"-"`
	Avatar    *string      `json:"avatar,omitempty"`
	Reason    string       `json:"-"`
}

// ModifyWebhookParams contains the parameters for modifying a webhook.
type ModifyWebhookParams struct {
	Name      string       `json:"name,omitempty"`
	Avatar    *string      `json:"avatar,omitempty"`
	ChannelID snowflake.ID `json:"channel_id,string,omitempty"`
}

// ExecuteWebhookParams contains the parameters for executing a webhook.
type ExecuteWebhookParams struct {
	Content         string                    `json:"content,omitempty"`
	Username        string                    `json:"username,omitempty"`
	AvatarURL       string                    `json:"avatar_url,omitempty"`
	TTS             bool                      `json:"tts,omitempty"`
	Embeds          []messages.Embed          `json:"embeds,omitempty"`
	Components      []components.Component    `json:"components,omitempty"`
	Flags           int                       `json:"flags,omitempty"`
	Attachments     []messages.Attachment     `json:"attachments,omitempty"`
	ThreadName      string                    `json:"thread_name,omitempty"`
	AppliedTags     []string                  `json:"applied_tags,omitempty"`
	Poll            *messages.Poll            `json:"poll,omitempty"`
	AllowedMentions *messages.AllowedMentions `json:"allowed_mentions,omitempty"`
}
