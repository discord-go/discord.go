package rest

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/discord-go/discord.go/snowflake"
)

type ListEntitlementsParams struct {
	UserID         *snowflake.ID
	SKUIds         []snowflake.ID
	GuildID        *snowflake.ID
	Before         *snowflake.ID
	After          *snowflake.ID
	Limit          int
	ExcludeEnded   bool
	ExcludeDeleted bool
	OnlyActive     bool
}

func (p ListEntitlementsParams) query() string {
	values := url.Values{}
	if p.UserID != nil {
		values.Set("user_id", p.UserID.String())
	}
	if len(p.SKUIds) > 0 {
		ids := make([]string, len(p.SKUIds))
		for i, id := range p.SKUIds {
			ids[i] = id.String()
		}
		values.Set("sku_ids", strings.Join(ids, ","))
	}
	if p.GuildID != nil {
		values.Set("guild_id", p.GuildID.String())
	}
	if p.Before != nil {
		values.Set("before", p.Before.String())
	}
	if p.After != nil {
		values.Set("after", p.After.String())
	}
	if p.Limit > 0 {
		values.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.ExcludeEnded {
		values.Set("exclude_ended", "true")
	}
	if p.ExcludeDeleted {
		values.Set("exclude_deleted", "true")
	}
	if p.OnlyActive {
		values.Set("only_active", "true")
	}
	if encoded := values.Encode(); encoded != "" {
		return "?" + encoded
	}
	return ""
}

type SKU struct {
	ID            snowflake.ID `json:"id,string"`
	Type          int          `json:"type"`
	ApplicationID snowflake.ID `json:"application_id,string"`
	Name          string       `json:"name"`
	Slug          string       `json:"slug"`
	Flags         int          `json:"flags"`
}

type Entitlement struct {
	ID            snowflake.ID `json:"id,string"`
	SKUID         snowflake.ID `json:"sku_id,string"`
	ApplicationID snowflake.ID `json:"application_id,string"`
	UserID        snowflake.ID `json:"user_id,string,omitempty"`
	Type          int          `json:"type"`
	Deleted       bool         `json:"deleted"`
	StartsAt      *string      `json:"starts_at,omitempty"`
	EndsAt        *string      `json:"ends_at,omitempty"`
	GuildID       snowflake.ID `json:"guild_id,string,omitempty"`
}

type CreateTestEntitlementParams struct {
	SKUID     snowflake.ID `json:"sku_id,string"`
	OwnerID   snowflake.ID `json:"owner_id,string"`
	OwnerType int          `json:"owner_type"`
}

func (c *Client) ListSKUs(ctx context.Context, applicationID snowflake.ID) ([]SKU, error) {
	var skus []SKU
	err := c.Request(ctx, "GET", "/applications/"+applicationID.String()+"/skus", nil, &skus)
	return skus, err
}

func (c *Client) ListEntitlements(ctx context.Context, applicationID snowflake.ID) ([]Entitlement, error) {
	return c.ListEntitlementsWithParams(ctx, applicationID, ListEntitlementsParams{})
}

func (c *Client) ListEntitlementsWithParams(ctx context.Context, applicationID snowflake.ID, params ListEntitlementsParams) ([]Entitlement, error) {
	var entitlements []Entitlement
	err := c.Request(ctx, "GET", "/applications/"+applicationID.String()+"/entitlements"+params.query(), nil, &entitlements)
	return entitlements, err
}

func (c *Client) GetEntitlement(ctx context.Context, applicationID snowflake.ID, entitlementID snowflake.ID) (*Entitlement, error) {
	var entitlement Entitlement
	err := c.Request(ctx, "GET", "/applications/"+applicationID.String()+"/entitlements/"+entitlementID.String(), nil, &entitlement)
	if err != nil {
		return nil, err
	}
	return &entitlement, nil
}

func (c *Client) CreateTestEntitlement(ctx context.Context, applicationID snowflake.ID, params CreateTestEntitlementParams) (*Entitlement, error) {
	var entitlement Entitlement
	err := c.Request(ctx, "POST", "/applications/"+applicationID.String()+"/entitlements", params, &entitlement)
	if err != nil {
		return nil, err
	}
	return &entitlement, nil
}

func (c *Client) DeleteTestEntitlement(ctx context.Context, applicationID snowflake.ID, entitlementID snowflake.ID) error {
	return c.Request(ctx, "DELETE", "/applications/"+applicationID.String()+"/entitlements/"+entitlementID.String(), nil, nil)
}

func (c *Client) ConsumeEntitlement(ctx context.Context, applicationID snowflake.ID, entitlementID snowflake.ID) error {
	return c.Request(ctx, "POST", "/applications/"+applicationID.String()+"/entitlements/"+entitlementID.String()+"/consume", nil, nil)
}
