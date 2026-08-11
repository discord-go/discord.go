# Buttons

## Overview

`components.Button` has six current styles in this repository: primary,
secondary, success, danger, link, and premium. Application-owned buttons use a
`CustomID` and are routed by `bot.Router`; link buttons use a URL and navigate
without sending a component interaction.

## Tutorial: Interactive And Link Buttons

1. Build a custom-ID button with `SetCustomID`, `SetLabel`, and `SetStyle`.
2. Put it in an action row.
3. Register the exact ID with `router.Button`.
4. Acknowledge a click with `UpdateContent`, `Update`, or `DeferUpdate`.
5. Build link buttons with `SetURL` and `ButtonStyleLink`; do not register a
   route for them.

## Complete Runnable Example

Copy to `examples/buttons/main.go`, set `DISCORD_TOKEN`, and run it. Invoke
`/buttons`, then try both controls.

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
	router.Command("buttons", "Show button types", func(ctx *bot.InteractionContext) {
		custom := components.NewButtonBuilder().
			SetCustomID("buttons:done").
			SetLabel("Complete").
			SetStyle(components.ButtonStyleSuccess).
			Build()
		link := components.NewButtonBuilder().
			SetLabel("Discord developers").
			SetURL("https://discord.com/developers/docs").
			SetStyle(components.ButtonStyleLink).
			Build()
		row := components.NewActionRowBuilder().AddComponents(custom, link).Build()
		if err := ctx.ReplyComplex(&interactions.InteractionCallbackData{
			Content:    "The first button is routed by the bot. The second opens a URL.",
			Components: []components.Component{row},
		}); err != nil {
			log.Printf("button response: %v", err)
		}
	})

	router.Button("buttons:done", func(ctx *bot.InteractionContext) {
		if err := ctx.UpdateContent("Completed. The link button never reached the bot."); err != nil {
			log.Printf("button update: %v", err)
		}
	})

	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Explanation

The string passed to `SetCustomID` must equal the string passed to
`router.Button`. A custom ID is visible to clients and may be replayed, so it
is a routing hint rather than a security boundary. If a button belongs to a
specific user or resource, check that relationship inside the handler.

For IDs such as `ticket:close:123`, use `router.ButtonPrefix("ticket:close:",
handler)`, then read the suffix with `ctx.Suffix()` (which returns `"123"`),
and verify that the actor is allowed to close the ticket before performing
the operation.

## Slow Handlers

The component response deadline is short. Defer before slow work:

```go
router.Button("refresh", func(ctx *bot.InteractionContext) {
	if err := ctx.DeferUpdate(); err != nil {
		return
	}
	// Perform bounded work, then edit the original interaction response.
	if _, err := ctx.EditReply("Refreshed"); err != nil {
		log.Printf("edit button response: %v", err)
	}
})
```

`UpdateContent` is the simplest immediate acknowledgement. `Update` is useful
when the new content, embeds, or component state must be sent together.

## Common Mistakes

- Calling `Reply` from a button handler when the source message should change.
- Sleeping or doing REST work before `DeferUpdate`.
- Assigning both `CustomID` and `URL`; link buttons should be navigation only.
- Assuming disabled buttons enforce authorization.
- Forgetting to return/log an error from the acknowledgement.

## Expected Result

`/buttons` renders one application button and one link button. Clicking
`Complete` updates the original message. Clicking the link opens the Discord
developer documentation and does not invoke a router route.

## Related Pages

- [Action Rows](action-rows.md)
- [Interactions](interactions.md)
- [Collectors](../more-to-know/collectors.md)
