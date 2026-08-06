package components

// ComponentType represents the type of a component.
type ComponentType int

const (
	ComponentTypeActionRow         ComponentType = 1
	ComponentTypeButton            ComponentType = 2
	ComponentTypeStringSelect      ComponentType = 3
	ComponentTypeTextInput         ComponentType = 4
	ComponentTypeUserSelect        ComponentType = 5
	ComponentTypeRoleSelect        ComponentType = 6
	ComponentTypeMentionableSelect ComponentType = 7
	ComponentTypeChannelSelect     ComponentType = 8
	ComponentTypeSection           ComponentType = 9
	ComponentTypeTextDisplay       ComponentType = 10
	ComponentTypeThumbnail         ComponentType = 11
	ComponentTypeMediaGallery      ComponentType = 12
	ComponentTypeFile              ComponentType = 13
	ComponentTypeSeparator         ComponentType = 14
	ComponentTypeContainer         ComponentType = 17
)

// Component is an interface that all components must implement.
type Component interface {
	Type() ComponentType
}
