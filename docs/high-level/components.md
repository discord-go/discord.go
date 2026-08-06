# Components And Components V2

## Overview

Discord components are typed values in `discord.go/components`. Legacy action
rows contain interactive buttons or select menus; Components V2 adds text
displays, separators, sections, thumbnails, media galleries, files, and
containers. Builders make these values readable and ensure the serialized
payload carries each component's numeric `type`.

## Architecture

Every component implements `components.Component` with `Type()`. Builders return
concrete values from `Build`, and those values can be placed in
`interactions.InteractionCallbackData.Components` or `messages.MessageSend.Components`.
The `json.Marshaler` implementations add the Discord type field and recursively
serialize nested components.

Components V2 payloads must set `messages.FlagIsComponentsV2`. A container can
hold V2 children, while an action row is the normal wrapper for legacy
interactive components and modal text inputs. Routes are separate from payload
construction: use `Router.Button`, `Select`, `Modal`, or `Autocomplete` to
receive the resulting interaction.

## Quick Start

This complete program sends a Components V2 text display from `/panel`.

```go
package main

import (
	"log"
	"os"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/messages"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}

	router := bot.NewRouter()
	router.Command("panel", "Show a component panel", func(ctx *bot.InteractionContext) {
		text := components.NewTextDisplayBuilder().
			SetContent("This message uses Components V2.").Build()
		data := &interactions.InteractionCallbackData{
			Flags:      messages.FlagIsComponentsV2,
			Components: []components.Component{text},
		}
		if err := ctx.ReplyComplex(data); err != nil {
			log.Printf("reply: %v", err)
		}
	})

	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

Run it with `DISCORD_TOKEN=... go run .`. The application must be installed in
a guild where you can invoke the command.

## Creating/Configuration

Use the builder constructors such as `NewButtonBuilder`,
`NewStringSelectBuilder`, `NewActionRowBuilder`, `NewModalBuilder`, and
`NewContainerBuilder`. Builders are mutable and fluent; call `Build` when the
value is ready. For direct struct literals, ensure a component is a concrete
value implementing `Component` and not an untyped map.

When sending ordinary messages, use `MessageContext.ReplyComplex` with a
`messages.MessageSend`. When sending an interaction response, use
`InteractionContext.ReplyComplex` with `*interactions.InteractionCallbackData`.
For Components V2 files, use the multipart variants and an `attachment://`
URL, as shown in the [Components V2 example](../examples/interactions/components-v2.md).

## Using

### Basic: build a legacy action row

Build a button or select, put it in an action row, and send the row in a normal
message or interaction response. Interactive components require a `CustomID`;
link buttons instead use a URL and do not create an interaction.

### Intermediate: compose V2 layout

Use text displays, separators, sections with text-display children and an
accessory, media galleries, and containers. `SectionBuilder` accepts a thumbnail
or button accessory. `ContainerBuilder` accepts typed child components.

### Advanced: decode components

`components.Unmarshal(data)` returns a concrete `Component` for supported type
values and recursively decodes action rows, sections, and containers. Messages
also use this decoder while unmarshalling `Message.Components`.

## Common Patterns

- Keep custom IDs stable and route them with exact or prefix matching.
- Build immutable values once per response; do not reuse a builder concurrently.
- Use `FlagIsComponentsV2` whenever a response is a Components V2 payload.
- Keep interactive children in action rows for legacy layouts and use V2 layout
  components only where their payload rules permit them.
- Pair component construction with a route in the same feature registration code.

## Best Practices

### Use builders for complex layouts

Why: nested component interfaces are easy to get wrong with raw struct literals.

Pros: readable construction, typed nesting, and automatic type serialization.

Cons: builders are mutable and introduce a small amount of setup code.

### Separate rendering from routing

Why: a component's `CustomID` is the protocol key while a route owns behavior.

Pros: the same layout can be rendered in several contexts and route tests are
small.

Cons: changing an ID requires updating both the builder and route.

### Validate payload mode before sending

Why: Components V2 has message flag and content constraints enforced by Discord.

Pros: errors are caught near the response construction.

Cons: the Go types do not perform every Discord semantic validation, so API
errors still need handling.

## Common Mistakes

Incorrect: omitting the Components V2 flag.

```go
_ = ctx.ReplyComplex(&interactions.InteractionCallbackData{
	Components: []components.Component{text},
})
```

Correct:

```go
_ = ctx.ReplyComplex(&interactions.InteractionCallbackData{
	Flags:      messages.FlagIsComponentsV2,
	Components: []components.Component{text},
})
```

Incorrect: expecting a link button to reach a route.

```go
router.Button("docs", handler)
button := components.NewButtonBuilder().SetURL("https://example.com").Build()
```

Correct: use a custom ID for an interactive button, or treat a URL button as a
navigation-only component.

```go
router.Button("docs", handler)
button := components.NewButtonBuilder().
	SetCustomID("docs").SetLabel("Docs").
	SetStyle(components.ButtonStylePrimary).Build()
```

## API Walkthrough

- `Component` requires `Type() ComponentType`; `ComponentType` constants cover
  action rows, buttons, five select types, text inputs, and Components V2 types.
- Legacy values are `ActionRow`, `Button`, `StringSelect`, `UserSelect`,
  `RoleSelect`, `MentionableSelect`, `ChannelSelect`, `SelectOption`, and
  `TextInput`. Their public fields mirror Discord payload fields and their
  `Type` and `MarshalJSON` methods provide typed serialization.
- V2 values are `TextDisplay`, `Separator`, `Section`, `Thumbnail`,
  `MediaGalleryItem`, `MediaGallery`, `File`, and `Container`.
- Builders include `NewTextDisplayBuilder`, `NewSeparatorBuilder`,
  `NewThumbnailBuilder`, `NewThumbnailBuilderWithURL`,
  `NewMediaGalleryItemBuilder`, `NewMediaGalleryItemBuilderWithURL`,
  `NewMediaGalleryBuilder`, `NewFileBuilder`, `NewButtonBuilder`,
  `NewChannelSelectBuilder`, `NewSelectOptionBuilder`, `NewStringSelectBuilder`,
  `NewRoleSelectBuilder`, `NewUserSelectBuilder`, `NewTextInputBuilder`,
  `NewModalBuilder`, `NewActionRowBuilder`, `NewSectionBuilder`, and
  `NewContainerBuilder`.
- Builder methods include `SetContent`, `SetDivider`, `SetSpacing`, `SetURL`,
  `AddItems`, `AddItemBuilders`, `SetCustomID`, `SetCustomId`, `SetLabel`,
  `SetURL`, `SetStyle`, `SetDisabled`, `SetPlaceholder`, `SetChannelTypes`,
  `SetMinValues`, `SetMaxValues`, `SetLabel`, `SetValue`, `SetDescription`,
  `SetDefault`, `SetRequired`, `SetMinLength`, `SetMaxLength`,
  `AddOptions`, `AddComponents`, `AddTextDisplayComponents`,
  `AddTextDisplayBuilders`, `SetThumbnailAccessory`, `SetButtonAccessory`,
  `SetAccentColor`, `AddMediaGalleryComponents`, `AddSectionComponents`,
  `AddSeparatorComponents`, `AddFileComponents`, and
  `AddActionRowComponents`. Each builder's `Build` returns its concrete value.
- `ModalData` stores `CustomID`, `Title`, and `[]Component`; `ModalBuilder` has
  `SetCustomID`, `SetCustomId`, `SetTitle`, `AddComponents`, and `Build`.
- `Unmarshal([]byte) (Component, error)` decodes supported component JSON and
  recursively handles nested children.
- `ChannelSelectMenuBuilder`, `StringSelectMenuBuilder`,
  `RoleSelectMenuBuilder`, and `UserSelectMenuBuilder` are aliases with matching
  constructor functions: `NewChannelSelectMenuBuilder`,
  `NewStringSelectMenuBuilder`, `NewRoleSelectMenuBuilder`, and
  `NewUserSelectMenuBuilder`.

## Examples

- [Components V2](../examples/interactions/components-v2.md)
- [Buttons](buttons.md)
- [Modals](modals.md)
- [Components low-level reference](../low-level/components/README.md)

## Related APIs

- [`buttons.md`](buttons.md) for button-specific routing.
- [`modals.md`](modals.md) for text inputs and submissions.
- [`interactions.md`](interactions.md) for response data and updates.
- [`embeds.md`](embeds.md) for regular rich messages.
