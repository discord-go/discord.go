package interactions

import (
	"encoding/json"
	"testing"
)

func TestOptionGetters(t *testing.T) {
	raw := []byte(`{
		"name": "cmd",
		"type": 1,
		"options": [
			{"name": "str", "type": 3, "value": "hello"},
			{"name": "num", "type": 4, "value": 42},
			{"name": "flag", "type": 5, "value": true},
			{"name": "ratio", "type": 10, "value": 3.14},
			{"name": "user", "type": 6, "value": "9876543210"}
		]
	}`)
	var data ApplicationCommandInteractionData
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(data.Options) != 5 {
		t.Fatalf("expected 5 options, got %d", len(data.Options))
	}

	byName := func(name string) *ApplicationCommandInteractionDataOption {
		for i := range data.Options {
			if data.Options[i].Name == name {
				return &data.Options[i]
			}
		}
		return nil
	}

	if got := byName("str").String(); got != "hello" {
		t.Errorf("String = %q, want hello", got)
	}
	if got := byName("num").Int(); got != 42 {
		t.Errorf("Int = %d, want 42", got)
	}
	if got := byName("flag").Bool(); !got {
		t.Error("Bool = false, want true")
	}
	if got := byName("ratio").Float(); got != 3.14 {
		t.Errorf("Float = %v, want 3.14", got)
	}
	if got := byName("user").Snowflake(); got.String() != "9876543210" {
		t.Errorf("Snowflake = %s, want 9876543210", got)
	}
}

func TestOptionGettersNilAndWrongTypes(t *testing.T) {
	raw := []byte(`[
		{"name": "str", "type": 3, "value": 42},
		{"name": "num", "type": 4, "value": "not-a-number"},
		{"name": "flag", "type": 5, "value": "yes"},
		{"name": "empty", "type": 3}
	]`)
	var options []ApplicationCommandInteractionDataOption
	if err := json.Unmarshal(raw, &options); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got := options[0].String(); got != "" {
		t.Errorf("String on int value = %q, want empty", got)
	}
	if got := options[1].Int(); got != 0 {
		t.Errorf("Int on non-numeric string = %d, want 0", got)
	}
	if got := options[2].Bool(); got {
		t.Error("Bool on non-bool string = true, want false")
	}
	if got := options[3].String(); got != "" {
		t.Errorf("String on nil value = %q, want empty", got)
	}

	var nilOption *ApplicationCommandInteractionDataOption
	if got := nilOption.String(); got != "" {
		t.Errorf("String on nil option = %q, want empty", got)
	}
	if got := nilOption.Int(); got != 0 {
		t.Errorf("Int on nil option = %d, want 0", got)
	}
	if got := nilOption.Bool(); got {
		t.Error("Bool on nil option = true, want false")
	}
	if got := nilOption.Snowflake(); got != 0 {
		t.Errorf("Snowflake on nil option = %d, want 0", got)
	}
}

func TestSubcommandExtraction(t *testing.T) {
	raw := []byte(`{
		"name": "cmd",
		"type": 1,
		"options": [
			{"name": "sub", "type": 1, "options": [
				{"name": "arg1", "type": 3, "value": "x"}
			]}
		]
	}`)
	var data ApplicationCommandInteractionData
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	root := &data.Options[0]
	if got := root.Subcommand(); got != "sub" {
		t.Errorf("Subcommand on root = %q, want sub", got)
	}
	if got := root.NestedOptions()[0].Subcommand(); got != "" {
		t.Errorf("Subcommand on leaf = %q, want empty", got)
	}

	// A subcommand group containing a subcommand.
	groupRaw := []byte(`{
		"name": "cmd",
		"type": 1,
		"options": [
			{"name": "grp", "type": 2, "options": [
				{"name": "sub", "type": 1, "options": [
					{"name": "arg1", "type": 3, "value": "x"}
				]}
			]}
		]
	}`)
	var groupData ApplicationCommandInteractionData
	if err := json.Unmarshal(groupRaw, &groupData); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	group := &groupData.Options[0]
	if got := group.SubcommandGroup(); got != "grp" {
		t.Errorf("SubcommandGroup = %q, want grp", got)
	}
	if got := group.NestedOptions()[0].Subcommand(); got != "sub" {
		t.Errorf("Subcommand = %q, want sub", got)
	}
}
