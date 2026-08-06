# Slash Command Permissions

## Overview

Discord supplies the invoking member's channel permissions and the bot's application permissions in interaction payloads. Middleware turns those values into reusable policy checks.

## Quick Start

```go
package main
import (
    "github.com/discord-go/discord.go/bot"
    "github.com/discord-go/discord.go/permissions"
)
func register(r *bot.Router) { r.Command("moderate", "Moderate a member", func(c *bot.InteractionContext) { _ = c.Reply("allowed") }).RequirePermissions(permissions.ModerateMembers).RequireBotPermissions(permissions.ModerateMembers) }
```

## Best Practices

Check both user and bot permissions, enforce hierarchy for role/member changes, and use `GuildOnly` for commands that require guild state. Return ephemeral denial messages.

## Related Pages

- [Permissions](../more-to-know/permissions.md)
- [Middleware](../../high-level/permissions.md)
