package rest

import (
	"context"
	"net/url"

	"github.com/discord-go/discord.go/channels"
	"github.com/discord-go/discord.go/snowflake"
)

type GetInviteParams struct {
	WithCounts            bool
	WithExpiration        bool
	GuildScheduledEventID *snowflake.ID
}

func (p GetInviteParams) query() string {
	values := url.Values{}
	if p.WithCounts {
		values.Set("with_counts", "true")
	}
	if p.WithExpiration {
		values.Set("with_expiration", "true")
	}
	if p.GuildScheduledEventID != nil {
		values.Set("guild_scheduled_event_id", p.GuildScheduledEventID.String())
	}
	if encoded := values.Encode(); encoded != "" {
		return "?" + encoded
	}
	return ""
}

func (c *Client) GetInvite(ctx context.Context, code string) (*channels.Invite, error) {
	return c.GetInviteWithOptions(ctx, code, GetInviteParams{})
}

func (c *Client) GetInviteWithOptions(ctx context.Context, code string, params GetInviteParams) (*channels.Invite, error) {
	var invite channels.Invite
	err := c.Request(ctx, "GET", "/invites/"+code+params.query(), nil, &invite)
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

func (c *Client) DeleteInvite(ctx context.Context, code string) (*channels.Invite, error) {
	var invite channels.Invite
	err := c.Request(ctx, "DELETE", "/invites/"+code, nil, &invite)
	if err != nil {
		return nil, err
	}
	return &invite, nil
}
