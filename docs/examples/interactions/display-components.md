# Display Components

## Overview

Components V2 provides display-oriented primitives in addition to legacy
action rows: `TextDisplay`, `Separator`, `Section`, `Thumbnail`,
`MediaGallery`, `File`, and `Container`. In `discord.go`, build these with the
typed builders and set `messages.FlagIsComponentsV2` on the response.

## Tutorial: Build A V2 Layout

1. Build text displays and other visual components.
2. Compose them into a `Section` or `Container`.
3. Add an action row when the layout needs an interactive control.
4. Put the top-level components in `InteractionCallbackData.Components`.
5. Set `Flags: messages.FlagIsComponentsV2`.

V2 payloads are not ordinary content messages. Do not omit the flag, and do not
assume a legacy `content` field is the right place for the primary layout.

## Complete Runnable Example

Copy to `examples/display-components/main.go`, set `DISCORD_TOKEN`, and run it.
Invoke `/display` to render a V2 container with a button accessory.

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
	router.Command("display", "Show a Components V2 layout", func(ctx *bot.InteractionContext) {
		heading := components.NewTextDisplayBuilder().
			SetContent("# Components V2\nA typed display layout built with discord.go.").
			Build()
		separator := components.NewSeparatorBuilder().
			SetDivider(true).
			SetSpacing(components.SeparatorSpacingSmall).
			Build()
		status := components.NewTextDisplayBuilder().
			SetContent("The status is **ready**. Use the button to acknowledge this panel.").
			Build()
		section := components.NewSectionBuilder().
			AddTextDisplayComponents(status).
			SetButtonAccessory(components.NewButtonBuilder().
				SetCustomID("display:ack").
				SetLabel("Acknowledge").
				SetStyle(components.ButtonStylePrimary).
				Build()).
			Build()
		container := components.NewContainerBuilder().
			SetAccentColor(0x5865F2).
			AddTextDisplayComponents(heading).
			AddSeparatorComponents(separator).
			AddSectionComponents(section).
			Build()

		if err := ctx.ReplyComplex(&interactions.InteractionCallbackData{
			Flags:      messages.FlagIsComponentsV2,
			Components: []components.Component{container},
		}); err != nil {
			log.Printf("display response: %v", err)
		}
	})

	router.Button("display:ack", func(ctx *bot.InteractionContext) {
		if err := ctx.UpdateContent("The panel was acknowledged."); err != nil {
			log.Printf("display update: %v", err)
		}
	})

	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Components V2 And Legacy Components

V2 display components can be composed with the V2 builders. Interactive menus
and buttons remain action-row components in this repository. A V2 `Section` can
use a button or thumbnail accessory, and a V2 `Container` can contain sections,
separators, text displays, galleries, files, and action rows through its typed
builder methods.

When attaching a file, build a `File` component with an
`attachment://filename` URL and upload a matching `rest.File` in the same
multipart response. `examples/v2_components/main.go` is the repository's
source-backed file example.

## Common Mistakes

Wrong:

```go
_ = ctx.ReplyComplex(&interactions.InteractionCallbackData{
	Components: []components.Component{container},
})
```

Correct:

```go
_ = ctx.ReplyComplex(&interactions.InteractionCallbackData{
	Flags:      messages.FlagIsComponentsV2,
	Components: []components.Component{container},
})
```

Also avoid using unstable user input as attachment names or external media
URLs. Validate names, sizes, and URLs before constructing a multipart request.

## Expected Result

`/display` returns a V2 container with a heading, divider, status section, and
button accessory. Clicking the button acknowledges the component interaction.

## Related Pages

- [Buttons](buttons.md)
- [Action Rows](action-rows.md)
- [Canvas Alternatives](../more-to-know/canvas-alternatives.md)
- [Embeds](../more-to-know/embeds.md)
