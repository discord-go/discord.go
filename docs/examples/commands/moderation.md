# Moderation

## Overview

The moderation example combines slash commands, guild-only middleware, invoker and bot permission checks, deferred interaction responses, audit-log reasons, and direct REST operations. It implements `/ban`, `/kick`, `/timeout`, and `/unban` against a guild member target.

This is an operational example, not a recommendation to grant broad permissions by default. Run it with a test bot and test guild first, and add an audit trail appropriate to your organization.

## Prerequisites

- Go `1.26.4` or newer.
- `DISCORD_TOKEN` set to a bot token and never committed.
- `Guilds`, `GuildMessages`, and `GuildMembers` enabled in the Portal for the source example.
- The bot installed with View Channel, Send Messages, and the exact moderation permissions used: Ban Members, Kick Members, and Moderate Members as applicable.
- A test guild where the bot's role is above the target members and Discord hierarchy rules permit the action.

## Architecture

The router applies `bot.GuildOnly()` globally, then each command adds the least invoker permission needed with `bot.RequirePermissions`. A command handler reads snowflake IDs and options from `InteractionContext`, calls `Defer` immediately, performs the action through `ctx.Bot.Rest`, and sends an embed follow-up. `rest.WithReason` attaches the reason to the Discord audit-log header.

## Quick Start

From the repository root:

```bash
export DISCORD_TOKEN='replace-with-a-bot-token'
go run ./docs/examples/code/moderation
```

Test `/timeout` on a disposable test account before trying ban or kick. Use a reason and confirm the bot role hierarchy first.

## Complete Runnable Example

[`examples/moderation/main.go`](../code/moderation/main.go) is the complete runnable program. It includes imports, all command option definitions, middleware, event handlers, REST calls, embeds, timeout clamping, token validation, and `b.Run()`.

The source is the authoritative runnable example. The production adjustments below explain how to harden its reference behavior without claiming that a short excerpt is a standalone program.

## Explanation

Permission middleware checks the invoking member, but Discord still enforces the bot's own permissions and role hierarchy. Both checks matter. `ctx.GetUserID("user")` returns a `snowflake.ID`, and `ctx.GuildID()` returns a `snowflake.ID` (zero for DMs); `GuildOnly` prevents commands from running outside a guild.

The source defers before each REST request. Its `rest.WithReason(context.Background(), reason)` call demonstrates audit reasons; a production service should derive that context from a bounded operation context instead of using an unbounded background context.

## Basic Usage

- Register a guild-only command with `router.Command(...).Use(bot.GuildOnly())`.
- Add `bot.RequirePermissions` for the invoker and `bot.RequireBotPermissions` for the bot.
- Read target IDs with `GetUserID` and reasons with `GetStringOption`.
- Call `ctx.Defer()` before the REST operation.
- Send success or failure with `Followup` or `FollowupEmbed`.

## Intermediate Usage

- Clamp user-provided timeout durations to Discord's supported range.
- Use `rest.WithReason` for every action that should appear in the audit log.
  Attach the reason to the context before the REST call:
  `apiCtx = rest.WithReason(apiCtx, reason)`. Without it, Discord's audit log
  shows "no reason provided" even when the bot logs the reason to its own
  channel.
- Use `AddGuildMemberRole` and `RemoveGuildMemberRole` for single-role
  changes instead of fetch-modify-resend via `ModifyGuildMember`.
- Use ephemeral replies for validation and permission errors that should not clutter a channel.
- Handle REST errors without leaking internal request details to users.
- Count successful and rejected moderation actions for operational review.

## Advanced Usage

- Derive REST contexts from `ctx.Context()` with a deadline, for example ten seconds, and pass that context to `ctx.Bot.Rest`.
- Re-fetch or validate the target member when the action is sensitive or the interaction may be delayed.
- Add an application-owned authorization policy in addition to Discord permissions, such as moderator roles, two-person approval, or protected-user rules.
- Make retries safe and avoid repeating a ban, kick, or timeout after an ambiguous network failure without checking current state.
- Store immutable action records with actor, target, guild, reason, request ID, result, and timestamp.

## Common Patterns

- Defer first, then do bounded work.
- Use a stable default reason only when the command permits an omitted reason.
- Return a user-safe error and log the detailed REST error with IDs.
- Use `CommunicationDisabledUntil` for timeouts and compute it with `time.Now().Add`.
- Refuse actions against the bot owner, guild owner, protected roles, or the bot itself according to application policy.

## Best Practices

- Grant only the moderation permissions the deployed commands need.
- Keep the bot role below roles it must never moderate and above roles it may moderate.
- Treat target IDs and reasons as untrusted input; validate membership, length, and content policy.
- Meet the initial interaction deadline with `Defer`, then use explicit REST deadlines.
- Prefer an explicit shutdown context so pending moderation calls and handlers finish or cancel predictably.
- Test failure paths: missing permissions, hierarchy conflicts, unknown users, rate limits, and timeouts.

## Common Mistakes with wrong/correct examples

### Wrong

```go
router.Command("ban", "Ban a member", handleBan)
```

### Correct

```go
router.Command("ban", "Ban a member", handleBan,
	interactions.ApplicationCommandOption{
		Type:        interactions.ApplicationCommandOptionTypeUser,
		Name:        "user",
		Description: "The member to ban",
		Required:    true,
	},
).Use(bot.GuildOnly()).Use(bot.RequirePermissions(permissions.BanMembers))
```

### Wrong

```go
apiCtx := rest.WithReason(context.Background(), reason)
_ = ctx.Bot.Rest.CreateGuildBan(apiCtx, *ctx.GuildID, userID, rest.CreateBanParams{})
```

### Correct

```go
apiCtx, cancel := context.WithTimeout(ctx.Context(), 10*time.Second)
defer cancel()
apiCtx = rest.WithReason(apiCtx, reason)
err := ctx.Bot.Rest.CreateGuildBan(apiCtx, ctx.GuildID(), userID, rest.CreateBanParams{
	DeleteMessageSeconds: 86400,
})
```

### Wrong

```go
_ = ctx.Reply("Starting moderation action")
slowRESTCall()
```

### Correct

```go
if err := ctx.Defer(); err != nil {
	return
}
slowRESTCall()
_, _ = ctx.Followup("Moderation action finished")
```

The corrected fragments use current APIs but remain excerpts. Use the linked source for a complete file.

## Expected Result

The bot registers four guild-only moderation commands. Authorized users receive deferred responses followed by success or failure embeds. Unauthorized users are rejected by middleware, REST audit reasons are sent with actions, and READY/GUILD_CREATE logs show service startup and guild discovery.

## Related Pages

- [Examples Overview](README.md)
- [Slash Commands](slash-commands.md)
- [Full Template](../advanced/full-template.md)
- [Complete source: `examples/moderation/main.go`](../code/moderation/main.go)
- [REST client source](../../rest/client.go)
