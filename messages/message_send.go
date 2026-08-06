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
