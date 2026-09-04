package messages

import (
	"encoding/json"
	"testing"

	"github.com/discord-go/discord.go/channels"
	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/roles"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
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

func TestMessage_UnmarshalJSON_Member(t *testing.T) {
	data := []byte(`{
		"id": "123456789012345678",
		"channel_id": "876543210987654321",
		"content": "Hello",
		"author": {"id": "42", "username": "ada", "discriminator": "0"},
		"member": {
			"user": {"id": "42", "username": "ada", "discriminator": "0"},
			"roles": ["111111111111111111", "222222222222222222"],
			"joined_at": "2026-01-01T00:00:00+00:00",
			"nick": "ada-dev",
			"permissions": "1024"
		}
	}`)

	var m Message
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}
	if m.Member == nil {
		t.Fatal("Expected Member to not be nil")
	}
	if m.Member.Nick == nil || *m.Member.Nick != "ada-dev" {
		t.Errorf("Expected nick ada-dev, got %v", m.Member.Nick)
	}
	if len(m.Member.Roles) != 2 {
		t.Fatalf("Expected 2 member roles, got %d", len(m.Member.Roles))
	}
	if m.Member.Roles[0] != snowflake.ID(111111111111111111) {
		t.Errorf("Expected role 111111111111111111, got %v", m.Member.Roles[0])
	}
}

func TestMessage_UnmarshalJSON_MemberAbsent(t *testing.T) {
	data := []byte(`{
		"id": "123456789012345678",
		"channel_id": "876543210987654321",
		"content": "Hello"
	}`)
	var m Message
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}
	if m.Member != nil {
		t.Error("Expected Member to be nil when absent")
	}
}

func TestMessage_CleanContent(t *testing.T) {
	globalName := "Alice"
	username := "alice"
	channelName := "general"
	msg := Message{
		Content: "hi <@123> and <@!456> and <@&789> in <#101>",
		Mentions: []users.User{
			{ID: snowflake.ID(123), Username: username, GlobalName: &globalName},
			{ID: snowflake.ID(456), Username: "bob"},
		},
		MentionRoles: []roles.Role{
			{ID: snowflake.ID(789), Name: "Mods"},
		},
		MentionChannels: []channels.Channel{
			{ID: snowflake.ID(101), Name: &channelName},
		},
	}
	got := msg.CleanContent()
	want := "hi @Alice and @bob and @Mods in #general"
	if got != want {
		t.Errorf("CleanContent = %q, want %q", got, want)
	}
}

func TestMessage_CleanContentUnknownMentions(t *testing.T) {
	msg := Message{Content: "hi <@999> and <@&888> and <#777>"}
	got := msg.CleanContent()
	if got != "hi <@999> and <@&888> and <#777>" {
		t.Errorf("CleanContent = %q, want original tokens preserved", got)
	}
}

// TestMessage_CleanContentGatewayRoleMentions covers the path every
// production message takes: JSON decode via UnmarshalJSON, which rebuilds
// MentionRoles as ID-only roles (the gateway sends IDs, not names). The
// resolvable-name substitution must not mangle <@&id> into a bare "@".
func TestMessage_CleanContentGatewayRoleMentions(t *testing.T) {
	raw := []byte(`{"content":"ping <@&789>","mention_roles":["789"],"mentions":[]}`)
	var msg Message
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(msg.MentionRoles) != 1 || msg.MentionRoles[0].Name != "" {
		t.Fatalf("MentionRoles = %+v, want one ID-only role", msg.MentionRoles)
	}
	if got := msg.CleanContent(); got != "ping <@&789>" {
		t.Errorf("CleanContent = %q, want %q", got, "ping <@&789>")
	}
}

func TestMessage_CleanContentEdgeCases(t *testing.T) {
	var nilMsg *Message
	if got := nilMsg.CleanContent(); got != "" {
		t.Errorf("nil message CleanContent = %q, want empty", got)
	}
	if got := (&Message{}).CleanContent(); got != "" {
		t.Errorf("empty message CleanContent = %q, want empty", got)
	}
	// Unclosed tag and plain text with no mentions.
	if got := (&Message{Content: "a <@123"}).CleanContent(); got != "a <@123" {
		t.Errorf("unclosed token CleanContent = %q, want original", got)
	}
	if got := (&Message{Content: "plain text"}).CleanContent(); got != "plain text" {
		t.Errorf("plain text CleanContent = %q, want unchanged", got)
	}
}

func TestMessage_CleanContentPrefersGlobalName(t *testing.T) {
	msg := Message{
		Content: "<@42>",
		Mentions: []users.User{
			{ID: snowflake.ID(42), Username: "alice", GlobalName: ptr("Alice")},
		},
	}
	if got := msg.CleanContent(); got != "@Alice" {
		t.Errorf("CleanContent = %q, want @Alice", got)
	}
}

func ptr[T any](v T) *T { return &v }
