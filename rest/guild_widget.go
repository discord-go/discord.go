package rest

import (
	"context"
	"net/url"

	"github.com/discord-go/discord.go/guilds"
	"github.com/discord-go/discord.go/snowflake"
)

// ModifyGuildWidgetParams controls which widget settings are changed.
// ChannelID is omitted when nil; point to a nil *snowflake.ID to clear the
// channel, or to a non-nil *snowflake.ID to set it.
type ModifyGuildWidgetParams struct {
	Enabled   *bool          `json:"enabled,omitempty"`
	ChannelID **snowflake.ID `json:"channel_id,omitempty"`
}

// GetGuildWidgetSettings returns widget settings. It requires MANAGE_GUILD.
func (c *Client) GetGuildWidgetSettings(ctx context.Context, guildID snowflake.ID) (*guilds.WidgetSettings, error) {
	var settings guilds.WidgetSettings
	if err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/widget", nil, &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

// ModifyGuildWidget updates widget settings. It requires MANAGE_GUILD and
// supports an audit-log reason supplied with WithReason.
func (c *Client) ModifyGuildWidget(ctx context.Context, guildID snowflake.ID, params ModifyGuildWidgetParams) (*guilds.WidgetSettings, error) {
	var settings guilds.WidgetSettings
	if err := c.Request(ctx, "PATCH", "/guilds/"+guildID.String()+"/widget", params, &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

// GetGuildWidgetImage fetches the public PNG widget without authentication.
// An empty style uses Discord's default; supported styles include shield and
// banner1 through banner4.
func (c *Client) GetGuildWidgetImage(ctx context.Context, guildID snowflake.ID, style string) ([]byte, error) {
	path := "/guilds/" + guildID.String() + "/widget.png"
	if style != "" {
		path += "?" + url.Values{"style": {style}}.Encode()
	}
	var image []byte
	err := c.requestWithDecoder(ctx, "GET", path, nil, false, func(body []byte) error {
		image = body
		return nil
	})
	return image, err
}
