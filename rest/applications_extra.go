package rest

import (
	"context"

	"github.com/discord-go/discord.go/application"
	"github.com/discord-go/discord.go/snowflake"
)

type ApplicationRoleConnectionMetadata struct {
	Type                     int               `json:"type"`
	Key                      string            `json:"key"`
	Name                     string            `json:"name"`
	NameLocalizations        map[string]string `json:"name_localizations,omitempty"`
	Description              string            `json:"description"`
	DescriptionLocalizations map[string]string `json:"description_localizations,omitempty"`
}

type ApplicationRoleConnection struct {
	PlatformName     string            `json:"platform_name,omitempty"`
	PlatformUsername string            `json:"platform_username,omitempty"`
	Metadata         map[string]string `json:"metadata"`
}

func (c *Client) GetCurrentApplication(ctx context.Context) (*application.Application, error) {
	var result application.Application
	err := c.Request(ctx, "GET", "/applications/@me", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) ModifyCurrentApplication(ctx context.Context, params any) (*application.Application, error) {
	var result application.Application
	err := c.Request(ctx, "PATCH", "/applications/@me", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetApplicationActivityInstance(ctx context.Context, applicationID snowflake.ID, instanceID string) (*application.ActivityInstance, error) {
	var result application.ActivityInstance
	err := c.Request(ctx, "GET", "/applications/"+applicationID.String()+"/activity-instances/"+instanceID, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetApplicationRoleConnectionMetadata(ctx context.Context, applicationID snowflake.ID) ([]ApplicationRoleConnectionMetadata, error) {
	var result []ApplicationRoleConnectionMetadata
	err := c.Request(ctx, "GET", "/applications/"+applicationID.String()+"/role-connections/metadata", nil, &result)
	return result, err
}

func (c *Client) UpdateApplicationRoleConnectionMetadata(ctx context.Context, applicationID snowflake.ID, metadata []ApplicationRoleConnectionMetadata) ([]ApplicationRoleConnectionMetadata, error) {
	var result []ApplicationRoleConnectionMetadata
	err := c.Request(ctx, "PUT", "/applications/"+applicationID.String()+"/role-connections/metadata", metadata, &result)
	return result, err
}
