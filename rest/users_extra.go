package rest

import (
	"context"

	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

type UserConnection struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Revoked      bool     `json:"revoked"`
	Integrations []string `json:"integrations,omitempty"`
	Verified     bool     `json:"verified"`
	FriendSync   bool     `json:"friend_sync"`
	ShowActivity bool     `json:"show_activity"`
	TwoWayLink   bool     `json:"two_way_link"`
	Visibility   int      `json:"visibility"`
}

func (c *Client) GetCurrentUserGuildMember(ctx context.Context, guildID snowflake.ID) (*users.Member, error) {
	var result users.Member
	err := c.Request(ctx, "GET", "/users/@me/guilds/"+guildID.String()+"/member", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetCurrentUserConnections(ctx context.Context) ([]UserConnection, error) {
	var result []UserConnection
	err := c.Request(ctx, "GET", "/users/@me/connections", nil, &result)
	return result, err
}

func (c *Client) GetCurrentUserRoleConnection(ctx context.Context, applicationID snowflake.ID) (*ApplicationRoleConnection, error) {
	var result ApplicationRoleConnection
	err := c.Request(ctx, "GET", "/users/@me/applications/"+applicationID.String()+"/role-connection", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateCurrentUserRoleConnection(ctx context.Context, applicationID snowflake.ID, connection ApplicationRoleConnection) (*ApplicationRoleConnection, error) {
	var result ApplicationRoleConnection
	err := c.Request(ctx, "PUT", "/users/@me/applications/"+applicationID.String()+"/role-connection", connection, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteCurrentUserRoleConnection(ctx context.Context, applicationID snowflake.ID) error {
	return c.Request(ctx, "DELETE", "/users/@me/applications/"+applicationID.String()+"/role-connection", nil, nil)
}
