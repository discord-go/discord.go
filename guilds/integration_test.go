package guilds

import (
	"encoding/json"
	"testing"
)

func TestIntegrationUnmarshalJSON(t *testing.T) {
	data := []byte(`{
		"id": "123456789012345678",
		"name": "Integration Name",
		"type": "twitch",
		"enabled": true,
		"account": {
			"id": "12345",
			"name": "Twitch User"
		},
		"role_id": "987654321098765432",
		"application": {
			"id": "111111111111111111",
			"name": "Application Name",
			"description": "App description"
		}
	}`)

	var i Integration
	if err := json.Unmarshal(data, &i); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if i.ID.String() != "123456789012345678" {
		t.Errorf("expected 123456789012345678, got %v", i.ID)
	}
	if i.RoleID == nil || i.RoleID.String() != "987654321098765432" {
		t.Errorf("expected 987654321098765432, got %v", i.RoleID)
	}
	if i.Application == nil || i.Application.ID.String() != "111111111111111111" {
		t.Errorf("expected application ID 111111111111111111, got %v", i.Application)
	}

	// Test invalid JSON syntax for Integration
	if err := i.UnmarshalJSON([]byte(`{`)); err == nil {
		t.Error("expected error for invalid JSON syntax")
	}

	// Test invalid IDs inside JSON
	if err := json.Unmarshal([]byte(`{"id": "invalid"}`), &i); err == nil {
		t.Error("expected error for invalid id")
	}
	if err := json.Unmarshal([]byte(`{"role_id": "invalid"}`), &i); err == nil {
		t.Error("expected error for invalid role_id")
	}

	// Test IntegrationApplication separately
	var app IntegrationApplication
	appData := []byte(`{
		"id": "222222222222222222",
		"name": "App 2",
		"description": "Desc"
	}`)
	if err := json.Unmarshal(appData, &app); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app.ID.String() != "222222222222222222" {
		t.Errorf("expected 222222222222222222, got %v", app.ID)
	}

	// Test invalid Application JSON syntax
	if err := app.UnmarshalJSON([]byte(`{`)); err == nil {
		t.Error("expected error for invalid JSON syntax")
	}
	if err := json.Unmarshal([]byte(`{"id": "invalid"}`), &app); err == nil {
		t.Error("expected error for invalid id")
	}
}
