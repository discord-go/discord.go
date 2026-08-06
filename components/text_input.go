package components

import "encoding/json"

type TextInputStyle int

const (
	TextInputStyleShort     TextInputStyle = 1
	TextInputStyleParagraph TextInputStyle = 2
)

type TextInput struct {
	CustomID    string         `json:"custom_id"`
	Style       TextInputStyle `json:"style"`
	Label       string         `json:"label"`
	MinLength   *int           `json:"min_length,omitempty"`
	MaxLength   *int           `json:"max_length,omitempty"`
	Required    *bool          `json:"required,omitempty"` // Optional in spec, but default true
	Value       string         `json:"value,omitempty"`
	Placeholder string         `json:"placeholder,omitempty"`
}

func (t TextInput) Type() ComponentType { return ComponentTypeTextInput }

func (t TextInput) MarshalJSON() ([]byte, error) {
	type alias TextInput
	return json.Marshal(struct {
		Type ComponentType `json:"type"`
		alias
	}{
		Type:  t.Type(),
		alias: (alias)(t),
	})
}
