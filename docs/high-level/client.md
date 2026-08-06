# Bot Client

## Overview

`bot.Bot` is the top-level object for a high-level application. It owns the
gateway connection, typed event dispatch, command router, REST client, optional
cache and store, presence, jobs, and lifecycle state. Create one bot per
process and register handlers before starting it.

## Architecture

`bot.New` applies functional options to a `Bot`. The bot creates a default
`rest.Client` unless `WithRESTClient` supplies one. `Start` creates either one
gateway client or a shard manager, then dispatches each event to registered
handlers in separate goroutines. Context helpers retain a reference to the bot
and its REST client, so a `MessageContext` or `InteractionContext` can perform
common operations without rebuilding a request.

`Bot.Rest` is public for direct endpoint access. `Bot.Store` is application-owned
persistence and is not a Discord cache.

## Quick Start

This complete program creates a client, logs READY, and registers a command.

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
	router.Command("ping", "Check the client", func(ctx *bot.InteractionContext) {
		if err := ctx.Reply("The client is running"); err != nil {
			log.Printf("reply: %v", err)
		}
	})

	client := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	client.OnReady(func(ctx *bot.ReadyContext) {
		log.Printf("ready as %s", ctx.User.Username)
	})
	if err := client.Run(); err != nil {
		log.Fatal(err)
	}
}
```

Run it with `DISCORD_TOKEN=... go run .`. The bot token must belong to an
application installed with the `bot` and `applications.commands` scopes.

## Creating/Configuration

The constructor is `bot.New(token string, opts ...bot.Option) *bot.Bot`.
Common options are:

- `WithIntents(intents.Intent)` sets gateway subscriptions.
- `WithPrefix(string)`, `WithBotName(string)`, and `WithMentionTriggers(bool)`
  control text-command triggers.
- `WithRouter(*bot.Router)` attaches slash, prefix, and component routing.
- `WithRESTClient(*rest.Client)` replaces the default REST transport.
- `WithCache(cache.Cache)` enables cache hydration and typed helpers.
- `WithStore(storage.Store)` attaches application persistence.
- `WithLogger(*log.Logger)` and `WithErrorHandler(bot.ErrorHandler)` configure
  diagnostics.
- `WithGatewayCompression(bool)`, `WithShards(int)`, and
  `WithMaxHandlerConcurrency(int)` tune runtime behavior.
- `WithCommandSync(bot.CommandSyncConfig)` or
  `WithGuildCommandSync(snowflake.ID)` controls command registration.

For JSON and environment-driven construction, see
[`configuration.md`](configuration.md).

## Using

### Basic: inspect the client

`State`, `IsReady`, `WaitReady`, `User`, `AppID`, `ReadyAt`, and `Uptime` are
safe to call while the bot is running. `User` returns a copy of the READY user.

### Intermediate: use direct REST

For example, `client.Rest.GetCurrentUser(ctx)` performs a request not tied to an
event. Always pass a context with cancellation or a deadline for long work.

### Advanced: replace transports

Use `WithRESTClient` for a custom HTTP client or limiter, and
`WithConnectionFactory` for a gateway proxy or test connection. These options
are useful for infrastructure integration, not normal bot setup.

## Common Patterns

- Register all handlers before `Run`; registration methods are safe but setup is
  easier to reason about when it is single-threaded.
- Keep one shared router and pass it to `WithRouter`.
- Use `WithMaxHandlerConcurrency` when unbounded event-handler goroutines could
  overload an external service.
- Keep `Bot.Store` for application records and `WithCache` for disposable Discord
  resource snapshots.

## Best Practices

### Request least privilege

Why: intents determine which gateway events Discord sends and some are
privileged.

Pros: less data, simpler privacy review, and lower event volume.

Cons: a missing intent can make a handler appear broken until the portal and
code are both updated.

### Make lifecycle ownership explicit

Why: `Run` owns OS signals, while `RunContext` and `Start` let a service own
cancellation.

Pros: clean tests and predictable shutdown.

Cons: the caller must wait for `Stop` or `Wait` instead of abandoning the
process.

### Keep the public REST client centralized

Why: the bot's client already carries authentication, rate limiting, and base
URL configuration.

Pros: consistent requests and fewer leaked tokens.

Cons: code that needs a different auth mode must intentionally use another
`rest.Client`.

## Common Mistakes

Incorrect: constructing a bot with no token and expecting `Run` to discover it.

```go
client := bot.New("")
_ = client.Run()
```

Correct: load the token before construction.

```go
token := os.Getenv("DISCORD_TOKEN")
if token == "" {
	log.Fatal("DISCORD_TOKEN is required")
}
client := bot.New(token)
```

Incorrect: treating `Start` as a blocking call.

```go
_ = client.Start(context.Background())
// The process exits immediately.
```

Correct: wait for readiness and termination, or use `Run`.

```go
if err := client.Start(ctx); err != nil {
	return err
}
if err := client.WaitReady(ctx); err != nil {
	return err
}
return client.Wait()
```

## API Walkthrough

- `bot.Bot` contains `Rest *rest.Client` and `Store storage.Store` for direct
  access to configured clients.
- `bot.New` creates the object; `bot.WithIntents`, `WithPrefix`, `WithBotName`,
  `WithMentionTriggers`, `WithGatewayCompression`, `WithPresence`, `WithShards`,
  `WithRouter`, `WithRESTClient`, `WithCache`, `WithStore`, `WithLogger`,
  `WithErrorHandler`, `WithGatewayURL`, `WithConnectionFactory`,
  `WithMaxHandlerConcurrency`, `WithCommandSync`, `WithCommandSyncDisabled`,
  and `WithGuildCommandSync` return `bot.Option` values.
- `bot.Bot.OnReady`, `OnMessageCreate`, `OnMessageUpdate`, `OnMessageDelete`,
  `OnInteraction`, `OnInteractionCreate`, `OnMessageReactionAdd`, `OnGuildCreate`,
  `OnGuildUpdate`, `OnGuildDelete`, `OnChannelCreate`, `OnChannelUpdate`,
  `OnGuildAuditLogEntryCreate`, `OnRawEvent`, and `OnError` register handlers.
- `OnEvent` and `On` return an unsubscribe function for named dispatch events;
  `OnceEvent` and `Once` remove the handler after its first call.
- `State`, `Done`, `Wait`, `Start`, `Run`, `RunContext`, `Stop`, `WaitReady`,
  `IsReady`, `AppID`, `User`, `ReadyAt`, `Uptime`, `Stats`, `GatewayLatency`,
  and `APILatency` expose lifecycle and runtime state.
- `BotState` has `BotStateStopped`, `BotStateStarting`, `BotStateRunning`, and
  `BotStateStopping`.
- `CommandSyncConfig` contains `Mode`, `GuildID`, and `Timeout`; modes are
  `CommandSyncGlobal`, `CommandSyncGuild`, and `CommandSyncDisabled`.

## Examples

- [Basic client](../examples/setup/basic-client.md)
- [Full template](../examples/advanced/full-template.md)
- [Configuration guide](configuration.md)

## Related APIs

- [`lifecycle.md`](lifecycle.md) for service startup and shutdown.
- [`commands.md`](commands.md) for routing.
- [`errors.md`](errors.md) for diagnostics.
- [`../low-level/client/README.md`](../low-level/client/README.md) for lower-level client concepts.
- [`../low-level/rest/README.md`](../low-level/rest/README.md) for direct REST calls.
