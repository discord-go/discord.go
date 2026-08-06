# Components

## Overview

The `components` package models the JSON tree used in message and interaction
payloads. It supports legacy action rows, buttons, string and entity select
menus, text inputs, and Components V2 text displays, separators, sections,
thumbnails, media galleries, files, and containers. Every concrete component
implements `Component`, whose only method is `Type() ComponentType`.

## Architecture

Concrete components implement custom `MarshalJSON` methods that add their
numeric `type` field. Builders keep a private component value and return a
value from `Build`; they mutate the builder in place and return the builder for
chaining. `Unmarshal(data)` reads `type` and recursively decodes supported
nested children into the `Component` interface.

Legacy components use type values 1 through 8: action row, button, string
select, text input, user select, role select, mentionable select, and channel
select. Components V2 uses 9 through 14 and 17 for section, text display,
thumbnail, media gallery, file, separator, and container. Builders do not
perform all Discord validation, such as component count, custom ID length, or
the style-specific requirement that a button URL and custom ID are mutually
exclusive.

## Quick Start

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/discord-go/discord.go/components"
)

func main() {
	button := components.NewButtonBuilder().
		SetStyle(components.ButtonStylePrimary).
		SetCustomID("confirm").
		SetLabel("Confirm").
		Build()
	row := components.NewActionRowBuilder().AddComponents(button).Build()

	data, err := json.Marshal(row)
	if err != nil {
		panic(err)
	}
	decoded, err := components.Unmarshal(data)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data), decoded.Type() == components.ComponentTypeActionRow)
}
```

## Creating Legacy Components

`ButtonBuilder` has `SetCustomID`, `SetLabel`, `SetURL`, `SetStyle`, and
`SetDisabled`; styles are Primary, Secondary, Success, Danger, Link, and
Premium. `ActionRowBuilder.AddComponents` accepts any `Component`. String
selects use `StringSelectBuilder.AddOptions`, which accepts `SelectOption` or
`*SelectOptionBuilder`; option builders expose label, value, description, and
default. User, role, mentionable, and channel select builders expose custom ID,
placeholder, min/max values, disabled, and channel types where applicable.
The `SetCustomId` spelling and `*MenuBuilder` aliases are compatibility names
for the same methods and types.

Text inputs use `TextInputBuilder` with custom ID, style (Short or Paragraph),
label, placeholder, value, required, and min/max lengths. Discord expects a
text input inside an action row when it is used in a modal. `ModalBuilder`
collects component rows and returns `ModalData` with custom ID, title, and
components; it deliberately avoids importing the interactions package.

## Creating Components V2

`TextDisplayBuilder` sets content. `SeparatorBuilder` sets divider and
`SeparatorSpacingSmall` or `SeparatorSpacingLarge`. `ThumbnailBuilder` and
`MediaGalleryItemBuilder` set media URLs; media gallery builders accept values
or item builders. `FileBuilder` sets the file URL. `SectionBuilder` accepts
text displays and either a thumbnail or button accessory. `ContainerBuilder`
sets an accent color and accepts action rows, files, galleries, sections,
separators, and text displays. The package does not upload referenced media;
use attachment URLs and the multipart tools in [`../rest/uploads.md`](../rest/uploads.md).

## Using Components

Put legacy components in `messages.MessageSend.Components` or
`interactions.InteractionCallbackData.Components`. Components V2 messages must
also carry `messages.FlagIsComponentsV2`. Serialize the final parent payload,
not individual children, so nested interface values receive their custom
marshalers. `Unmarshal` is the safe entry point for a single component JSON
value and handles nested action rows, sections, and containers.

## Common Patterns

Build reusable controls as values, then put them into an action row or V2
container. Use stable custom IDs that encode an operation and a short record
key; treat interaction custom IDs as untrusted input when they return. Keep
select option values small and validate them in the interaction handler.

## Best Practices

Call `Build` once a builder is configured and do not share a mutable builder
between goroutines. Validate Discord limits in application code because the
builders primarily provide shape and discoverability. Keep the selected
component type aligned with the interaction response type and use explicit
pointer values for optional min/max fields when constructing structs directly.

## Common Mistakes

A `Component` interface cannot be JSON-decoded with plain `encoding/json`
without a concrete type; use `components.Unmarshal`. A link button uses a URL,
while an interactive button uses a custom ID. A text input is not a top-level
modal child in Discord's legacy format. V2 components are not automatically
enabled by constructing a `Container`; the message flag is separate.

## API Walkthrough

The exported models are `ActionRow`, `Button`, `ChannelSelect`, `ChannelType`,
`MentionableSelect`,
`RoleSelect`, `StringSelect`, `UserSelect`, `TextInput`, `SelectOption`,
`TextDisplay`, `Separator`, `Section`, `Thumbnail`, `MediaGalleryItem`,
`MediaGallery`, `File`, and `Container`, plus `Component`, `ComponentType`,
all component/style constants, and `ModalData`. Builders are
`NewActionRowBuilder`, `NewButtonBuilder`, `NewChannelSelectBuilder`,
`NewStringSelectBuilder`, `NewRoleSelectBuilder`, `NewUserSelectBuilder`,
`NewTextInputBuilder`, `NewSelectOptionBuilder`, `NewModalBuilder`,
`NewTextDisplayBuilder`, `NewSeparatorBuilder`, `NewThumbnailBuilder`,
`NewMediaGalleryItemBuilder`, `NewMediaGalleryBuilder`, `NewFileBuilder`,
`NewSectionBuilder`, and `NewContainerBuilder`, with the documented setters,
adders, `Build`, `Type`, and marshal methods. The menu-builder aliases and
constructors are also exported. `Unmarshal` decodes a component.

## Examples

The Quick Start program is complete and runnable. For a full interaction
response, combine its component with [`../interactions/`](../interactions/README.md)
and for a message send combine it with [`../messages/`](../messages/README.md).

## Related APIs

- [`../messages/`](../messages/README.md)
- [`../interactions/`](../interactions/README.md)
- [`../rest/uploads.md`](../rest/uploads.md)
