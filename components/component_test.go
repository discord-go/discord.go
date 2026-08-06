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
