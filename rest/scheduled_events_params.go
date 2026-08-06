package rest

import (
	"time"

	"github.com/discord-go/discord.go/guilds"
	"github.com/discord-go/discord.go/snowflake"
)

// CreateScheduledEventParams represents the official scheduled-event payload.
type CreateScheduledEventParams struct {
	ChannelID          *snowflake.ID                        `json:"channel_id,string,omitempty"`
	EntityMetadata     *guilds.ScheduledEventEntityMetadata `json:"entity_metadata,omitempty"`
	Name               string                               `json:"name"`
	PrivacyLevel       int                                  `json:"privacy_level"`
	ScheduledStartTime time.Time                            `json:"scheduled_start_time"`
	ScheduledEndTime   *time.Time                           `json:"scheduled_end_time,omitempty"`
	Description        *string                              `json:"description,omitempty"`
	EntityType         int                                  `json:"entity_type"`
	Image              *string                              `json:"image,omitempty"`
	RecurrenceRule     *guilds.ScheduledEventRecurrenceRule `json:"recurrence_rule,omitempty"`
}

// ModifyScheduledEventParams represents the official scheduled-event modify payload.
type ModifyScheduledEventParams struct {
	ChannelID          *snowflake.ID                        `json:"channel_id,string,omitempty"`
	EntityMetadata     *guilds.ScheduledEventEntityMetadata `json:"entity_metadata,omitempty"`
	Name               *string                              `json:"name,omitempty"`
	PrivacyLevel       *int                                 `json:"privacy_level,omitempty"`
	ScheduledStartTime *time.Time                           `json:"scheduled_start_time,omitempty"`
	ScheduledEndTime   *time.Time                           `json:"scheduled_end_time,omitempty"`
	Description        *string                              `json:"description,omitempty"`
	EntityType         *int                                 `json:"entity_type,omitempty"`
	Status             *int                                 `json:"status,omitempty"`
	Image              *string                              `json:"image,omitempty"`
	RecurrenceRule     *guilds.ScheduledEventRecurrenceRule `json:"recurrence_rule,omitempty"`
}
