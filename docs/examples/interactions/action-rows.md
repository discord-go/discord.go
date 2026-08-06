# Action Rows

## Overview

An action row is the legacy Discord container for interactive components. In
`discord.go`, `components.NewActionRowBuilder` accepts buttons and select menus
for messages, and text inputs for modals. The row itself is sent as a
`components.Component` inside `InteractionCallbackData.Components` or
`messages.MessageSend.Components`.

## Tutorial: Compose A Row

1. Build each child component.
2. Add the children to one `ActionRowBuilder`.
3. Put the built row in the response's `Components` slice.
4. Register routes whose IDs exactly match custom-ID children.
5. Update or disable the row after an interaction when the workflow is done.

Legacy action rows have Discord composition limits. Keep interactive controls
in rows, and do not use an action row as a substitute for a Components V2
container. See [Display Components](display-components.md) for V2 layouts.

## Complete Runnable Example

Copy this to `examples/action-rows/main.go`, set `DISCORD_TOKEN`, and run it.
Invoke `/row` and press either button.

```go
package main

import (
	"log"
	"os"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/interactions"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}

	router := bot.NewRouter()
	router.Command("row", "Show two buttons in one action row", func(ctx *bot.InteractionContext) {
		approve := components.NewButtonBuilder().
			SetCustomID("row:approve").
			SetLabel("Approve").
			SetStyle(components.ButtonStyleSuccess).
			Build()
		cancel := components.NewButtonBuilder().
			SetCustomID("row:cancel").
			SetLabel("Cancel").
			SetStyle(components.ButtonStyleDanger).
			Build()
		row := components.NewActionRowBuilder().AddComponents(approve, cancel).Build()
		if err := ctx.ReplyComplex(&interactions.InteractionCallbackData{
			Content:    "Choose one action.",
			Components: []components.Component{row},
		}); err != nil {
			log.Printf("send action row: %v", err)
		}
	})

	router.Button("row:approve", func(ctx *bot.InteractionContext) {
		if err := ctx.UpdateContent("Approved."); err != nil {
			log.Printf("approve response: %v", err)
		}
	})
	router.Button("row:cancel", func(ctx *bot.InteractionContext) {
		if err := ctx.UpdateContent("Cancelled."); err != nil {
			log.Printf("cancel response: %v", err)
		}
	})

	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Explanation

`ActionRow` implements `components.Component`, so the built value can be placed
directly into the callback. A row can contain multiple buttons, but each
custom-ID button must have a unique ID within the message. A link button uses
`SetURL` and `ButtonStyleLink` and does not produce a routed interaction.

For a modal, put each `TextInput` in its own action row before passing the rows
to `NewModalBuilder`. The same row type is used, but the parent payload is a
modal callback rather than a message callback.

## Basic Usage

```go
row := components.NewActionRowBuilder().AddComponents(
	components.NewButtonBuilder().
		SetCustomID("refresh").
		SetLabel("Refresh").
		SetStyle(components.ButtonStyleSecondary).
		Build(),
).Build()
```

Use `[]components.Component{row}` in a response. Do not pass the builder
pointer itself; call `Build` so the payload contains the concrete component.

## Common Mistakes

Wrong:

```go
row := components.NewActionRowBuilder().AddComponents(
	components.NewButtonBuilder().SetLabel("Save").Build(),
).Build()
```

The button has no `CustomID` and is not a link, so it cannot generate a useful
application interaction. Correct it with either `SetCustomID` or `SetURL`:

```go
button := components.NewButtonBuilder().
	SetCustomID("save").
	SetLabel("Save").
	SetStyle(components.ButtonStylePrimary).
	Build()
row := components.NewActionRowBuilder().AddComponents(button).Build()
```

Another common error is to put a V2 display component into an old row without
checking Discord's component composition rules. Use a V2 `Container` for
display-only layout and reserve action rows for supported interactive controls.

## Patterns And Best Practices

- Keep IDs namespaced, such as `settings:save` and `settings:cancel`.
- Use `ButtonPrefix` only when the suffix is parsed and authorized.
- Build a fresh row when disabling controls after completion.
- Use `DeferUpdate` before slow work triggered by a component.
- Do not rely on a disabled visual state for authorization; recheck ownership.

## Expected Result

`/row` sends two buttons in one action row. Each route edits the original
message, proving that the row is only a layout container and the child custom
IDs determine dispatch.

## Related Pages

- [Buttons](buttons.md)
- [Select Menus](select-menus.md)
- [Modals](modals.md)
- [Interactions](interactions.md)
