# Command Responses

## Overview

An interaction needs an initial response within Discord's deadline. After that response, use edits or followups. Prefix commands use channel message helpers and can send complex payloads or files.

## Quick Start

```go
package main
import "github.com/discord-go/discord.go/bot"
func handle(c *bot.InteractionContext) { if err := c.Defer(); err != nil { return }; _, _ = c.Followup("Finished") }
```

## Common Patterns

Use `ReplyEphemeral` for private validation failures, `Defer` for slow work, `EditReply` for progress completion, and `FollowupComplexWithFiles` for generated results.

## Common Mistakes

Calling `Reply` after `Defer` creates a second initial response. Call `Followup` or `EditReply` instead.

## Related Pages

- [Interactions](../interactions/interactions.md)
- [Errors](../../high-level/errors.md)
