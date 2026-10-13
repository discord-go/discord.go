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
func TestSelectOptionBuilderEmojiWireFormat(t *testing.T) {
	unicode := NewSelectOptionBuilder().
		SetLabel("Development").
		SetValue("dev").
		SetEmoji(map[string]interface{}{"name": "✅"}).
		Build()
	custom := NewSelectOptionBuilder().
		SetLabel("Staging").
		SetValue("stage").
		SetEmoji(map[string]interface{}{"id": "123456789012345678", "name": "check"}).
		Build()
	plain := NewSelectOptionBuilder().
		SetLabel("Production").
		SetValue("prod").
		Build()

	for _, tc := range []struct {
		name string
		opt  SelectOption
		want string
	}{
		{
			name: "unicode emoji",
			opt:  unicode,
			want: `{"label":"Development","value":"dev","emoji":{"name":"✅"}}`,
		},
		{
			name: "custom emoji",
			opt:  custom,
			want: `{"label":"Staging","value":"stage","emoji":{"id":"123456789012345678","name":"check"}}`,
		},
		{
			name: "no emoji omits the field",
			opt:  plain,
			want: `{"label":"Production","value":"prod"}`,
		},
	} {
		got, err := json.Marshal(tc.opt)
		if err != nil {
			t.Fatalf("%s: marshal: %v", tc.name, err)
		}
		if string(got) != tc.want {
			t.Errorf("%s: got %s, want %s", tc.name, got, tc.want)
		}
	}

	menu := NewStringSelectBuilder().
		SetCustomID("select:environment").
		AddOptions(unicode, custom, plain).
		Build()
	encoded, err := json.Marshal(menu)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"type":3,"custom_id":"select:environment","options":[` +
		`{"label":"Development","value":"dev","emoji":{"name":"✅"}},` +
		`{"label":"Staging","value":"stage","emoji":{"id":"123456789012345678","name":"check"}},` +
		`{"label":"Production","value":"prod"}]}`
	if string(encoded) != want {
		t.Errorf("select menu: got %s, want %s", encoded, want)
	}

	var decoded StringSelect
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Options) != 3 {
		t.Fatalf("decoded %d options, want 3", len(decoded.Options))
	}
	first, ok := decoded.Options[0].Emoji.(map[string]interface{})
	if !ok || first["name"] != "✅" {
		t.Errorf("unicode emoji round-trip = %#v", decoded.Options[0].Emoji)
	}
	second, ok := decoded.Options[1].Emoji.(map[string]interface{})
	if !ok || second["id"] != "123456789012345678" || second["name"] != "check" {
		t.Errorf("custom emoji round-trip = %#v", decoded.Options[1].Emoji)
	}
	if decoded.Options[2].Emoji != nil {
		t.Errorf("option without emoji decoded %#v, want nil", decoded.Options[2].Emoji)
	}
}
