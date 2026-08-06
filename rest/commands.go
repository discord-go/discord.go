package rest

import (
	"context"
	"net/url"

	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/snowflake"
)

// Application Commands
//
// Global commands are available in every guild that adds your app.
// Note: Global commands have a 1-hour propagation delay for updates.
// Guild commands are specific to the guild you specify when making them.
// Note: Guild commands update instantly (or near-instant).

func (c *Client) GetGlobalApplicationCommands(ctx context.Context, applicationID snowflake.ID) ([]interactions.ApplicationCommand, error) {
	return c.GetGlobalApplicationCommandsWithOptions(ctx, applicationID, false)
}

func (c *Client) GetGlobalApplicationCommandsWithOptions(ctx context.Context, applicationID snowflake.ID, withLocalizations bool) ([]interactions.ApplicationCommand, error) {
	var commands []interactions.ApplicationCommand
	query := ""
	if withLocalizations {
		query = "?" + url.Values{"with_localizations": {"true"}}.Encode()
	}
	err := c.Request(ctx, "GET", "/applications/"+applicationID.String()+"/commands"+query, nil, &commands)
	if err != nil {
		return nil, err
	}
	return commands, nil
}

func (c *Client) CreateGlobalApplicationCommand(ctx context.Context, applicationID snowflake.ID, params CreateCommandParams) (*interactions.ApplicationCommand, error) {
	var cmd interactions.ApplicationCommand
	err := c.Request(ctx, "POST", "/applications/"+applicationID.String()+"/commands", params, &cmd)
	if err != nil {
		return nil, err
	}
	return &cmd, nil
}

func (c *Client) GetGlobalApplicationCommand(ctx context.Context, applicationID, commandID snowflake.ID) (*interactions.ApplicationCommand, error) {
	var cmd interactions.ApplicationCommand
	err := c.Request(ctx, "GET", "/applications/"+applicationID.String()+"/commands/"+commandID.String(), nil, &cmd)
	if err != nil {
		return nil, err
	}
	return &cmd, nil
}

func (c *Client) EditGlobalApplicationCommand(ctx context.Context, applicationID, commandID snowflake.ID, params CreateCommandParams) (*interactions.ApplicationCommand, error) {
	var cmd interactions.ApplicationCommand
	err := c.Request(ctx, "PATCH", "/applications/"+applicationID.String()+"/commands/"+commandID.String(), params, &cmd)
	if err != nil {
		return nil, err
	}
	return &cmd, nil
}

func (c *Client) DeleteGlobalApplicationCommand(ctx context.Context, applicationID, commandID snowflake.ID) error {
	return c.Request(ctx, "DELETE", "/applications/"+applicationID.String()+"/commands/"+commandID.String(), nil, nil)
}

func (c *Client) BulkOverwriteGlobalCommands(ctx context.Context, applicationID snowflake.ID, commands []CreateCommandParams) ([]interactions.ApplicationCommand, error) {
	var cmds []interactions.ApplicationCommand
	err := c.Request(ctx, "PUT", "/applications/"+applicationID.String()+"/commands", commands, &cmds)
	if err != nil {
		return nil, err
	}
	return cmds, nil
}

func (c *Client) GetGuildApplicationCommands(ctx context.Context, applicationID, guildID snowflake.ID) ([]interactions.ApplicationCommand, error) {
	return c.GetGuildApplicationCommandsWithOptions(ctx, applicationID, guildID, false)
}

func (c *Client) GetGuildApplicationCommandsWithOptions(ctx context.Context, applicationID, guildID snowflake.ID, withLocalizations bool) ([]interactions.ApplicationCommand, error) {
	var commands []interactions.ApplicationCommand
	query := ""
	if withLocalizations {
		query = "?" + url.Values{"with_localizations": {"true"}}.Encode()
	}
	err := c.Request(ctx, "GET", "/applications/"+applicationID.String()+"/guilds/"+guildID.String()+"/commands"+query, nil, &commands)
	if err != nil {
		return nil, err
	}
	return commands, nil
}

func (c *Client) CreateGuildApplicationCommand(ctx context.Context, applicationID, guildID snowflake.ID, params CreateCommandParams) (*interactions.ApplicationCommand, error) {
	var cmd interactions.ApplicationCommand
	err := c.Request(ctx, "POST", "/applications/"+applicationID.String()+"/guilds/"+guildID.String()+"/commands", params, &cmd)
	if err != nil {
		return nil, err
	}
	return &cmd, nil
}

func (c *Client) GetGuildApplicationCommand(ctx context.Context, applicationID, guildID, commandID snowflake.ID) (*interactions.ApplicationCommand, error) {
	var cmd interactions.ApplicationCommand
	err := c.Request(ctx, "GET", "/applications/"+applicationID.String()+"/guilds/"+guildID.String()+"/commands/"+commandID.String(), nil, &cmd)
	if err != nil {
		return nil, err
	}
	return &cmd, nil
}

func (c *Client) EditGuildApplicationCommand(ctx context.Context, applicationID, guildID, commandID snowflake.ID, params CreateCommandParams) (*interactions.ApplicationCommand, error) {
	var cmd interactions.ApplicationCommand
	err := c.Request(ctx, "PATCH", "/applications/"+applicationID.String()+"/guilds/"+guildID.String()+"/commands/"+commandID.String(), params, &cmd)
	if err != nil {
		return nil, err
	}
	return &cmd, nil
}

func (c *Client) DeleteGuildApplicationCommand(ctx context.Context, applicationID, guildID, commandID snowflake.ID) error {
	return c.Request(ctx, "DELETE", "/applications/"+applicationID.String()+"/guilds/"+guildID.String()+"/commands/"+commandID.String(), nil, nil)
}

func (c *Client) BulkOverwriteGuildCommands(ctx context.Context, applicationID, guildID snowflake.ID, commands []CreateCommandParams) ([]interactions.ApplicationCommand, error) {
	var cmds []interactions.ApplicationCommand
	err := c.Request(ctx, "PUT", "/applications/"+applicationID.String()+"/guilds/"+guildID.String()+"/commands", commands, &cmds)
	if err != nil {
		return nil, err
	}
	return cmds, nil
}
