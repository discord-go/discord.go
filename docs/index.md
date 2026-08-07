<p align="center">
  <img src="/logo.svg" alt="discord.go" width="160">
</p>

<h1 align="center">discord.go Documentation</h1>

## Overview

This is the entry point for learning and operating `discord.go`. The library is
split into a protocol-oriented foundation and an application-oriented bot
facade. The foundation gives you typed Discord API objects and exact control
over Gateway, REST, voice, cache, authentication, rate limits, serialization,
and uploads. The facade turns those primitives into commands, middleware,
interaction contexts, collectors, lifecycle hooks, and developer-friendly
helpers.

Use the high-level guides for a normal bot. Use the low-level guides when you
are writing an adapter, building infrastructure, testing Discord payloads,
implementing a resource not yet wrapped by `bot`, or controlling the Gateway or
voice protocol directly. The examples show how both layers fit together in a
complete application.

## Architecture

Discord sends Gateway events over a long-lived WebSocket and accepts resource
operations over REST. `gateway` and `rest` are independent transport layers.
Resource packages such as `guilds`, `channels`, `users`, and `messages` contain
typed request and response models. `components`, `interactions`, and `voice`
model specialized Discord protocols.

`bot.Bot` composes these pieces. Its lifecycle owns a Gateway client, its
dispatcher creates typed contexts, and its router selects command or component
handlers. Context methods use the bot's REST client and the event's context.
Optional caches and storage remain explicit so applications can choose their
memory and persistence strategy.

## Quick Start

### Prerequisites

- Go `1.26.4` or newer.
- A Discord application and bot token.
- `Guilds` enabled for slash commands; add `GuildMessages` and
  `MessageContent` for prefix commands.

### Complete Example

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
    router.Command("ping", "Check bot status", func(ctx *bot.InteractionContext) {
        if err := ctx.Reply("Pong!"); err != nil {
            log.Printf("reply: %v", err)
        }
    })

    client := bot.New(token,
        bot.WithIntents(intents.Guilds),
        bot.WithRouter(router),
    )
    client.OnReady(func(ctx *bot.ReadyContext) {
        log.Printf("ready as %s", ctx.User.Username)
    })
    if err := client.Run(); err != nil {
        log.Fatal(err)
    }
}
```

Run it with:

```bash
export DISCORD_TOKEN='replace-with-a-bot-token'
go run ./docs/examples/code/ping
```

## Creating A Project

Start with `bot.New` and add options for intents, a router, cache, storage,
logging, command synchronization, presence, compression, or shards. Keep the
token in the environment. Use guild command synchronization during development
because global command changes can take a while to propagate.

## Using The Documentation

- Start with [`high-level/`](high-level/README.md) when writing application code.
- Read [`high-level/commands.md`](high-level/commands.md) before registering commands.
- Read [`high-level/interactions.md`](high-level/interactions.md) before handling buttons, menus, modals, or autocomplete.
- Read [`low-level/rest/`](low-level/rest/README.md) when calling resource endpoints directly.
- Read [`low-level/gateway/`](low-level/gateway/README.md) when handling raw events, reconnects, compression, or shards.
- Use [`examples/`](examples/README.md) to copy a complete program shape, then return to the relevant API guide for production hardening.

## Common Patterns

Always use contexts with deadlines for REST calls and collectors. Acknowledge
interactions within Discord's response window, then defer before slow work.
Scope component collectors by user, guild, message, and custom ID. Treat event
delivery as eventually consistent and make persistence writes idempotent.

## Best Practices

The high-level facade minimizes boilerplate and protects applications from
manual protocol mistakes. The low-level API is more flexible but requires you
to handle authentication, response deadlines, rate limits, and payload details
yourself. Keep REST and Gateway work behind application services, share cache
and storage intentionally, and test request bodies with local HTTP servers
before using production credentials.

## Common Mistakes

Do not commit a token, use `context.Background()` for unbounded network work,
send a second initial interaction response after `Defer`, or use global command
sync while iterating rapidly. Correct these by reading secrets from the
environment, deriving bounded contexts, using followups after deferral, and
using `WithGuildCommandSync` during development.

## API Walkthrough

The package map is maintained in [`low-level/README.md`](low-level/README.md)
and [`high-level/README.md`](high-level/README.md). Each package page explains
its constructors, exported types, options, behavior, edge cases, examples, and
related APIs. Resource endpoint groups are listed in
[`low-level/rest/endpoints.md`](low-level/rest/endpoints.md).

## Examples

Use [`examples/first-bot.md`](examples/first-bot.md) for a step-by-step tutorial,
[`examples/setup/basic-client.md`](examples/setup/basic-client.md) for the first bot,
[`examples/commands/slash-commands.md`](examples/commands/slash-commands.md) for command routing,
[`examples/interactions/components-v2.md`](examples/interactions/components-v2.md) for modern messages,
[`high-level/components-v2-guide`](high-level/components-v2-guide) for the V2 deep-dive,
[`high-level/voice-guide`](high-level/voice-guide) for voice,
[`high-level/security`](high-level/security) for security practices,
[`high-level/middleware-guide`](high-level/middleware-guide) for middleware,
[`high-level/error-handling`](high-level/error-handling) for error handling,
[`high-level/testing`](high-level/testing) for testing bots,
[`high-level/performance`](high-level/performance) for performance tuning,
[`high-level/anti-patterns`](high-level/anti-patterns) for common mistakes,
[`examples/best-practices.md`](examples/best-practices.md) for consolidated best practices,
[`examples/troubleshooting.md`](examples/troubleshooting.md) for troubleshooting,
[`examples/glossary.md`](examples/glossary.md) for term definitions,
and [`examples/advanced/full-template.md`](examples/advanced/full-template.md) for a larger
application layout.

## Related APIs

- [`high-level/client.md`](high-level/client.md)
- [`high-level/lifecycle.md`](high-level/lifecycle.md)
- [`high-level/security.md`](high-level/security.md)
- [`low-level/rest/README.md`](low-level/rest/README.md)
- [`low-level/gateway/README.md`](low-level/gateway/README.md)
- [`examples/README.md`](examples/README.md)
