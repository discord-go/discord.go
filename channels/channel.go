package channels

import (
	"time"

	"github.com/discord-go/discord.go/permissions"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

type Channel struct {
	ID                            snowflake.ID      `json:"id,string"`
	Type                          ChannelType       `json:"type"`
	GuildID                       snowflake.ID      `json:"guild_id,string,omitempty"`
	Position                      *int              `json:"position,omitempty"`
	PermissionOverwrites          []Overwrite       `json:"permission_overwrites,omitempty"`
	Name                          *string           `json:"name,omitempty"`
	Topic                         *string           `json:"topic,omitempty"`
	NSFW                          *bool             `json:"nsfw,omitempty"`
	LastMessageID                 *snowflake.ID     `json:"last_message_id,string,omitempty"`
	Bitrate                       *int              `json:"bitrate,omitempty"`
	UserLimit                     *int              `json:"user_limit,omitempty"`
	RateLimitPerUser              *int              `json:"rate_limit_per_user,omitempty"`
	Recipients                    []users.User      `json:"recipients,omitempty"`
	Icon                          *string           `json:"icon,omitempty"`
	OwnerID                       *snowflake.ID     `json:"owner_id,string,omitempty"`
	ApplicationID                 *snowflake.ID     `json:"application_id,string,omitempty"`
	Managed                       *bool             `json:"managed,omitempty"`
	ParentID                      *snowflake.ID     `json:"parent_id,string,omitempty"`
	LastPinTimestamp              *time.Time        `json:"last_pin_timestamp,omitempty"`
	RTCRegion                     *string           `json:"rtc_region,omitempty"`
	VideoQualityMode              *VideoQualityMode `json:"video_quality_mode,omitempty"`
	MessageCount                  *int              `json:"message_count,omitempty"`
	MemberCount                   *int              `json:"member_count,omitempty"`
	ThreadMetadata                *ThreadMetadata   `json:"thread_metadata,omitempty"`
	Member                        *ThreadMember     `json:"member,omitempty"`
	DefaultAutoArchiveDuration    *int              `json:"default_auto_archive_duration,omitempty"`
	Permissions                   *string           `json:"permissions,omitempty"`
	Flags                         *int              `json:"flags,omitempty"`
	TotalMessageSent              *int              `json:"total_message_sent,omitempty"`
	AvailableTags                 []ForumTag        `json:"available_tags,omitempty"`
	AppliedTags                   []string          `json:"applied_tags,omitempty"`
	DefaultReactionEmoji          *DefaultReaction  `json:"default_reaction_emoji,omitempty"`
	DefaultThreadRateLimitPerUser *int              `json:"default_thread_rate_limit_per_user,omitempty"`
	DefaultSortOrder              *int              `json:"default_sort_order,omitempty"`
	DefaultForumLayout            *int              `json:"default_forum_layout,omitempty"`
}

type Overwrite struct {
	ID    snowflake.ID           `json:"id,string"`
	Type  int                    `json:"type"`
	Allow permissions.Permission `json:"allow,string"`
	Deny  permissions.Permission `json:"deny,string"`
}

type ForumTag struct {
	ID        snowflake.ID  `json:"id,string"`
	Name      string        `json:"name"`
	Moderated bool          `json:"moderated"`
	EmojiID   *snowflake.ID `json:"emoji_id,string,omitempty"`
	EmojiName *string       `json:"emoji_name,omitempty"`
}

type DefaultReaction struct {
	EmojiID   *snowflake.ID `json:"emoji_id,string,omitempty"`
	EmojiName *string       `json:"emoji_name,omitempty"`
}
