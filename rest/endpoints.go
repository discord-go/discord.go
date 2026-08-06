package rest

import (
	"context"
	"strconv"

	"github.com/discord-go/discord.go/guilds"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/snowflake"
)

// CreateMessageRequest is the payload for creating a message.
type CreateMessageRequest struct {
	Content string `json:"content"`
}

// CreateMessage creates a message in a channel.
func (c *Client) CreateMessage(ctx context.Context, channelID snowflake.ID, content string) (*messages.Message, error) {
	req := CreateMessageRequest{Content: content}
	var msg messages.Message
	err := c.Request(ctx, "POST", "/channels/"+channelID.String()+"/messages", req, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// CreateMessageComplex creates a message in a channel with advanced parameters like embeds and components.
func (c *Client) CreateMessageComplex(ctx context.Context, channelID snowflake.ID, send messages.MessageSend) (*messages.Message, error) {
	var msg messages.Message
	// If there are files, we should use RequestMultipart, but we'll use Request for JSON payloads first.
	err := c.Request(ctx, "POST", "/channels/"+channelID.String()+"/messages", send, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// CreateMessageComplexWithFiles sends a message using multipart/form-data.
func (c *Client) CreateMessageComplexWithFiles(ctx context.Context, channelID snowflake.ID, send messages.MessageSend, files []File) (*messages.Message, error) {
	send.Attachments = make([]messages.AttachmentSend, 0, len(files))
	for index, file := range files {
		send.Attachments = append(send.Attachments, messages.AttachmentSend{ID: strconv.Itoa(index), Filename: file.Name})
	}
	var msg messages.Message
	err := c.RequestMultipart(ctx, "POST", "/channels/"+channelID.String()+"/messages", send, files, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// GetGuild gets a guild by its ID.
func (c *Client) GetGuild(ctx context.Context, guildID snowflake.ID) (*guilds.Guild, error) {
	return c.GetGuildWithOptions(ctx, guildID, false)
}

// GetGuildWithOptions fetches a guild and optionally includes approximate counts.
func (c *Client) GetGuildWithOptions(ctx context.Context, guildID snowflake.ID, withCounts bool) (*guilds.Guild, error) {
	var g guilds.Guild
	path := "/guilds/" + guildID.String()
	if withCounts {
		path += "?with_counts=true"
	}
	err := c.Request(ctx, "GET", path, nil, &g)
	if err != nil {
		return nil, err
	}
	return &g, nil
}
