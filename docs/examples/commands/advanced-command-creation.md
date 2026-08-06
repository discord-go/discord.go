# Advanced Command Creation

## Overview

Advanced commands combine option schemas, context-menu types, categories, validation, permissions, cooldowns, and middleware. The router keeps these policies close to the command definition.

## Quick Start

```go
package main
import (
    "errors"
    "time"
    "github.com/discord-go/discord.go/bot"
    "github.com/discord-go/discord.go/interactions"
    "github.com/discord-go/discord.go/permissions"
)
func register(r *bot.Router) {
    r.Command("lookup", "Look up a user", func(c *bot.InteractionContext) { _ = c.Reply(c.GetUserID("user").String()) }, interactions.ApplicationCommandOption{Type: interactions.ApplicationCommandOptionTypeUser, Name: "user", Description: "Target", Required: true}).RequirePermissions(permissions.ManageGuild).Cooldown(time.Second).Use(bot.Validate(func(c *bot.InteractionContext) error { if !c.InGuild() { return errors.New("guild only") }; return nil }))
}
```

## Common Patterns

Use `ContextCommand`, `UserCommand`, and `MessageCommand` for context menus. Use `InCategory` for help systems and `RangeCommands` to build deterministic command indexes.

## Common Mistakes

Do not hide required options in handler code. Put requiredness and descriptions in `ApplicationCommandOption` so Discord validates input before invocation.

## Related Pages

- [Commands](README.md)
- [Permissions](../more-to-know/permissions.md)
