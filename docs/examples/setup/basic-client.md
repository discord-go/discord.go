# Basic Client

## Overview

This guide starts a complete `discord.go` bot, listens for READY, handles a direct message event, and registers both slash and prefix commands. It is the recommended first example because it shows the normal high-level path without requiring application code to decode Gateway payloads.

## Prerequisites

- Go `1.26.4` or newer.
- `DISCORD_TOKEN` containing a bot token, supplied by the environment.
- A test guild where the bot can view channels and send messages.
- `Guilds` and `GuildMessages` enabled for the application. `MessageContent` is also required by the prefix and direct content checks in this exact source example.
- The Message Content privileged intent enabled in the Developer Portal.

## Architecture

`bot.New` creates the Gateway and REST-backed bot. `bot.WithIntents` controls the Gateway subscription, `bot.WithRouter` attaches slash and prefix dispatch, and `OnMessageCreate` registers a direct handler. The router registers slash commands automatically when READY is received. Replies use context helpers that create bounded REST requests.

## Quick Start

From the repository root:

```bash
export DISCORD_TOKEN='replace-with-a-bot-token'
go run ./docs/examples/code/ping
```

Try `/ping`, `!ping`, or a message containing exactly `ping` in a channel visible to the bot.

## Complete Runnable Example

[`examples/ping/main.go`](../code/ping/main.go) is the complete runnable program. It has imports and `main`, reads `DISCORD_TOKEN`, configures `intents.Guilds|intents.GuildMessages|intents.MessageContent`, creates a router, registers handlers, and ends with `b.Run()`.

Do not copy only a handler from that file and call it a complete program. The token check, imports, bot options, and lifecycle call are required.

## Explanation

The source uses `router.Command("ping", ...)` for a slash command and `router.Prefix("ping", ...)` for `!ping`. A `MessageContext` supplies `Content` and `Reply`; an `InteractionContext` supplies interaction-specific reply methods. `OnMessageCreate` is useful for event-driven behavior that is not a command. Bot-authored messages are filtered before this handler is called.

The example builds an embed with `messages.NewEmbedBuilder` and returns errors from all reply calls to the log. This is preferable to silently ignoring failed REST requests.

## Basic Usage

- Replace the example handler with a small command that responds with `ctx.Reply`.
- Add a READY handler with `b.OnReady` for startup logs and metrics.
- Use `bot.WithPrefix` to change `!` to an application-specific prefix.
- Remove `MessageContent` and the content handler if the bot only needs slash commands.

## Intermediate Usage

- Add global or per-command middleware with `router.Use` and `Command.Use`.
- Use `bot.WithGuildCommandSync(guildID)` while developing commands in one guild.
- Use `bot.RunContext` in a service that already owns cancellation instead of relying on process signals.
- Add `bot.WithMaxHandlerConcurrency` when unbounded parallel handlers could exhaust downstream resources.

## Advanced Usage

- Use `bot.Start(ctx)`, `bot.WaitReady(ctx)`, `bot.Done()`, and `bot.Stop(ctx)` for explicit lifecycle control.
- Use `bot.WithCache` or `bot.WithStore` only when the application has a defined cache or persistence policy.
- Use `bot.WithShards(0)` for automatic shard sizing when the application has grown beyond one Gateway session.
- Use `bot.OnError` and `Bot.Stats()` to expose handler panics, event counts, command syncs, and lifecycle health.

## Common Patterns

- Keep the token in `DISCORD_TOKEN`, `TOKEN`, or a secret manager and pass it to `bot.New` only at startup.
- Select the smallest intent set. `MessageContent` is privileged and unnecessary for slash-only bots.
- Return or log errors from `Reply`, `ReplyEmbed`, and `ReplyComplex`.
- Use `ctx.Bot.Rest` for APIs not exposed by a context helper, with a `context.WithTimeout` deadline.
- Let `b.Run()` handle SIGINT/SIGTERM for a simple executable; use `RunContext` for an orchestrated service.

## Best Practices

- Keep test and production applications separate so command synchronization cannot overwrite the wrong command set.
- Do not log the token or embed untrusted input as raw administrative instructions.
- Configure Discord permissions separately from Gateway intents; intents do not grant channel permissions.
- Use structured logs for READY, disconnect, reconnect, and REST failures.
- Set a shutdown deadline and allow active handlers to finish before the process exits.

## Common Mistakes with wrong/correct examples

### Wrong

```go
token := "YOUR_BOT_TOKEN"
b := bot.New(token)
```

### Correct

```go
token := os.Getenv("DISCORD_TOKEN")
if token == "" {
    log.Fatal("DISCORD_TOKEN is required")
}
b := bot.New(token, bot.WithIntents(intents.Guilds))
```

### Wrong

```go
router.Command("Ping", "", handler)
```

### Correct

```go
router.Command("ping", "Check whether the bot is online", handler)
```

Command names must be lower-case and descriptions for chat-input commands must be non-empty. The corrected fragment still belongs inside a complete program.

## Expected Result

The process logs a READY message, the bot appears online, `/ping` returns the Pong embed, `!ping` returns a prefix response, and a direct `ping` message receives a response. Ctrl+C causes `b.Run()` to stop the Gateway and wait for active handlers.

## Related Pages

- [Examples Overview](../README.md)
- [Slash Commands](../commands/slash-commands.md)
- [Gateway](../more-to-know/gateway.md)
- [Full Template](../advanced/full-template.md)
- [Complete source: `examples/ping/main.go`](../code/ping/main.go)
