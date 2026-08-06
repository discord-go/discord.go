# Permissions

## Overview

`permissions.Permission` is a 64-bit bitfield. Combine constants with
`permissions.Build` or `NewBuilder`; use `Has` for any-bit checks and `HasAll`
for all-bit checks. Router middleware can check both the invoking member and
the bot's permissions from an interaction.

## Tutorial: Protect A Command

1. Restrict the route with `bot.GuildOnly`.
2. Build the exact permissions the member needs.
3. Add `RequirePermissions` for the member.
4. Add `RequireBotPermissions` for what the bot must do.
5. Recheck resource-specific authorization immediately before REST mutation.

Permissions are channel-aware in Discord. The interaction's `Member.Permissions`
and `AppPermissions` reflect the invocation context, but changes can happen
after the interaction arrives, so REST errors still need handling.

## Complete Runnable Example

Copy to `examples/permissions/main.go`, set `DISCORD_TOKEN`, and run it.
Invoke `/moderate` in a test guild.

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

	required := permissions.Build(permissions.ManageMessages, permissions.ReadMessageHistory)
	router := bot.NewRouter()
	router.Command("moderate", "Run a permission-protected action", func(ctx *bot.InteractionContext) {
		if err := ctx.Reply("You and the bot passed the permission checks."); err != nil {
			log.Printf("permission response: %v", err)
		}
	}).
		Use(bot.GuildOnly()).
		Use(bot.RequirePermissions(required)).
		Use(bot.RequireBotPermissions(permissions.SendMessages))

	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Bitfield Operations

```go
required := permissions.NewBuilder(permissions.ViewChannel).
	Add(permissions.SendMessages, permissions.EmbedLinks).
	Build()

if !memberPermissions.HasAll(required) {
	return
}
if memberPermissions.Has(permissions.BanMembers | permissions.KickMembers) {
	// At least one of the two permissions is present.
}
```

`RequirePermissions` uses `HasAll`. Use `RequireAnyPermissions` only when any
one of the requested bits is intentionally sufficient. Avoid
`permissions.Administrator` unless the product genuinely requires it.

## Prefix Commands

Prefix permission middleware needs a configured cache implementing
`cache.MemberCache`:

```go
router.Prefix("purge", handler).
	Use(bot.PrefixGuildOnly()).
	Use(bot.RequirePrefixPermissions(permissions.ManageMessages))
```

Reject a cache miss rather than treating it as permission. Fetch the member
explicitly if the operation justifies the additional REST request.

## Common Mistakes

- Using `Has` when all requested permissions are required.
- Checking the member but not the bot.
- Treating a role name or cache entry as proof of effective permissions.
- Granting Administrator to avoid understanding a missing bit.
- Performing a mutation after authorization without handling a 403 response.

## Expected Result

Only members with both `ManageMessages` and `ReadMessageHistory` can invoke the
command, and the bot must have `SendMessages` in that channel.

## Related Pages

- [Partials And Cache](partials-cache.md)
- [Audit Logs](audit-logs.md)
- [Gateway Intents](gateway-intents.md)
