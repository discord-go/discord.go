package rest

import (
	"github.com/discord-go/discord.go/channels"
	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/snowflake"
)

// ModifyChannelParams contains the parameters for modifying a channel.
type ModifyChannelParams struct {
	Name                          *string                    `json:"name,omitempty"`
	Type                          *channels.ChannelType      `json:"type,omitempty"`
	Position                      *int                       `json:"position,omitempty"`
	Topic                         *string                    `json:"topic,omitempty"`
	NSFW                          *bool                      `json:"nsfw,omitempty"`
	RateLimitPerUser              *int                       `json:"rate_limit_per_user,omitempty"`
	Bitrate                       *int                       `json:"bitrate,omitempty"`
	UserLimit                     *int                       `json:"user_limit,omitempty"`
	PermissionOverwrites          *[]channels.Overwrite      `json:"permission_overwrites,omitempty"`
	ParentID                      *snowflake.ID              `json:"parent_id,string,omitempty"`
	RTCRegion                     *string                    `json:"rtc_region,omitempty"`
	VideoQualityMode              *channels.VideoQualityMode `json:"video_quality_mode,omitempty"`
	Flags                         *int                       `json:"flags,omitempty"`
	AvailableTags                 *[]channels.ForumTag       `json:"available_tags,omitempty"`
	AppliedTags                   *[]string                  `json:"applied_tags,omitempty"`
	DefaultReactionEmoji          *channels.DefaultReaction  `json:"default_reaction_emoji,omitempty"`
	DefaultAutoArchiveDuration    *int                       `json:"default_auto_archive_duration,omitempty"`
	DefaultThreadRateLimitPerUser *int                       `json:"default_thread_rate_limit_per_user,omitempty"`
	DefaultSortOrder              *int                       `json:"default_sort_order,omitempty"`
	DefaultForumLayout            *int                       `json:"default_forum_layout,omitempty"`
}

// GetMessagesParams contains the query parameters for retrieving channel messages.
type GetMessagesParams struct {
	Around *snowflake.ID `json:"around,omitempty"`
	Before *snowflake.ID `json:"before,omitempty"`
	After  *snowflake.ID `json:"after,omitempty"`
	Limit  *int          `json:"limit,omitempty"`
}

// QueryString returns the query string representation of the parameters.
func (p GetMessagesParams) QueryString() string {
	var parts []string
	if p.Around != nil {
		parts = append(parts, "around="+p.Around.String())
	}
	if p.Before != nil {
		parts = append(parts, "before="+p.Before.String())
	}
	if p.After != nil {
		parts = append(parts, "after="+p.After.String())
	}
	if p.Limit != nil {
		parts = append(parts, "limit="+itoa(*p.Limit))
	}
	if len(parts) == 0 {
		return ""
	}
	return "?" + joinStrings(parts, "&")
}

// EditMessageParams contains the parameters for editing a message.
type EditMessageParams struct {
	Content         *string                   `json:"content,omitempty"`
	Embeds          *[]messages.Embed         `json:"embeds,omitempty"`
	Components      *[]components.Component   `json:"components,omitempty"`
	Attachments     *[]messages.Attachment    `json:"attachments,omitempty"`
	Flags           *int                      `json:"flags,omitempty"`
	AllowedMentions *messages.AllowedMentions `json:"allowed_mentions,omitempty"`
	Poll            *messages.Poll            `json:"poll,omitempty"`
}

// CreateInviteParams contains the parameters for creating a channel invite.
type CreateInviteParams struct {
	MaxAge              *int          `json:"max_age,omitempty"`
	MaxUses             *int          `json:"max_uses,omitempty"`
	Temporary           *bool         `json:"temporary,omitempty"`
	Unique              *bool         `json:"unique,omitempty"`
	TargetType          *int          `json:"target_type,omitempty"`
	TargetUserID        *snowflake.ID `json:"target_user_id,string,omitempty"`
	TargetApplicationID *snowflake.ID `json:"target_application_id,string,omitempty"`
	RoleIDs             snowflake.IDs `json:"role_ids,omitempty"`
}

// GetAnswerVotersParams contains the query parameters for retrieving poll answer voters.
type GetAnswerVotersParams struct {
	After snowflake.ID `json:"after,omitempty"`
	Limit int          `json:"limit,omitempty"`
}

// QueryString returns the query string representation of the parameters.
func (p GetAnswerVotersParams) QueryString() string {
	var parts []string
	if p.After != 0 {
		parts = append(parts, "after="+p.After.String())
	}
	if p.Limit != 0 {
		parts = append(parts, "limit="+itoa(p.Limit))
	}
	if len(parts) == 0 {
		return ""
	}
	return "?" + joinStrings(parts, "&")
}

// itoa converts an int to a string without importing strconv in this file.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	digits := make([]byte, 0, 10)
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}
	if neg {
		digits = append(digits, '-')
	}
	// reverse
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	return string(digits)
}

// joinStrings joins strings with a separator.
func joinStrings(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for _, p := range parts[1:] {
		result += sep + p
	}
	return result
}
