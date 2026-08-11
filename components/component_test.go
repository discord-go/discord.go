package components

import (
	"testing"
)

func TestComponentTypes(t *testing.T) {
	comps := []Component{
		ActionRow{},
		Button{},
		StringSelect{},
		TextInput{},
		UserSelect{},
		RoleSelect{},
		MentionableSelect{},
		ChannelSelect{},
	}

	expected := []ComponentType{
		ComponentTypeActionRow,
		ComponentTypeButton,
		ComponentTypeStringSelect,
		ComponentTypeTextInput,
		ComponentTypeUserSelect,
		ComponentTypeRoleSelect,
		ComponentTypeMentionableSelect,
		ComponentTypeChannelSelect,
	}

	for i, c := range comps {
		if c.Type() != expected[i] {
			t.Errorf("expected %d, got %d", expected[i], c.Type())
		}
	}
}

func TestNewSelectOption(t *testing.T) {
	opt := NewSelectOption("Yes", "yes")
	if opt.Label != "Yes" {
		t.Errorf("expected label 'Yes', got %q", opt.Label)
	}
	if opt.Value != "yes" {
		t.Errorf("expected value 'yes', got %q", opt.Value)
	}
	if opt.Description != "" {
		t.Errorf("expected empty description, got %q", opt.Description)
	}
	if opt.Default {
		t.Error("expected Default to be false")
	}

	// Verify it can be used in a select menu.
	selectMenu := StringSelect{
		CustomID: "vote",
		Options:  []SelectOption{NewSelectOption("Yes", "yes"), NewSelectOption("No", "no")},
	}
	if len(selectMenu.Options) != 2 {
		t.Fatalf("expected 2 options, got %d", len(selectMenu.Options))
	}
}
