package rest

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"time"

	"github.com/discord-go/discord.go/channels"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

// GetChannel gets a channel by its ID.
func (c *Client) GetChannel(ctx context.Context, channelID snowflake.ID) (*channels.Channel, error) {
	var ch channels.Channel
	err := c.Request(ctx, "GET", "/channels/"+channelID.String(), nil, &ch)
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

// ModifyChannel modifies a channel's settings.
func (c *Client) ModifyChannel(ctx context.Context, channelID snowflake.ID, params ModifyChannelParams) (*channels.Channel, error) {
	var ch channels.Channel
	err := c.Request(ctx, "PATCH", "/channels/"+channelID.String(), params, &ch)
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

// DeleteChannel deletes a channel.
func (c *Client) DeleteChannel(ctx context.Context, channelID snowflake.ID) error {
	return c.Request(ctx, "DELETE", "/channels/"+channelID.String(), nil, nil)
}

// GetChannelMessages retrieves messages from a channel.
func (c *Client) GetChannelMessages(ctx context.Context, channelID snowflake.ID, params GetMessagesParams) ([]messages.Message, error) {
	path := "/channels/" + channelID.String() + "/messages" + params.QueryString()
	var msgs []messages.Message
	err := c.Request(ctx, "GET", path, nil, &msgs)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

// GetChannelMessage retrieves a single message from a channel.
func (c *Client) GetChannelMessage(ctx context.Context, channelID, messageID snowflake.ID) (*messages.Message, error) {
	var msg messages.Message
	err := c.Request(ctx, "GET", "/channels/"+channelID.String()+"/messages/"+messageID.String(), nil, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// EditMessage edits a previously sent message.
func (c *Client) EditMessage(ctx context.Context, channelID, messageID snowflake.ID, params EditMessageParams) (*messages.Message, error) {
	var msg messages.Message
	err := c.Request(ctx, "PATCH", "/channels/"+channelID.String()+"/messages/"+messageID.String(), params, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// DeleteMessage deletes a message.
func (c *Client) DeleteMessage(ctx context.Context, channelID, messageID snowflake.ID) error {
	return c.Request(ctx, "DELETE", "/channels/"+channelID.String()+"/messages/"+messageID.String(), nil, nil)
}

// BulkDeleteMessages deletes multiple messages at once (2-100 messages).
func (c *Client) BulkDeleteMessages(ctx context.Context, channelID snowflake.ID, messageIDs []snowflake.ID) error {
	body := struct {
		Messages snowflake.IDs `json:"messages"`
	}{Messages: snowflake.IDs(messageIDs)}
	return c.Request(ctx, "POST", "/channels/"+channelID.String()+"/messages/bulk-delete", body, nil)
}

// CreateReaction adds a reaction to a message.
func (c *Client) CreateReaction(ctx context.Context, channelID, messageID snowflake.ID, emoji string) error {
	path := "/channels/" + channelID.String() + "/messages/" + messageID.String() + "/reactions/" + url.PathEscape(emoji) + "/@me"
	return c.Request(ctx, "PUT", path, nil, nil)
}

// DeleteOwnReaction removes the bot's own reaction from a message.
func (c *Client) DeleteOwnReaction(ctx context.Context, channelID, messageID snowflake.ID, emoji string) error {
	path := "/channels/" + channelID.String() + "/messages/" + messageID.String() + "/reactions/" + url.PathEscape(emoji) + "/@me"
	return c.Request(ctx, "DELETE", path, nil, nil)
}

// GetReactions gets the users who reacted with a specific emoji on a message.
func (c *Client) GetReactions(ctx context.Context, channelID, messageID snowflake.ID, emoji string) ([]users.User, error) {
	path := "/channels/" + channelID.String() + "/messages/" + messageID.String() + "/reactions/" + url.PathEscape(emoji)
	var u []users.User
	err := c.Request(ctx, "GET", path, nil, &u)
	if err != nil {
		return nil, err
	}
	return u, nil
}

type GetReactionsParams struct {
	After *snowflake.ID
	Limit int
	Type  int
}

func (p GetReactionsParams) query() string {
	values := url.Values{}
	if p.After != nil {
		values.Set("after", p.After.String())
	}
	if p.Limit > 0 {
		values.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Type != 0 {
		values.Set("type", strconv.Itoa(p.Type))
	}
	if encoded := values.Encode(); encoded != "" {
		return "?" + encoded
	}
	return ""
}

func (c *Client) GetReactionsPage(ctx context.Context, channelID, messageID snowflake.ID, emoji string, params GetReactionsParams) ([]users.User, error) {
	path := "/channels/" + channelID.String() + "/messages/" + messageID.String() + "/reactions/" + url.PathEscape(emoji) + params.query()
	var result []users.User
	err := c.Request(ctx, "GET", path, nil, &result)
	return result, err
}

func (c *Client) DeleteUserReaction(ctx context.Context, channelID, messageID snowflake.ID, emoji string, userID snowflake.ID) error {
	path := "/channels/" + channelID.String() + "/messages/" + messageID.String() + "/reactions/" + url.PathEscape(emoji) + "/" + userID.String()
	return c.Request(ctx, "DELETE", path, nil, nil)
}

func (c *Client) DeleteAllReactions(ctx context.Context, channelID, messageID snowflake.ID) error {
	return c.Request(ctx, "DELETE", "/channels/"+channelID.String()+"/messages/"+messageID.String()+"/reactions", nil, nil)
}

func (c *Client) DeleteAllReactionsForEmoji(ctx context.Context, channelID, messageID snowflake.ID, emoji string) error {
	path := "/channels/" + channelID.String() + "/messages/" + messageID.String() + "/reactions/" + url.PathEscape(emoji)
	return c.Request(ctx, "DELETE", path, nil, nil)
}

func (c *Client) CrosspostMessage(ctx context.Context, channelID, messageID snowflake.ID) (*messages.Message, error) {
	var message messages.Message
	err := c.Request(ctx, "POST", "/channels/"+channelID.String()+"/messages/"+messageID.String()+"/crosspost", nil, &message)
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// TriggerTypingIndicator triggers the typing indicator in a channel.
func (c *Client) TriggerTypingIndicator(ctx context.Context, channelID snowflake.ID) error {
	return c.Request(ctx, "POST", "/channels/"+channelID.String()+"/typing", nil, nil)
}

// GetPinnedMessages retrieves the pinned messages in a channel.
func (c *Client) GetPinnedMessages(ctx context.Context, channelID snowflake.ID) ([]messages.Message, error) {
	var msgs []messages.Message
	err := c.Request(ctx, "GET", "/channels/"+channelID.String()+"/messages/pins", nil, &msgs)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

type ChannelMessagePin struct {
	PinnedAt time.Time        `json:"pinned_at"`
	Message  messages.Message `json:"message"`
}

type ChannelMessagePins struct {
	Items   []ChannelMessagePin `json:"items"`
	HasMore bool                `json:"has_more"`
}

type GetPinnedMessagesParams struct {
	Before *snowflake.ID
	Limit  int
}

func (p GetPinnedMessagesParams) query() string {
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

func (c *Client) GetChannelMessagePins(ctx context.Context, channelID snowflake.ID, params GetPinnedMessagesParams) (*ChannelMessagePins, error) {
	var result ChannelMessagePins
	err := c.Request(ctx, "GET", "/channels/"+channelID.String()+"/messages/pins"+params.query(), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// PinMessage pins a message in a channel.
func (c *Client) PinMessage(ctx context.Context, channelID, messageID snowflake.ID) error {
	return c.Request(ctx, "PUT", "/channels/"+channelID.String()+"/messages/pins/"+messageID.String(), nil, nil)
}

// UnpinMessage unpins a message in a channel.
func (c *Client) UnpinMessage(ctx context.Context, channelID, messageID snowflake.ID) error {
	return c.Request(ctx, "DELETE", "/channels/"+channelID.String()+"/messages/pins/"+messageID.String(), nil, nil)
}

// GetChannelInvites gets the invites for a channel.
func (c *Client) GetChannelInvites(ctx context.Context, channelID snowflake.ID) ([]channels.Invite, error) {
	var invites []channels.Invite
	err := c.Request(ctx, "GET", "/channels/"+channelID.String()+"/invites", nil, &invites)
	if err != nil {
		return nil, err
	}
	return invites, nil
}

// CreateChannelInvite creates a new invite for a channel.
func (c *Client) CreateChannelInvite(ctx context.Context, channelID snowflake.ID, params CreateInviteParams) (*channels.Invite, error) {
	var invite channels.Invite
	err := c.Request(ctx, "POST", "/channels/"+channelID.String()+"/invites", params, &invite)
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

type FollowedChannel struct {
	ChannelID snowflake.ID `json:"channel_id,string"`
	WebhookID snowflake.ID `json:"webhook_id,string"`
}

type FollowChannelParams struct {
	WebhookChannelID snowflake.ID `json:"webhook_channel_id,string"`
}

func (c *Client) FollowNewsChannel(ctx context.Context, channelID snowflake.ID, params FollowChannelParams) (*FollowedChannel, error) {
	var result FollowedChannel
	err := c.Request(ctx, "POST", "/channels/"+channelID.String()+"/followers", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

type SendSoundboardSoundParams struct {
	SoundID       snowflake.ID  `json:"sound_id,string"`
	SourceGuildID *snowflake.ID `json:"source_guild_id,string,omitempty"`
}

func (c *Client) SendSoundboardSound(ctx context.Context, channelID snowflake.ID, params SendSoundboardSoundParams) error {
	return c.Request(ctx, "POST", "/channels/"+channelID.String()+"/send-soundboard-sound", params, nil)
}

type VoiceChannelStatusUpdateParams struct {
	Status string `json:"status"`
}

func (c *Client) ModifyVoiceChannelStatus(ctx context.Context, channelID snowflake.ID, params VoiceChannelStatusUpdateParams) error {
	return c.Request(ctx, "PUT", "/channels/"+channelID.String()+"/voice-status", params, nil)
}

// GetAnswerVoters gets the users who voted for a specific answer in a poll.
func (c *Client) GetAnswerVoters(ctx context.Context, channelID snowflake.ID, messageID snowflake.ID, answerID int, params GetAnswerVotersParams) ([]users.User, error) {
	result, err := c.GetAnswerVotersResponse(ctx, channelID, messageID, answerID, params)
	if err != nil {
		return nil, err
	}
	return result.Users, nil
}

type PollAnswerDetailsResponse struct {
	Users []users.User `json:"users"`
}

func (p *PollAnswerDetailsResponse) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '[' {
		return json.Unmarshal(data, &p.Users)
	}
	type alias PollAnswerDetailsResponse
	return json.Unmarshal(data, (*alias)(p))
}

func (c *Client) GetAnswerVotersResponse(ctx context.Context, channelID, messageID snowflake.ID, answerID int, params GetAnswerVotersParams) (*PollAnswerDetailsResponse, error) {
	path := "/channels/" + channelID.String() + "/polls/" + messageID.String() + "/answers/" + itoa(answerID) + params.QueryString()
	var result PollAnswerDetailsResponse
	err := c.Request(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// EndPoll ends a poll immediately.
func (c *Client) EndPoll(ctx context.Context, channelID snowflake.ID, messageID snowflake.ID) (*messages.Message, error) {
	var msg messages.Message
	err := c.Request(ctx, "POST", "/channels/"+channelID.String()+"/polls/"+messageID.String()+"/expire", nil, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}
