package rest

import (
	"context"

	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/voice"
)

func (c *Client) ListVoiceRegions(ctx context.Context) ([]voice.VoiceRegion, error) {
	var result []voice.VoiceRegion
	err := c.Request(ctx, "GET", "/voice/regions", nil, &result)
	return result, err
}

// ListGuildVoiceRegions returns the guild's available voice regions, including
// VIP regions when applicable.
func (c *Client) ListGuildVoiceRegions(ctx context.Context, guildID snowflake.ID) ([]voice.VoiceRegion, error) {
	var result []voice.VoiceRegion
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/regions", nil, &result)
	return result, err
}
