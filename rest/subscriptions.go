package rest

import (
	"context"
	"net/url"
	"strconv"

	"github.com/discord-go/discord.go/snowflake"
)

type Subscription struct {
	ID                 snowflake.ID  `json:"id,string"`
	UserID             snowflake.ID  `json:"user_id,string"`
	SKU                snowflake.ID  `json:"sku_id,string"`
	ApplicationID      snowflake.ID  `json:"application_id,string"`
	Status             int           `json:"status"`
	CurrentPeriodStart string        `json:"current_period_start"`
	CurrentPeriodEnd   string        `json:"current_period_end"`
	CanceledAt         *string       `json:"canceled_at,omitempty"`
	Country            string        `json:"country,omitempty"`
	PaymentSourceID    *snowflake.ID `json:"payment_source_id,string,omitempty"`
}

type SubscriptionListParams struct {
	Before *snowflake.ID
	After  *snowflake.ID
	Limit  int
	UserID *snowflake.ID
}

func (p SubscriptionListParams) query() string {
	values := url.Values{}
	if p.Before != nil {
		values.Set("before", p.Before.String())
	}
	if p.After != nil {
		values.Set("after", p.After.String())
	}
	if p.Limit > 0 {
		values.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.UserID != nil {
		values.Set("user_id", p.UserID.String())
	}
	if encoded := values.Encode(); encoded != "" {
		return "?" + encoded
	}
	return ""
}

func (c *Client) ListSKUSubscriptions(ctx context.Context, skuID snowflake.ID, params SubscriptionListParams) ([]Subscription, error) {
	var result []Subscription
	err := c.Request(ctx, "GET", "/skus/"+skuID.String()+"/subscriptions"+params.query(), nil, &result)
	return result, err
}

func (c *Client) GetSKUSubscription(ctx context.Context, skuID, subscriptionID snowflake.ID, userID *snowflake.ID) (*Subscription, error) {
	values := url.Values{}
	if userID != nil {
		values.Set("user_id", userID.String())
	}
	path := "/skus/" + skuID.String() + "/subscriptions/" + subscriptionID.String()
	if encoded := values.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var result Subscription
	err := c.Request(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
