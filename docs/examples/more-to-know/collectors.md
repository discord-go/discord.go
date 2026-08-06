# Collectors

## Overview

Collectors are bounded waits for one matching event. `Bot.AwaitInteraction`
supports buttons, selects, and modals; `Bot.AwaitMessage` supports a message
workflow. Both accept a `context.Context`, return the first match, and remove
their registration when the result arrives or the context ends.

## Tutorial: Wait For One Button

1. Send the prompt and capture the initiating user's ID.
2. Create a timeout context.
3. Filter by component kind, custom ID, and actor.
4. Acknowledge the returned interaction with `Update`, `Reply`, or a defer.
5. Handle cancellation without logging expected timeouts as failures.

A collector is a workflow convenience, not a security boundary. Recheck
authorization after a match and before changing application state.

## Complete Runnable Example

Copy to `examples/collectors/main.go`, set `DISCORD_TOKEN`, and run it. Invoke
`/choose`, click `Next` as the initiating user, or wait for the 30-second
timeout.

```go
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/snowflake"
)

func actorID(ctx *bot.InteractionContext) snowflake.ID {
	if ctx.User != nil {
		return ctx.User.ID
	}
	if ctx.Member != nil && ctx.Member.User != nil {
		return ctx.Member.User.ID
	}
	return 0
}

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}

	router := bot.NewRouter()
	router.Command("choose", "Wait for a private button choice", func(ctx *bot.InteractionContext) {
		owner := actorID(ctx)
		if owner == 0 {
			_ = ctx.ReplyEphemeral("Could not identify the command user.")
			return
		}
		button := components.NewButtonBuilder().
			SetCustomID("collector:next").
			SetLabel("Next").
			SetStyle(components.ButtonStylePrimary).
			Build()
		row := components.NewActionRowBuilder().AddComponents(button).Build()
		if err := ctx.ReplyComplex(&interactions.InteractionCallbackData{
			Content:    "Only the initiating user can click this within 30 seconds.",
			Components: []components.Component{row},
		}); err != nil {
			log.Printf("collector prompt: %v", err)
			return
		}

		waitCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		go func() {
			defer cancel()
			selected, err := ctx.Bot.AwaitInteraction(waitCtx, func(candidate *bot.InteractionContext) bool {
				return candidate.IsButton() && candidate.CustomID() == "collector:next" && actorID(candidate) == owner
			})
			if err != nil {
				if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
					log.Printf("collector: %v", err)
				}
				return
			}
			if err := selected.UpdateContent("Choice received; this collector is closed."); err != nil {
				log.Printf("collector update: %v", err)
			}
		}()
	})

	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Collector Design

Use a filter that is cheap and specific: check event kind, custom ID, actor,
guild, channel, and message where available. Use `context.WithCancel` when a
workflow completes through another path, and cancel all workflow contexts on
shutdown. A process-wide unbounded collector is a memory and authorization
risk.

`AwaitMessage` follows the same shape:

```go
message, err := b.AwaitMessage(waitCtx, func(candidate *bot.MessageContext) bool {
	return candidate.GuildID == guildID && candidate.Author != nil && candidate.Author.ID == owner
})
```

## Common Mistakes

- Passing `context.Background()` to a user-facing collector.
- Matching only a custom ID shared by several messages.
- Assuming the first matching event is authorized.
- Forgetting to acknowledge the collected component.
- Starting a goroutine without a timeout or shutdown context.

## Expected Result

`/choose` starts one scoped collector. Only the initiating user's matching
button completes it; cancellation or timeout removes the collector.

## Related Pages

- [Buttons](../interactions/buttons.md)
- [Interactions](../interactions/interactions.md)
- [Common Errors](common-errors.md)
