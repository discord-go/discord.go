# Gateway

## Overview

The high-level bot exposes typed handlers for common events and a generic event API for every named Gateway dispatch. This guide consumes `MESSAGE_DELETE` with `bot.On`, copies the raw payload through `EventContext`, and decodes it into the repository's typed `events.MessageDelete` model.

The same pattern works for a Discord event that has not yet acquired a typed helper. Use the typed helper when it exists; use the generic route when the event is newer, application-specific, or needs fields not covered by a convenience context.

## Prerequisites

- Go `1.26.4` or newer.
- `DISCORD_TOKEN` set to a bot token.
- A guild where the bot can receive message events.
- `Guilds` and `GuildMessages` enabled in the Developer Portal and selected in `bot.WithIntents`.
- Permission to view the test channel. Gateway intents and channel permissions are separate controls.

## Architecture

The Gateway dispatcher receives a dispatch name and JSON data. `Bot.On` registers an event subscription, normalizes the name to upper case, and passes an `EventContext` to the handler. `EventContext.Decode` calls JSON unmarshalling into a caller-owned value, while `Raw` returns a copy for logging or forwarding. The returned unsubscribe function removes the subscription and is safe to call more than once.

## Quick Start

Save the complete program below as a temporary `main.go` outside this repository's source tree, or use the equivalent code in your application:

```bash
export DISCORD_TOKEN='replace-with-a-bot-token'
go run ./path/to/your/gateway-example
```

For a repository-root program under `examples/gateway`, the command would be `go run ./examples/gateway`.

## Complete Runnable Example

This complete program includes imports and `main`. It runs until SIGINT or SIGTERM through `b.Run()`.

```go
package main

import (
	"log"
	"os"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/events"
	"github.com/discord-go/discord.go/intents"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}

	b := bot.New(token,
		bot.WithIntents(intents.Guilds|intents.GuildMessages),
	)

	unsubscribe := b.On("MESSAGE_DELETE", func(event *bot.EventContext) {
		var payload events.MessageDelete
		if err := event.Decode(&payload); err != nil {
			log.Printf("decode %s: %v", event.Name, err)
			return
		}
		log.Printf("message %s deleted from channel %s", payload.ID.String(), payload.ChannelID.String())
	})
	defer unsubscribe()

	b.OnError(func(err error) {
		log.Printf("bot error: %v", err)
	})

	log.Println("listening for MESSAGE_DELETE; press Ctrl+C to stop")
	if err := b.Run(); err != nil {
		log.Fatalf("bot stopped: %v", err)
	}
}
```

## Explanation

`MESSAGE_DELETE` also has the typed `b.OnMessageDelete` helper. The generic form is useful here because it demonstrates a stable fallback for any dispatch name. The event data is already copied by the bot before the handler runs, so decoding it does not depend on a mutable network buffer.

Gateway handlers can run concurrently with other handlers. Do not mutate shared application state without synchronization, and do not block the dispatch path on unbounded work.

## Basic Usage

- Register `b.On("EVENT_NAME", handler)` before `b.Run()`.
- Decode into an `events` type or a small application-owned struct.
- Use `b.Once("EVENT_NAME", handler)` for one-time initialization.
- Keep the unsubscribe function when a subscription has a bounded lifetime.
- Use `event.Raw()` when forwarding the exact JSON is more appropriate than decoding it.

## Intermediate Usage

- Use `b.OnRawEvent` when one callback should observe every dispatch.
- Use `b.OnVoiceServerUpdate` and `b.OnVoiceStateUpdate` for the voice control-plane events while still decoding their payloads yourself.
- Use typed handlers such as `OnMessageDelete` when the repository already models the event.
- Send slow work to a bounded worker queue and preserve the event name and relevant IDs in logs.

## Advanced Usage

- Use `Start`, `Done`, and `Stop` when the service needs an explicit shutdown deadline.
- Use `WithMaxHandlerConcurrency` to protect memory and downstream APIs under event bursts.
- Track `Bot.Stats()` and `OnError` for event volume, handler panics, and disconnect health.
- Validate untrusted payload fields before writing to storage or invoking administrative actions.
- Keep subscriptions idempotent during reconnects. READY can occur again after a fresh identify, so do not register duplicate handlers from a READY callback.

## Common Patterns

- Normalize and document the event name in one place.
- Decode only the fields needed by the handler.
- Return early on decode errors; never use a partially populated payload as an authorization decision.
- Use `context.WithTimeout` for downstream HTTP, database, or queue operations started by a Gateway handler.
- Redact raw payloads when they may contain user-generated content or sensitive application data.

## Best Practices

- Request only the intents needed for the events consumed by the application.
- Treat a Gateway event as at-least-once application input around reconnect and resume behavior.
- Make handlers safe to run concurrently and safe to retry where practical.
- Log stable IDs and event names, not full raw payloads by default.
- Unsubscribe temporary handlers and stop worker queues during shutdown.

## Common Mistakes with wrong/correct examples

### Wrong

```go
b.On("message_delete", func(event *bot.EventContext) {
	var payload events.MessageDelete
	_ = json.Unmarshal(event.Data, &payload)
})
```

### Correct

```go
unsubscribe := b.On("MESSAGE_DELETE", func(event *bot.EventContext) {
	var payload events.MessageDelete
	if err := event.Decode(&payload); err != nil {
		return
	}
	log.Printf("deleted %s", payload.ID.String())
})
defer unsubscribe()
```

The first fragment also omits imports and error handling. The second is an excerpt of the complete program above, not a standalone file.

### Wrong

```go
b.On("MESSAGE_DELETE", func(event *bot.EventContext) {
	for {
		// Unbounded work blocks this handler forever.
	}
})
```

### Correct

```go
b.On("MESSAGE_DELETE", func(event *bot.EventContext) {
	go enqueueDelete(event.Name, event.Raw())
})
```

In production, `enqueueDelete` must itself use bounded queues and shutdown cancellation.

## Expected Result

The bot connects and logs each decodable `MESSAGE_DELETE` payload as a message ID and channel ID. Malformed payloads are logged and ignored. On Ctrl+C, `b.Run()` closes the Gateway and waits for active handlers before returning.

## Related Pages

- [Examples Overview](README.md)
- [Basic Client](../setup/basic-client.md)
- [Voice](../voice/index.md)
- [Complete event model: `events/message_events.go`](../../events/message_events.go)
