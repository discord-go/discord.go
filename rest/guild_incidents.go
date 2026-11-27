package rest

import (
	"context"
	"time"

	"github.com/discord-go/discord.go/guilds"
	"github.com/discord-go/discord.go/snowflake"
)

// ModifyGuildIncidentActionsParams sets, clears, or leaves incident actions
// unchanged. A nil outer pointer omits the field, a pointer to nil clears it,
// and a pointer to a non-nil timestamp sets it. Discord allows at most 24
// hours into the future.
type ModifyGuildIncidentActionsParams struct {
	InvitesDisabledUntil **time.Time `json:"invites_disabled_until,omitempty"`
	DMsDisabledUntil     **time.Time `json:"dms_disabled_until,omitempty"`
}

// ModifyGuildIncidentActions updates a guild's incident actions. It requires
// MANAGE_GUILD and returns the resulting incident data.
func (c *Client) ModifyGuildIncidentActions(ctx context.Context, guildID snowflake.ID, params ModifyGuildIncidentActionsParams) (*guilds.IncidentsData, error) {
	var result guilds.IncidentsData
	if err := c.Request(ctx, "PUT", "/guilds/"+guildID.String()+"/incident-actions", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
