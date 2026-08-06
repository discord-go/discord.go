package rest

import (
	"context"
	"net/url"
	"strconv"

	"github.com/discord-go/discord.go/channels"
	"github.com/discord-go/discord.go/snowflake"
)

// ActiveThreads represents the response from List Active Threads.
type ActiveThreads struct {
	Threads []channels.Channel      `json:"threads"`
	Members []channels.ThreadMember `json:"members"`
}

type ArchivedThreads struct {
	Threads []channels.Channel      `json:"threads"`
	Members []channels.ThreadMember `json:"members"`
	HasMore bool                    `json:"has_more"`
}

type ThreadMembersParams struct {
	WithMember bool
	After      *snowflake.ID
	Limit      int
}

func (p ThreadMembersParams) query() string {
	values := url.Values{}
	if p.WithMember {
		values.Set("with_member", "true")
	}
	if p.After != nil {
		values.Set("after", p.After.String())
	}
	if p.Limit > 0 {
		values.Set("limit", strconv.Itoa(p.Limit))
	}
	if encoded := values.Encode(); encoded != "" {
		return "?" + encoded
	}
	return ""
}

type ArchivedThreadsParams struct {
	Before *snowflake.ID
	Limit  int
}

func (p ArchivedThreadsParams) query() string {
	values := url.Values{}
	if p.Before != nil {
		values.Set("before", p.Before.String())
	}
	if p.Limit > 0 {
		values.Set("limit", strconv.Itoa(p.Limit))
	}
	if encoded := values.Encode(); encoded != "" {
		return "?" + encoded
	}
	return ""
}

// StartThreadWithMessage creates a new thread from an existing message.
func (c *Client) StartThreadWithMessage(ctx context.Context, channelID, messageID snowflake.ID, params StartThreadWithMessageParams) (*channels.Channel, error) {
	var thread channels.Channel
	err := c.Request(ctx, "POST", "/channels/"+channelID.String()+"/messages/"+messageID.String()+"/threads", params, &thread)
	if err != nil {
		return nil, err
	}
	return &thread, nil
}

// StartThread creates a new thread that is not connected to an existing message.
func (c *Client) StartThread(ctx context.Context, channelID snowflake.ID, params StartThreadParams) (*channels.Channel, error) {
	var thread channels.Channel
	err := c.Request(ctx, "POST", "/channels/"+channelID.String()+"/threads", params, &thread)
	if err != nil {
		return nil, err
	}
	return &thread, nil
}

// JoinThread adds the current user to a thread.
func (c *Client) JoinThread(ctx context.Context, threadID snowflake.ID) error {
	return c.Request(ctx, "PUT", "/channels/"+threadID.String()+"/thread-members/@me", nil, nil)
}

// AddThreadMember adds another member to a thread.
func (c *Client) AddThreadMember(ctx context.Context, threadID, userID snowflake.ID) error {
	return c.Request(ctx, "PUT", "/channels/"+threadID.String()+"/thread-members/"+userID.String(), nil, nil)
}

func (c *Client) LeaveThread(ctx context.Context, threadID snowflake.ID) error {
	return c.Request(ctx, "DELETE", "/channels/"+threadID.String()+"/thread-members/@me", nil, nil)
}

func (c *Client) RemoveThreadMember(ctx context.Context, threadID, userID snowflake.ID) error {
	return c.Request(ctx, "DELETE", "/channels/"+threadID.String()+"/thread-members/"+userID.String(), nil, nil)
}

func (c *Client) GetThreadMember(ctx context.Context, threadID, userID snowflake.ID, params ThreadMembersParams) (*channels.ThreadMember, error) {
	var member channels.ThreadMember
	err := c.Request(ctx, "GET", "/channels/"+threadID.String()+"/thread-members/"+userID.String()+params.query(), nil, &member)
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (c *Client) ListThreadMembers(ctx context.Context, threadID snowflake.ID, params ThreadMembersParams) ([]channels.ThreadMember, error) {
	var members []channels.ThreadMember
	err := c.Request(ctx, "GET", "/channels/"+threadID.String()+"/thread-members"+params.query(), nil, &members)
	return members, err
}

func (c *Client) ListPublicArchivedThreads(ctx context.Context, channelID snowflake.ID, params ArchivedThreadsParams) (*ArchivedThreads, error) {
	var result ArchivedThreads
	err := c.Request(ctx, "GET", "/channels/"+channelID.String()+"/threads/archived/public"+params.query(), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) ListPrivateArchivedThreads(ctx context.Context, channelID snowflake.ID, params ArchivedThreadsParams) (*ArchivedThreads, error) {
	var result ArchivedThreads
	err := c.Request(ctx, "GET", "/channels/"+channelID.String()+"/threads/archived/private"+params.query(), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) ListJoinedPrivateArchivedThreads(ctx context.Context, channelID snowflake.ID, params ArchivedThreadsParams) (*ArchivedThreads, error) {
	var result ArchivedThreads
	err := c.Request(ctx, "GET", "/channels/"+channelID.String()+"/users/@me/threads/archived/private"+params.query(), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListActiveThreads returns all active threads in the guild.
func (c *Client) ListActiveThreads(ctx context.Context, guildID snowflake.ID) (*ActiveThreads, error) {
	var active ActiveThreads
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/threads/active", nil, &active)
	if err != nil {
		return nil, err
	}
	return &active, nil
}
