package users

import (
	"encoding/json"
	"testing"
)

// TestPresenceUpdateOmittedCustomStatus covers the September 2026 privacy
// change: the custom status (activity type 4) may be left out of a
// PRESENCE_UPDATE payload entirely, so handlers must not require it.
func TestPresenceUpdateOmittedCustomStatus(t *testing.T) {
	raw := `{
		"user": {"id": "1"},
		"guild_id": "2",
		"status": "online",
		"activities": [{"name": "Discord", "type": 0, "created_at": 1756000000}],
		"client_status": {"desktop": "online"}
	}`
	var presence PresenceUpdate
	if err := json.Unmarshal([]byte(raw), &presence); err != nil {
		t.Fatal(err)
	}
	if presence.Status != "online" {
		t.Errorf("status = %q", presence.Status)
	}
	if len(presence.Activities) != 1 {
		t.Fatalf("len(activities) = %d", len(presence.Activities))
	}
	for _, activity := range presence.Activities {
		if activity.Type == 4 {
			t.Error("expected no custom status activity in this payload")
		}
	}
}

// TestPresenceUpdateEmptyActivities covers payloads that arrive with an empty
// or absent activities array after the same privacy change.
func TestPresenceUpdateEmptyActivities(t *testing.T) {
	for _, raw := range []string{
		`{"user":{"id":"1"},"guild_id":"2","status":"dnd","activities":[],"client_status":{}}`,
		`{"user":{"id":"1"},"guild_id":"2","status":"idle","client_status":{}}`,
	} {
		var presence PresenceUpdate
		if err := json.Unmarshal([]byte(raw), &presence); err != nil {
			t.Fatalf("unmarshal %s: %v", raw, err)
		}
		if len(presence.Activities) != 0 {
			t.Errorf("activities = %v", presence.Activities)
		}
		if presence.Status == "" {
			t.Errorf("status missing in %s", raw)
		}
	}
}

// TestPresenceUpdateCustomStatusWhenPresent documents the shape Discord still
// sends when the custom status is included: activity type 4 with the text in
// State and the name "Custom Status".
func TestPresenceUpdateCustomStatusWhenPresent(t *testing.T) {
	raw := `{
		"user": {"id": "1"},
		"guild_id": "2",
		"status": "online",
		"activities": [{"name": "Custom Status", "type": 4, "state": "shipping", "created_at": 1756000000}],
		"client_status": {}
	}`
	var presence PresenceUpdate
	if err := json.Unmarshal([]byte(raw), &presence); err != nil {
		t.Fatal(err)
	}
	if len(presence.Activities) != 1 {
		t.Fatalf("len(activities) = %d", len(presence.Activities))
	}
	custom := presence.Activities[0]
	if custom.Type != 4 {
		t.Errorf("type = %d, want 4", custom.Type)
	}
	if custom.State == nil || *custom.State != "shipping" {
		t.Errorf("state = %v", custom.State)
	}
}
