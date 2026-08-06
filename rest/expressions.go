package rest

import (
	"context"

	"github.com/discord-go/discord.go/emojis"
	"github.com/discord-go/discord.go/snowflake"
)

type CreateEmojiParams struct {
	Name          string        `json:"name"`
	Image         string        `json:"image"`
	Roles         snowflake.IDs `json:"roles,omitempty"`
	RequireColons *bool         `json:"require_colons,omitempty"`
}

type ModifyEmojiParams struct {
	Name  *string       `json:"name,omitempty"`
	Roles snowflake.IDs `json:"roles,omitempty"`
}

func (c *Client) ListGuildEmojis(ctx context.Context, guildID snowflake.ID) ([]emojis.Emoji, error) {
	var result []emojis.Emoji
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/emojis", nil, &result)
	return result, err
}

func (c *Client) GetGuildEmoji(ctx context.Context, guildID, emojiID snowflake.ID) (*emojis.Emoji, error) {
	var result emojis.Emoji
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/emojis/"+emojiID.String(), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CreateGuildEmoji(ctx context.Context, guildID snowflake.ID, params CreateEmojiParams) (*emojis.Emoji, error) {
	var result emojis.Emoji
	err := c.Request(ctx, "POST", "/guilds/"+guildID.String()+"/emojis", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) ModifyGuildEmoji(ctx context.Context, guildID, emojiID snowflake.ID, params ModifyEmojiParams) (*emojis.Emoji, error) {
	var result emojis.Emoji
	err := c.Request(ctx, "PATCH", "/guilds/"+guildID.String()+"/emojis/"+emojiID.String(), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteGuildEmoji(ctx context.Context, guildID, emojiID snowflake.ID) error {
	return c.Request(ctx, "DELETE", "/guilds/"+guildID.String()+"/emojis/"+emojiID.String(), nil, nil)
}

type ApplicationEmojiList struct {
	Items []emojis.Emoji `json:"items"`
}

func (c *Client) ListApplicationEmojis(ctx context.Context, applicationID snowflake.ID) ([]emojis.Emoji, error) {
	var result ApplicationEmojiList
	err := c.Request(ctx, "GET", "/applications/"+applicationID.String()+"/emojis", nil, &result)
	return result.Items, err
}

func (c *Client) CreateApplicationEmoji(ctx context.Context, applicationID snowflake.ID, params CreateEmojiParams) (*emojis.Emoji, error) {
	var result emojis.Emoji
	err := c.Request(ctx, "POST", "/applications/"+applicationID.String()+"/emojis", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetApplicationEmoji(ctx context.Context, applicationID, emojiID snowflake.ID) (*emojis.Emoji, error) {
	var result emojis.Emoji
	err := c.Request(ctx, "GET", "/applications/"+applicationID.String()+"/emojis/"+emojiID.String(), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) ModifyApplicationEmoji(ctx context.Context, applicationID, emojiID snowflake.ID, params ModifyEmojiParams) (*emojis.Emoji, error) {
	var result emojis.Emoji
	err := c.Request(ctx, "PATCH", "/applications/"+applicationID.String()+"/emojis/"+emojiID.String(), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteApplicationEmoji(ctx context.Context, applicationID, emojiID snowflake.ID) error {
	return c.Request(ctx, "DELETE", "/applications/"+applicationID.String()+"/emojis/"+emojiID.String(), nil, nil)
}
