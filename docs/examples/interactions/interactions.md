# Interactions

## Overview

An interaction is a Discord callback for a slash command, context-menu command,
button, select menu, autocomplete request, modal submission, or ping. The
high-level API decodes the interaction into `bot.InteractionContext`. A
`bot.Router` handles known routes, while `Bot.OnInteraction` is useful for
cross-cutting logging and metrics.

## Tutorial: Route And Acknowledge

1. Create a `bot.Router`.
2. Register a slash command with `router.Command`.
3. Register a component route with `router.Button` or `router.Select`.
4. Send one initial response from each handler.
5. Use `Defer` and then `Followup` for work that cannot finish immediately.

`InteractionContext` exposes `IsCommand`, `IsButton`, `IsSelectMenu`,
`IsModalSubmit`, `CommandName`, `CustomID`, `Values`, and typed option helpers.
The router normally performs the type checks, so handlers can stay focused on
application work.

## Complete Runnable Example

Copy this program to `examples/interactions/main.go`, then run
`go run ./examples/interactions`. Invoke `/interactions`, click the button, and
inspect the log line for the interaction type.

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
	router.Command("interactions", "Show an interaction lifecycle example", func(ctx *bot.InteractionContext) {
		button := components.NewButtonBuilder().
			SetCustomID("interaction:ack").
			SetLabel("Acknowledge").
			SetStyle(components.ButtonStylePrimary).
			Build()
		row := components.NewActionRowBuilder().AddComponents(button).Build()
		if err := ctx.ReplyComplex(&interactions.InteractionCallbackData{
			Content:    "This command has been acknowledged. Click the button for an update.",
			Components: []components.Component{row},
		}); err != nil {
			log.Printf("command response: %v", err)
		}
	})

	router.Button("interaction:ack", func(ctx *bot.InteractionContext) {
		if err := ctx.UpdateContent("The component interaction updated the original message."); err != nil {
			log.Printf("component response: %v", err)
		}
	})

	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	b.OnInteraction(func(ctx *bot.InteractionContext) {
		log.Printf("interaction type=%d command=%q custom_id=%q", ctx.Type, ctx.CommandName(), ctx.CustomID())
	})
	b.OnReady(func(ctx *bot.ReadyContext) {
		log.Printf("ready as %s", ctx.User.Username)
	})
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## What Happens

The slash command is an `InteractionTypeApplicationCommand`. The button click
is an `InteractionTypeMessageComponent`; its `CustomID` is the routing key.
`UpdateContent` sends an update callback, so the message containing the button
is edited instead of a second message being created.

For slow work, acknowledge first:

```go
if err := ctx.Defer(); err != nil {
	return
}
// Run bounded work, then send a follow-up.
_, err := ctx.Followup("The operation finished.")
```

## API Choices

- Use `router.Command` for slash and context-menu commands.
- Use `router.Button`, `ButtonPrefix`, `Select`, `SelectPrefix`, and `Modal` for
  application-owned component routes.
- Use `Bot.OnInteraction` when every interaction needs observation or a custom
  dispatch layer.
- Use `ctx.Options`, `GetStringOption`, `GetIntOption`, `GetUserID`, and related
  helpers for command data.
- Use `ctx.FocusedOption` and `ctx.Autocomplete` for autocomplete handlers.

## Common Mistakes

Wrong:

```go
router.Button("save", func(ctx *bot.InteractionContext) {
	// Slow work before acknowledgement can make the interaction expire.
	// time.Sleep(5 * time.Second)
	_ = ctx.UpdateContent("saved")
})
```

Correct:

```go
router.Button("save", func(ctx *bot.InteractionContext) {
	if err := ctx.DeferUpdate(); err != nil {
		return
	}
	// Perform bounded work and edit the original message through a REST helper.
	_, _ = ctx.EditReply("saved")
})
```

Do not use `Reply` after `Update`, `DeferUpdate`, or `ShowModalBuilder`. Those
methods already acknowledge the interaction.

## Production Checklist

- **Verify incoming interaction webhooks with `interactions.VerifyRequest`, not
  `VerifySignature`.** `VerifyRequest` enforces both the Ed25519 signature and
  timestamp freshness to prevent replay attacks. If you use
  `interactions.Server` as your HTTP handler, this is handled automatically.
- Validate guild, channel, actor, message, and resource ownership in each route.
- Use a timeout for external work and make the final operation idempotent.
- Use guild-scoped command synchronization while developing fast-changing
  commands; global synchronization can take up to an hour to propagate.
- Log IDs and operation names, not tokens or sensitive modal content.

## Expected Result

`/interactions` produces a message with a button. The button updates that
message, and the process logs both the command and component interaction types.

## Related Pages

- [Buttons](buttons.md)
- [Action Rows](action-rows.md)
- [Select Menus](select-menus.md)
- [Modals](modals.md)
- [Display Components](display-components.md)
