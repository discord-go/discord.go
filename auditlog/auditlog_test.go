package auditlog

import (
	"encoding/json"
	"testing"
)

func TestAuditLogJSON(t *testing.T) {
	data := []byte(`{
		"application_commands": [{"id": "123"}],
		"audit_log_entries": [{
			"target_id": "456",
			"changes": [
				{"new_value": "new", "old_value": "old", "key": "name"}
			],
			"user_id": "789",
			"id": "101112",
			"action_type": 1,
			"options": {
				"channel_id": "131415",
				"count": "1"
			},
			"reason": "because"
		}],
		"auto_moderation_rules": [{"id": "161718"}],
		"guild_scheduled_events": [{"id": "192021"}],
		"integrations": [{"id": "222324"}],
		"threads": [{"id": "252627"}],
		"users": [{"id": "282930"}],
		"webhooks": [{"id": "313233"}]
	}`)

	var al AuditLog
	err := json.Unmarshal(data, &al)
	if err != nil {
		t.Fatalf("Failed to unmarshal AuditLog: %v", err)
	}

	if len(al.ApplicationCommands) != 1 {
		t.Errorf("Expected 1 application command, got %d", len(al.ApplicationCommands))
	}
	if len(al.AuditLogEntries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(al.AuditLogEntries))
	}
	if al.AuditLogEntries[0].ActionType != GUILD_UPDATE {
		t.Errorf("Expected ActionType GUILD_UPDATE, got %v", al.AuditLogEntries[0].ActionType)
	}
	if len(al.AuditLogEntries[0].Changes) != 1 {
		t.Errorf("Expected 1 change, got %d", len(al.AuditLogEntries[0].Changes))
	}
	if al.AuditLogEntries[0].Changes[0].Key != "name" {
		t.Errorf("Expected key 'name', got %s", al.AuditLogEntries[0].Changes[0].Key)
	}
	if al.AuditLogEntries[0].Changes[0].NewValue != "new" {
		t.Errorf("Expected new_value 'new', got %v", al.AuditLogEntries[0].Changes[0].NewValue)
	}
	if al.AuditLogEntries[0].Changes[0].OldValue != "old" {
		t.Errorf("Expected old_value 'old', got %v", al.AuditLogEntries[0].Changes[0].OldValue)
	}
	if al.AuditLogEntries[0].Options.Count != "1" {
		t.Errorf("Expected options.count '1', got %s", al.AuditLogEntries[0].Options.Count)
	}
	if al.AuditLogEntries[0].Reason != "because" {
		t.Errorf("Expected reason 'because', got %s", al.AuditLogEntries[0].Reason)
	}
}
