package rest

import (
	"encoding/json"
	"time"

	"github.com/discord-go/discord.go/channels"
	"github.com/discord-go/discord.go/permissions"
	"github.com/discord-go/discord.go/snowflake"
)

// CreateGuildParams are the parameters for creating a guild.
type CreateGuildParams struct {
	Name                        string                     `json:"name"`
	Region                      string                     `json:"region,omitempty"`
	Icon                        string                     `json:"icon,omitempty"`
	VerificationLevel           *int                       `json:"verification_level,omitempty"`
	DefaultMessageNotifications *int                       `json:"default_message_notifications,omitempty"`
	ExplicitContentFilter       *int                       `json:"explicit_content_filter,omitempty"`
	Roles                       []CreateGuildRoleParams    `json:"roles,omitempty"`
	Channels                    []CreateGuildChannelParams `json:"channels,omitempty"`
	AFKChannelID                *snowflake.ID              `json:"afk_channel_id,string,omitempty"`
	AFKTimeout                  *int                       `json:"afk_timeout,omitempty"`
	SystemChannelID             *snowflake.ID              `json:"system_channel_id,string,omitempty"`
	SystemChannelFlags          *int                       `json:"system_channel_flags,omitempty"`
}

// CreateGuildRoleParams is a role within CreateGuildParams.
type CreateGuildRoleParams struct {
	Name        string                 `json:"name,omitempty"`
	Permissions permissions.Permission `json:"permissions,string,omitempty"`
	Color       int                    `json:"color,omitempty"`
	Hoist       bool                   `json:"hoist,omitempty"`
	Mentionable bool                   `json:"mentionable,omitempty"`
}

// CreateGuildChannelParams is used for creating a guild channel within a guild,
// or as part of CreateGuildParams.
//
// Discord limits channel names to 100 characters. Names longer than 100
// characters will be rejected by the API with a 400 error.
type CreateGuildChannelParams struct {
	Name                 string               `json:"name"`
	Type                 channels.ChannelType `json:"type,omitempty"`
	Topic                string               `json:"topic,omitempty"`
	Bitrate              int                  `json:"bitrate,omitempty"`
	UserLimit            int                  `json:"user_limit,omitempty"`
	RateLimitPerUser     int                  `json:"rate_limit_per_user,omitempty"`
	Position             int                  `json:"position,omitempty"`
	PermissionOverwrites []channels.Overwrite `json:"permission_overwrites,omitempty"`
	ParentID             *snowflake.ID        `json:"parent_id,string,omitempty"`
	NSFW                 bool                 `json:"nsfw,omitempty"`
}

// ModifyGuildParams are the parameters for modifying a guild.
type ModifyGuildParams struct {
	Name                        *string       `json:"name,omitempty"`
	Region                      *string       `json:"region,omitempty"`
	VerificationLevel           *int          `json:"verification_level,omitempty"`
	DefaultMessageNotifications *int          `json:"default_message_notifications,omitempty"`
	ExplicitContentFilter       *int          `json:"explicit_content_filter,omitempty"`
	AFKChannelID                *snowflake.ID `json:"afk_channel_id,string,omitempty"`
	AFKTimeout                  *int          `json:"afk_timeout,omitempty"`
	Icon                        *string       `json:"icon,omitempty"`
	OwnerID                     *snowflake.ID `json:"owner_id,string,omitempty"`
	Splash                      *string       `json:"splash,omitempty"`
	DiscoverySplash             *string       `json:"discovery_splash,omitempty"`
	Banner                      *string       `json:"banner,omitempty"`
	SystemChannelID             *snowflake.ID `json:"system_channel_id,string,omitempty"`
	SystemChannelFlags          *int          `json:"system_channel_flags,omitempty"`
	RulesChannelID              *snowflake.ID `json:"rules_channel_id,string,omitempty"`
	PublicUpdatesChannelID      *snowflake.ID `json:"public_updates_channel_id,string,omitempty"`
	PreferredLocale             *string       `json:"preferred_locale,omitempty"`
	Description                 *string       `json:"description,omitempty"`
	PremiumProgressBarEnabled   *bool         `json:"premium_progress_bar_enabled,omitempty"`
	SafetyAlertsChannelID       *snowflake.ID `json:"safety_alerts_channel_id,string,omitempty"`
}

// ListMembersParams are the query parameters for listing guild members.
type ListMembersParams struct {
	Limit int           `json:"limit,omitempty"`
	After *snowflake.ID `json:"after,omitempty"`
}

// ModifyMemberParams are the parameters for modifying a guild member.
type ModifyMemberParams struct {
	Nick                       *string        `json:"nick,omitempty"`
	Roles                      []snowflake.ID `json:"roles,omitempty"`
	Mute                       *bool          `json:"mute,omitempty"`
	Deaf                       *bool          `json:"deaf,omitempty"`
	ChannelID                  *snowflake.ID  `json:"channel_id,string,omitempty"`
	CommunicationDisabledUntil *time.Time     `json:"communication_disabled_until,omitempty"`
}

func (p ModifyMemberParams) MarshalJSON() ([]byte, error) {
	type payload struct {
		Nick                       *string       `json:"nick,omitempty"`
		Roles                      snowflake.IDs `json:"roles,omitempty"`
		Mute                       *bool         `json:"mute,omitempty"`
		Deaf                       *bool         `json:"deaf,omitempty"`
		ChannelID                  *snowflake.ID `json:"channel_id,string,omitempty"`
		CommunicationDisabledUntil *time.Time    `json:"communication_disabled_until,omitempty"`
	}
	return json.Marshal(payload{p.Nick, snowflake.IDs(p.Roles), p.Mute, p.Deaf, p.ChannelID, p.CommunicationDisabledUntil})
}

// CreateBanParams are the parameters for creating a guild ban.
type CreateBanParams struct {
	DeleteMessageSeconds int `json:"delete_message_seconds,omitempty"`
}

// CreateRoleParams are the parameters for creating a guild role.
type CreateRoleParams struct {
	Name         string                 `json:"name,omitempty"`
	Permissions  permissions.Permission `json:"permissions,string,omitempty"`
	Color        int                    `json:"color,omitempty"`
	Hoist        bool                   `json:"hoist,omitempty"`
	Icon         *string                `json:"icon,omitempty"`
	UnicodeEmoji *string                `json:"unicode_emoji,omitempty"`
	Mentionable  bool                   `json:"mentionable,omitempty"`
}

// ModifyRoleParams are the parameters for modifying a guild role.
type ModifyRoleParams struct {
	Name         *string                 `json:"name,omitempty"`
	Permissions  *permissions.Permission `json:"permissions,string,omitempty"`
	Color        *int                    `json:"color,omitempty"`
	Hoist        *bool                   `json:"hoist,omitempty"`
	Icon         *string                 `json:"icon,omitempty"`
	UnicodeEmoji *string                 `json:"unicode_emoji,omitempty"`
	Mentionable  *bool                   `json:"mentionable,omitempty"`
}

// PruneParams are the parameters for getting or beginning a guild prune.
type PruneParams struct {
	Days              int            `json:"days,omitempty"`
	ComputePruneCount *bool          `json:"compute_prune_count,omitempty"`
	IncludeRoles      []snowflake.ID `json:"include_roles,omitempty"`
}

func (p PruneParams) MarshalJSON() ([]byte, error) {
	type payload struct {
		Days              int           `json:"days,omitempty"`
		ComputePruneCount *bool         `json:"compute_prune_count,omitempty"`
		IncludeRoles      snowflake.IDs `json:"include_roles,omitempty"`
	}
	return json.Marshal(payload{p.Days, p.ComputePruneCount, snowflake.IDs(p.IncludeRoles)})
}

// AuditLogParams are the query parameters for fetching an audit log.
type AuditLogParams struct {
	UserID     *snowflake.ID `json:"user_id,omitempty"`
	ActionType *int          `json:"action_type,omitempty"`
	Before     *snowflake.ID `json:"before,omitempty"`
	After      *snowflake.ID `json:"after,omitempty"`
	Limit      int           `json:"limit,omitempty"`
}
