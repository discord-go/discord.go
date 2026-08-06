package components

import (
	"encoding/json"
	"testing"
)

func TestComponentsV2BuildersAndNestedUnmarshal(t *testing.T) {
	text := NewTextDisplayBuilder().SetContent("Components V2").Build()
	separator := NewSeparatorBuilder().SetDivider(true).SetSpacing(SeparatorSpacingSmall).Build()
	thumbnail := NewThumbnailBuilder().SetURL("https://example.com/image.png").Build()
	section := NewSectionBuilder().AddTextDisplayComponents(text).SetThumbnailAccessory(thumbnail).Build()
	selectMenu := NewChannelSelectBuilder().SetCustomID("channel_select").SetPlaceholder("Select a channel...").Build()
	row := NewActionRowBuilder().AddComponents(selectMenu).Build()
	gallery := NewMediaGalleryBuilder().AddItems(NewMediaGalleryItemBuilder().SetURL("https://example.com/one.png").Build()).Build()
	file := NewFileBuilder().SetURL("attachment://export.json").Build()
	container := NewContainerBuilder().SetAccentColor(0x5865F2).
		AddMediaGalleryComponents(gallery).
		AddSectionComponents(section).
		AddSeparatorComponents(separator).
		AddTextDisplayComponents(text).
		AddFileComponents(file).
		Build()

	send := struct {
		Flags      int         `json:"flags"`
		Components []Component `json:"components"`
	}{Flags: 32768, Components: []Component{text, separator, section, row, container}}
	encoded, err := json.Marshal(send)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(encoded) {
		t.Fatal("builder payload is not valid JSON")
	}
	componentJSON, err := json.Marshal(send.Components)
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Components []json.RawMessage `json:"components"`
	}
	if err := json.Unmarshal(append(append([]byte(`{"components":`), componentJSON...), '}'), &decoded); err != nil {
		t.Fatal(err)
	}
	components := make([]Component, 0, len(decoded.Components))
	for _, raw := range decoded.Components {
		component, decodeErr := Unmarshal(raw)
		if decodeErr != nil {
			t.Fatal(decodeErr)
		}
		components = append(components, component)
	}
	if len(components) != 5 {
		t.Fatalf("decoded components = %d, want 5", len(decoded.Components))
	}
	if components[4].Type() != ComponentTypeContainer {
		t.Fatalf("last component type = %d, want container", components[4].Type())
	}
	containerValue, ok := components[4].(Container)
	if !ok || len(containerValue.Components) != 5 {
		t.Fatalf("decoded container = %#v", components[4])
	}
}
