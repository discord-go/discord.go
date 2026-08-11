package rest

import (
	"context"
	"net/url"
	"strconv"

	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/permissions"
	"github.com/discord-go/discord.go/snowflake"
)

type PermissionOverwriteParams struct {
	Allow permissions.Permission `json:"allow,string"`
	Deny  permissions.Permission `json:"deny,string"`
	Type  int                    `json:"type"`
}

func (c *Client) EditChannelPermission(ctx context.Context, channelID, overwriteID snowflake.ID, params PermissionOverwriteParams) error {
	return c.Request(ctx, "PUT", "/channels/"+channelID.String()+"/permissions/"+overwriteID.String(), params, nil)
}

func (c *Client) DeleteChannelPermission(ctx context.Context, channelID, overwriteID snowflake.ID) error {
	return c.Request(ctx, "DELETE", "/channels/"+channelID.String()+"/permissions/"+overwriteID.String(), nil, nil)
}

type MessageSearchParams struct {
	Content       string
	AuthorID      snowflake.ID
	Mentions      snowflake.ID
	Has           string
	In            snowflake.ID
	Limit         int
	Offset        int
	Before        *snowflake.ID
	After         *snowflake.ID
	Pinned        *bool
	WithoutEmbeds *bool
	SortBy        string
	SortOrder     string
}

func (p MessageSearchParams) query() string {
	values := url.Values{}
	if p.Content != "" {
		values.Set("content", p.Content)
	}
	if p.AuthorID != 0 {
		values.Set("author_id", p.AuthorID.String())
	}
	if p.Mentions != 0 {
		values.Set("mentions", p.Mentions.String())
	}
	if p.Has != "" {
		values.Set("has", p.Has)
	}
	if p.In != 0 {
		values.Set("in", p.In.String())
	}
	if p.Limit > 0 {
		values.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Offset > 0 {
		values.Set("offset", strconv.Itoa(p.Offset))
	}
	if p.Before != nil {
		values.Set("before", p.Before.String())
	}
	if p.After != nil {
		values.Set("after", p.After.String())
	}
	if p.Pinned != nil {
		values.Set("pinned", strconv.FormatBool(*p.Pinned))
	}
	if p.WithoutEmbeds != nil {
		values.Set("embed", strconv.FormatBool(!*p.WithoutEmbeds))
	}
	if p.SortBy != "" {
		values.Set("sort_by", p.SortBy)
	}
	if p.SortOrder != "" {
		values.Set("sort_order", p.SortOrder)
	}
	if encoded := values.Encode(); encoded != "" {
		return "?" + encoded
	}
	return ""
}

type MessageSearchResult struct {
	AnalyticsID      string               `json:"analytics_id,omitempty"`
	Messages         [][]messages.Message `json:"messages"`
	Total            int                  `json:"total_results"`
	DoingDeepHistory bool                 `json:"doing_deep_historical_index"`
}

func (c *Client) SearchGuildMessages(ctx context.Context, guildID snowflake.ID, params MessageSearchParams) (*MessageSearchResult, error) {
	var result MessageSearchResult
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/messages/search"+params.query(), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

type VoiceStatusParams struct {
	Status string `json:"status"`
}

func (c *Client) SetVoiceChannelStatus(ctx context.Context, channelID snowflake.ID, status string) error {
	return c.Request(ctx, "PUT", "/channels/"+channelID.String()+"/voice-status", VoiceStatusParams{Status: status}, nil)
}
