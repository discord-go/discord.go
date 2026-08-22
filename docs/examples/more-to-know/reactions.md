# Reactions

## Overview

Reactions are message events plus REST operations. Enable
`intents.GuildMessageReactions` for guild reaction events, use
`Bot.OnMessageReactionAdd` for a typed handler, and call `b.Rest` methods to
add, list, or remove reactions. The event payload includes IDs and an emoji;
it does not provide every related resource inline.

## Tutorial: Seed And Observe A Reaction

1. Enable guild reaction intent in the Portal and `bot.WithIntents`.
2. Send a message and use `CreateReaction` when the bot should react itself.
3. Handle `ReactionContext` events without assuming the actor is cached.
4. Compare `Emoji.Name` or `Emoji.ID` safely; custom emoji names are optional.
5. Use a bounded context for REST calls and avoid reaction loops.

## Complete Runnable Example

Copy to `examples/reactions/main.go`, set `DISCORD_TOKEN`, and run it. Invoke
`/react`, then add a thumbs-up reaction to the created message.

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
	router.Command("react", "Create a message for a reaction workflow", func(ctx *bot.InteractionContext) {
		if err := ctx.Reply("React to this message with a thumbs-up."); err != nil {
			log.Printf("reaction prompt: %v", err)
			return
		}
		message, err := ctx.GetReply()
		if err != nil {
			log.Printf("fetch reaction prompt: %v", err)
			return
		}
		if err := ctx.Bot.Rest.CreateReaction(ctx.Context(), message.ChannelID, message.ID, "👍"); err != nil {
			log.Printf("seed reaction: %v", err)
		}
	})

	b := bot.New(token,
		bot.WithIntents(intents.Guilds|intents.GuildMessageReactions),
		bot.WithRouter(router),
	)
	b.OnMessageReactionAdd(func(ctx *bot.ReactionContext) {
		if ctx.Emoji.Name == nil || *ctx.Emoji.Name != "👍" {
			return
		}
		log.Printf("thumbs-up message=%s user=%s", ctx.MessageID.String(), ctx.UserID.String())
		if err := ctx.Bot.Rest.CreateReaction(ctx.Context(), ctx.ChannelID(), ctx.MessageID, "✅"); err != nil {
			log.Printf("reaction acknowledgement: %v", err)
		}
	})
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## REST Operations

- `CreateReaction(ctx, channelID, messageID, emoji)` adds the bot's reaction.
- `DeleteOwnReaction` removes the bot's reaction.
- `GetReactions` or `GetReactionsPage` lists users for one emoji.
- `DeleteUserReaction` removes a user's reaction and requires the appropriate
  moderation authority.
- `DeleteAllReactions` and `DeleteAllReactionsForEmoji` remove reactions with
  stronger permissions and should be guarded.

Unicode emoji can be passed as text. Custom emoji require the format Discord's
endpoint accepts, commonly `name:id`; validate or construct it from trusted
emoji data rather than accepting arbitrary path text.

## Common Mistakes

- Omitting `GuildMessageReactions` and expecting add events.
- Dereferencing `ctx.Emoji.Name` when it is nil for a custom emoji.
- Reacting to the bot's own acknowledgement and creating a loop.
- Treating the event's `UserID` as a full cached `users.User`.
- Performing bulk reaction deletion without a permission guard.

## Expected Result

`/react` sends a prompt and adds a bot thumbs-up. A user thumbs-up is logged and
gets a check-mark reaction from the bot.

## Related Pages

- [Gateway Intents](gateway-intents.md)
- [Partials And Cache](partials-cache.md)
- [Permissions](permissions.md)
