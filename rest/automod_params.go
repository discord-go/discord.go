package rest

import (
	"github.com/discord-go/discord.go/guilds"
	"github.com/discord-go/discord.go/snowflake"
)

type AutoModerationTriggerMetadata = guilds.AutoModerationTriggerMetadata
type AutoModerationActionMetadata = guilds.AutoModerationActionMetadata
type AutoModerationAction = guilds.AutoModerationAction

// CreateAutoModerationRuleParams represents the official Discord AutoMod
// create payload.
type CreateAutoModerationRuleParams struct {
	Name            string                                `json:"name"`
	EventType       int                                   `json:"event_type"`
	TriggerType     int                                   `json:"trigger_type"`
	TriggerMetadata *guilds.AutoModerationTriggerMetadata `json:"trigger_metadata,omitempty"`
	Actions         []guilds.AutoModerationAction         `json:"actions"`
	Enabled         *bool                                 `json:"enabled,omitempty"`
	ExemptRoles     snowflake.IDs                         `json:"exempt_roles,omitempty"`
	ExemptChannels  snowflake.IDs                         `json:"exempt_channels,omitempty"`
}

// ModifyAutoModerationRuleParams represents the official Discord AutoMod
// modify payload.
type ModifyAutoModerationRuleParams struct {
	Name            *string                               `json:"name,omitempty"`
	EventType       *int                                  `json:"event_type,omitempty"`
	TriggerMetadata *guilds.AutoModerationTriggerMetadata `json:"trigger_metadata,omitempty"`
	Actions         []guilds.AutoModerationAction         `json:"actions,omitempty"`
	Enabled         *bool                                 `json:"enabled,omitempty"`
	ExemptRoles     snowflake.IDs                         `json:"exempt_roles,omitempty"`
	ExemptChannels  snowflake.IDs                         `json:"exempt_channels,omitempty"`
}
