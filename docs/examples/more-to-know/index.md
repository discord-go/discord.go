# More To Know

These guides cover the Discord.js topics that are easy to get subtly wrong
when moving to `discord.go`: Gateway intents, cache and partial data, REST
resources, permissions, formatting, and lifecycle-safe workflows.

## Guides

- [Gateway Intents](gateway-intents.md) selects the minimum Gateway event set.
- [Partials And Cache](partials-cache.md) handles optional payloads and fetches.
- [Common Errors](common-errors.md) classifies direct and asynchronous errors.
- [Collectors](collectors.md) waits for one bounded event safely.
- [Formatters](formatters.md) builds mentions, timestamps, and safe text.
- [Embeds](embeds.md) creates and validates rich messages.
- [Canvas Alternatives](canvas-alternatives.md) creates images without a JS
  canvas dependency.
- [Permissions](permissions.md) checks member and bot permission bitfields.
- [Reactions](reactions.md) listens for reactions and calls reaction REST APIs.
- [Threads](threads.md) starts a thread from a message.
- [Webhooks](webhooks.md) creates and executes a webhook securely.
- [Audit Logs](audit-logs.md) reads and observes moderation history.

## General Prerequisites

- Go `1.26.4` or newer.
- A bot token in `DISCORD_TOKEN`.
- A test guild and channel where the bot has only the permissions required by
  the guide.
- Privileged intents enabled in the Developer Portal before selecting them in
  `bot.WithIntents`.

## API Shape

`bot.Bot` owns the Gateway and exposes the typed REST client as `b.Rest`.
Gateway events use `On...` handlers; resource operations use `context.Context`
and typed methods such as `GetAuditLog`, `StartThreadWithMessage`, and
`ExecuteWebhook`. The REST client has rate-limit handling, but application code
still needs deadlines and authorization checks.

## Tutorial: Add A Health Command

1. Select only `intents.Guilds` when the bot has no message or reaction handler.
2. Register a slash command with `bot.NewRouter`.
3. Reply immediately from the interaction handler.
4. Add a timeout when the command later performs REST work.

## Complete Runnable Example

Copy this to `examples/more-to-know-health/main.go`, set `DISCORD_TOKEN`, and
run `go run ./examples/more-to-know-health`.

```go
package main

import (
	"log"
	"os"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}
	router := bot.NewRouter()
	router.Command("health", "Show the bot health state", func(ctx *bot.InteractionContext) {
		if err := ctx.Reply("Gateway handler is running."); err != nil {
			log.Printf("health response: %v", err)
		}
	})
	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Security Baseline

- Never log tokens, webhook tokens, or complete user-submitted secrets.
- Do not treat a cache hit as authoritative permission data.
- Recheck authorization immediately before a destructive REST operation.
- Use `rest.WithReason` for moderation actions where an audit-log reason is
  useful.
- Treat IDs, mention text, webhook content, and uploaded files as untrusted.
