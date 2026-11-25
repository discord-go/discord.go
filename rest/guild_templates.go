package rest

import (
	"context"
	"net/url"

	"github.com/discord-go/discord.go/guilds"
	"github.com/discord-go/discord.go/snowflake"
)

// CreateGuildTemplateParams describes a new guild template.
type CreateGuildTemplateParams struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

// ModifyGuildTemplateParams describes template metadata to change. An empty
// description clears the existing description.
type ModifyGuildTemplateParams struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// GetGuildTemplate retrieves a template by its code.
func (c *Client) GetGuildTemplate(ctx context.Context, code string) (*guilds.Template, error) {
	var result guilds.Template
	err := c.Request(ctx, "GET", "/guilds/templates/"+url.PathEscape(code), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetGuildTemplates lists templates for a guild.
func (c *Client) GetGuildTemplates(ctx context.Context, guildID snowflake.ID) ([]guilds.Template, error) {
	var result []guilds.Template
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/templates", nil, &result)
	return result, err
}

// CreateGuildTemplate creates a template for a guild.
func (c *Client) CreateGuildTemplate(ctx context.Context, guildID snowflake.ID, params CreateGuildTemplateParams) (*guilds.Template, error) {
	var result guilds.Template
	err := c.Request(ctx, "POST", "/guilds/"+guildID.String()+"/templates", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// SyncGuildTemplate updates a template to match its source guild.
func (c *Client) SyncGuildTemplate(ctx context.Context, guildID snowflake.ID, code string) (*guilds.Template, error) {
	var result guilds.Template
	err := c.Request(ctx, "PUT", "/guilds/"+guildID.String()+"/templates/"+url.PathEscape(code), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ModifyGuildTemplate changes a template's metadata.
func (c *Client) ModifyGuildTemplate(ctx context.Context, guildID snowflake.ID, code string, params ModifyGuildTemplateParams) (*guilds.Template, error) {
	var result guilds.Template
	err := c.Request(ctx, "PATCH", "/guilds/"+guildID.String()+"/templates/"+url.PathEscape(code), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteGuildTemplate deletes a template and returns the deleted template.
func (c *Client) DeleteGuildTemplate(ctx context.Context, guildID snowflake.ID, code string) (*guilds.Template, error) {
	var result guilds.Template
	err := c.Request(ctx, "DELETE", "/guilds/"+guildID.String()+"/templates/"+url.PathEscape(code), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
