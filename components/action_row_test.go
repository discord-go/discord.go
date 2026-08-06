package components

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestActionRow_UnmarshalJSON(t *testing.T) {
	data := []byte(`{
		"type": 1,
		"components": [
			{
				"type": 2,
				"style": 1,
				"label": "Click me",
				"custom_id": "btn1"
			},
			{
				"type": 3,
				"custom_id": "sel1",
				"options": [
					{"label": "opt1", "value": "val1"}
				]
			},
			{
				"type": 4,
				"custom_id": "txt1",
				"style": 1,
				"label": "Text"
			},
			{
				"type": 5,
				"custom_id": "usr1"
			},
			{
				"type": 6,
				"custom_id": "role1"
			},
			{
				"type": 7,
				"custom_id": "ment1"
			},
			{
				"type": 8,
				"custom_id": "chan1"
			},
			{
				"type": 1,
				"components": []
			}
		]
	}`)

	var row ActionRow
	if err := json.Unmarshal(data, &row); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(row.Components) != 8 {
		t.Fatalf("expected 8 components, got %d", len(row.Components))
	}

	expectedTypes := []ComponentType{
		ComponentTypeButton,
		ComponentTypeStringSelect,
		ComponentTypeTextInput,
		ComponentTypeUserSelect,
		ComponentTypeRoleSelect,
		ComponentTypeMentionableSelect,
		ComponentTypeChannelSelect,
		ComponentTypeActionRow,
	}

	for i, comp := range row.Components {
		if comp.Type() != expectedTypes[i] {
			t.Errorf("component %d: expected type %d, got %d", i, expectedTypes[i], comp.Type())
		}
	}

	// Verify MarshalJSON
	marshaled, err := json.Marshal(row)
	if err != nil {
		t.Fatalf("unexpected error marshaling: %v", err)
	}
	var row2 ActionRow
	if err := json.Unmarshal(marshaled, &row2); err != nil {
		t.Fatalf("unexpected error unmarshaling marshaled: %v", err)
	}
	if !reflect.DeepEqual(row, row2) {
		t.Errorf("roundtrip failed: \nexpected %v\ngot %v", row, row2)
	}
}

func TestActionRow_UnmarshalJSON_Errors(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"invalid json", `{invalid`},
		{"wrong outer type", `{"type": 2, "components": []}`},
		{"invalid action row type format", `{"type": "string"}`},
		{"invalid component type format", `{"type": 1, "components": [{"type": "string"}]}`},
		{"unknown component type", `{"type": 1, "components": [{"type": 999}]}`},
		{"invalid button json", `{"type": 1, "components": [{"type": 2, "style": "invalid"}]}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var row ActionRow
			err := json.Unmarshal([]byte(tt.data), &row)
			if err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}
