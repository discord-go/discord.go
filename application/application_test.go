package application

import (
	"encoding/json"
	"testing"
)

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
