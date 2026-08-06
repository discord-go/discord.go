# Context Menus

## Overview

User and message context menus invoke commands against a selected Discord object rather than collecting slash options.

## Quick Start

```go
package main
import "github.com/discord-go/discord.go/bot"
func register(r *bot.Router) { r.UserCommand("Inspect user", func(c *bot.InteractionContext) { _ = c.Reply("Target: " + c.TargetID().String()) }); r.MessageCommand("Inspect message", func(c *bot.InteractionContext) { _ = c.Reply("Target: " + c.TargetID().String()) }) }
```

## Best Practices

Use `TargetID` as an identifier and fetch the current resource through `Bot.Rest` when details are required. Context commands are distinct remote command types and cannot have chat-input options.

## Related Pages

- [Commands](README.md)
- [Interactions](../interactions/interactions.md)
