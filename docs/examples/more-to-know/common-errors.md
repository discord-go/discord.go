# Common Errors

## Overview

`discord.go` returns errors for lifecycle operations, REST calls, interaction
responses, collectors, and asynchronous Gateway work. Handle direct errors at
the call site, use `errors.Is` for exported sentinels, and use `errors.As` for
typed errors such as `*rest.APIError` and `*bot.HandlerPanicError`.

## Tutorial: Preserve Error Identity

1. Check every initial interaction response.
2. Match sentinels with `errors.Is`, not string comparisons.
3. Inspect REST status and code with `errors.As`.
4. Install `WithErrorHandler` or `OnError` for asynchronous failures.
5. Treat cancellation during shutdown as expected.

## Complete Runnable Example

Copy to `examples/common-errors/main.go`, set `DISCORD_TOKEN`, and run it.
Invoke `/errors` to see a deliberately handled duplicate acknowledgement.

```go
package main

import (
	"errors"
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
	router.Command("errors", "Demonstrate error handling", func(ctx *bot.InteractionContext) {
		if err := ctx.Reply("The first acknowledgement succeeded."); err != nil {
			log.Printf("initial response: %v", err)
			return
		}
		if err := ctx.Reply("This is intentionally a second acknowledgement."); err != nil {
			if errors.Is(err, bot.ErrInteractionAlreadyResponded) {
				log.Printf("expected duplicate acknowledgement: %v", err)
				return
			}
			log.Printf("unexpected second response error: %v", err)
		}
	})

	b := bot.New(token,
		bot.WithIntents(intents.Guilds),
		bot.WithRouter(router),
		bot.WithErrorHandler(func(err error) {
			var panicErr *bot.HandlerPanicError
			if errors.As(err, &panicErr) {
				log.Printf("handler panic event=%s value=%v", panicErr.Event, panicErr.Value)
				return
			}
			log.Printf("asynchronous bot error: %v", err)
		}),
	)
	b.OnError(func(err error) { log.Printf("additional error observer: %v", err) })
	if err := b.Run(); err != nil {
		log.Printf("bot stopped: %v", err)
	}
}
```

## Error Categories

- `bot.ErrMissingToken`, `ErrBotAlreadyRunning`, and `ErrBotNotRunning` describe
  lifecycle state.
- `bot.ErrInteractionAlreadyResponded` means the initial callback was already
  accepted; use `Followup` or an edit instead.
- `bot.ErrCollectorClosed` describes a collector that was cancelled before a
  match.
- `*rest.APIError` exposes Discord's numeric code, HTTP status, message, and
  structured error details.
- `*bot.HandlerPanicError` identifies a recovered handler panic and its event.

## REST Classification

```go
var apiErr *rest.APIError
if errors.As(err, &apiErr) {
	log.Printf("discord status=%d code=%d message=%s", apiErr.HTTPStatus, apiErr.Code, apiErr.Message)
}
```

Wrap application context with `%w` so callers can still classify the original
error. Avoid matching human-readable error strings.

## Common Mistakes

- Ignoring `_ = ctx.Reply(...)` in production handlers.
- Trying a second initial response after a timeout without knowing whether the
  first request reached Discord.
- Logging webhook tokens inside a REST error or request URL.
- Treating every cancellation as an outage.
- Performing blocking alert delivery inside `OnError`.

## Expected Result

`/errors` logs the exported duplicate-response sentinel instead of comparing an
error string. The process also has an asynchronous error callback for runtime
failures.

## Related Pages

- [Interactions](../interactions/interactions.md)
- [Collectors](collectors.md)
- [Webhooks](webhooks.md)
