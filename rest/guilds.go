package rest

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/discord-go/discord.go/auditlog"
	"github.com/discord-go/discord.go/channels"
	"github.com/discord-go/discord.go/guilds"
	"github.com/discord-go/discord.go/roles"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

// GuildBan represents a guild ban.
type GuildBan struct {
	Reason string      `json:"reason"`
	User   *users.User `json:"user"`
}

type ListGuildBansParams struct {
	Before *snowflake.ID
	After  *snowflake.ID
	Limit  int
}

func (p ListGuildBansParams) query() string {
	query := url.Values{}
	if p.Before != nil {
		query.Set("before", p.Before.String())
	}
	if p.After != nil {
		query.Set("after", p.After.String())
	}
	if p.Limit > 0 {
		query.Set("limit", strconv.Itoa(p.Limit))
	}
	if encoded := query.Encode(); encoded != "" {
		return "?" + encoded
	}
	return ""
}

// CreateGuild creates a new guild.
func (c *Client) CreateGuild(ctx context.Context, params CreateGuildParams) (*guilds.Guild, error) {
	var g guilds.Guild
	err := c.Request(ctx, "POST", "/guilds", params, &g)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

// ModifyGuild modifies a guild.
func (c *Client) ModifyGuild(ctx context.Context, guildID snowflake.ID, params ModifyGuildParams) (*guilds.Guild, error) {
	var g guilds.Guild
	err := c.Request(ctx, "PATCH", "/guilds/"+guildID.String(), params, &g)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

// DeleteGuild deletes a guild permanently. User must be owner.
func (c *Client) DeleteGuild(ctx context.Context, guildID snowflake.ID) error {
	return c.Request(ctx, "DELETE", "/guilds/"+guildID.String(), nil, nil)
}

// GetGuildChannels returns a list of guild channel objects.
func (c *Client) GetGuildChannels(ctx context.Context, guildID snowflake.ID) ([]channels.Channel, error) {
	var chs []channels.Channel
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/channels", nil, &chs)
	if err != nil {
		return nil, err
	}
	return chs, nil
}

// CreateGuildChannel creates a new channel object for the guild.
func (c *Client) CreateGuildChannel(ctx context.Context, guildID snowflake.ID, params CreateGuildChannelParams) (*channels.Channel, error) {
	var ch channels.Channel
	err := c.Request(ctx, "POST", "/guilds/"+guildID.String()+"/channels", params, &ch)
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

// GetGuildMember returns a guild member object for the specified user.
func (c *Client) GetGuildMember(ctx context.Context, guildID, userID snowflake.ID) (*users.Member, error) {
	var m users.Member
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/members/"+userID.String(), nil, &m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// ListGuildMembers returns a list of guild member objects that are members of the guild.
func (c *Client) ListGuildMembers(ctx context.Context, guildID snowflake.ID, params ListMembersParams) ([]users.Member, error) {
	query := url.Values{}
	if params.Limit > 0 {
		query.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.After != nil {
		query.Set("after", params.After.String())
	}

	path := "/guilds/" + guildID.String() + "/members"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	var m []users.Member
	err := c.Request(ctx, "GET", path, nil, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// ModifyGuildMember modifies attributes of a guild member.
func (c *Client) ModifyGuildMember(ctx context.Context, guildID, userID snowflake.ID, params ModifyMemberParams) (*users.Member, error) {
	var m users.Member
	err := c.Request(ctx, "PATCH", "/guilds/"+guildID.String()+"/members/"+userID.String(), params, &m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// RemoveGuildMember removes a member from a guild.
func (c *Client) RemoveGuildMember(ctx context.Context, guildID, userID snowflake.ID) error {
	return c.Request(ctx, "DELETE", "/guilds/"+guildID.String()+"/members/"+userID.String(), nil, nil)
}

// GetGuildBans returns a list of ban objects for the users banned from this guild.
func (c *Client) GetGuildBans(ctx context.Context, guildID snowflake.ID) ([]GuildBan, error) {
	return c.GetGuildBansWithParams(ctx, guildID, ListGuildBansParams{})
}

func (c *Client) GetGuildBansWithParams(ctx context.Context, guildID snowflake.ID, params ListGuildBansParams) ([]GuildBan, error) {
	var bans []GuildBan
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/bans"+params.query(), nil, &bans)
	if err != nil {
		return nil, err
	}
	return bans, nil
}

// GetGuildBan returns a ban object for the given user or a 404 not found if the ban cannot be found.
func (c *Client) GetGuildBan(ctx context.Context, guildID, userID snowflake.ID) (*GuildBan, error) {
	var ban GuildBan
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/bans/"+userID.String(), nil, &ban)
	if err != nil {
		return nil, err
	}
	return &ban, nil
}

// CreateGuildBan creates a guild ban, and optionally deletes previous messages sent by the banned user.
func (c *Client) CreateGuildBan(ctx context.Context, guildID, userID snowflake.ID, params CreateBanParams) error {
	return c.Request(ctx, "PUT", "/guilds/"+guildID.String()+"/bans/"+userID.String(), params, nil)
}

// RemoveGuildBan removes the ban for a user.
func (c *Client) RemoveGuildBan(ctx context.Context, guildID, userID snowflake.ID) error {
	return c.Request(ctx, "DELETE", "/guilds/"+guildID.String()+"/bans/"+userID.String(), nil, nil)
}

// GetGuildRoles returns a list of role objects for the guild.
func (c *Client) GetGuildRoles(ctx context.Context, guildID snowflake.ID) ([]roles.Role, error) {
	var r []roles.Role
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/roles", nil, &r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// CreateGuildRole creates a new role for the guild.
func (c *Client) CreateGuildRole(ctx context.Context, guildID snowflake.ID, params CreateRoleParams) (*roles.Role, error) {
	var r roles.Role
	err := c.Request(ctx, "POST", "/guilds/"+guildID.String()+"/roles", params, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// ModifyGuildRole modifies a guild role.
func (c *Client) ModifyGuildRole(ctx context.Context, guildID, roleID snowflake.ID, params ModifyRoleParams) (*roles.Role, error) {
	var r roles.Role
	err := c.Request(ctx, "PATCH", "/guilds/"+guildID.String()+"/roles/"+roleID.String(), params, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// DeleteGuildRole deletes a guild role.
func (c *Client) DeleteGuildRole(ctx context.Context, guildID, roleID snowflake.ID) error {
	return c.Request(ctx, "DELETE", "/guilds/"+guildID.String()+"/roles/"+roleID.String(), nil, nil)
}

// GetGuildPruneCount returns the number of members that would be removed in a prune operation.
func (c *Client) GetGuildPruneCount(ctx context.Context, guildID snowflake.ID, params PruneParams) (int, error) {
	query := url.Values{}
	if params.Days > 0 {
		query.Set("days", strconv.Itoa(params.Days))
	}
	if len(params.IncludeRoles) > 0 {
		var roleStrs []string
		for _, r := range params.IncludeRoles {
			roleStrs = append(roleStrs, r.String())
		}
		query.Set("include_roles", strings.Join(roleStrs, ","))
	}

	path := "/guilds/" + guildID.String() + "/prune"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	var resp struct {
		Pruned int `json:"pruned"`
	}
	err := c.Request(ctx, "GET", path, nil, &resp)
	if err != nil {
		return 0, err
	}
	return resp.Pruned, nil
}

// BeginGuildPrune begins a member prune operation.
func (c *Client) BeginGuildPrune(ctx context.Context, guildID snowflake.ID, params PruneParams) (int, error) {
	var resp struct {
		Pruned int `json:"pruned"`
	}
	err := c.Request(ctx, "POST", "/guilds/"+guildID.String()+"/prune", params, &resp)
	if err != nil {
		return 0, err
	}
	return resp.Pruned, nil
}

// GetGuildInvites returns a list of invite objects (with invite metadata) for the guild.
func (c *Client) GetGuildInvites(ctx context.Context, guildID snowflake.ID) ([]channels.Invite, error) {
	var inv []channels.Invite
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/invites", nil, &inv)
	if err != nil {
		return nil, err
	}
	return inv, nil
}

// GetGuildIntegrations returns a list of integration objects for the guild.
func (c *Client) GetGuildIntegrations(ctx context.Context, guildID snowflake.ID) ([]guilds.Integration, error) {
	var integ []guilds.Integration
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/integrations", nil, &integ)
	if err != nil {
		return nil, err
	}
	return integ, nil
}

// GetGuildWidget returns a guild widget object.
func (c *Client) GetGuildWidget(ctx context.Context, guildID snowflake.ID) (*guilds.Widget, error) {
	var w guilds.Widget
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/widget.json", nil, &w)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

// GetAuditLog returns an audit log object for the guild.
func (c *Client) GetAuditLog(ctx context.Context, guildID snowflake.ID, params AuditLogParams) (*auditlog.AuditLog, error) {
	query := url.Values{}
	if params.UserID != nil {
		query.Set("user_id", params.UserID.String())
	}
	if params.ActionType != nil {
		query.Set("action_type", strconv.Itoa(*params.ActionType))
	}
	if params.Before != nil {
		query.Set("before", params.Before.String())
	}
	if params.After != nil {
		query.Set("after", params.After.String())
	}
	if params.Limit > 0 {
		query.Set("limit", strconv.Itoa(params.Limit))
	}

	path := "/guilds/" + guildID.String() + "/audit-logs"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	var a auditlog.AuditLog
	err := c.Request(ctx, "GET", path, nil, &a)
	if err != nil {
		return nil, err
	}
	return &a, nil
}
