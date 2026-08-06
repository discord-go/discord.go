# Deleting Commands

## Overview

Command synchronization uses bulk overwrite semantics: the commands supplied to the overwrite call become the active set. Removing a command from the router removes it from the next synchronization.

## Quick Start

```go
package main
import "github.com/discord-go/discord.go/bot"
func main() { r := bot.NewRouter(); r.Command("temporary", "Temporary command", func(c *bot.InteractionContext) {}); r.RemoveCommand("temporary") }
```

## Common Patterns

Use guild synchronization while testing deletions. Global command deletion can be delayed by Discord propagation. For surgical production changes, call the typed REST delete method with the application and command IDs.

## Common Mistakes

Do not assume removing a local handler deletes a remote command immediately. The remote registry changes only after synchronization.

## Related Pages

- [Deploying Commands](deploying-commands.md)
- [REST Endpoints](../../low-level/rest/endpoints.md)
