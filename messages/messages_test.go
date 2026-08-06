package messages

import (
	"encoding/json"
	"testing"

	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/snowflake"
)

func TestMessage_UnmarshalJSON(t *testing.T) {
	data := []byte(`{
		"id": "123456789012345678",
		"channel_id": "876543210987654321",
		"content": "Hello World",
		"mention_roles": ["111111111111111111"],
		"components": [
			{
				"type": 1,
				"components": [
					{
						"type": 2,
						"style": 1,
						"label": "Click Me",
						"custom_id": "click_one"
					}
				]
			},
			{
				"type": 3,
				"custom_id": "select_one"
			},
			{
				"type": 4,
				"custom_id": "text_input"
			},
			{
				"type": 5,
				"custom_id": "user_select"
			},
			{
				"type": 6,
				"custom_id": "role_select"
			},
			{
				"type": 7,
				"custom_id": "mentionable_select"
			},
			{
				"type": 8,
				"custom_id": "channel_select"
			},
			{
				"type": 2,
				"style": 1,
				"label": "Click Me Top Level",
				"custom_id": "click_top"
			}
		]
	}`)

	var m Message
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	if m.ID != snowflake.ID(123456789012345678) {
		t.Errorf("Expected ID 123456789012345678, got %v", m.ID)
	}

	if len(m.MentionRoles) != 1 {
		t.Fatalf("Expected 1 mention role, got %d", len(m.MentionRoles))
	}
	if m.MentionRoles[0].ID != snowflake.ID(111111111111111111) {
		t.Errorf("Expected mention role ID 111111111111111111, got %v", m.MentionRoles[0].ID)
	}

	if len(m.Components) != 8 {
		t.Fatalf("Expected 8 components, got %d", len(m.Components))
	}

	if m.Components[0].Type() != components.ComponentTypeActionRow {
		t.Errorf("Expected first component to be ActionRow, got %d", m.Components[0].Type())
	}
	if m.Components[1].Type() != components.ComponentTypeStringSelect {
		t.Errorf("Expected second component to be StringSelect, got %d", m.Components[1].Type())
	}
	if m.Components[2].Type() != components.ComponentTypeTextInput {
		t.Errorf("Expected third component to be TextInput, got %d", m.Components[2].Type())
	}
	if m.Components[3].Type() != components.ComponentTypeUserSelect {
		t.Errorf("Expected fourth component to be UserSelect, got %d", m.Components[3].Type())
	}
	if m.Components[4].Type() != components.ComponentTypeRoleSelect {
		t.Errorf("Expected fifth component to be RoleSelect, got %d", m.Components[4].Type())
	}
	if m.Components[5].Type() != components.ComponentTypeMentionableSelect {
		t.Errorf("Expected sixth component to be MentionableSelect, got %d", m.Components[5].Type())
	}
	if m.Components[6].Type() != components.ComponentTypeChannelSelect {
		t.Errorf("Expected seventh component to be ChannelSelect, got %d", m.Components[6].Type())
	}
	if m.Components[7].Type() != components.ComponentTypeButton {
		t.Errorf("Expected eighth component to be Button, got %d", m.Components[7].Type())
	}
}

func TestMessage_UnmarshalJSON_InvalidRole(t *testing.T) {
	data := []byte(`{
		"mention_roles": ["invalid_id"]
	}`)
	var m Message
	if err := json.Unmarshal(data, &m); err == nil {
		t.Error("Expected error unmarshaling invalid role ID")
	}
}

func TestMessage_UnmarshalJSON_InvalidBaseComponent(t *testing.T) {
	data := []byte(`{
		"components": [
			"invalid"
		]
	}`)
	var m Message
	if err := json.Unmarshal(data, &m); err == nil {
		t.Error("Expected error unmarshaling invalid base component")
	}
}

func TestMessage_UnmarshalJSON_UnknownComponent(t *testing.T) {
	data := []byte(`{
		"components": [
			{
				"type": 99
			}
		]
	}`)
	var m Message
	if err := json.Unmarshal(data, &m); err == nil {
		t.Error("Expected error unmarshaling unknown component")
	}
}

func TestMessage_UnmarshalJSON_InvalidComponentFields(t *testing.T) {
	// Let's pass valid type but invalid content that json.Unmarshal would reject, if any exist.
	// We'll just test a type mismatch in component to trigger error.
	data := []byte(`{
		"components": [
			{
				"type": 1,
				"components": "invalid"
			}
		]
	}`)
	var m Message
	if err := json.Unmarshal(data, &m); err == nil {
		t.Error("Expected error unmarshaling invalid component fields")
	}
}

func TestAllowedMentions_UnmarshalJSON(t *testing.T) {
	data := []byte(`{
		"parse": ["users", "roles", "everyone"],
		"roles": ["111111111111111111", "222222222222222222"],
		"users": ["333333333333333333"],
		"replied_user": true
	}`)
	var a AllowedMentions
	if err := json.Unmarshal(data, &a); err != nil {
		t.Fatalf("Failed to unmarshal AllowedMentions: %v", err)
	}

	if len(a.Roles) != 2 || a.Roles[0] != snowflake.ID(111111111111111111) || a.Roles[1] != snowflake.ID(222222222222222222) {
		t.Errorf("Parsed roles incorrectly: %v", a.Roles)
	}
	if len(a.Users) != 1 || a.Users[0] != snowflake.ID(333333333333333333) {
		t.Errorf("Parsed users incorrectly: %v", a.Users)
	}
}

func TestAllowedMentions_UnmarshalJSON_InvalidRole(t *testing.T) {
	data := []byte(`{
		"roles": ["invalid"]
	}`)
	var a AllowedMentions
	if err := json.Unmarshal(data, &a); err == nil {
		t.Error("Expected error unmarshaling invalid role in AllowedMentions")
	}
}

func TestAllowedMentions_UnmarshalJSON_InvalidUser(t *testing.T) {
	data := []byte(`{
		"users": ["invalid"]
	}`)
	var a AllowedMentions
	if err := json.Unmarshal(data, &a); err == nil {
		t.Error("Expected error unmarshaling invalid user in AllowedMentions")
	}
}

func TestAllowedMentions_UnmarshalJSON_InvalidBase(t *testing.T) {
	data := []byte(`"invalid"`)
	var a AllowedMentions
	if err := json.Unmarshal(data, &a); err == nil {
		t.Error("Expected error unmarshaling invalid base AllowedMentions")
	}
}

func TestMessage_UnmarshalJSON_InvalidBase(t *testing.T) {
	data := []byte(`"invalid"`)
	var m Message
	if err := json.Unmarshal(data, &m); err == nil {
		t.Error("Expected error unmarshaling invalid base Message")
	}
}

func TestMessage_UnmarshalJSON_NewFields(t *testing.T) {
	data := []byte(`{
		"id": "123456789012345678",
		"channel_id": "876543210987654321",
		"call": {
			"participants": ["111111111111111111", "222222222222222222"]
		},
		"poll": {},
		"message_snapshots": [{}],
		"interaction_metadata": {}
	}`)

	var m Message
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	if m.Call == nil {
		t.Fatal("Expected Call to not be nil")
	}
	if len(m.Call.Participants) != 2 {
		t.Errorf("Expected 2 participants, got %d", len(m.Call.Participants))
	}
	if m.Call.Participants[0] != snowflake.ID(111111111111111111) {
		t.Errorf("Expected participant ID 111111111111111111, got %v", m.Call.Participants[0])
	}

	if m.Poll == nil {
		t.Error("Expected Poll to not be nil")
	}
	if len(m.MessageSnapshots) != 1 {
		t.Errorf("Expected 1 message snapshot, got %d", len(m.MessageSnapshots))
	}
	if m.InteractionMetadata == nil {
		t.Error("Expected InteractionMetadata to not be nil")
	}
}
