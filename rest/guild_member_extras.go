package rest

import (
	"context"
	"net/url"
	"strconv"

	"github.com/discord-go/discord.go/channels"
	"github.com/discord-go/discord.go/roles"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
	"github.com/discord-go/discord.go/voice"
)

type SearchGuildMembersParams struct {
	Query string
	Limit int
}

func (p SearchGuildMembersParams) query() string {
	values := url.Values{}
	if p.Query != "" {
		values.Set("query", p.Query)
	}
	if p.Limit > 0 {
		values.Set("limit", strconv.Itoa(p.Limit))
	}
	if encoded := values.Encode(); encoded != "" {
		return "?" + encoded
	}
	return ""
}

type AddGuildMemberParams struct {
	AccessToken string        `json:"access_token"`
	Nick        string        `json:"nick,omitempty"`
	Roles       snowflake.IDs `json:"roles,omitempty"`
	Mute        bool          `json:"mute,omitempty"`
	Deaf        bool          `json:"deaf,omitempty"`
}

type ModifyCurrentMemberParams struct {
	Nick   *string `json:"nick,omitempty"`
	Avatar *string `json:"avatar,omitempty"`
	Banner *string `json:"banner,omitempty"`
	Bio    *string `json:"bio,omitempty"`
}

type ModifyVoiceStateParams struct {
	ChannelID               *snowflake.ID `json:"channel_id,string,omitempty"`
	Suppress                *bool         `json:"suppress,omitempty"`
	RequestToSpeakTimestamp *string       `json:"request_to_speak_timestamp,omitempty"`
}

type BulkBanParams struct {
	UserIDs              snowflake.IDs `json:"user_ids"`
	DeleteMessageSeconds int           `json:"delete_message_seconds,omitempty"`
}

type BulkBanResult struct {
	BannedUsers snowflake.IDs `json:"banned_users"`
	FailedUsers snowflake.IDs `json:"failed_users"`
}

type GuildChannelPosition struct {
	ID              snowflake.ID  `json:"id,string"`
	Position        *int          `json:"position,omitempty"`
	LockPermissions *bool         `json:"lock_permissions,omitempty"`
	ParentID        *snowflake.ID `json:"parent_id,string,omitempty"`
}

func (c *Client) SearchGuildMembers(ctx context.Context, guildID snowflake.ID, params SearchGuildMembersParams) ([]users.Member, error) {
	var result []users.Member
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/members/search"+params.query(), nil, &result)
	return result, err
}

func (c *Client) AddGuildMember(ctx context.Context, guildID, userID snowflake.ID, params AddGuildMemberParams) (*users.Member, error) {
	var result users.Member
	err := c.Request(ctx, "PUT", "/guilds/"+guildID.String()+"/members/"+userID.String(), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) ModifyCurrentGuildMember(ctx context.Context, guildID snowflake.ID, params ModifyCurrentMemberParams) (*users.Member, error) {
	var result users.Member
	err := c.Request(ctx, "PATCH", "/guilds/"+guildID.String()+"/members/@me", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) AddGuildMemberRole(ctx context.Context, guildID, userID, roleID snowflake.ID) error {
	return c.Request(ctx, "PUT", "/guilds/"+guildID.String()+"/members/"+userID.String()+"/roles/"+roleID.String(), nil, nil)
}

func (c *Client) RemoveGuildMemberRole(ctx context.Context, guildID, userID, roleID snowflake.ID) error {
	return c.Request(ctx, "DELETE", "/guilds/"+guildID.String()+"/members/"+userID.String()+"/roles/"+roleID.String(), nil, nil)
}

func (c *Client) GetGuildRole(ctx context.Context, guildID, roleID snowflake.ID) (*roles.Role, error) {
	var result roles.Role
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/roles/"+roleID.String(), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) ModifyGuildChannelPositions(ctx context.Context, guildID snowflake.ID, positions []GuildChannelPosition) ([]channels.Channel, error) {
	var result []channels.Channel
	err := c.Request(ctx, "PATCH", "/guilds/"+guildID.String()+"/channels", positions, &result)
	return result, err
}

func (c *Client) ModifyGuildRolePositions(ctx context.Context, guildID snowflake.ID, positions []RolePosition) ([]roles.Role, error) {
	var result []roles.Role
	err := c.Request(ctx, "PATCH", "/guilds/"+guildID.String()+"/roles", positions, &result)
	return result, err
}

type RolePosition struct {
	ID       snowflake.ID `json:"id,string"`
	Position *int         `json:"position,omitempty"`
}

func (c *Client) GetGuildRoleMemberCounts(ctx context.Context, guildID snowflake.ID) (map[string]int, error) {
	var result map[string]int
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/roles/member-counts", nil, &result)
	return result, err
}

func (c *Client) BulkBanGuildMembers(ctx context.Context, guildID snowflake.ID, params BulkBanParams) (*BulkBanResult, error) {
	var result BulkBanResult
	err := c.Request(ctx, "POST", "/guilds/"+guildID.String()+"/bulk-ban", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetCurrentGuildVoiceState(ctx context.Context, guildID snowflake.ID) (*voice.VoiceState, error) {
	var result voice.VoiceState
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/voice-states/@me", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetGuildVoiceState(ctx context.Context, guildID, userID snowflake.ID) (*voice.VoiceState, error) {
	var result voice.VoiceState
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/voice-states/"+userID.String(), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) ModifyCurrentGuildVoiceState(ctx context.Context, guildID snowflake.ID, params ModifyVoiceStateParams) error {
	return c.Request(ctx, "PATCH", "/guilds/"+guildID.String()+"/voice-states/@me", params, nil)
}

func (c *Client) ModifyGuildVoiceState(ctx context.Context, guildID, userID snowflake.ID, params ModifyVoiceStateParams) error {
	return c.Request(ctx, "PATCH", "/guilds/"+guildID.String()+"/voice-states/"+userID.String(), params, nil)
}
