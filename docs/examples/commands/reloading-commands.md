# Reloading Commands

## Overview

Go does not dynamically reload compiled command functions like a JavaScript module loader. The safe equivalent is to build a new router in a controlled process, validate it, then restart or replace the bot during deployment.

## Architecture

Command registration is deterministic Go code. `CommandE`, `MustCommand`, `Validate`, and `Commands` make the registry inspectable before the Gateway starts. This avoids partially reloaded handlers and stale function pointers.

## Quick Start

```go
package main
import (
    "log"
    "github.com/discord-go/discord.go/bot"
)
func buildRouter() *bot.Router {
    r := bot.NewRouter()
    r.MustCommand("ping", "Check status", func(c *bot.InteractionContext) { _ = c.Reply("Pong") })
    if err := r.Validate(); err != nil { panic(err) }
    return r
}
func main() { log.Printf("loaded %d commands", buildRouter().CommandCount()) }
```

## Best Practices

Validate before replacing a running process, deploy command changes through a test guild first, and keep command handlers in separate Go packages that expose registration functions.

## Related Pages

- [Project Setup](project-setup.md)
- [Deploying Commands](deploying-commands.md)
