# Gateway Intents

## Overview

Gateway intents tell Discord which event families a bot wants. `discord.go`
represents them as `intents.Intent` bit flags passed to `bot.WithIntents`.
Request only what the bot consumes. Some intents are privileged and must also
be enabled for the application in the Developer Portal.

## Tutorial: Select The Minimum Set

1. List the event handlers and router features the bot actually uses.
2. Map each feature to an intent.
3. Enable privileged intents in the Portal when required.
4. Pass the bitwise OR of the selected constants to `bot.WithIntents`.
5. Test in a guild and inspect missing-event behavior before deploying.

Interaction-only bots usually need `intents.Guilds`. Prefix commands and
message-create handlers need `GuildMessages`; reading message content needs the
privileged `MessageContent` intent. Reactions use `GuildMessageReactions` or
`DirectMessageReactions` depending on the event source.

## Complete Runnable Example

Copy to `examples/gateway-intents/main.go`, set `DISCORD_TOKEN`, and run it.
Send `!hello` in a guild channel where the bot can read and send messages.

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
	router.Prefix("hello", func(ctx *bot.MessageContext, args []string) {
		if _, err := ctx.Reply("Hello from a message-create handler."); err != nil {
			log.Printf("prefix response: %v", err)
		}
	})

	b := bot.New(token,
		bot.WithPrefix("!"),
		bot.WithIntents(intents.Guilds|intents.GuildMessages|intents.MessageContent),
		bot.WithRouter(router),
	)
	b.OnMessageCreate(func(ctx *bot.MessageContext) {
		log.Printf("message channel=%s content_length=%d", ctx.ChannelID.String(), len(ctx.Content))
	})
	b.OnReady(func(ctx *bot.ReadyContext) {
		log.Printf("ready as %s", ctx.User.Username)
	})
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Intent Reference

- `Guilds`: guild create/update/delete and guild-scoped interaction context.
- `GuildMembers`: member events and member cache hydration.
- `GuildPresences`: presence updates.
- `GuildMessages`: guild message create, update, and delete events.
- `GuildMessageReactions`: guild reaction events.
- `GuildMessageTyping`: guild typing events.
- `DirectMessages`, `DirectMessageReactions`, and `DirectMessageTyping`: DM
  equivalents.
- `MessageContent`: message body and prefix-command content; privileged.
- `GuildVoiceStates`: voice-state events.

The constants live in `discord.go/intents`; combine them with `|`. Omitting an
intent can look like a handler bug because the Gateway simply does not send the
event family.

## Common Mistakes

- Enabling every intent by default.
- Selecting `MessageContent` without enabling it in the Portal.
- Expecting a slash-command bot to receive all message events.
- Using `GuildMessageReactions` for DMs.
- Assuming a cache can hydrate data for an event whose intent was never enabled.

## Expected Result

The example receives guild message content, logs message-create events, and
responds to `!hello`. Removing `MessageContent` or the Portal setting prevents
the prefix workflow from working as written.

## Related Pages

- [Partials And Cache](partials-cache.md)
- [Reactions](reactions.md)
- [Interactions](../interactions/interactions.md)
