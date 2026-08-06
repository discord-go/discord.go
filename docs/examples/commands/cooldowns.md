# Cooldowns

## Overview

Cooldowns prevent one user from repeatedly invoking expensive or disruptive commands. In `discord.go`, cooldowns are middleware, so the policy runs before the handler and shares the same lifecycle as other command checks.

## Architecture

`Command.Cooldown` uses interaction identity, guild identity, and command name. `PrefixCooldown` provides the equivalent prefix policy. A rejected invocation receives an explanatory response and the handler is not called.

## Quick Start

```go
package main
import (
    "log"
    "os"
    "time"
    "github.com/discord-go/discord.go/bot"
    "github.com/discord-go/discord.go/intents"
)
func main() {
    r := bot.NewRouter()
    r.Command("slow", "Run a limited command", func(c *bot.InteractionContext) { _ = c.Reply("accepted") }).Cooldown(10*time.Second)
    b := bot.New(os.Getenv("DISCORD_TOKEN"), bot.WithIntents(intents.Guilds), bot.WithRouter(r))
    if err := b.Run(); err != nil { log.Fatal(err) }
}
```

## API Walkthrough

Use `Cooldown(duration)` as middleware or chain `.Cooldown(duration)` on a command. Use `PrefixCooldown` on a `PrefixCommand`. Cooldown state is process-local; use a shared storage-backed policy when running multiple processes.

## Common Mistakes

Do not use a global mutex in a handler and assume it is a cooldown. Use middleware so rejected requests never execute business logic.

## Related Pages

- [Permissions](../more-to-know/permissions.md)
- [Commands](README.md)
