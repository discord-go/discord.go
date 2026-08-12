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

// ExecuteWebhookParamsBuilder provides a fluent API for constructing
// ExecuteWebhookParams. It mirrors the pattern of MessageSendBuilder.
type ExecuteWebhookParamsBuilder struct {
	params ExecuteWebhookParams
}

// NewExecuteWebhookParamsBuilder creates a new ExecuteWebhookParamsBuilder.
func NewExecuteWebhookParamsBuilder() *ExecuteWebhookParamsBuilder {
	return &ExecuteWebhookParamsBuilder{}
}

// SetContent sets the message content.
func (b *ExecuteWebhookParamsBuilder) SetContent(content string) *ExecuteWebhookParamsBuilder {
	b.params.Content = content
	return b
}

// SetUsername overrides the default webhook username for this execution.
func (b *ExecuteWebhookParamsBuilder) SetUsername(username string) *ExecuteWebhookParamsBuilder {
	b.params.Username = username
	return b
}

// SetAvatarURL overrides the default webhook avatar for this execution.
func (b *ExecuteWebhookParamsBuilder) SetAvatarURL(avatarURL string) *ExecuteWebhookParamsBuilder {
	b.params.AvatarURL = avatarURL
	return b
}

// SetTTS sets whether the message is text-to-speech.
func (b *ExecuteWebhookParamsBuilder) SetTTS(tts bool) *ExecuteWebhookParamsBuilder {
	b.params.TTS = tts
	return b
}

// AddEmbed adds an embed to the message.
func (b *ExecuteWebhookParamsBuilder) AddEmbed(embed messages.Embed) *ExecuteWebhookParamsBuilder {
	b.params.Embeds = append(b.params.Embeds, embed)
	return b
}

// SetEmbeds replaces the embeds on the message.
func (b *ExecuteWebhookParamsBuilder) SetEmbeds(embeds []messages.Embed) *ExecuteWebhookParamsBuilder {
	b.params.Embeds = embeds
	return b
}

// AddComponent adds a component (action row, button, etc.) to the message.
func (b *ExecuteWebhookParamsBuilder) AddComponent(component components.Component) *ExecuteWebhookParamsBuilder {
	b.params.Components = append(b.params.Components, component)
	return b
}

// SetComponents replaces the components on the message.
func (b *ExecuteWebhookParamsBuilder) SetComponents(components []components.Component) *ExecuteWebhookParamsBuilder {
	b.params.Components = components
	return b
}

// SetFlags sets the message flags (e.g. ephemeral).
func (b *ExecuteWebhookParamsBuilder) SetFlags(flags int) *ExecuteWebhookParamsBuilder {
	b.params.Flags = flags
	return b
}

// AddAttachment adds an attachment metadata entry.
func (b *ExecuteWebhookParamsBuilder) AddAttachment(attachment messages.Attachment) *ExecuteWebhookParamsBuilder {
	b.params.Attachments = append(b.params.Attachments, attachment)
	return b
}

// SetThreadName sets the name of the thread to create in a forum channel.
func (b *ExecuteWebhookParamsBuilder) SetThreadName(threadName string) *ExecuteWebhookParamsBuilder {
	b.params.ThreadName = threadName
	return b
}

// AddAppliedTag adds a tag to the message in a forum channel.
func (b *ExecuteWebhookParamsBuilder) AddAppliedTag(tagID string) *ExecuteWebhookParamsBuilder {
	b.params.AppliedTags = append(b.params.AppliedTags, tagID)
	return b
}

// SetPoll sets the poll on the message.
func (b *ExecuteWebhookParamsBuilder) SetPoll(poll *messages.Poll) *ExecuteWebhookParamsBuilder {
	b.params.Poll = poll
	return b
}

// SetAllowedMentions sets the allowed mention configuration.
func (b *ExecuteWebhookParamsBuilder) SetAllowedMentions(mentions *messages.AllowedMentions) *ExecuteWebhookParamsBuilder {
	b.params.AllowedMentions = mentions
	return b
}

// Build returns the constructed ExecuteWebhookParams.
func (b *ExecuteWebhookParamsBuilder) Build() ExecuteWebhookParams {
	return b.params
}
