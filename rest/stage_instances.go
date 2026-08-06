package rest

import (
	"context"

	"github.com/discord-go/discord.go/guilds"
	"github.com/discord-go/discord.go/snowflake"
)

type CreateStageInstanceParams struct {
	ChannelID             snowflake.ID  `json:"channel_id,string"`
	Topic                 string        `json:"topic"`
	PrivacyLevel          int           `json:"privacy_level,omitempty"`
	GuildScheduledEventID *snowflake.ID `json:"guild_scheduled_event_id,string,omitempty"`
	SendStartNotification *bool         `json:"send_start_notification,omitempty"`
}

type ModifyStageInstanceParams struct {
	Topic        *string `json:"topic,omitempty"`
	PrivacyLevel *int    `json:"privacy_level,omitempty"`
}

func (c *Client) GetStageInstance(ctx context.Context, channelID snowflake.ID) (*guilds.StageInstance, error) {
	var instance guilds.StageInstance
	err := c.Request(ctx, "GET", "/stage-instances/"+channelID.String(), nil, &instance)
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

func (c *Client) CreateStageInstance(ctx context.Context, params CreateStageInstanceParams) (*guilds.StageInstance, error) {
	var instance guilds.StageInstance
	err := c.Request(ctx, "POST", "/stage-instances", params, &instance)
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

func (c *Client) ModifyStageInstance(ctx context.Context, channelID snowflake.ID, params ModifyStageInstanceParams) (*guilds.StageInstance, error) {
	var instance guilds.StageInstance
	err := c.Request(ctx, "PATCH", "/stage-instances/"+channelID.String(), params, &instance)
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

func (c *Client) DeleteStageInstance(ctx context.Context, channelID snowflake.ID) error {
	return c.Request(ctx, "DELETE", "/stage-instances/"+channelID.String(), nil, nil)
}
