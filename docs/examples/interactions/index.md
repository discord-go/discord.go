# Interactions

These guides adapt the Discord.js interaction topics to the current `discord.go`
API. They use `bot.Router` for application commands and component routes,
`components` builders for payloads, and `bot.InteractionContext` for replies.

## Before You Start

- Use Go `1.26.4` or newer, as declared by the repository `go.mod`.
- Set `DISCORD_TOKEN` to a bot token. Do not put a token in source code.
- Install the application in a test guild with the `applications.commands` scope.
- Use `intents.Guilds` for the interaction-only examples. Buttons, menus, and
  modals arrive as interactions and do not require message-content intent.
- Run a copied example from the repository root with `go run ./examples/name`.

## Topic Map

- [Interactions](interactions.md) explains interaction types, routing, and the
  one-initial-response rule.
- [Buttons](buttons.md) creates custom-ID and link buttons.
- [Action Rows](action-rows.md) composes legacy interactive components.
- [Select Menus](select-menus.md) handles string and user selections.
- [Modals](modals.md) opens forms and reads submitted text inputs.
- [Display Components](display-components.md) builds Components V2 layouts.

## Acknowledgement Rule

Discord expects the initial interaction response quickly. In `discord.go`, use
exactly one of `Reply`, `ReplyEphemeral`, `ReplyComplex`, `Defer`,
`DeferUpdate`, `Update`, `ShowModalBuilder`, or another initial response method.
After a deferral, use `EditReply` or `Followup`; after a component update, do
not call a second initial `Reply`. A duplicate acknowledgement returns
`bot.ErrInteractionAlreadyResponded`.

## Tutorial: First Interaction

1. Read `DISCORD_TOKEN` from the environment.
2. Register a command on `bot.NewRouter()`.
3. Create the bot with `intents.Guilds` and `bot.WithRouter`.
4. Reply from the handler and start `b.Run()`.

## Complete Runnable Example

Copy this to `examples/interaction-ping/main.go`, set `DISCORD_TOKEN`, and run
`go run ./examples/interaction-ping`.

```go
package main

import (
	"log"
	"os"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}
	router := bot.NewRouter()
	router.Command("ping", "Check whether the bot is online", func(ctx *bot.InteractionContext) {
		if err := ctx.Reply("Pong"); err != nil {
			log.Printf("ping response: %v", err)
		}
	})
	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Shared Safety Rules

- Treat custom IDs, selected values, and modal values as untrusted input.
- Authorize a component at click time; rendering a button is not authorization.
- Keep custom IDs stable and short. Put sensitive or long-lived state in server
  storage rather than exposing it in the ID.
- Defer before slow database or REST work, then use a bounded context.
- Return or log every response error. A silently ignored failed acknowledgement
  makes the next recovery attempt ambiguous.
