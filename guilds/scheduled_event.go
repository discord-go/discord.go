package guilds

import "github.com/discord-go/discord.go/snowflake"

// ScheduledEvent represents a scheduled event in a guild.
type ScheduledEvent struct {
	ID                 snowflake.ID                  `json:"id,string"`
	GuildID            snowflake.ID                  `json:"guild_id,string"`
	ChannelID          *snowflake.ID                 `json:"channel_id,string,omitempty"`
	CreatorID          snowflake.ID                  `json:"creator_id,string,omitempty"`
	Name               string                        `json:"name"`
	Description        *string                       `json:"description,omitempty"`
	ScheduledStartTime string                        `json:"scheduled_start_time"`
	ScheduledEndTime   *string                       `json:"scheduled_end_time,omitempty"`
	PrivacyLevel       int                           `json:"privacy_level"`
	Status             int                           `json:"status"`
	EntityType         int                           `json:"entity_type"`
	EntityID           *snowflake.ID                 `json:"entity_id,string,omitempty"`
	EntityMetadata     *ScheduledEventEntityMetadata `json:"entity_metadata,omitempty"`
	UserCount          *int                          `json:"user_count,omitempty"`
	Image              *string                       `json:"image,omitempty"`
}

type ScheduledEventEntityMetadata struct {
	Location string `json:"location,omitempty"`
}

type ScheduledEventRecurrenceRule struct {
	Start      string `json:"start"`
	End        string `json:"end,omitempty"`
	Frequency  int    `json:"frequency"`
	Interval   int    `json:"interval"`
	ByWeekday  []int  `json:"by_weekday,omitempty"`
	ByNWeekday []struct {
		N       int `json:"n"`
		Weekday int `json:"day"`
	} `json:"by_n_weekday,omitempty"`
	ByMonth    []int `json:"by_month,omitempty"`
	ByMonthDay []int `json:"by_month_day,omitempty"`
	ByYearDay  []int `json:"by_year_day,omitempty"`
	Count      int   `json:"count,omitempty"`
}
