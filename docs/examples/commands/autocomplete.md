# Autocomplete

## Overview

Autocomplete interactions receive a focused option and must be acknowledged with choices rather than a normal message response.

## Quick Start

```go
package main
import (
    "github.com/discord-go/discord.go/bot"
    "github.com/discord-go/discord.go/interactions"
)
func register(r *bot.Router) {
    r.Command("search", "Search records", func(c *bot.InteractionContext) { _ = c.Reply("search") }, interactions.ApplicationCommandOption{Type: interactions.ApplicationCommandOptionTypeString, Name: "query", Description: "Search text", Autocomplete: true})
    r.Autocomplete("search", func(c *bot.InteractionContext) { _ = c.Autocomplete(interactions.ApplicationCommandOptionChoice{Name: c.GetStringOption("query"), Value: c.GetStringOption("query")}) })
}
```

## Best Practices

Return a small, fast result set. Use `FocusedOption` and do not perform long database or HTTP work inside Discord's autocomplete response window.

## Related Pages

- [Interactions](../interactions/interactions.md)
- [Slash Commands](../commands/slash-commands.md)
