# Errors And Recovery

## Overview

The bot error path reports gateway failures, payload decoding errors, command
sync failures, handler panics, scheduled-job panics, and errors returned by
internal callbacks. The default handler logs through the configured logger.
Applications should register `OnError` for metrics, alerting, and structured
logging while still handling errors returned directly from response methods.

## Architecture

`WithErrorHandler` sets the primary callback; `OnError` adds callbacks. Runtime
dispatch recovers panics from user event handlers and reports a
`*bot.HandlerPanicError` containing the event and recovered value. A panic in an
error callback is suppressed so error reporting cannot crash the dispatch loop.

Some APIs return errors directly because the caller can act immediately:
`Start`, `Stop`, REST helpers, interaction replies, and presence methods. The
error callback is for asynchronous failures where returning an error is not
possible.

## Quick Start

This complete program installs a logger and classifies recovered handler
panics. The command is intentionally safe unless `PANIC_COMMAND` is set.

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
	logger := log.Default()
	router := bot.NewRouter()
	router.Command("diagnostics", "Exercise error reporting", func(ctx *bot.InteractionContext) {
		if os.Getenv("PANIC_COMMAND") == "1" {
			panic("intentional test panic")
		}
		if err := ctx.Reply("No error was generated"); err != nil {
			logger.Printf("reply: %v", err)
		}
	})

	b := bot.New(token,
		bot.WithIntents(intents.Guilds),
		bot.WithRouter(router),
		bot.WithLogger(logger),
		bot.WithErrorHandler(func(err error) {
			var panicErr *bot.HandlerPanicError
			if errors.As(err, &panicErr) {
				logger.Printf("handler panic in %s: %v", panicErr.Event, panicErr.Value)
				return
			}
			logger.Printf("bot error: %v", err)
		}),
	)
	b.OnError(func(err error) { logger.Printf("metric error=%v", err) })
	if err := b.Run(); err != nil {
		logger.Printf("run stopped: %v", err)
	}
}
```

Set `PANIC_COMMAND=1` only in a test guild. Handler panics are recovered and
reported, but the command's interaction still needs a response in normal code.

## Creating/Configuration

Use `WithLogger(*log.Logger)` for lifecycle diagnostics and
`WithErrorHandler(bot.ErrorHandler)` for the primary asynchronous error path.
Call `OnError` to add observers without replacing the default callback. The
`ErrorHandler` type is `func(error)`.

For direct operations, check the returned error. Use `errors.Is` for sentinel
errors such as `ErrMissingToken`, `ErrBotAlreadyRunning`,
`ErrBotNotRunning`, and `ErrInteractionAlreadyResponded`; use `errors.As` for
`*HandlerPanicError` and REST `*rest.APIError`.

## Using

### Basic: log and return direct errors

Do not discard an initial interaction response error. Log it or report it and
stop trying to send a second initial response unless the specific error path is
known to be retryable.

### Intermediate: classify errors

Use `errors.Is` and `errors.As` to distinguish cancellation, API status, panic,
and application failures. Attach event names and command names to application
logs around the API call.

### Advanced: recover operationally

Count `HandlerPanicError`, command sync failures, and gateway disconnects
separately. Keep the bot alive after an isolated handler panic, but alert when
panic counts or repeated gateway errors exceed an operational threshold.

## Common Patterns

- Use the default logger for human-readable startup diagnostics and `OnError`
  for structured metrics.
- Wrap REST errors with operation context using `%w`.
- Treat `context.Canceled` during deployment as expected.
- Include `HandlerPanicError.Event` in alerts.
- Test both direct error returns and asynchronous error callbacks.

## Best Practices

### Do not use error callbacks as normal control flow

Why: asynchronous callbacks cannot return a result to the original caller.

Pros: clear ownership and better local recovery.

Cons: error reporting code exists in more than one place.

### Preserve error identity

Why: sentinel and typed errors support reliable classification.

Pros: callers can use `errors.Is` and `errors.As` instead of string matching.

Cons: wrapping requires `%w`, and converting to a string loses useful fields.

### Keep callbacks non-blocking

Why: error callbacks run from runtime paths that should continue dispatching.

Pros: an alerting outage does not stop the bot.

Cons: durable delivery may require an asynchronous queue and its own shutdown.

## Common Mistakes

Incorrect: ignoring an initial response error.

```go
_ = ctx.Reply("done")
```

Correct:

```go
if err := ctx.Reply("done"); err != nil {
	log.Printf("reply error: %v", err)
	return
}
```

In normal code, register `OnError` during setup rather than inside a handler:

```go
b.OnError(func(err error) { log.Printf("bot error: %v", err) })
```

Incorrect: matching error strings.

```go
if err.Error() == "bot: interaction already responded" {
	// fragile
}
```

Correct:

```go
if errors.Is(err, bot.ErrInteractionAlreadyResponded) {
	// handle a duplicate acknowledgement
}
```

## API Walkthrough

- `ErrorHandler` is `func(error)`; `WithErrorHandler` sets the default callback
  and `OnError` registers additional callbacks.
- `WithLogger(*log.Logger) bot.Option` chooses the lifecycle and default error
  logger; nil restores the standard logger.
- `HandlerPanicError` has `Event` and `Value` fields and implements `Error()`.
- `ErrMissingToken`, `ErrBotAlreadyRunning`, `ErrBotNotRunning`,
  `ErrInteractionAlreadyResponded`, and `ErrCollectorClosed` are exported
  sentinel errors for common state failures.
- `rest.APIError` has `Code`, `Message`, `Errors`, and `HTTPStatus`; use
  `errors.As` to inspect Discord REST failures.
- `rest.CaptchaError` embeds `APIError` and adds CAPTCHA challenge fields.
- `Bot.Stats()` exposes `HandlerPanics` and event counters for health reporting.

## Examples

- [Full template](../examples/advanced/full-template.md)
- [Basic client](../examples/setup/basic-client.md)
- [REST errors](../low-level/rest/requests.md)

## Related APIs

- [`lifecycle.md`](lifecycle.md) for shutdown error ownership.
- [`client.md`](client.md) for logger and handler options.
- [`../low-level/rest/README.md`](../low-level/rest/README.md) for API errors and rate limits.
