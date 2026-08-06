# Parsing Options

## Overview

Interaction option values arrive from Discord's JSON payload. `InteractionContext` provides typed accessors that preserve large snowflakes and safely handle absent values.

## Quick Start

```go
package main
import "github.com/discord-go/discord.go/bot"
func handle(c *bot.InteractionContext) { name := c.GetStringOption("name"); count := c.GetIntOption("count"); target := c.GetUserID("target"); _ = c.Reply(name + " " + target.String() + " " + string(rune(count))) }
```

## API Walkthrough

Use `GetStringOption`, `GetIntOption`, `GetFloatOption`, `GetBoolOption`, `GetUserID`, `GetRoleID`, `GetChannelID`, `GetOption`, `HasOption`, `Subcommand`, and `FocusedOption`. Missing values return zero values; use `HasOption` when zero is meaningful.

## Common Mistakes

Do not decode Discord IDs through `float64`. JSON numbers can exceed JavaScript's safe integer range; the context uses `json.Number` internally.

## Related Pages

- [Interactions](../interactions/interactions.md)
- [Slash Commands](README.md)
