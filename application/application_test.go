package application

import (
	"encoding/json"
	"reflect"
	"strconv"
	"testing"
)

func TestApplicationWebhookEventTypeFormats(t *testing.T) {
	var app Application
	if err := json.Unmarshal([]byte(`{"event_webhooks_types":["APPLICATION_AUTHORIZED","ENTITLEMENT_CREATE"]}`), &app); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(app.EventWebhookTypeNames, []string{"APPLICATION_AUTHORIZED", "ENTITLEMENT_CREATE"}) {
		t.Fatalf("event type names = %v", app.EventWebhookTypeNames)
	}
	encoded, err := json.Marshal(app)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &raw); err != nil {
		t.Fatal(err)
	}
	if string(raw["event_webhooks_types"]) != `["APPLICATION_AUTHORIZED","ENTITLEMENT_CREATE"]` {
		t.Fatalf("encoded event type names = %s", raw["event_webhooks_types"])
	}
	if err := json.Unmarshal([]byte(`{"event_webhooks_types":[1,2]}`), &app); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(app.EventWebhooksTypes, []int{1, 2}) || app.EventWebhookTypeNames != nil {
		t.Fatalf("legacy event types = %v, names = %v", app.EventWebhooksTypes, app.EventWebhookTypeNames)
	}
	encoded, err = json.Marshal(app)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encoded, &raw); err != nil {
		t.Fatal(err)
	}
	if string(raw["event_webhooks_types"]) != `[1,2]` {
		t.Fatalf("encoded legacy types = %s", raw["event_webhooks_types"])
	}
	if err := json.Unmarshal([]byte(`{"event_webhooks_types":null}`), &app); err != nil || app.EventWebhooksTypes != nil || app.EventWebhookTypeNames != nil {
		t.Fatalf("null event types: legacy %v, names %v, error %v", app.EventWebhooksTypes, app.EventWebhookTypeNames, err)
	}
	if err := json.Unmarshal([]byte(`{"event_webhooks_types":[{}]}`), &app); err == nil {
		t.Fatal("expected invalid event type error")
	}
}

func TestApplicationFlagsNewFormats(t *testing.T) {
	for _, tt := range []struct {
		payload string
		want    int
		bad     bool
	}{
		{`{"flags_new":"8192"}`, 8192, false},
		{`{"flags_new":8192}`, 8192, false},
		{`{"flags_new":null}`, 0, false},
		{`{}`, 0, false},
		{`{"flags_new":"invalid"}`, 0, true},
		{`{"flags_new":"18446744073709551616"}`, 0, true},
	} {
		var a Application
		err := json.Unmarshal([]byte(tt.payload), &a)
		if (err != nil) != tt.bad || (!tt.bad && a.FlagsNew != tt.want) {
			t.Fatalf("decode %s: flags %d, error %v", tt.payload, a.FlagsNew, err)
		}
	}
	if strconv.IntSize == 64 {
		var a Application
		if err := json.Unmarshal([]byte(`{"flags_new":"1099511627776"}`), &a); err != nil || int64(a.FlagsNew) != 1<<40 {
			t.Fatalf("high-bit flags = %d, error %v", a.FlagsNew, err)
		}
	}
}

func TestApplicationUnmarshal(t *testing.T) {
	data := []byte(`{
		"id": "123456789012345678",
		"name": "Test App",
		"icon": "icon_hash",
		"description": "A test application",
		"rpc_origins": ["http://localhost:8080"],
		"bot_public": true,
		"bot_require_code_grant": false,
		"bot": {
			"id": "123456789012345679",
			"username": "testbot",
			"discriminator": "1234",
			"avatar": null
		},
		"terms_of_service_url": "https://example.com/tos",
		"privacy_policy_url": "https://example.com/privacy",
		"owner": {
			"id": "123456789012345680",
			"username": "owner",
			"discriminator": "5678",
			"avatar": null
		},
		"verify_key": "verify_key_hash",
		"team": {
			"icon": null,
			"id": "123456789012345681",
			"members": [
				{
					"membership_state": 2,
					"permissions": ["*"],
					"team_id": "123456789012345681",
					"user": {
						"id": "123456789012345680",
						"username": "owner",
						"discriminator": "5678",
						"avatar": null
					},
					"role": "admin"
				}
			],
			"name": "Test Team",
			"owner_user_id": "123456789012345680"
		},
		"guild_id": "123456789012345682",
		"primary_sku_id": "123456789012345683",
		"slug": "test-app",
		"cover_image": "cover_hash",
		"flags": 56,
		"approximate_guild_count": 100,
		"redirect_uris": ["https://example.com/callback"],
		"interactions_endpoint_url": "https://example.com/interactions",
		"role_connections_verification_url": "https://example.com/role-connections",
		"tags": ["test", "bot"],
		"custom_install_url": "https://example.com/install"
	}`)

	var app Application
	err := json.Unmarshal(data, &app)
	if err != nil {
		t.Fatalf("Failed to unmarshal Application: %v", err)
	}

	if app.ID != 123456789012345678 {
		t.Errorf("Expected ID 123456789012345678, got %d", app.ID)
	}
	if app.Name != "Test App" {
		t.Errorf("Expected Name 'Test App', got '%s'", app.Name)
	}
	if app.Team == nil {
		t.Fatalf("Expected Team to not be nil")
	}
	if len(app.Team.Members) != 1 {
		t.Fatalf("Expected 1 team member, got %d", len(app.Team.Members))
	}
	if app.Team.Members[0].Role != "admin" {
		t.Errorf("Expected team member role 'admin', got '%s'", app.Team.Members[0].Role)
	}
	if app.GuildID != 123456789012345682 {
		t.Errorf("Expected GuildID 123456789012345682, got %d", app.GuildID)
	}
}

func TestMembershipState(t *testing.T) {
	if MembershipStateInvited != 1 {
		t.Errorf("Expected MembershipStateInvited to be 1")
	}
	if MembershipStateAccepted != 2 {
		t.Errorf("Expected MembershipStateAccepted to be 2")
	}
}
