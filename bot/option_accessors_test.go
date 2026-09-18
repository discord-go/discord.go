package bot

import (
	"encoding/json"
	"testing"

	"github.com/discord-go/discord.go/interactions"
)

// TestInteractionContext_OptionAccessors pins the Option* accessor family and
// the coercion contracts it inherits from the previous Get* implementation.
func TestInteractionContext_OptionAccessors(t *testing.T) {
	raw := []byte(`{
		"type": 2,
		"token": "t",
		"application_id": "1",
		"data": {
			"id": "10",
			"name": "cmd",
			"options": [
				{"name": "flag", "type": 5, "value": true},
				{"name": "text", "type": 3, "value": "hello"},
				{"name": "count", "type": 4, "value": 7},
				{"name": "ratio", "type": 10, "value": 2.5},
				{"name": "target", "type": 6, "value": "9007199254740993"},
				{"name": "channel", "type": 7, "value": "1234567890123456789"},
				{"name": "role", "type": 8, "value": "9876543210987654321"}
			]
		}
	}`)
	var interaction interactions.Interaction
	if err := json.Unmarshal(raw, &interaction); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	ctx := newInteractionContext(BaseContext{}, &interaction)

	if !ctx.OptionBool("flag") {
		t.Error("OptionBool(flag) = false, want true")
	}
	if ctx.OptionBool("missing") {
		t.Error("OptionBool(missing) = true, want false")
	}
	if got := ctx.OptionString("text"); got != "hello" {
		t.Errorf("OptionString(text) = %q, want %q", got, "hello")
	}
	if got := ctx.OptionString("missing"); got != "" {
		t.Errorf("OptionString(missing) = %q, want empty", got)
	}
	if got := ctx.OptionInt("count"); got != 7 {
		t.Errorf("OptionInt(count) = %d, want 7", got)
	}
	if got := ctx.OptionFloat("ratio"); got != 2.5 {
		t.Errorf("OptionFloat(ratio) = %v, want 2.5", got)
	}
	if got := ctx.OptionSnowflake("target"); got.String() != "9007199254740993" {
		t.Errorf("OptionSnowflake(target) = %s, want 9007199254740993 (precision must be preserved)", got)
	}
	if got := ctx.OptionUser("target"); got.String() != "9007199254740993" {
		t.Errorf("OptionUser(target) = %s, want 9007199254740993", got)
	}
	if got := ctx.OptionChannel("channel"); got.String() != "1234567890123456789" {
		t.Errorf("OptionChannel(channel) = %s, want 1234567890123456789", got)
	}
	if got := ctx.OptionRole("role"); got.String() != "9876543210987654321" {
		t.Errorf("OptionRole(role) = %s, want 9876543210987654321", got)
	}
	if ctx.Option("missing") != nil {
		t.Error("Option(missing) != nil, want nil")
	}
	if ctx.Option("text") == nil {
		t.Fatal("Option(text) = nil, want the option")
	}
	if !ctx.HasOption("text") {
		t.Error("HasOption(text) = false, want true")
	}
	if ctx.HasOption("missing") {
		t.Error("HasOption(missing) = true, want false")
	}
}

// TestInteractionContext_GetAliasesEquivalentToOptionAccessors pins that
// every deprecated Get* name returns exactly what its Option* replacement
// returns, so existing code keeps working unchanged.
func TestInteractionContext_GetAliasesEquivalentToOptionAccessors(t *testing.T) {
	raw := []byte(`{
		"type": 2,
		"token": "t",
		"application_id": "1",
		"data": {
			"id": "10",
			"name": "cmd",
			"options": [
				{"name": "flag", "type": 5, "value": true},
				{"name": "text", "type": 3, "value": "hello"},
				{"name": "count", "type": 4, "value": 42},
				{"name": "ratio", "type": 10, "value": 0.75},
				{"name": "target", "type": 6, "value": "9007199254740993"},
				{"name": "channel", "type": 7, "value": "111"},
				{"name": "role", "type": 8, "value": "222"}
			]
		}
	}`)
	var interaction interactions.Interaction
	if err := json.Unmarshal(raw, &interaction); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	ctx := newInteractionContext(BaseContext{}, &interaction)

	if ctx.GetOption("text") != ctx.Option("text") {
		t.Error("GetOption(text) != Option(text)")
	}
	if ctx.GetStringOption("text") != ctx.OptionString("text") {
		t.Error("GetStringOption(text) != OptionString(text)")
	}
	if ctx.GetIntOption("count") != ctx.OptionInt("count") {
		t.Error("GetIntOption(count) != OptionInt(count)")
	}
	if ctx.GetFloatOption("ratio") != ctx.OptionFloat("ratio") {
		t.Error("GetFloatOption(ratio) != OptionFloat(ratio)")
	}
	if ctx.GetBoolOption("flag") != ctx.OptionBool("flag") {
		t.Error("GetBoolOption(flag) != OptionBool(flag)")
	}
	if ctx.GetString("text") != ctx.OptionString("text") {
		t.Error("GetString(text) != OptionString(text)")
	}
	if ctx.GetInt("count") != ctx.OptionInt("count") {
		t.Error("GetInt(count) != OptionInt(count)")
	}
	if ctx.GetFloat("ratio") != ctx.OptionFloat("ratio") {
		t.Error("GetFloat(ratio) != OptionFloat(ratio)")
	}
	if ctx.GetBool("flag") != ctx.OptionBool("flag") {
		t.Error("GetBool(flag) != OptionBool(flag)")
	}
	if ctx.GetSnowflake("target") != ctx.OptionSnowflake("target") {
		t.Error("GetSnowflake(target) != OptionSnowflake(target)")
	}
	if ctx.GetUserID("target") != ctx.OptionUser("target") {
		t.Error("GetUserID(target) != OptionUser(target)")
	}
	if ctx.GetRoleID("role") != ctx.OptionRole("role") {
		t.Error("GetRoleID(role) != OptionRole(role)")
	}
	if ctx.GetChannelID("channel") != ctx.OptionChannel("channel") {
		t.Error("GetChannelID(channel) != OptionChannel(channel)")
	}
}
