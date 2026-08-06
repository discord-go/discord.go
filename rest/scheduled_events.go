package rest

import (
	"context"
	"net/url"
	"strconv"

	"github.com/discord-go/discord.go/guilds"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

// ScheduledEventUsersParams controls scheduled-event subscriber listing.
type ScheduledEventUsersParams struct {
	Limit         int
	WithMember    bool
	WithUserCount bool
	After         *snowflake.ID
}

type ScheduledEventListParams struct{ WithUserCount bool }

func (p ScheduledEventListParams) query() string {
	if p.WithUserCount {
		return "?with_user_count=true"
	}
	return ""
}

func (p ScheduledEventUsersParams) query() string {
	values := url.Values{}
	if p.Limit > 0 {
		values.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.WithMember {
		values.Set("with_member", "true")
	}
	if p.WithUserCount {
		values.Set("with_user_count", "true")
	}
	if p.After != nil {
		values.Set("after", p.After.String())
	}
	encoded := values.Encode()
	if encoded == "" {
		return ""
	}
	return "?" + encoded
}

// ListScheduledEvents lists all scheduled events for a guild.
func (c *Client) ListScheduledEvents(ctx context.Context, guildID snowflake.ID) ([]guilds.ScheduledEvent, error) {
	return c.ListScheduledEventsWithOptions(ctx, guildID, ScheduledEventListParams{})
}

func (c *Client) ListScheduledEventsWithOptions(ctx context.Context, guildID snowflake.ID, params ScheduledEventListParams) ([]guilds.ScheduledEvent, error) {
	var events []guilds.ScheduledEvent
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/scheduled-events"+params.query(), nil, &events)
	return events, err
}

// GetScheduledEvent gets a single scheduled event.
func (c *Client) GetScheduledEvent(ctx context.Context, guildID, eventID snowflake.ID) (*guilds.ScheduledEvent, error) {
	return c.GetScheduledEventWithOptions(ctx, guildID, eventID, ScheduledEventListParams{})
}

func (c *Client) GetScheduledEventWithOptions(ctx context.Context, guildID, eventID snowflake.ID, params ScheduledEventListParams) (*guilds.ScheduledEvent, error) {
	var event guilds.ScheduledEvent
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/scheduled-events/"+eventID.String()+params.query(), nil, &event)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// CreateScheduledEvent creates a new scheduled event.
func (c *Client) CreateScheduledEvent(ctx context.Context, guildID snowflake.ID, params CreateScheduledEventParams) (*guilds.ScheduledEvent, error) {
	var event guilds.ScheduledEvent
	err := c.Request(ctx, "POST", "/guilds/"+guildID.String()+"/scheduled-events", params, &event)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// ModifyScheduledEvent modifies an existing scheduled event.
func (c *Client) ModifyScheduledEvent(ctx context.Context, guildID, eventID snowflake.ID, params ModifyScheduledEventParams) (*guilds.ScheduledEvent, error) {
	var event guilds.ScheduledEvent
	err := c.Request(ctx, "PATCH", "/guilds/"+guildID.String()+"/scheduled-events/"+eventID.String(), params, &event)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// DeleteScheduledEvent deletes a scheduled event.
func (c *Client) DeleteScheduledEvent(ctx context.Context, guildID, eventID snowflake.ID) error {
	return c.Request(ctx, "DELETE", "/guilds/"+guildID.String()+"/scheduled-events/"+eventID.String(), nil, nil)
}

// ListScheduledEventUsers lists users subscribed to a scheduled event.
func (c *Client) ListScheduledEventUsers(ctx context.Context, guildID, eventID snowflake.ID, params ScheduledEventUsersParams) ([]users.Member, error) {
	var members []users.Member
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/scheduled-events/"+eventID.String()+"/users"+params.query(), nil, &members)
	return members, err
}
