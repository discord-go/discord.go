package components

import "encoding/json"

type SelectOption struct {
	Label       string      `json:"label"`
	Value       string      `json:"value"`
	Description string      `json:"description,omitempty"`
	Emoji       interface{} `json:"emoji,omitempty"`
	Default     bool        `json:"default,omitempty"`
}

type StringSelect struct {
	CustomID    string         `json:"custom_id"`
	Options     []SelectOption `json:"options"`
	Placeholder string         `json:"placeholder,omitempty"`
	MinValues   *int           `json:"min_values,omitempty"`
	MaxValues   *int           `json:"max_values,omitempty"`
	Disabled    bool           `json:"disabled,omitempty"`
}

func (s StringSelect) Type() ComponentType { return ComponentTypeStringSelect }

func (s StringSelect) MarshalJSON() ([]byte, error) {
	type alias StringSelect
	return json.Marshal(struct {
		Type ComponentType `json:"type"`
		alias
	}{
		Type:  s.Type(),
		alias: (alias)(s),
	})
}

type UserSelect struct {
	CustomID    string `json:"custom_id"`
	Placeholder string `json:"placeholder,omitempty"`
	MinValues   *int   `json:"min_values,omitempty"`
	MaxValues   *int   `json:"max_values,omitempty"`
	Disabled    bool   `json:"disabled,omitempty"`
}

func (s UserSelect) Type() ComponentType { return ComponentTypeUserSelect }

func (s UserSelect) MarshalJSON() ([]byte, error) {
	type alias UserSelect
	return json.Marshal(struct {
		Type ComponentType `json:"type"`
		alias
	}{
		Type:  s.Type(),
		alias: (alias)(s),
	})
}

type RoleSelect struct {
	CustomID    string `json:"custom_id"`
	Placeholder string `json:"placeholder,omitempty"`
	MinValues   *int   `json:"min_values,omitempty"`
	MaxValues   *int   `json:"max_values,omitempty"`
	Disabled    bool   `json:"disabled,omitempty"`
}

func (s RoleSelect) Type() ComponentType { return ComponentTypeRoleSelect }

func (s RoleSelect) MarshalJSON() ([]byte, error) {
	type alias RoleSelect
	return json.Marshal(struct {
		Type ComponentType `json:"type"`
		alias
	}{
		Type:  s.Type(),
		alias: (alias)(s),
	})
}

type MentionableSelect struct {
	CustomID    string `json:"custom_id"`
	Placeholder string `json:"placeholder,omitempty"`
	MinValues   *int   `json:"min_values,omitempty"`
	MaxValues   *int   `json:"max_values,omitempty"`
	Disabled    bool   `json:"disabled,omitempty"`
}

func (s MentionableSelect) Type() ComponentType { return ComponentTypeMentionableSelect }

func (s MentionableSelect) MarshalJSON() ([]byte, error) {
	type alias MentionableSelect
	return json.Marshal(struct {
		Type ComponentType `json:"type"`
		alias
	}{
		Type:  s.Type(),
		alias: (alias)(s),
	})
}

type ChannelType int // Simplified

type ChannelSelect struct {
	CustomID     string        `json:"custom_id"`
	ChannelTypes []ChannelType `json:"channel_types,omitempty"`
	Placeholder  string        `json:"placeholder,omitempty"`
	MinValues    *int          `json:"min_values,omitempty"`
	MaxValues    *int          `json:"max_values,omitempty"`
	Disabled     bool          `json:"disabled,omitempty"`
}

func (s ChannelSelect) Type() ComponentType { return ComponentTypeChannelSelect }

func (s ChannelSelect) MarshalJSON() ([]byte, error) {
	type alias ChannelSelect
	return json.Marshal(struct {
		Type ComponentType `json:"type"`
		alias
	}{
		Type:  s.Type(),
		alias: (alias)(s),
	})
}
