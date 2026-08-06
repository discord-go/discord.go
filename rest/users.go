package rest

import (
	"context"

	"github.com/discord-go/discord.go/channels"
	"github.com/discord-go/discord.go/guilds"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

func (c *Client) GetCurrentUser(ctx context.Context) (*users.User, error) {
	var user users.User
	err := c.Request(ctx, "GET", "/users/@me", nil, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *Client) GetUser(ctx context.Context, userID snowflake.ID) (*users.User, error) {
	var user users.User
	err := c.Request(ctx, "GET", "/users/"+userID.String(), nil, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *Client) ModifyCurrentUser(ctx context.Context, params ModifyUserParams) (*users.User, error) {
	var user users.User
	err := c.Request(ctx, "PATCH", "/users/@me", params, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *Client) GetCurrentUserGuilds(ctx context.Context, params ListGuildsParams) ([]guilds.Guild, error) {
	var gs []guilds.Guild
	err := c.Request(ctx, "GET", "/users/@me/guilds"+params.QueryString(), nil, &gs)
	if err != nil {
		return nil, err
	}
	return gs, nil
}

func (c *Client) LeaveGuild(ctx context.Context, guildID snowflake.ID) error {
	return c.Request(ctx, "DELETE", "/users/@me/guilds/"+guildID.String(), nil, nil)
}

// createDMRequest is the payload for creating a DM.
type createDMRequest struct {
	RecipientID snowflake.ID `json:"recipient_id,string"`
}

func (c *Client) CreateDM(ctx context.Context, recipientID snowflake.ID) (*channels.Channel, error) {
	var ch channels.Channel
	req := createDMRequest{RecipientID: recipientID}
	err := c.Request(ctx, "POST", "/users/@me/channels", req, &ch)
	if err != nil {
		return nil, err
	}
	return &ch, nil
}
