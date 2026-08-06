package rest

import (
	"context"

	"github.com/discord-go/discord.go/voice"
)

func (c *Client) ListVoiceRegions(ctx context.Context) ([]voice.VoiceRegion, error) {
	var result []voice.VoiceRegion
	err := c.Request(ctx, "GET", "/voice/regions", nil, &result)
	return result, err
}
