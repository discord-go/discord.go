package roles

import (
	"encoding/json"

	"github.com/discord-go/discord.go/snowflake"
)

// RoleTags represents tags related to a role.
// https://discord.com/developers/docs/topics/permissions#role-object-role-tags-structure
type RoleTags struct {
	BotID                 *snowflake.ID `json:"bot_id,string,omitempty"`
	IntegrationID         *snowflake.ID `json:"integration_id,string,omitempty"`
	PremiumSubscriber     bool          `json:"-"`
	SubscriptionListingID *snowflake.ID `json:"subscription_listing_id,string,omitempty"`
	AvailableForPurchase  bool          `json:"-"`
	GuildConnections      bool          `json:"-"`
}

// UnmarshalJSON unmarshals the JSON data into a RoleTags object, properly handling
// the presence-based null fields.
func (t *RoleTags) UnmarshalJSON(b []byte) error {
	type roleTags RoleTags
	var tags roleTags
	if err := json.Unmarshal(b, &tags); err != nil {
		return err
	}
	*t = RoleTags(tags)

	var m map[string]json.RawMessage
	_ = json.Unmarshal(b, &m)

	_, t.PremiumSubscriber = m["premium_subscriber"]
	_, t.AvailableForPurchase = m["available_for_purchase"]
	_, t.GuildConnections = m["guild_connections"]

	return nil
}
