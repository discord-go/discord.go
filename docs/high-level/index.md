# High-Level Developer Guides

## Overview

The high-level API is the application-facing layer of discord.go. It turns
gateway events into typed contexts, routes commands and components, provides
message and interaction response helpers, and owns the bot lifecycle. A first
application normally needs `bot`, `intents`, and one or more resource packages;
it does not need to decode gateway payloads itself.

This guide assumes a Go module can import the repository module as `discord.go`.
For protocol details, continue with the [low-level guides](../low-level/README.md).

## Architecture

The main flow is:

1. `bot.New` creates a `Bot` around a gateway connection and a `rest.Client`.
2. Functional options configure intents, routing, presence, cache, logging, and
   sharding.
3. `Start`, `Run`, or `RunContext` opens the gateway and dispatches events in
   handler goroutines.
4. A `Router` maps slash commands, prefix commands, buttons, selects, modals,
   and autocomplete requests to handlers.
5. Typed contexts expose both the event data and convenience methods such as
   `Reply`, `Defer`, `Fetch`, and `Update`.

The `Bot.Rest` field remains the escape hatch for REST endpoints that do not
have a context helper. See the [REST request guide](../low-level/rest/requests.md)
and [endpoint reference](../low-level/rest/endpoints.md) when using it directly.

## Quick Start

Create `main.go` in the module root and run it with a bot token. The program is
complete and connects to Discord, registers `/ping`, and waits for a signal.

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
			log.Printf("reply: %v", err)
		}
	})

	b := bot.New(token,
		bot.WithIntents(intents.Guilds),
		bot.WithRouter(router),
	)
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

Install the application in a test guild with the `applications.commands` scope,
export `DISCORD_TOKEN`, and run `go run .`. Global command registration can take
up to an hour to propagate; use `bot.WithGuildCommandSync(guildID)` while
developing.

## Creating/Configuration

Start with `bot.New(token, opts...)`. Useful options include
`WithIntents`, `WithPrefix`, `WithBotName`, `WithMentionTriggers`, `WithRouter`,
`WithPresence`, `WithCache`, `WithStore`, `WithLogger`, `WithErrorHandler`,
`WithGatewayCompression`, `WithShards`, `WithMaxHandlerConcurrency`, and
`WithCommandSync`. The `bot.Config` helpers are covered in
[`configuration.md`](configuration.md).

Keep the token outside source control. Request only the gateway intents the
application actually uses, and enable privileged intents in the Discord
Developer Portal before requesting them in code.

## Using

### Basic: handle a command

Register a command on a `Router`, attach it with `WithRouter`, and reply through
the interaction context. Errors from response methods should be logged or sent
to the bot error handler.

### Intermediate: add a component flow

Send a button or select with `InteractionCallbackData`, then register a route
with `Router.Button`, `Router.Select`, or their prefix variants. Use
`InteractionContext.Update` when the triggering message should change.

### Advanced: combine middleware, cache, jobs, and REST

Use command middleware for authorization and cooldowns, `WithCache` for cheap
lookups, `Every` for lifecycle-owned periodic work, and `Bot.Rest` for an
endpoint without a convenience method. Keep network work behind a context with
a meaningful deadline.

## Common Patterns

- Use guild command sync for local development and global sync for release.
- Reply immediately when possible; call `Defer` before work that may exceed the
  interaction response window.
- Scope component IDs with a stable prefix such as `ticket:close:`.
- Treat cache hits as hints and use `Fetch*` when fresh data is required.
- Register `OnError` and log `Bot.Stats()` so operational failures are visible.

## Best Practices

### Choose the high-level layer first

Why: it supplies lifecycle, routing, and typed contexts in one place.

Pros: less protocol code, consistent error handling, and faster development.

Cons: protocol-specific features may still require `Bot.Rest`, `gateway`, or
`voice` directly.

### Use explicit contexts and bounded work

Why: gateway handlers run concurrently and REST calls can outlive an event.

Pros: cancellation is predictable and shutdown is graceful.

Cons: every asynchronous operation needs deliberate context ownership.

### Validate before production sync

Why: Discord rejects invalid names, descriptions, and option definitions.

Pros: `Router.Validate` or `CommandE` fails locally before a REST request.

Cons: validation does not replace Discord-side permission and installation checks.

## Common Mistakes

Incorrect: starting a router without attaching it.

```go
b := bot.New(token)
```

Correct:

```go
b := bot.New(token, bot.WithRouter(router))
```

Incorrect: sending two initial interaction responses.

```go
_ = ctx.Reply("one")
_ = ctx.Reply("two")
```

Correct: send one initial response, then use a follow-up.

```go
if err := ctx.Reply("one"); err != nil {
	return
}
_, _ = ctx.Followup("two")
```

## API Walkthrough

- `bot.New(string, ...bot.Option) *bot.Bot` constructs a configured bot.
- `bot.Option` is `func(*bot.Bot)` and is applied during construction.
- `bot.NewRouter() *bot.Router` creates command and interaction registries.
- `(*bot.Bot).OnReady(bot.ReadyHandler)` observes readiness.
- `(*bot.Bot).OnMessageCreate(bot.MessageHandler)` observes user messages.
- `(*bot.Bot).OnInteraction(bot.InteractionHandler)` observes all interactions.
- `(*bot.Bot).Run() error`, `RunContext(context.Context) error`, and
  `Start(context.Context) error` begin execution with different ownership of
  signals and blocking.
- `(*bot.Bot).Stop(context.Context) error`, `Wait() error`, and
  `Done() <-chan struct{}` provide shutdown coordination.
- `(*bot.Bot).Rest *rest.Client` exposes the authenticated REST client.
- `(*bot.Bot).State() bot.BotState`, `IsReady() bool`, `WaitReady(context.Context) error`,
  `AppID() snowflake.ID`, and `Stats() bot.BotStats` expose runtime state.

The individual pages below document the feature-specific APIs and complete
examples. These are tutorials, not package summaries.

## Examples

- [Basic client](../examples/setup/basic-client.md)
- [Slash commands](../examples/commands/slash-commands.md)
- [Buttons](../examples/interactions/buttons.md)
- [Modals](../examples/interactions/modals.md)
- [Collectors](../examples/more-to-know/collectors.md)
- [Components V2](../examples/interactions/components-v2.md)
- [Full template](../examples/advanced/full-template.md)

## Related APIs

- [`client.md`](client.md) for construction and runtime state.
- [`commands.md`](commands.md) for command and route registration.
- [`interactions.md`](interactions.md) for responses and option access.
- [`components.md`](components.md) for legacy and V2 components.
- [`buttons.md`](buttons.md) for custom-ID button flows.
- [`modals.md`](modals.md) for modal forms.
- [`collectors.md`](collectors.md) for one-shot waits and jobs.
- [`permissions.md`](permissions.md) for authorization middleware.
- [`lifecycle.md`](lifecycle.md) for startup, shutdown, jobs, and reconnects.
- [`presence.md`](presence.md) for status and latency.
- [`caching.md`](caching.md) for cache-backed lookups.
- [`resources.md`](resources.md) for typed REST helpers.
- [`embeds.md`](embeds.md) for rich message content.
- [`voice.md`](voice.md) for gateway voice state.
- [`errors.md`](errors.md) for runtime error handling.
- [`configuration.md`](configuration.md) for JSON and environment setup.
- [`../low-level/gateway/README.md`](../low-level/gateway/README.md) for gateway control.
- [`../low-level/rest/README.md`](../low-level/rest/README.md) for direct REST usage.
