package rest

import (
	"context"

	"github.com/discord-go/discord.go/snowflake"
)

type ApplicationCommandPermission struct {
	ID         snowflake.ID `json:"id,string"`
	Type       int          `json:"type"`
	Permission bool         `json:"permission"`
}

type GuildApplicationCommandPermissions struct {
	ID            snowflake.ID                   `json:"id,string"`
	ApplicationID snowflake.ID                   `json:"application_id,string"`
	GuildID       snowflake.ID                   `json:"guild_id,string"`
	Permissions   []ApplicationCommandPermission `json:"permissions"`
}

func (c *Client) GetGuildApplicationCommandPermissions(ctx context.Context, applicationID, guildID, commandID snowflake.ID) (*GuildApplicationCommandPermissions, error) {
	var result GuildApplicationCommandPermissions
	err := c.Request(ctx, "GET", "/applications/"+applicationID.String()+"/guilds/"+guildID.String()+"/commands/"+commandID.String()+"/permissions", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetGuildApplicationCommandsPermissions(ctx context.Context, applicationID, guildID snowflake.ID) ([]GuildApplicationCommandPermissions, error) {
	var result []GuildApplicationCommandPermissions
	err := c.Request(ctx, "GET", "/applications/"+applicationID.String()+"/guilds/"+guildID.String()+"/commands/permissions", nil, &result)
	return result, err
}

func (c *Client) EditGuildApplicationCommandPermissions(ctx context.Context, applicationID, guildID, commandID snowflake.ID, permissions []ApplicationCommandPermission) (*GuildApplicationCommandPermissions, error) {
	payload := struct {
		Permissions []ApplicationCommandPermission `json:"permissions"`
	}{Permissions: permissions}
	var result GuildApplicationCommandPermissions
	err := c.Request(ctx, "PUT", "/applications/"+applicationID.String()+"/guilds/"+guildID.String()+"/commands/"+commandID.String()+"/permissions", payload, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
