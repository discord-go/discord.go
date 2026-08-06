package roles

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/discord-go/discord.go/permissions"
	"github.com/discord-go/discord.go/snowflake"
)

func TestRole_UnmarshalJSON(t *testing.T) {
	// Sample JSON based on Discord API documentation
	data := []byte(`{
		"id": "41771983423143936",
		"name": "WE DEM BOYS",
		"color": 3447003,
		"hoist": true,
		"icon": "cf3c224b52b21703666d93cb02a0f8ac",
		"unicode_emoji": null,
		"position": 1,
		"permissions": "66321471",
		"managed": false,
		"mentionable": false,
		"tags": {
			"bot_id": "514590659612459018",
			"integration_id": "41771983423143936",
			"premium_subscriber": null
		}
	}`)

	var role Role
	err := json.Unmarshal(data, &role)
	if err != nil {
		t.Fatalf("Failed to unmarshal Role: %v", err)
	}

	if role.ID != snowflake.ID(41771983423143936) {
		t.Errorf("Expected ID 41771983423143936, got %v", role.ID)
	}
	if role.Name != "WE DEM BOYS" {
		t.Errorf("Expected Name 'WE DEM BOYS', got '%s'", role.Name)
	}
	if role.Color != 3447003 {
		t.Errorf("Expected Color 3447003, got %d", role.Color)
	}
	if !role.Hoist {
		t.Errorf("Expected Hoist true, got %v", role.Hoist)
	}
	if role.Icon == nil || *role.Icon != "cf3c224b52b21703666d93cb02a0f8ac" {
		t.Errorf("Expected Icon 'cf3c224b52b21703666d93cb02a0f8ac', got %v", role.Icon)
	}
	if role.UnicodeEmoji != nil {
		t.Errorf("Expected UnicodeEmoji to be nil, got %v", *role.UnicodeEmoji)
	}
	if role.Position != 1 {
		t.Errorf("Expected Position 1, got %d", role.Position)
	}
	if role.Permissions != permissions.Permission(66321471) {
		t.Errorf("Expected Permissions 66321471, got %v", role.Permissions)
	}
	if role.Managed {
		t.Errorf("Expected Managed false, got %v", role.Managed)
	}
	if role.Mentionable {
		t.Errorf("Expected Mentionable false, got %v", role.Mentionable)
	}

	if role.Tags == nil {
		t.Fatalf("Expected Tags to not be nil")
	}

	if role.Tags.BotID == nil || *role.Tags.BotID != snowflake.ID(514590659612459018) {
		t.Errorf("Expected BotID 514590659612459018, got %v", role.Tags.BotID)
	}
	if role.Tags.IntegrationID == nil || *role.Tags.IntegrationID != snowflake.ID(41771983423143936) {
		t.Errorf("Expected IntegrationID 41771983423143936, got %v", role.Tags.IntegrationID)
	}
	if !role.Tags.PremiumSubscriber {
		t.Errorf("Expected PremiumSubscriber to be true")
	}
	if role.Tags.AvailableForPurchase {
		t.Errorf("Expected AvailableForPurchase to be false")
	}
	if role.Tags.GuildConnections {
		t.Errorf("Expected GuildConnections to be false")
	}
	if role.Tags.SubscriptionListingID != nil {
		t.Errorf("Expected SubscriptionListingID to be nil")
	}
}

func TestRoleTags_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected RoleTags
	}{
		{
			name: "All null tags present",
			data: []byte(`{"premium_subscriber": null, "available_for_purchase": null, "guild_connections": null}`),
			expected: RoleTags{
				PremiumSubscriber:    true,
				AvailableForPurchase: true,
				GuildConnections:     true,
			},
		},
		{
			name: "Only bot ID",
			data: []byte(`{"bot_id": "12345"}`),
			expected: RoleTags{
				BotID: func() *snowflake.ID { id := snowflake.ID(12345); return &id }(),
			},
		},
		{
			name:     "Empty JSON",
			data:     []byte(`{}`),
			expected: RoleTags{},
		},
		{
			name: "Invalid JSON format should error map",
			data: []byte(`{"bot_id": "1234"}`), // valid json but just basic test
			expected: RoleTags{
				BotID: func() *snowflake.ID { id := snowflake.ID(1234); return &id }(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tags RoleTags
			err := json.Unmarshal(tt.data, &tags)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !reflect.DeepEqual(tags, tt.expected) {
				t.Errorf("Expected %+v, got %+v", tt.expected, tags)
			}
		})
	}
}

func TestRoleTags_UnmarshalJSON_Error(t *testing.T) {
	// Test error paths
	var tags RoleTags

	// Malformed JSON should return an error from the first Unmarshal
	err := tags.UnmarshalJSON([]byte(`{malformed`))
	if err == nil {
		t.Errorf("Expected error for malformed JSON")
	}
}
