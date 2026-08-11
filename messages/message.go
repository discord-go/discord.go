package messages

import (
	"encoding/json"
	"time"

	"github.com/discord-go/discord.go/channels"
	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/emojis"
	"github.com/discord-go/discord.go/roles"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

// MessageType represents the type of message.
type MessageType int

const (
	MessageTypeDefault                                 MessageType = 0
	MessageTypeRecipientAdd                            MessageType = 1
	MessageTypeRecipientRemove                         MessageType = 2
	MessageTypeCall                                    MessageType = 3
	MessageTypeChannelNameChange                       MessageType = 4
	MessageTypeChannelIconChange                       MessageType = 5
	MessageTypeChannelPinnedMessage                    MessageType = 6
	MessageTypeUserJoin                                MessageType = 7
	MessageTypeGuildBoost                              MessageType = 8
	MessageTypeGuildBoostTier1                         MessageType = 9
	MessageTypeGuildBoostTier2                         MessageType = 10
	MessageTypeGuildBoostTier3                         MessageType = 11
	MessageTypeChannelFollowAdd                        MessageType = 12
	MessageTypeGuildDiscoveryDisqualified              MessageType = 14
	MessageTypeGuildDiscoveryRequalified               MessageType = 15
	MessageTypeGuildDiscoveryGracePeriodInitialWarning MessageType = 16
	MessageTypeGuildDiscoveryGracePeriodFinalWarning   MessageType = 17
	MessageTypeThreadCreated                           MessageType = 18
	MessageTypeReply                                   MessageType = 19
	MessageTypeChatInputCommand                        MessageType = 20
	MessageTypeThreadStarterMessage                    MessageType = 21
	MessageTypeGuildInviteReminder                     MessageType = 22
	MessageTypeContextMenuCommand                      MessageType = 23
	MessageTypeAutoModerationAction                    MessageType = 24
	MessageTypeRoleSubscriptionPurchase                MessageType = 25
	MessageTypeInteractionPremiumUpsell                MessageType = 26
	MessageTypeStageStart                              MessageType = 27
	MessageTypeStageEnd                                MessageType = 28
	MessageTypeStageSpeaker                            MessageType = 29
	MessageTypeStageTopic                              MessageType = 31
	MessageTypeGuildApplicationPremiumSubscription     MessageType = 32
)

// MessageActivity represents a message activity.
type MessageActivity struct {
	Type    int    `json:"type"`
	PartyID string `json:"party_id,omitempty"`
}

// MessageReference represents a message reference.
type MessageReference struct {
	MessageID       snowflake.ID `json:"message_id,string,omitempty"`
	ChannelID       snowflake.ID `json:"channel_id,string,omitempty"`
	GuildID         snowflake.ID `json:"guild_id,string,omitempty"`
	FailIfNotExists bool         `json:"fail_if_not_exists,omitempty"`
}

// Message represents a Discord message.
type Message struct {
	ID                   snowflake.ID           `json:"id,string"`
	ChannelID            snowflake.ID           `json:"channel_id,string"`
	Author               *users.User            `json:"author"`
	Content              string                 `json:"content"`
	Timestamp            time.Time              `json:"timestamp"`
	EditedTimestamp      *time.Time             `json:"edited_timestamp"`
	TTS                  bool                   `json:"tts"`
	MentionEveryone      bool                   `json:"mention_everyone"`
	Mentions             []users.User           `json:"mentions"`
	MentionRoles         []roles.Role           `json:"mention_roles"`
	MentionChannels      []channels.Channel     `json:"mention_channels,omitempty"`
	Attachments          []Attachment           `json:"attachments"`
	Embeds               []Embed                `json:"embeds"`
	Reactions            []Reaction             `json:"reactions,omitempty"`
	Nonce                interface{}            `json:"nonce,omitempty"`
	Pinned               bool                   `json:"pinned"`
	WebhookID            snowflake.ID           `json:"webhook_id,string,omitempty"`
	Type                 MessageType            `json:"type"`
	Activity             *MessageActivity       `json:"activity,omitempty"`
	ApplicationID        snowflake.ID           `json:"application_id,string,omitempty"`
	MessageReference     *MessageReference      `json:"message_reference,omitempty"`
	Flags                int                    `json:"flags,omitempty"`
	ReferencedMessage    *Message               `json:"referenced_message,omitempty"`
	Interaction          interface{}            `json:"interaction,omitempty"`
	Thread               *channels.Channel      `json:"thread,omitempty"`
	Components           []components.Component `json:"components,omitempty"`
	StickerItems         []emojis.StickerItem   `json:"sticker_items,omitempty"`
	Position             int                    `json:"position,omitempty"`
	RoleSubscriptionData interface{}            `json:"role_subscription_data,omitempty"`
	Poll                 *Poll                  `json:"poll,omitempty"`
	MessageSnapshots     []MessageSnapshot      `json:"message_snapshots,omitempty"`
	InteractionMetadata  *InteractionMetadata   `json:"interaction_metadata,omitempty"`
	Call                 *MessageCall           `json:"call,omitempty"`
}

// UnmarshalJSON unmarshals the Message and handles components properly.
func (m *Message) UnmarshalJSON(data []byte) error {
	type Alias Message
	aux := &struct {
		Components   []json.RawMessage `json:"components"`
		MentionRoles []string          `json:"mention_roles"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.MentionRoles != nil {
		m.MentionRoles = make([]roles.Role, len(aux.MentionRoles))
		for i, r := range aux.MentionRoles {
			id, err := snowflake.Parse(r)
			if err != nil {
				return err
			}
			m.MentionRoles[i] = roles.Role{ID: id}
		}
	}

	if len(aux.Components) > 0 {
		m.Components = make([]components.Component, len(aux.Components))
		for i, compRaw := range aux.Components {
			component, err := components.Unmarshal(compRaw)
			if err != nil {
				return err
			}
			m.Components[i] = component
		}
	}

	return nil
}

// FirstEmbed returns the first embed on the message, or false if the
// message has no embeds. This is a safe accessor that avoids index-out-of-
// range panics when a message's embeds were edited or removed externally.
func (m *Message) FirstEmbed() (Embed, bool) {
	if m == nil || len(m.Embeds) == 0 {
		return Embed{}, false
	}
	return m.Embeds[0], true
}
