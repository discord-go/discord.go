# Lifecycle And Shutdown

## Overview

The bot lifecycle separates connection startup, Discord readiness, run
termination, and handler shutdown. Use `Run` for a conventional executable,
`RunContext` for a service that owns cancellation, or `Start` plus `Stop` and
`Wait` when the application needs explicit orchestration.

## Architecture

`Start(ctx)` creates a child run context, opens a gateway connection or shard
manager, and returns after the gateway loop has been launched. READY later marks
the bot ready and closes the internal readiness channel. Dispatch handlers run
in goroutines and are tracked by a wait group. `Stop` cancels the run, closes
connections, cancels jobs, waits for the run to finish, and waits for active
handlers.

Reconnect, resume, invalidation, and disconnect callbacks expose gateway state
changes. A resumed session does not necessarily mean a fresh READY event, so
applications that rebuild state should choose the appropriate callback.

## Quick Start

This complete program uses explicit startup and graceful shutdown with a signal
context.

```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	b := bot.New(token, bot.WithIntents(intents.Guilds))
	b.OnReady(func(ctx *bot.ReadyContext) {
		log.Printf("ready as %s", ctx.User.Username)
	})
	b.OnDisconnect(func() { log.Println("gateway disconnected") })

	if err := b.Start(ctx); err != nil {
		log.Fatal(err)
	}
	if err := b.WaitReady(ctx); err != nil {
		log.Fatal(err)
	}
	<-ctx.Done()

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer stopCancel()
	if err := b.Stop(stopCtx); err != nil {
		log.Fatal(err)
	}
}
```

`Run` provides the same signal behavior with less orchestration. In a service,
use the service's cancellation context instead of installing a second signal
handler.

## Creating/Configuration

Register lifecycle callbacks with `OnReconnect`, `OnResume`, `OnInvalidated`,
and `OnDisconnect`. Register `OnReady` before starting. Use
`WithShards(count)` for explicit sharding or `WithShards(0)` to ask Discord for
the recommended count. `WithMaxHandlerConcurrency` limits active handler
goroutines; it does not serialize event order.

## Using

### Basic: `Run`

`Run()` blocks until SIGINT, SIGTERM, or a fatal gateway error, then performs a
graceful shutdown. It is the best default for a standalone process.

### Intermediate: `RunContext`

`RunContext(ctx)` starts the bot and stops it when `ctx` is cancelled. It returns
the gateway or shutdown error after waiting for the run.

### Advanced: `Start` and state observation

Use `Start` when a supervisor controls the process. `Done` closes when the run
terminates, `Wait` waits for it and active handlers, and `State` reports the
current `BotState`. `WaitReady` should gate work that requires `AppID`, `User`,
or a gateway session.

## Common Patterns

- Call `WaitReady` before sending presence or joining voice.
- Use a shutdown timeout so a stuck handler cannot block process termination
  forever.
- Register `Every` jobs after startup; `Stop` cancels them automatically.
- Use `OnResume` for session-resume metrics and `OnReady` for fresh identify
  initialization.
- Treat `context.Canceled` and `context.DeadlineExceeded` as expected shutdown
  causes when wrapping errors.

## Best Practices

### Prefer `RunContext` in services

Why: the service owns cancellation and can coordinate dependencies.

Pros: clean integration with supervisors and tests.

Cons: the caller must provide cancellation and inspect the returned error.

### Wait for readiness explicitly

Why: `Start` means the loop was launched, not that Discord sent READY.

Pros: avoids zero application IDs and unavailable gateway operations.

Cons: startup has an additional wait and can fail or time out.

### Bound graceful shutdown

Why: a handler may block on an external dependency.

Pros: deployments complete predictably.

Cons: a deadline can stop waiting before every handler finishes; handlers should
honor their context to minimize this tradeoff.

## Common Mistakes

Incorrect: assuming `Start` means the bot is ready.

```go
_ = b.Start(ctx)
_ = b.SetStatus(ctx, "online")
```

Correct: wait for READY first.

```go
if err := b.Start(ctx); err != nil {
	return err
}
if err := b.WaitReady(ctx); err != nil {
	return err
}
```

Incorrect: exiting without stopping after a context cancellation.

```go
<-ctx.Done()
return nil
```

Correct: call `Stop` with a bounded context.

```go
<-ctx.Done()
stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
return b.Stop(stopCtx)
```

## API Walkthrough

- `Start(context.Context) error` launches a run asynchronously.
- `Run() error` owns SIGINT/SIGTERM handling; `RunContext(context.Context) error`
  is its non-signal service variant.
- `Stop(context.Context) error` cancels connections and jobs and waits for the
  run and active handlers.
- `Wait() error` waits for the current or most recent run; `Done() <-chan struct{}`
  closes when the run terminates.
- `State() BotState` returns `Stopped`, `Starting`, `Running`, or `Stopping`.
- `WaitReady(context.Context) error` waits for READY; `IsReady() bool` checks it.
- `OnReady`, `OnReconnect`, `OnResume`, `OnInvalidated`, and `OnDisconnect`
  register lifecycle callbacks.
- `Every(context.Context, time.Duration, func(context.Context)) func()` creates
  a cancellable run-owned job.
- `WithShards(int) Option` enables sharding; zero requests Discord's recommended
  count. `WithMaxHandlerConcurrency(int) Option` limits concurrent handlers.
- `BotStats` contains `StartedAt`, `EventsReceived`, `HandlerPanics`,
  `CommandSyncs`, `PrefixCommands`, and `SlashCommands`.

## Examples

- [Basic client](../examples/setup/basic-client.md)
- [Full template](../examples/advanced/full-template.md)
- [Gateway shards](../low-level/gateway/shards.md)

## Related APIs

- [`client.md`](client.md) for construction and runtime state.
- [`collectors.md`](collectors.md) for lifecycle-owned jobs.
- [`errors.md`](errors.md) for run and handler failures.
