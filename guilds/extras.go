package guilds

import (
	"github.com/discord-go/discord.go/emojis"
	"github.com/discord-go/discord.go/snowflake"
)

type GuildPreview struct {
	ID                       snowflake.ID   `json:"id,string"`
	Name                     string         `json:"name"`
	Icon                     *string        `json:"icon"`
	Splash                   *string        `json:"splash"`
	DiscoverySplash          *string        `json:"discovery_splash"`
	Emojis                   []emojis.Emoji `json:"emojis"`
	Features                 []Feature      `json:"features"`
	ApproximateMemberCount   int            `json:"approximate_member_count"`
	ApproximatePresenceCount int            `json:"approximate_presence_count"`
	Description              *string        `json:"description"`
}

type VanityURL struct {
	Code string `json:"code"`
	Uses int    `json:"uses"`
}

type Onboarding struct {
	GuildID           snowflake.ID       `json:"guild_id,string"`
	Prompts           []OnboardingPrompt `json:"prompts"`
	DefaultChannelIDs snowflake.IDs      `json:"default_channel_ids,omitempty"`
	Enabled           bool               `json:"enabled"`
	Mode              int                `json:"mode"`
}

type OnboardingPrompt struct {
	ID           snowflake.ID             `json:"id,string"`
	Type         int                      `json:"type"`
	Title        string                   `json:"title"`
	Options      []OnboardingPromptOption `json:"options"`
	SingleSelect bool                     `json:"single_select"`
	Required     bool                     `json:"required"`
	InOnboarding bool                     `json:"in_onboarding"`
}

type OnboardingPromptOption struct {
	ID          snowflake.ID  `json:"id,string"`
	ChannelIDs  snowflake.IDs `json:"channel_ids,omitempty"`
	RoleIDs     snowflake.IDs `json:"role_ids,omitempty"`
	Title       string        `json:"title"`
	Description *string       `json:"description,omitempty"`
	Emoji       *emojis.Emoji `json:"emoji,omitempty"`
	EmojiID     *snowflake.ID `json:"emoji_id,string,omitempty"`
	EmojiName   *string       `json:"emoji_name,omitempty"`
	Default     bool          `json:"default"`
}
