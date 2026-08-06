package components

import "encoding/json"

// TextDisplay represents a TextDisplay component (V2).
type TextDisplay struct {
	Content string `json:"content"`
}

func (c TextDisplay) Type() ComponentType { return ComponentTypeTextDisplay }
func (c TextDisplay) MarshalJSON() ([]byte, error) {
	type alias TextDisplay
	return json.Marshal(struct {
		Type ComponentType `json:"type"`
		alias
	}{Type: c.Type(), alias: (alias)(c)})
}

// Separator represents a Separator component (V2).
type Separator struct {
	Divider bool `json:"divider,omitempty"`
	Spacing int  `json:"spacing,omitempty"`
}

func (c Separator) Type() ComponentType { return ComponentTypeSeparator }
func (c Separator) MarshalJSON() ([]byte, error) {
	type alias Separator
	return json.Marshal(struct {
		Type ComponentType `json:"type"`
		alias
	}{Type: c.Type(), alias: (alias)(c)})
}

// Section represents a Section component (V2).
type Section struct {
	Components []Component `json:"components"`
	Accessory  Component   `json:"accessory,omitempty"`
}

func (c Section) Type() ComponentType { return ComponentTypeSection }
func (c Section) MarshalJSON() ([]byte, error) {
	type alias Section
	return json.Marshal(struct {
		Type ComponentType `json:"type"`
		alias
	}{Type: c.Type(), alias: (alias)(c)})
}

// Thumbnail represents a Thumbnail component (V2).
type Thumbnail struct {
	Media struct {
		URL string `json:"url"`
	} `json:"media"`
}

func (c Thumbnail) Type() ComponentType { return ComponentTypeThumbnail }
func (c Thumbnail) MarshalJSON() ([]byte, error) {
	type alias Thumbnail
	return json.Marshal(struct {
		Type ComponentType `json:"type"`
		alias
	}{Type: c.Type(), alias: (alias)(c)})
}

// MediaGalleryItem represents an item in a MediaGallery.
type MediaGalleryItem struct {
	Media struct {
		URL string `json:"url"`
	} `json:"media"`
}

// MediaGallery represents a MediaGallery component (V2).
type MediaGallery struct {
	Items []MediaGalleryItem `json:"items"`
}

func (c MediaGallery) Type() ComponentType { return ComponentTypeMediaGallery }
func (c MediaGallery) MarshalJSON() ([]byte, error) {
	type alias MediaGallery
	return json.Marshal(struct {
		Type ComponentType `json:"type"`
		alias
	}{Type: c.Type(), alias: (alias)(c)})
}

// Container represents a Container component (V2).
type Container struct {
	Components  []Component `json:"components"`
	AccentColor int         `json:"accent_color,omitempty"`
}

func (c Container) Type() ComponentType { return ComponentTypeContainer }
func (c Container) MarshalJSON() ([]byte, error) {
	type alias Container
	return json.Marshal(struct {
		Type ComponentType `json:"type"`
		alias
	}{Type: c.Type(), alias: (alias)(c)})
}

// File represents a File component (V2).
type File struct {
	File struct {
		URL string `json:"url"`
	} `json:"file"`
}

func (c File) Type() ComponentType { return ComponentTypeFile }
func (c File) MarshalJSON() ([]byte, error) {
	type alias File
	return json.Marshal(struct {
		Type ComponentType `json:"type"`
		alias
	}{Type: c.Type(), alias: (alias)(c)})
}
