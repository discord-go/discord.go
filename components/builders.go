package components

// SeparatorSpacingSize controls the amount of space around a separator.
type SeparatorSpacingSize int

const (
	SeparatorSpacingSmall SeparatorSpacingSize = 1
	SeparatorSpacingLarge SeparatorSpacingSize = 2
)

// TextDisplayBuilder builds a Components V2 text display.
type TextDisplayBuilder struct{ component TextDisplay }

func NewTextDisplayBuilder() *TextDisplayBuilder { return &TextDisplayBuilder{} }
func (b *TextDisplayBuilder) SetContent(content string) *TextDisplayBuilder {
	b.component.Content = content
	return b
}
func (b *TextDisplayBuilder) Build() TextDisplay          { return b.component }
func (b TextDisplayBuilder) Type() ComponentType          { return b.component.Type() }
func (b TextDisplayBuilder) MarshalJSON() ([]byte, error) { return b.component.MarshalJSON() }

// SeparatorBuilder builds a Components V2 separator.
type SeparatorBuilder struct{ component Separator }

func NewSeparatorBuilder() *SeparatorBuilder { return &SeparatorBuilder{} }
func (b *SeparatorBuilder) SetDivider(divider bool) *SeparatorBuilder {
	b.component.Divider = divider
	return b
}
func (b *SeparatorBuilder) SetSpacing(spacing SeparatorSpacingSize) *SeparatorBuilder {
	b.component.Spacing = int(spacing)
	return b
}
func (b *SeparatorBuilder) Build() Separator            { return b.component }
func (b SeparatorBuilder) Type() ComponentType          { return b.component.Type() }
func (b SeparatorBuilder) MarshalJSON() ([]byte, error) { return b.component.MarshalJSON() }

// ThumbnailBuilder builds a section thumbnail accessory.
type ThumbnailBuilder struct{ component Thumbnail }

func NewThumbnailBuilder() *ThumbnailBuilder { return &ThumbnailBuilder{} }
func NewThumbnailBuilderWithURL(url string) *ThumbnailBuilder {
	return NewThumbnailBuilder().SetURL(url)
}
func (b *ThumbnailBuilder) SetURL(url string) *ThumbnailBuilder {
	b.component.Media.URL = url
	return b
}
func (b *ThumbnailBuilder) Build() Thumbnail            { return b.component }
func (b ThumbnailBuilder) Type() ComponentType          { return b.component.Type() }
func (b ThumbnailBuilder) MarshalJSON() ([]byte, error) { return b.component.MarshalJSON() }

// MediaGalleryItemBuilder builds one media gallery item.
type MediaGalleryItemBuilder struct{ item MediaGalleryItem }

func NewMediaGalleryItemBuilder() *MediaGalleryItemBuilder { return &MediaGalleryItemBuilder{} }
func NewMediaGalleryItemBuilderWithURL(url string) *MediaGalleryItemBuilder {
	return NewMediaGalleryItemBuilder().SetURL(url)
}
func (b *MediaGalleryItemBuilder) SetURL(url string) *MediaGalleryItemBuilder {
	b.item.Media.URL = url
	return b
}
func (b *MediaGalleryItemBuilder) Build() MediaGalleryItem { return b.item }

// MediaGalleryBuilder builds a Components V2 media gallery.
type MediaGalleryBuilder struct{ component MediaGallery }

func NewMediaGalleryBuilder() *MediaGalleryBuilder { return &MediaGalleryBuilder{} }
func (b *MediaGalleryBuilder) AddItems(items ...interface{}) *MediaGalleryBuilder {
	for _, item := range items {
		switch value := item.(type) {
		case MediaGalleryItem:
			b.component.Items = append(b.component.Items, value)
		case *MediaGalleryItemBuilder:
			if value != nil {
				b.component.Items = append(b.component.Items, value.Build())
			}
		}
	}
	return b
}
func (b *MediaGalleryBuilder) AddItemBuilders(items ...*MediaGalleryItemBuilder) *MediaGalleryBuilder {
	for _, item := range items {
		if item != nil {
			b.component.Items = append(b.component.Items, item.Build())
		}
	}
	return b
}
func (b *MediaGalleryBuilder) Build() MediaGallery         { return b.component }
func (b MediaGalleryBuilder) Type() ComponentType          { return b.component.Type() }
func (b MediaGalleryBuilder) MarshalJSON() ([]byte, error) { return b.component.MarshalJSON() }

// FileBuilder builds a Components V2 file reference.
type FileBuilder struct{ component File }

func NewFileBuilder() *FileBuilder { return &FileBuilder{} }
func (b *FileBuilder) SetURL(url string) *FileBuilder {
	b.component.File.URL = url
	return b
}
func (b *FileBuilder) Build() File                 { return b.component }
func (b FileBuilder) Type() ComponentType          { return b.component.Type() }
func (b FileBuilder) MarshalJSON() ([]byte, error) { return b.component.MarshalJSON() }

// ButtonBuilder builds a button component.
type ButtonBuilder struct{ component Button }

func NewButtonBuilder() *ButtonBuilder { return &ButtonBuilder{} }
func (b *ButtonBuilder) SetCustomID(id string) *ButtonBuilder {
	b.component.CustomID = id
	return b
}
func (b *ButtonBuilder) SetLabel(label string) *ButtonBuilder {
	b.component.Label = label
	return b
}
func (b *ButtonBuilder) SetURL(url string) *ButtonBuilder {
	b.component.URL = url
	return b
}
func (b *ButtonBuilder) SetStyle(style ButtonStyle) *ButtonBuilder {
	b.component.Style = style
	return b
}
func (b *ButtonBuilder) SetDisabled(disabled bool) *ButtonBuilder {
	b.component.Disabled = disabled
	return b
}
func (b *ButtonBuilder) Build() Button               { return b.component }
func (b ButtonBuilder) Type() ComponentType          { return b.component.Type() }
func (b ButtonBuilder) MarshalJSON() ([]byte, error) { return b.component.MarshalJSON() }

// ChannelSelectBuilder builds a channel select menu.
type ChannelSelectBuilder struct{ component ChannelSelect }

func NewChannelSelectBuilder() *ChannelSelectBuilder { return &ChannelSelectBuilder{} }
func (b *ChannelSelectBuilder) SetCustomID(id string) *ChannelSelectBuilder {
	b.component.CustomID = id
	return b
}
func (b *ChannelSelectBuilder) SetCustomId(id string) *ChannelSelectBuilder {
	return b.SetCustomID(id)
}
func (b *ChannelSelectBuilder) SetPlaceholder(placeholder string) *ChannelSelectBuilder {
	b.component.Placeholder = placeholder
	return b
}
func (b *ChannelSelectBuilder) SetChannelTypes(types ...ChannelType) *ChannelSelectBuilder {
	b.component.ChannelTypes = append([]ChannelType(nil), types...)
	return b
}
func (b *ChannelSelectBuilder) SetMinValues(value int) *ChannelSelectBuilder {
	b.component.MinValues = &value
	return b
}
func (b *ChannelSelectBuilder) SetMaxValues(value int) *ChannelSelectBuilder {
	b.component.MaxValues = &value
	return b
}
func (b *ChannelSelectBuilder) SetDisabled(disabled bool) *ChannelSelectBuilder {
	b.component.Disabled = disabled
	return b
}
func (b *ChannelSelectBuilder) Build() ChannelSelect        { return b.component }
func (b ChannelSelectBuilder) Type() ComponentType          { return b.component.Type() }
func (b ChannelSelectBuilder) MarshalJSON() ([]byte, error) { return b.component.MarshalJSON() }

// SelectOptionBuilder builds one string-select option.
type SelectOptionBuilder struct{ option SelectOption }

func NewSelectOptionBuilder() *SelectOptionBuilder { return &SelectOptionBuilder{} }
func (b *SelectOptionBuilder) SetLabel(value string) *SelectOptionBuilder {
	b.option.Label = value
	return b
}
func (b *SelectOptionBuilder) SetValue(value string) *SelectOptionBuilder {
	b.option.Value = value
	return b
}
func (b *SelectOptionBuilder) SetDescription(value string) *SelectOptionBuilder {
	b.option.Description = value
	return b
}
func (b *SelectOptionBuilder) SetDefault(value bool) *SelectOptionBuilder {
	b.option.Default = value
	return b
}
func (b *SelectOptionBuilder) Build() SelectOption { return b.option }

// StringSelectBuilder builds a string select menu.
type StringSelectBuilder struct{ component StringSelect }

func NewStringSelectBuilder() *StringSelectBuilder { return &StringSelectBuilder{} }
func (b *StringSelectBuilder) SetCustomID(value string) *StringSelectBuilder {
	b.component.CustomID = value
	return b
}
func (b *StringSelectBuilder) SetCustomId(value string) *StringSelectBuilder {
	return b.SetCustomID(value)
}
func (b *StringSelectBuilder) SetPlaceholder(value string) *StringSelectBuilder {
	b.component.Placeholder = value
	return b
}
func (b *StringSelectBuilder) AddOptions(options ...interface{}) *StringSelectBuilder {
	for _, option := range options {
		switch value := option.(type) {
		case SelectOption:
			b.component.Options = append(b.component.Options, value)
		case *SelectOptionBuilder:
			if value != nil {
				b.component.Options = append(b.component.Options, value.Build())
			}
		}
	}
	return b
}
func (b *StringSelectBuilder) SetMinValues(value int) *StringSelectBuilder {
	b.component.MinValues = &value
	return b
}
func (b *StringSelectBuilder) SetMaxValues(value int) *StringSelectBuilder {
	b.component.MaxValues = &value
	return b
}
func (b *StringSelectBuilder) SetDisabled(value bool) *StringSelectBuilder {
	b.component.Disabled = value
	return b
}
func (b *StringSelectBuilder) Build() StringSelect         { return b.component }
func (b StringSelectBuilder) Type() ComponentType          { return b.component.Type() }
func (b StringSelectBuilder) MarshalJSON() ([]byte, error) { return b.component.MarshalJSON() }

// RoleSelectBuilder builds a role select menu.
type RoleSelectBuilder struct{ component RoleSelect }

func NewRoleSelectBuilder() *RoleSelectBuilder { return &RoleSelectBuilder{} }
func (b *RoleSelectBuilder) SetCustomID(value string) *RoleSelectBuilder {
	b.component.CustomID = value
	return b
}
func (b *RoleSelectBuilder) SetCustomId(value string) *RoleSelectBuilder { return b.SetCustomID(value) }
func (b *RoleSelectBuilder) SetPlaceholder(value string) *RoleSelectBuilder {
	b.component.Placeholder = value
	return b
}
func (b *RoleSelectBuilder) SetMinValues(value int) *RoleSelectBuilder {
	b.component.MinValues = &value
	return b
}
func (b *RoleSelectBuilder) SetMaxValues(value int) *RoleSelectBuilder {
	b.component.MaxValues = &value
	return b
}
func (b *RoleSelectBuilder) SetDisabled(value bool) *RoleSelectBuilder {
	b.component.Disabled = value
	return b
}
func (b *RoleSelectBuilder) Build() RoleSelect           { return b.component }
func (b RoleSelectBuilder) Type() ComponentType          { return b.component.Type() }
func (b RoleSelectBuilder) MarshalJSON() ([]byte, error) { return b.component.MarshalJSON() }

// UserSelectBuilder builds a user select menu.
type UserSelectBuilder struct{ component UserSelect }

func NewUserSelectBuilder() *UserSelectBuilder { return &UserSelectBuilder{} }
func (b *UserSelectBuilder) SetCustomID(value string) *UserSelectBuilder {
	b.component.CustomID = value
	return b
}
func (b *UserSelectBuilder) SetCustomId(value string) *UserSelectBuilder { return b.SetCustomID(value) }
func (b *UserSelectBuilder) SetPlaceholder(value string) *UserSelectBuilder {
	b.component.Placeholder = value
	return b
}
func (b *UserSelectBuilder) SetMinValues(value int) *UserSelectBuilder {
	b.component.MinValues = &value
	return b
}
func (b *UserSelectBuilder) SetMaxValues(value int) *UserSelectBuilder {
	b.component.MaxValues = &value
	return b
}
func (b *UserSelectBuilder) SetDisabled(value bool) *UserSelectBuilder {
	b.component.Disabled = value
	return b
}
func (b *UserSelectBuilder) Build() UserSelect           { return b.component }
func (b UserSelectBuilder) Type() ComponentType          { return b.component.Type() }
func (b UserSelectBuilder) MarshalJSON() ([]byte, error) { return b.component.MarshalJSON() }

// TextInputBuilder builds a modal text input. Text inputs should be wrapped in
// an ActionRowBuilder before being passed to a modal.
type TextInputBuilder struct{ component TextInput }

func NewTextInputBuilder() *TextInputBuilder { return &TextInputBuilder{} }
func (b *TextInputBuilder) SetCustomID(value string) *TextInputBuilder {
	b.component.CustomID = value
	return b
}
func (b *TextInputBuilder) SetCustomId(value string) *TextInputBuilder { return b.SetCustomID(value) }
func (b *TextInputBuilder) SetStyle(value TextInputStyle) *TextInputBuilder {
	b.component.Style = value
	return b
}
func (b *TextInputBuilder) SetLabel(value string) *TextInputBuilder {
	b.component.Label = value
	return b
}
func (b *TextInputBuilder) SetPlaceholder(value string) *TextInputBuilder {
	b.component.Placeholder = value
	return b
}
func (b *TextInputBuilder) SetValue(value string) *TextInputBuilder {
	b.component.Value = value
	return b
}
func (b *TextInputBuilder) SetRequired(value bool) *TextInputBuilder {
	b.component.Required = &value
	return b
}
func (b *TextInputBuilder) SetMinLength(value int) *TextInputBuilder {
	b.component.MinLength = &value
	return b
}
func (b *TextInputBuilder) SetMaxLength(value int) *TextInputBuilder {
	b.component.MaxLength = &value
	return b
}
func (b *TextInputBuilder) Build() TextInput            { return b.component }
func (b TextInputBuilder) Type() ComponentType          { return b.component.Type() }
func (b TextInputBuilder) MarshalJSON() ([]byte, error) { return b.component.MarshalJSON() }

// ModalData is the interaction callback data required to display a modal.
type ModalData struct {
	CustomID   string
	Title      string
	Components []Component
}

// ModalBuilder builds a modal callback payload without coupling components to
// the interactions package.
type ModalBuilder struct{ data ModalData }

func NewModalBuilder() *ModalBuilder                           { return &ModalBuilder{} }
func (b *ModalBuilder) SetCustomID(value string) *ModalBuilder { b.data.CustomID = value; return b }
func (b *ModalBuilder) SetCustomId(value string) *ModalBuilder { return b.SetCustomID(value) }
func (b *ModalBuilder) SetTitle(value string) *ModalBuilder    { b.data.Title = value; return b }
func (b *ModalBuilder) AddComponents(value ...Component) *ModalBuilder {
	b.data.Components = append(b.data.Components, value...)
	return b
}
func (b *ModalBuilder) Build() ModalData { return b.data }

// ActionRowBuilder builds a legacy action row containing interactive components.
type ActionRowBuilder struct{ component ActionRow }

func NewActionRowBuilder() *ActionRowBuilder { return &ActionRowBuilder{} }
func (b *ActionRowBuilder) AddComponents(components ...Component) *ActionRowBuilder {
	b.component.Components = append(b.component.Components, components...)
	return b
}
func (b *ActionRowBuilder) Build() ActionRow            { return b.component }
func (b ActionRowBuilder) Type() ComponentType          { return b.component.Type() }
func (b ActionRowBuilder) MarshalJSON() ([]byte, error) { return b.component.MarshalJSON() }

// SectionBuilder builds a Components V2 section.
type SectionBuilder struct{ component Section }

func NewSectionBuilder() *SectionBuilder { return &SectionBuilder{} }
func (b *SectionBuilder) AddTextDisplayComponents(components ...Component) *SectionBuilder {
	b.component.Components = append(b.component.Components, components...)
	return b
}
func (b *SectionBuilder) AddTextDisplayBuilders(builders ...*TextDisplayBuilder) *SectionBuilder {
	for _, builder := range builders {
		if builder != nil {
			b.component.Components = append(b.component.Components, builder.Build())
		}
	}
	return b
}
func (b *SectionBuilder) SetThumbnailAccessory(accessory Component) *SectionBuilder {
	b.component.Accessory = accessory
	return b
}
func (b *SectionBuilder) SetButtonAccessory(accessory Component) *SectionBuilder {
	b.component.Accessory = accessory
	return b
}
func (b *SectionBuilder) Build() Section              { return b.component }
func (b SectionBuilder) Type() ComponentType          { return b.component.Type() }
func (b SectionBuilder) MarshalJSON() ([]byte, error) { return b.component.MarshalJSON() }

// ContainerBuilder builds a Components V2 container.
type ContainerBuilder struct{ component Container }

func NewContainerBuilder() *ContainerBuilder { return &ContainerBuilder{} }
func (b *ContainerBuilder) SetAccentColor(color int) *ContainerBuilder {
	b.component.AccentColor = color
	return b
}
func (b *ContainerBuilder) AddMediaGalleryComponents(components ...Component) *ContainerBuilder {
	b.component.Components = append(b.component.Components, components...)
	return b
}
func (b *ContainerBuilder) AddSectionComponents(components ...Component) *ContainerBuilder {
	b.component.Components = append(b.component.Components, components...)
	return b
}
func (b *ContainerBuilder) AddSeparatorComponents(components ...Component) *ContainerBuilder {
	b.component.Components = append(b.component.Components, components...)
	return b
}
func (b *ContainerBuilder) AddTextDisplayComponents(components ...Component) *ContainerBuilder {
	b.component.Components = append(b.component.Components, components...)
	return b
}
func (b *ContainerBuilder) AddFileComponents(components ...Component) *ContainerBuilder {
	b.component.Components = append(b.component.Components, components...)
	return b
}
func (b *ContainerBuilder) AddActionRowComponents(components ...Component) *ContainerBuilder {
	b.component.Components = append(b.component.Components, components...)
	return b
}
func (b *ContainerBuilder) Build() Container            { return b.component }
func (b ContainerBuilder) Type() ComponentType          { return b.component.Type() }
func (b ContainerBuilder) MarshalJSON() ([]byte, error) { return b.component.MarshalJSON() }

// ChannelSelectMenuBuilder is an alias matching Discord.js terminology.
type ChannelSelectMenuBuilder = ChannelSelectBuilder

// NewChannelSelectMenuBuilder creates a channel select menu builder.
func NewChannelSelectMenuBuilder() *ChannelSelectBuilder { return NewChannelSelectBuilder() }

// StringSelectMenuBuilder is an alias matching Discord.js terminology.
type StringSelectMenuBuilder = StringSelectBuilder

func NewStringSelectMenuBuilder() *StringSelectBuilder { return NewStringSelectBuilder() }

// RoleSelectMenuBuilder is an alias matching Discord.js terminology.
type RoleSelectMenuBuilder = RoleSelectBuilder

func NewRoleSelectMenuBuilder() *RoleSelectBuilder { return NewRoleSelectBuilder() }

// UserSelectMenuBuilder is an alias matching Discord.js terminology.
type UserSelectMenuBuilder = UserSelectBuilder

func NewUserSelectMenuBuilder() *UserSelectBuilder { return NewUserSelectBuilder() }
