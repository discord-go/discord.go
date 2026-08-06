package rest

import (
	"context"

	"github.com/discord-go/discord.go/gateway"
)

type GatewayInfo struct {
	URL string `json:"url"`
}

type GatewayBotInfo struct {
	URL               string                    `json:"url"`
	Shards            int                       `json:"shards"`
	SessionStartLimit gateway.SessionStartLimit `json:"session_start_limit"`
}

func (c *Client) GetGateway(ctx context.Context) (*GatewayInfo, error) {
	var result GatewayInfo
	err := c.Request(ctx, "GET", "/gateway", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetGatewayBot(ctx context.Context) (*GatewayBotInfo, error) {
	var result GatewayBotInfo
	err := c.Request(ctx, "GET", "/gateway/bot", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
