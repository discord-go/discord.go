# Handling Events

## Overview

Events are Gateway notifications that are not necessarily commands. discord.go exposes typed helpers such as `OnReady`, `OnGuildCreate`, and `OnMessageCreate`, plus generic subscriptions through `On` and `OnRawEvent`. This page adapts Discord.js's Handling Events topic using the typed high-level API first.

## Architecture

The Gateway dispatcher decodes Discord dispatches and invokes registered handlers. Typed contexts embed the corresponding `events` model and include REST convenience methods where appropriate. Generic `EventContext` contains the normalized event name and a copied JSON payload, which can be decoded into a repository model or an application-owned struct. Handlers may run concurrently, so shared state needs synchronization.

## Prerequisites

- A bot token and installed application.
- `Guilds` enabled for guild events.
- `GuildMessages` and `MessageContent` enabled for the message example.
- A test guild and channel the bot can view.

## Quick Start

This complete program logs READY, guild joins, and exact `ping` messages. It also shows a generic event observer without decoding untrusted payloads into application state:

```bash
export DISCORD_TOKEN='replace-with-a-bot-token'
```

```go
package main

import (
	"context"
	"encoding/json"
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

	b := bot.New(token, bot.WithIntents(intents.Guilds|intents.GuildMessages|intents.MessageContent))
	b.OnReady(func(ctx *bot.ReadyContext) {
		log.Printf("ready as %s", ctx.User.Username)
	})
	b.OnGuildCreate(func(ctx *bot.GuildContext) {
		log.Printf("guild available: %s (%s)", ctx.Name, ctx.ID)
	})
	b.OnMessageCreate(func(ctx *bot.MessageContext) {
		if ctx.Content != "ping" {
			return
		}
		if _, err := ctx.Reply("Pong from an event handler."); err != nil {
			log.Printf("message reply: %v", err)
		}
	})
	b.OnRawEvent(func(_ context.Context, name string, data json.RawMessage) {
		if name == "READY" {
			log.Printf("observed raw %s payload (%d bytes)", name, len(data))
		}
	})
	b.OnError(func(err error) {
		log.Printf("runtime error: %v", err)
	})
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Creating/Using

Register handlers before `b.Run`. Prefer the typed helper whose payload matches the event. Use `b.On("EVENT_NAME", handler)` or `b.OnEvent` when an event has no typed convenience API. Call `event.Decode(&value)` to parse generic data and retain the returned unsubscribe function for temporary subscriptions. `b.Once` removes a generic subscription after its first invocation.

## Common Patterns

- Use `OnReady` for logs and metrics, not command registration side effects.
- Use `OnMessageCreate` for message behavior and let the bot filter bot-authored messages.
- Use `OnRawEvent` for metrics or forwarding every dispatch.
- Decode only the fields needed by a generic handler.
- Put event work onto a bounded queue when it involves slow external systems.

## Best Practices

- Request only the intents required by registered handlers.
- Enable privileged intents in the Portal and in `bot.WithIntents`.
- Treat events as concurrent input and protect shared maps with a mutex or another synchronization strategy.
- Use context deadlines for REST, database, and HTTP work launched from handlers.
- Log event names and stable IDs instead of full user-generated payloads.

## Common Mistakes

### Incorrect

```go
b.On("message_create", handler)
```

### Correct

```go
b.OnMessageCreate(handler)
```

Typed helpers avoid event-name spelling errors. Generic names are normalized to upper-case, but using the documented helper is clearer.

### Incorrect

```go
b.OnGuildCreate(func(ctx *bot.GuildContext) {
	sharedCount++
})
```

### Correct

```go
var mu sync.Mutex
var sharedCount int
b.OnGuildCreate(func(ctx *bot.GuildContext) {
	mu.Lock()
	sharedCount++
	mu.Unlock()
})
```

Handlers can overlap; synchronize application-owned state.

## API Walkthrough

- `b.OnReady` handles `READY` with `*bot.ReadyContext`.
- `b.OnGuildCreate` handles guild availability with `*bot.GuildContext`.
- `b.OnMessageCreate` handles message creation with `*bot.MessageContext`.
- `b.On` and `b.OnEvent` register generic named dispatch handlers.
- `EventContext.Decode` unmarshals a copied payload.
- `EventContext.Raw` returns a copied raw payload.
- `b.OnRawEvent` observes every dispatch.
- The function returned by `b.On` unsubscribes safely.
- `b.OnError` reports Gateway and handler failures.

## Examples

- [Gateway](../gateway.md) decodes `MESSAGE_DELETE` through a generic event route.
- [Basic Client](../basic-client.md) combines commands and message events.
- [Voice](../voice.md) demonstrates voice lifecycle events.
- [Event models](../../low-level/events/README.md) documents low-level payload types.

## Related Pages

- [Main File](main-file.md)
- [Handling Commands](handling-commands.md)
- [Gateway](../gateway.md)
- [Intents](../../low-level/intents/README.md)
