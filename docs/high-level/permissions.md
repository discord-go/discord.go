# Permissions And Middleware

## Overview

The high-level middleware package provides reusable authorization, validation,
guild, role, owner, and cooldown guards. It works with slash, context-menu, and
prefix commands. The `permissions.Permission` type is a 64-bit bitfield; combine
flags with `permissions.Build` or `permissions.NewBuilder`.

## Architecture

Slash-command middleware receives `*bot.InteractionContext`. Member permissions
come from the interaction's `Member.Permissions`, while bot permissions come
from the interaction's `AppPermissions` string. Prefix middleware receives a
`*bot.MessageContext`; `RequirePrefixPermissions` looks up the author through
the configured cache.

Middleware wraps a handler. Router middleware runs for all slash commands,
command middleware runs for one command, and prefix middleware applies to one
prefix command. Authorization middleware responds ephemerally when it rejects
a slash interaction.

## Quick Start

This complete program protects `/audit` with member and bot permissions.

```go
package main

import (
	"log"
	"os"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/permissions"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}

	router := bot.NewRouter()
	router.Command("audit", "Show an audit entry", func(ctx *bot.InteractionContext) {
		if err := ctx.Reply("Authorized"); err != nil {
			log.Printf("reply: %v", err)
		}
	}).
		Use(bot.GuildOnly()).
		RequirePermissions(permissions.ViewAuditLog).
		RequireBotPermissions(permissions.ViewAuditLog)

	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

Grant the bot the required permission in the test guild and remember that
Discord's application-command visibility and channel overwrites still apply.

## Creating/Configuration

Build a permission set with `permissions.Build(permissions.ViewAuditLog,
permissions.SendMessages)` or `permissions.NewBuilder(...).Add(...).Build()`.
Pass it to `RequirePermissions`, `RequireAnyPermissions`, or
`RequireBotPermissions`.

Use `RequireRole` and `RequireAnyRole` with `snowflake.ID` values, and
`RequireOwners` with an application-owned allowlist. `OwnerOnly` checks the
current guild owner. `GuildOnly` rejects DMs. `Validate` turns a returned error
into an ephemeral response. Prefix permission checks need `WithCache` and a
cache implementing `MemberCache`.

## Using

### Basic: all required permissions

`RequirePermissions(perms)` calls `HasAll`, so a member must contain every bit.
Use `RequireAnyPermissions` when any one of several bits is sufficient.

### Intermediate: layered policy

A moderation command often uses `GuildOnly`, member permission checks, bot
permission checks, `RequireRole`, and `Cooldown`. The first rejecting middleware
stops the chain.

### Advanced: prefix authorization and custom checks

Use `PrefixGuildOnly` and `RequirePrefixPermissions` for text commands. Use
`Validate(func(*InteractionContext) error)` for resource-specific checks such as
"the target is below the moderator"; return a user-safe message.

## Common Patterns

- Check both invoker and bot permissions for actions the bot must perform.
- Use `Administrator` sparingly; exact permissions communicate intent better.
- Put global logging or tracing middleware on `Router.Use` and authorization on
  the command that needs it.
- Keep owner IDs in configuration, not source code, when deployments vary.
- Use `Permission.Has` for any-bit checks and `HasAll` for all-bit checks.

## Best Practices

### Check the bot separately

Why: a user can be authorized while the bot lacks the channel permission.

Pros: failures are explained before a REST request and avoid predictable 403s.

Cons: `AppPermissions` is interaction-provided and cannot replace handling
permission changes that happen after invocation.

### Prefer least privilege

Why: broad permissions increase blast radius.

Pros: safer installations and clearer audits.

Cons: more middleware and REST error paths must be handled when guild policy
changes.

### Treat prefix cache checks as fallible

Why: a missing member cache entry is not proof that a user lacks permission.

Pros: rejecting on cache miss avoids accidental authorization.

Cons: a cache miss can deny a valid user; fetch the member explicitly when the
operation justifies the network request.

## Common Mistakes

Incorrect: using `Has` when all permissions are required.

```go
if perms.Has(permissions.BanMembers | permissions.KickMembers) {
	// This allows only one of the two bits.
}
```

Correct:

```go
if perms.HasAll(permissions.BanMembers | permissions.KickMembers) {
	// Both bits are present.
}
```

Incorrect: checking only the user's permission for a ban command.

```go
router.Command("ban", "Ban a member", handleBan).
	RequirePermissions(permissions.BanMembers)
```

Correct: check the bot too.

```go
router.Command("ban", "Ban a member", handleBan).
	RequirePermissions(permissions.BanMembers).
	RequireBotPermissions(permissions.BanMembers)
```

## API Walkthrough

- `permissions.Permission` is a `uint64` bitfield. Constants include
  `Administrator`, channel send/read/manage flags, moderation flags, voice
  flags, thread flags, poll flags, and application-command flags.
- `Permission.Add` and `Remove` mutate a bitfield; `Has` checks any matching bit
  and `HasAll` checks every requested bit.
- `permissions.Build(...Permission) Permission` combines flags.
- `NewBuilder(initial ...Permission) *Builder`, `Builder.Add`, `Remove`, and
  `Build` provide fluent composition.
- `permissions.Overwrite` contains `ID`, `Type`, `Allow`, and `Deny` (all
  `Permission` typed). `channels.Overwrite` also uses `permissions.Permission`
  for `Allow` and `Deny`, so the two types are compatible without manual
  string conversion.
- `permissions.Calculate(memberID, guildID, guildOwnerID, baseRolePermissions,
  memberRoleIDs, memberRolePermissions, overwrites) Permission` applies owner,
  administrator, everyone, role, and member overwrite precedence.
- `RequirePermissions`, `RequireAnyPermissions`, `RequireBotPermissions`,
  `RequireRole`, `RequireAnyRole`, `RequireOwners`, `GuildOnly`, `OwnerOnly`,
  and `Validate` return slash-command `Middleware`.
- `PrefixGuildOnly`, `RequirePrefixPermissions`, `PrefixCooldown`, and the
  prefix command configuration methods operate on prefix handlers.
- `Cooldown(time.Duration)` returns per-user interaction middleware. The
  cooldown key includes user, guild, and command name.
- `InteractionContext.MemberPermissions()` and `BotPermissions()` expose the
  permissions used by custom policy code.

## Examples

- [Moderation](../examples/commands/moderation.md)
- [Slash command middleware](../examples/commands/slash-commands.md)
- [Permissions low-level guide](../low-level/permissions/README.md)

## Related APIs

- [`commands.md`](commands.md) for middleware attachment.
- [`caching.md`](caching.md) for prefix member lookups.
- [`../low-level/permissions/README.md`](../low-level/permissions/README.md) for bitfields.
