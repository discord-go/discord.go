package guilds

import (
	"github.com/discord-go/discord.go/snowflake"
)

type AutoModerationTriggerMetadata struct {
	KeywordFilter                []string `json:"keyword_filter,omitempty"`
	RegexPatterns                []string `json:"regex_patterns,omitempty"`
	Presets                      []int    `json:"presets,omitempty"`
	AllowList                    []string `json:"allow_list,omitempty"`
	MentionTotalLimit            int      `json:"mention_total_limit,omitempty"`
	MentionRaidProtectionEnabled bool     `json:"mention_raid_protection_enabled,omitempty"`
}

type AutoModerationActionMetadata struct {
	ChannelID       *snowflake.ID `json:"channel_id,string,omitempty"`
	DurationSeconds int           `json:"duration_seconds,omitempty"`
	CustomMessage   string        `json:"custom_message,omitempty"`
}

type AutoModerationAction struct {
	Type     int                           `json:"type"`
	Metadata *AutoModerationActionMetadata `json:"metadata,omitempty"`
}

// AutoModerationRule represents an auto moderation rule.
type AutoModerationRule struct {
	ID              snowflake.ID                   `json:"id,string"`
	GuildID         snowflake.ID                   `json:"guild_id,string"`
	Name            string                         `json:"name"`
	CreatorID       snowflake.ID                   `json:"creator_id,string"`
	EventType       int                            `json:"event_type"`
	TriggerType     int                            `json:"trigger_type"`
	TriggerMetadata *AutoModerationTriggerMetadata `json:"trigger_metadata,omitempty"`
	Actions         []AutoModerationAction         `json:"actions"`
	Enabled         bool                           `json:"enabled"`
	ExemptRoles     snowflake.IDs                  `json:"exempt_roles,omitempty"`
	ExemptChannels  snowflake.IDs                  `json:"exempt_channels,omitempty"`
}
