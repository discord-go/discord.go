# Slash Commands

## Overview

Slash commands are registered on a `bot.Router` and dispatched as `*bot.InteractionContext`. The source-backed example covers string, integer, boolean, and user options; global and per-command middleware; guild-only and permission checks; public, ephemeral, embed, deferred, and follow-up responses; and automatic registration after READY.

## Prerequisites

- Go `1.26.4` or newer.
- `DISCORD_TOKEN` set to a bot token.
- An installed application in a test guild.
- `Guilds` and `GuildMessages` enabled for the source example. Slash-only applications may be able to use only `Guilds`, depending on the commands they implement.
- For `/kick`, the invoking member must have `KickMembers`, and the bot must have the corresponding channel and guild permission.

## Architecture

The router owns command definitions, options, middleware, validation, and command synchronization. Discord sends an interaction after the user invokes a command. `InteractionContext` parses options and owns the initial callback, deferral, follow-up, embed, and ephemeral response methods. Middleware runs before the command handler and can reject a request without executing business logic.

## Quick Start

The commands in this guide are source-backed:

```bash
export DISCORD_TOKEN='replace-with-a-bot-token'
go run ./docs/examples/code/slash_commands
```

Try `/hello`, `/userinfo`, `/kick`, or `/serverinfo` after the READY log appears. Global registration can take time to propagate.

## Complete Runnable Example

[`examples/slash_commands/main.go`](../code/slash_commands/main.go) is the complete runnable program. It includes `package main`, imports, token validation, all option definitions, middleware, handlers, and the bot lifecycle. The command above runs that exact file through the module.

## Explanation

`router.Command` takes a lower-case name, a non-empty description, a handler, and zero or more `interactions.ApplicationCommandOption` values. The handler uses `GetStringOption`, `GetIntOption`, `GetBoolOption`, and `GetUserID` rather than decoding raw interaction JSON.

The example uses `bot.GuildOnly()` and `bot.RequirePermissions(permissions.KickMembers)` as per-command middleware. `ctx.Defer()` acknowledges a slow interaction, after which `ctx.Followup` sends the result. A handler must not attempt a second initial response after `Reply` or `Defer`; the context returns `bot.ErrInteractionAlreadyResponded` for that mistake.

## Basic Usage

- Register a command with `router.Command`.
- Read optional strings as empty values and apply an application default.
- Use `Reply` for a public initial response and `ReplyEphemeral` for a private one.
- Attach the router with `bot.WithRouter(router)`.
- Validate a static registry with `router.Validate()` or use `CommandE`/`MustCommand` during startup.

## Intermediate Usage

- Add logging or authorization once with `router.Use`.
- Chain per-command middleware with `Command.Use`.
- Use `GuildOnly` for commands that dereference `ctx.GuildID` or `ctx.Member`.
- Use `Defer` before REST, database, or network work and report the result with `Followup` or `FollowupEmbed`.
- Use a guild-scoped sync option during development to avoid waiting for global propagation.

## Advanced Usage

- Use `UserCommand`, `MessageCommand`, and `TargetID` for context-menu actions.
- Use `Cooldown` for per-user interaction throttling.
- Use `RequireBotPermissions` as well as invoker permission middleware so missing bot permissions become a controlled response.
- Use `WithCommandSync` to set global, guild, or disabled synchronization and a finite sync timeout.
- Instrument `Bot.Stats()` and `OnError`; command handlers can panic or fail independently of the Gateway.

## Common Patterns

- Treat option values as untrusted and validate ranges, lengths, and target relationships.
- Use `ctx.MemberPermissions()` and middleware for authorization, not a user-supplied option.
- Keep the first response small and edit or follow up after expensive work.
- Make administrative commands guild-only and record an audit reason in REST requests.
- Use a distinct test bot and guild for global command development.

## Best Practices

- Keep command descriptions and option descriptions meaningful and within Discord limits.
- Prefer ephemeral error details when revealing permission or lookup information.
- Defer within Discord's initial interaction response window; a normal REST timeout does not extend that deadline.
- Never use `context.Background()` for unbounded production REST work; derive a timeout from the operation or service context.
- Handle failures from the initial response, deferral, follow-up, and command synchronization separately.

## Common Mistakes with wrong/correct examples

### Wrong

```go
router.Command("hello", "", func(ctx *bot.InteractionContext) {
	go slowLookup()
	_ = ctx.Reply("finished")
})
```

### Correct

```go
router.Command("hello", "Run a greeting", func(ctx *bot.InteractionContext) {
	if err := ctx.Defer(); err != nil {
		return
	}
	// Perform bounded work, then use a follow-up.
	_, _ = ctx.Followup("finished")
})
```

### Wrong

```go
router.Command("kick", "Kick a member", handler)
```

### Correct

```go
router.Command("kick", "Kick a member", handler).
	Use(bot.GuildOnly()).
	Use(bot.RequirePermissions(permissions.KickMembers)).
```

The corrections are excerpts; the complete source is the linked program.

## Expected Result

The router synchronizes the registered commands after READY. `/hello` reads its options and replies, `/userinfo` demonstrates a deferral and follow-up, `/kick` is rejected outside a guild or without permission, and `/serverinfo` returns an embed for guild invocations.

## Related Pages

- [Examples Overview](README.md)
- [Basic Client](basic-client.md)
- [Buttons](buttons.md)
- [Autocomplete](autocomplete.md)
- [Moderation](moderation.md)
- [Complete source: `examples/slash_commands/main.go`](../code/slash_commands/main.go)
