package messages

import (
	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/snowflake"
)

// MessageSend represents the payload for sending a message.
type MessageSend struct {
	Content           string                 `json:"content,omitempty"`
	Nonce             string                 `json:"nonce,omitempty"`
	TTS               bool                   `json:"tts,omitempty"`
	Embeds            []Embed                `json:"embeds,omitempty"`
	AllowedMentions   *AllowedMentions       `json:"allowed_mentions,omitempty"`
	MessageReference  *MessageReference      `json:"message_reference,omitempty"`
	Components        []components.Component `json:"components,omitempty"`
	StickerIDs        []snowflake.ID         `json:"sticker_ids,omitempty"`
	PayloadJSON       string                 `json:"payload_json,omitempty"`
	Attachments       []AttachmentSend       `json:"attachments,omitempty"`
	Flags             int                    `json:"flags,omitempty"`
	EnforceNonce      bool                   `json:"enforce_nonce,omitempty"`
	Poll              *Poll                  `json:"poll,omitempty"`
	SharedClientTheme *string                `json:"shared_client_theme,omitempty"`
}

// MessageFlags for message payloads
const (
	FlagCrossposted           = 1 << 0
	FlagIsCrosspost           = 1 << 1
	FlagSuppressEmbeds        = 1 << 2
	FlagSourceMessageDeleted  = 1 << 3
	FlagUrgent                = 1 << 4
	FlagHasThread             = 1 << 5
	FlagEphemeral             = 1 << 6
	FlagLoading               = 1 << 7
	FlagFailedToMentionRoles  = 1 << 8
	FlagSuppressNotifications = 1 << 12
	FlagIsComponentsV2        = 1 << 15
)

// AttachmentSend represents an attachment to be sent.
type AttachmentSend struct {
	ID          string `json:"id"`
	Filename    string `json:"filename,omitempty"`
	Description string `json:"description,omitempty"`
	Title       string `json:"title,omitempty"`
}

// MessageSendBuilder provides a fluent interface for constructing
// MessageSend payloads, matching the EmbedBuilder and ActionRowBuilder
// patterns.
type MessageSendBuilder struct {
	send MessageSend
}

// NewMessageSendBuilder creates a new MessageSendBuilder.
func NewMessageSendBuilder() *MessageSendBuilder {
	return &MessageSendBuilder{
		send: MessageSend{},
	}
}

// SetContent sets the text content of the message.
func (b *MessageSendBuilder) SetContent(content string) *MessageSendBuilder {
	b.send.Content = content
	return b
}

// SetTTS sets whether the message is text-to-speech.
func (b *MessageSendBuilder) SetTTS(tts bool) *MessageSendBuilder {
	b.send.TTS = tts
	return b
}

// SetEmbeds replaces the embeds on the message.
func (b *MessageSendBuilder) SetEmbeds(embeds ...Embed) *MessageSendBuilder {
	b.send.Embeds = embeds
	return b
}

// AddEmbed appends a single embed to the message.
func (b *MessageSendBuilder) AddEmbed(embed Embed) *MessageSendBuilder {
	b.send.Embeds = append(b.send.Embeds, embed)
	return b
}

// SetComponents replaces the components on the message.
func (b *MessageSendBuilder) SetComponents(comps ...components.Component) *MessageSendBuilder {
	b.send.Components = comps
	return b
}

// AddComponent appends a single component to the message.
func (b *MessageSendBuilder) AddComponent(comp components.Component) *MessageSendBuilder {
	b.send.Components = append(b.send.Components, comp)
	return b
}

// SetFlags sets the message flags (e.g. FlagEphemeral, FlagIsComponentsV2).
func (b *MessageSendBuilder) SetFlags(flags int) *MessageSendBuilder {
	b.send.Flags = flags
	return b
}

// AddFlag adds a single flag to the message flags.
func (b *MessageSendBuilder) AddFlag(flag int) *MessageSendBuilder {
	b.send.Flags |= flag
	return b
}

// SetAllowedMentions sets the allowed mentions for the message.
func (b *MessageSendBuilder) SetAllowedMentions(am *AllowedMentions) *MessageSendBuilder {
	b.send.AllowedMentions = am
	return b
}

// SetMessageReference sets the message reference for replies.
func (b *MessageSendBuilder) SetMessageReference(ref *MessageReference) *MessageSendBuilder {
	b.send.MessageReference = ref
	return b
}

// SetNonce sets the nonce for message deduplication.
func (b *MessageSendBuilder) SetNonce(nonce string) *MessageSendBuilder {
	b.send.Nonce = nonce
	return b
}

// SetEnforceNonce sets whether the nonce should be enforced.
func (b *MessageSendBuilder) SetEnforceNonce(enforce bool) *MessageSendBuilder {
	b.send.EnforceNonce = enforce
	return b
}

// SetPoll sets the poll on the message.
func (b *MessageSendBuilder) SetPoll(poll *Poll) *MessageSendBuilder {
	b.send.Poll = poll
	return b
}

// AddAttachment appends an attachment to the message.
func (b *MessageSendBuilder) AddAttachment(attachment AttachmentSend) *MessageSendBuilder {
	b.send.Attachments = append(b.send.Attachments, attachment)
	return b
}

// SetStickerIDs sets the sticker IDs for the message.
func (b *MessageSendBuilder) SetStickerIDs(ids ...snowflake.ID) *MessageSendBuilder {
	b.send.StickerIDs = ids
	return b
}

// Build returns the constructed MessageSend.
func (b *MessageSendBuilder) Build() MessageSend {
	return b.send
}
