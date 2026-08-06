package components

import (
	"encoding/json"
	"fmt"
)

// Unmarshal decodes any supported Discord component, including nested
// Components V2 children.
func Unmarshal(data []byte) (Component, error) {
	var header struct {
		Type ComponentType `json:"type"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return nil, err
	}
	var component Component
	switch header.Type {
	case ComponentTypeActionRow:
		component = &ActionRow{}
	case ComponentTypeButton:
		component = &Button{}
	case ComponentTypeStringSelect:
		component = &StringSelect{}
	case ComponentTypeTextInput:
		component = &TextInput{}
	case ComponentTypeUserSelect:
		component = &UserSelect{}
	case ComponentTypeRoleSelect:
		component = &RoleSelect{}
	case ComponentTypeMentionableSelect:
		component = &MentionableSelect{}
	case ComponentTypeChannelSelect:
		component = &ChannelSelect{}
	case ComponentTypeTextDisplay:
		component = &TextDisplay{}
	case ComponentTypeSeparator:
		component = &Separator{}
	case ComponentTypeSection:
		component = &Section{}
	case ComponentTypeThumbnail:
		component = &Thumbnail{}
	case ComponentTypeMediaGallery:
		component = &MediaGallery{}
	case ComponentTypeFile:
		component = &File{}
	case ComponentTypeContainer:
		component = &Container{}
	default:
		return nil, fmt.Errorf("components: unknown component type: %d", header.Type)
	}
	if err := json.Unmarshal(data, component); err != nil {
		return nil, err
	}
	return componentValue(component), nil
}

func componentValue(component Component) Component {
	switch value := component.(type) {
	case *ActionRow:
		return *value
	case *Button:
		return *value
	case *StringSelect:
		return *value
	case *TextInput:
		return *value
	case *UserSelect:
		return *value
	case *RoleSelect:
		return *value
	case *MentionableSelect:
		return *value
	case *ChannelSelect:
		return *value
	case *TextDisplay:
		return *value
	case *Separator:
		return *value
	case *Section:
		return *value
	case *Thumbnail:
		return *value
	case *MediaGallery:
		return *value
	case *File:
		return *value
	case *Container:
		return *value
	default:
		return component
	}
}

type nestedComponentJSON struct {
	Type       ComponentType     `json:"type"`
	Components []json.RawMessage `json:"components,omitempty"`
}

func decodeNested(raw []json.RawMessage) ([]Component, error) {
	components := make([]Component, 0, len(raw))
	for _, data := range raw {
		component, err := Unmarshal(data)
		if err != nil {
			return nil, err
		}
		components = append(components, component)
	}
	return components, nil
}

func (c *Section) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type       ComponentType     `json:"type"`
		Components []json.RawMessage `json:"components"`
		Accessory  json.RawMessage   `json:"accessory"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.Type != ComponentTypeSection {
		return fmt.Errorf("components: expected section, got %d", raw.Type)
	}
	children, err := decodeNested(raw.Components)
	if err != nil {
		return err
	}
	c.Components = children
	if len(raw.Accessory) > 0 && string(raw.Accessory) != "null" {
		c.Accessory, err = Unmarshal(raw.Accessory)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Container) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type        ComponentType     `json:"type"`
		Components  []json.RawMessage `json:"components"`
		AccentColor int               `json:"accent_color,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.Type != ComponentTypeContainer {
		return fmt.Errorf("components: expected container, got %d", raw.Type)
	}
	children, err := decodeNested(raw.Components)
	if err != nil {
		return err
	}
	c.Components = children
	c.AccentColor = raw.AccentColor
	return nil
}
