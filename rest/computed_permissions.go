package rest

import (
	"context"

	"github.com/discord-go/discord.go/permissions"
	"github.com/discord-go/discord.go/snowflake"
)

type computedPermissionsResponse struct {
	Permissions permissions.Permission `json:"permissions,string"`
}

// GetGuildMemberPermissions returns a member's computed guild permissions,
// including their roles and @everyone.
func (c *Client) GetGuildMemberPermissions(ctx context.Context, guildID, userID snowflake.ID) (permissions.Permission, error) {
	var result computedPermissionsResponse
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/permissions/member/"+userID.String(), nil, &result)
	return result.Permissions, err
}

// GetGuildRolePermissions returns a role's guild permissions combined with
// @everyone; it does not include other roles or member-specific restrictions.
func (c *Client) GetGuildRolePermissions(ctx context.Context, guildID, roleID snowflake.ID) (permissions.Permission, error) {
	var result computedPermissionsResponse
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/permissions/role/"+roleID.String(), nil, &result)
	return result.Permissions, err
}

// GetChannelMemberPermissions returns computed permissions for a member in a
// channel, including roles and member-specific overwrites. Discord requires
// VIEW_CHANNEL and, for voice and stage channels, CONNECT.
func (c *Client) GetChannelMemberPermissions(ctx context.Context, channelID, userID snowflake.ID) (permissions.Permission, error) {
	var result computedPermissionsResponse
	err := c.Request(ctx, "GET", "/channels/"+channelID.String()+"/permissions/member/"+userID.String(), nil, &result)
	return result.Permissions, err
}

// GetChannelRolePermissions returns computed permissions for a role and
// @everyone in a channel. Other roles and member overwrites are excluded.
// Discord requires VIEW_CHANNEL and, for voice and stage channels, CONNECT.
func (c *Client) GetChannelRolePermissions(ctx context.Context, channelID, roleID snowflake.ID) (permissions.Permission, error) {
	var result computedPermissionsResponse
	err := c.Request(ctx, "GET", "/channels/"+channelID.String()+"/permissions/role/"+roleID.String(), nil, &result)
	return result.Permissions, err
}
