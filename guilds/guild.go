package guilds

import (
	"encoding/json"

	"github.com/discord-go/discord.go/emojis"
	"github.com/discord-go/discord.go/roles"
	"github.com/discord-go/discord.go/snowflake"
)

// Guild represents a Discord guild (server).
type Guild struct {
	ID                          snowflake.ID   `json:"id,string"`
	Name                        string         `json:"name"`
	Icon                        *string        `json:"icon"`
	IconHash                    *string        `json:"icon_hash,omitempty"`
	Splash                      *string        `json:"splash"`
	DiscoverySplash             *string        `json:"discovery_splash"`
	Owner                       *bool          `json:"owner,omitempty"`
	OwnerID                     snowflake.ID   `json:"owner_id,string"`
	Permissions                 *string        `json:"permissions,omitempty"`
	Region                      *string        `json:"region,omitempty"` // Deprecated
	AFKChannelID                *snowflake.ID  `json:"afk_channel_id,string"`
	AFKTimeout                  int            `json:"afk_timeout"`
	WidgetEnabled               *bool          `json:"widget_enabled,omitempty"`
	WidgetChannelID             *snowflake.ID  `json:"widget_channel_id,string"`
	VerificationLevel           int            `json:"verification_level"`
	DefaultMessageNotifications int            `json:"default_message_notifications"`
	ExplicitContentFilter       int            `json:"explicit_content_filter"`
	Roles                       []roles.Role   `json:"roles"`
	Emojis                      []emojis.Emoji `json:"emojis"`
	Features                    []Feature      `json:"features"`
	MFALevel                    int            `json:"mfa_level"`
	ApplicationID               *snowflake.ID  `json:"application_id,string"`
	SystemChannelID             *snowflake.ID  `json:"system_channel_id,string"`
	SystemChannelFlags          int            `json:"system_channel_flags"`
	RulesChannelID              *snowflake.ID  `json:"rules_channel_id,string"`
	MaxPresences                *int           `json:"max_presences,omitempty"`
	MaxMembers                  *int           `json:"max_members,omitempty"`
	VanityURLCode               *string        `json:"vanity_url_code"`
	Description                 *string        `json:"description"`
	Banner                      *string        `json:"banner"`
	PremiumTier                 int            `json:"premium_tier"`
	PremiumSubscriptionCount    *int           `json:"premium_subscription_count,omitempty"`
	PreferredLocale             string         `json:"preferred_locale"`
	PublicUpdatesChannelID      *snowflake.ID  `json:"public_updates_channel_id,string"`
	MaxVideoChannelUsers        *int           `json:"max_video_channel_users,omitempty"`
	MaxStageVideoChannelUsers   *int           `json:"max_stage_video_channel_users,omitempty"`
	ApproximateMemberCount      *int           `json:"approximate_member_count,omitempty"`
	ApproximatePresenceCount    *int           `json:"approximate_presence_count,omitempty"`
	WelcomeScreen               *WelcomeScreen `json:"welcome_screen,omitempty"`
	NSFWLevel                   int            `json:"nsfw_level"`
	PremiumProgressBarEnabled   bool           `json:"premium_progress_bar_enabled"`
	SafetyAlertsChannelID       *snowflake.ID  `json:"safety_alerts_channel_id,string"`
	Unavailable                 bool           `json:"unavailable"`
}

// UnmarshalJSON unmarshals Guild, correctly handling snowflakes where the API returns null strings.
// Some fields like afk_channel_id can be sent as null.
func (g *Guild) UnmarshalJSON(data []byte) error {
	type Alias Guild
	aux := &struct {
		ID                     string  `json:"id"`
		OwnerID                string  `json:"owner_id"`
		AFKChannelID           *string `json:"afk_channel_id"`
		WidgetChannelID        *string `json:"widget_channel_id"`
		ApplicationID          *string `json:"application_id"`
		SystemChannelID        *string `json:"system_channel_id"`
		RulesChannelID         *string `json:"rules_channel_id"`
		PublicUpdatesChannelID *string `json:"public_updates_channel_id"`
		SafetyAlertsChannelID  *string `json:"safety_alerts_channel_id"`
		*Alias
	}{
		Alias: (*Alias)(g),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.ID != "" {
		id, err := snowflake.Parse(aux.ID)
		if err != nil {
			return err
		}
		g.ID = id
	}
	if aux.OwnerID != "" {
		id, err := snowflake.Parse(aux.OwnerID)
		if err != nil {
			return err
		}
		g.OwnerID = id
	}

	parseSnowflakePtr := func(s *string) (*snowflake.ID, error) {
		if s == nil || *s == "" {
			return nil, nil
		}
		id, err := snowflake.Parse(*s)
		if err != nil {
			return nil, err
		}
		return &id, nil
	}

	var err error
	g.AFKChannelID, err = parseSnowflakePtr(aux.AFKChannelID)
	if err != nil {
		return err
	}
	g.WidgetChannelID, err = parseSnowflakePtr(aux.WidgetChannelID)
	if err != nil {
		return err
	}
	g.ApplicationID, err = parseSnowflakePtr(aux.ApplicationID)
	if err != nil {
		return err
	}
	g.SystemChannelID, err = parseSnowflakePtr(aux.SystemChannelID)
	if err != nil {
		return err
	}
	g.RulesChannelID, err = parseSnowflakePtr(aux.RulesChannelID)
	if err != nil {
		return err
	}
	g.PublicUpdatesChannelID, err = parseSnowflakePtr(aux.PublicUpdatesChannelID)
	if err != nil {
		return err
	}
	g.SafetyAlertsChannelID, err = parseSnowflakePtr(aux.SafetyAlertsChannelID)
	if err != nil {
		return err
	}

	return nil
}

// UnavailableGuild represents a guild that is unavailable due to an outage.
type UnavailableGuild struct {
	ID          snowflake.ID `json:"id,string"`
	Unavailable bool         `json:"unavailable"`
}

// WelcomeScreen represents the welcome screen of a Community guild.
type WelcomeScreen struct {
	Description     *string                `json:"description"`
	WelcomeChannels []WelcomeScreenChannel `json:"welcome_channels"`
}

// WelcomeScreenChannel represents a channel in a welcome screen.
type WelcomeScreenChannel struct {
	ChannelID   snowflake.ID  `json:"channel_id,string"`
	Description string        `json:"description"`
	EmojiID     *snowflake.ID `json:"emoji_id,string"`
	EmojiName   *string       `json:"emoji_name"`
}

// UnmarshalJSON unmarshals WelcomeScreenChannel
func (w *WelcomeScreenChannel) UnmarshalJSON(data []byte) error {
	type Alias WelcomeScreenChannel
	aux := &struct {
		ChannelID string  `json:"channel_id"`
		EmojiID   *string `json:"emoji_id"`
		*Alias
	}{
		Alias: (*Alias)(w),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.ChannelID != "" {
		id, err := snowflake.Parse(aux.ChannelID)
		if err != nil {
			return err
		}
		w.ChannelID = id
	}
	if aux.EmojiID != nil && *aux.EmojiID != "" {
		id, err := snowflake.Parse(*aux.EmojiID)
		if err != nil {
			return err
		}
		w.EmojiID = &id
	}

	return nil
}

// MaxEmojis returns the maximum number of custom emojis (both static and animated individually) the guild can have based on its premium tier.
func (g *Guild) MaxEmojis() int {
	switch g.PremiumTier {
	case 1:
		return 100
	case 2:
		return 150
	case 3:
		return 250
	default:
		return 50
	}
}

// MaxStickers returns the maximum number of custom stickers the guild can have based on its premium tier.
func (g *Guild) MaxStickers() int {
	switch g.PremiumTier {
	case 1:
		return 15
	case 2:
		return 30
	case 3:
		return 60
	default:
		return 5
	}
}
