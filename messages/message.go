package messages

import (
	"encoding/json"
	"strings"
	"time"
	"unicode"

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
	// Member is the guild member who sent the message. The gateway includes
	// it in MESSAGE_CREATE payloads when the GuildMembers intent is enabled.
	// REST responses may omit it, so treat a nil value as "member unknown".
	Member *users.Member `json:"member,omitempty"`
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

// unicode category sets used by SanitizeContent.
var (
	// Zero-width and invisible formatting characters: zero-width space,
	// word joiner, soft hyphen, BOM, and the invisible tag characters
	// Discord uses for super reactions. ZWJ (U+200D), ZWNJ (U+200C), and
	// variation selector-16 (U+FE0F) are deliberately NOT here: they are
	// required for emoji composition (family emoji, keycaps, hearts) and
	// Persian/Arabic orthography, so stripping them corrupts visible text.
	zeroWidthRunes = &unicode.RangeTable{
		R16: []unicode.Range16{
			{Lo: 0x00ad, Hi: 0x00ad, Stride: 1}, // soft hyphen
			{Lo: 0x180e, Hi: 0x180e, Stride: 1}, // mongolian vowel separator
			{Lo: 0x200b, Hi: 0x200b, Stride: 1}, // zero-width space
			{Lo: 0x2028, Hi: 0x2029, Stride: 1}, // line/paragraph separator
			{Lo: 0x2060, Hi: 0x2064, Stride: 1}, // word joiner, invisible ops
			{Lo: 0xfe00, Hi: 0xfe0e, Stride: 1}, // variation selectors 1-15
			{Lo: 0xfeff, Hi: 0xfeff, Stride: 1}, // BOM / zero-width no-break space
			{Lo: 0xfff9, Hi: 0xfffb, Stride: 1}, // interlinear annotation
		},
		R32: []unicode.Range32{
			{Lo: 0x13430, Hi: 0x1345f, Stride: 1}, // egyptian format controls
			{Lo: 0x1bca0, Hi: 0x1bca3, Stride: 1}, // shorthand format controls
			{Lo: 0x1d173, Hi: 0x1d17a, Stride: 1}, // musical format controls
			{Lo: 0xe0000, Hi: 0xe0fff, Stride: 1}, // tags / variation selectors 17+
		},
	}
	// Bidi formatting characters that can reorder text visually
	// (trojan-source style). Includes the isolates and pop directional.
	bidiRunes = &unicode.RangeTable{
		R16: []unicode.Range16{
			{Lo: 0x061c, Hi: 0x061c, Stride: 1}, // arabic letter mark
			{Lo: 0x200e, Hi: 0x200f, Stride: 1}, // LRM, RLM
			{Lo: 0x202a, Hi: 0x202e, Stride: 1}, // LRE, RLE, PDF, LRO, RLO
			{Lo: 0x2066, Hi: 0x2069, Stride: 1}, // LRI, RLI, FSI, PDI
		},
	}
)

// SanitizeContent returns the message content with invisible formatting
// removed: zero-width characters, bidi control characters (including
// trojan-source bidi overrides), and other unprintable format characters.
// Visible text, emoji, and legitimate whitespace are preserved. Use it
// before displaying, logging, or matching user-supplied content, because
// hidden characters can make two visually identical strings compare
// unequal or flip the rendering direction of surrounding text.
func (m *Message) SanitizeContent() string {
	if m == nil {
		return ""
	}
	if !strings.ContainsFunc(m.Content, func(r rune) bool {
		return unicode.Is(zeroWidthRunes, r) || unicode.Is(bidiRunes, r)
	}) {
		return m.Content
	}
	var b strings.Builder
	b.Grow(len(m.Content))
	for _, r := range m.Content {
		if unicode.Is(zeroWidthRunes, r) || unicode.Is(bidiRunes, r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// CleanContent returns the message content with user, role, and channel
// mentions replaced by their display names when the name is known from the
// message's mention data: <@123> becomes @username, <@&456> becomes
// @rolename, and <#789> becomes #channelname. Mentions whose name is not
// carried by the payload (gateway role mentions arrive as bare IDs) are left
// as-is, and the original content is returned unchanged when the message
// has no mention data. The returned string is safe to display but is not
// a substitute for escaping user input when re-sending it to Discord.
func (m *Message) CleanContent() string {
	if m == nil {
		return ""
	}
	content := m.Content
	if content == "" {
		return ""
	}

	userNames := make(map[string]string, len(m.Mentions))
	for i := range m.Mentions {
		u := &m.Mentions[i]
		name := u.Username
		if u.GlobalName != nil && *u.GlobalName != "" {
			name = *u.GlobalName
		}
		userNames[u.ID.String()] = name
	}

	roleNames := make(map[string]string, len(m.MentionRoles))
	for i := range m.MentionRoles {
		// Gateway payloads carry role IDs only; a role decoded from JSON has
		// an empty name. Substituting it would mangle <@&id> into a bare "@",
		// so leave the token intact when the name is unknown.
		if m.MentionRoles[i].Name == "" {
			continue
		}
		roleNames[m.MentionRoles[i].ID.String()] = m.MentionRoles[i].Name
	}

	channelNames := make(map[string]string, len(m.MentionChannels))
	for i := range m.MentionChannels {
		name := m.MentionChannels[i].Name
		if name == nil {
			continue
		}
		channelNames[m.MentionChannels[i].ID.String()] = *name
	}

	var b strings.Builder
	b.Grow(len(content) + 16)
	for i := 0; i < len(content); {
		if content[i] != '<' {
			b.WriteByte(content[i])
			i++
			continue
		}
		end := strings.IndexByte(content[i:], '>')
		if end == -1 {
			b.WriteString(content[i:])
			break
		}
		end += i
		token := content[i : end+1]
		inner := content[i+1 : end]
		switch {
		case strings.HasPrefix(inner, "@&"):
			if name, ok := roleNames[inner[2:]]; ok {
				b.WriteString("@" + name)
			} else {
				b.WriteString(token)
			}
		case strings.HasPrefix(inner, "@!"):
			if name, ok := userNames[inner[2:]]; ok {
				b.WriteString("@" + name)
			} else {
				b.WriteString(token)
			}
		case strings.HasPrefix(inner, "@"):
			if name, ok := userNames[inner[1:]]; ok {
				b.WriteString("@" + name)
			} else {
				b.WriteString(token)
			}
		case strings.HasPrefix(inner, "#"):
			if name, ok := channelNames[inner[1:]]; ok {
				b.WriteString("#" + name)
			} else {
				b.WriteString(token)
			}
		default:
			b.WriteString(token)
		}
		i = end + 1
	}
	return b.String()
}
