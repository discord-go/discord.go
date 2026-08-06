package components

import "encoding/json"

type ButtonStyle int

const (
	ButtonStylePrimary   ButtonStyle = 1
	ButtonStyleSecondary ButtonStyle = 2
	ButtonStyleSuccess   ButtonStyle = 3
	ButtonStyleDanger    ButtonStyle = 4
	ButtonStyleLink      ButtonStyle = 5
	ButtonStylePremium   ButtonStyle = 6
)

type Button struct {
	Style    ButtonStyle `json:"style"`
	Label    string      `json:"label,omitempty"`
	Emoji    interface{} `json:"emoji,omitempty"` // simplified emoji
	CustomID string      `json:"custom_id,omitempty"`
	URL      string      `json:"url,omitempty"`
	Disabled bool        `json:"disabled,omitempty"`
	SKUID    string      `json:"sku_id,omitempty"` // used for ButtonStylePremium
}

func (b Button) Type() ComponentType {
	return ComponentTypeButton
}

func (b Button) MarshalJSON() ([]byte, error) {
	type alias Button
	return json.Marshal(struct {
		Type ComponentType `json:"type"`
		alias
	}{
		Type:  b.Type(),
		alias: (alias)(b),
	})
}
