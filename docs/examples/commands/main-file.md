# Main File

## Overview

The main file is the executable composition root. It should load configuration, create the router, register commands and observers, configure intents, and start the bot. This page adapts the Discord.js Main File topic to a Go lifecycle with explicit context support.

## Architecture

`bot.RunContext` starts the Gateway and runs until its context is canceled or the Gateway stops. Command synchronization occurs as part of the bot's READY handling. `OnReady` is for logging and startup observations; definitions should already exist. `signal.NotifyContext` translates process shutdown into the context passed to the bot.

## Prerequisites

- A Go module and `discord.go` dependency.
- `DISCORD_TOKEN` set in the environment.
- An installed application with at least `View Channel` and `Send Messages`.
- Familiarity with `context.Context` and process signals.

## Quick Start

Save this complete program as `main.go`:

```bash

```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}

	router := bot.NewRouter()
	router.Command("status", "Show bot status", func(ctx *bot.InteractionContext) {
		if err := ctx.Reply("Online."); err != nil {
			log.Printf("status reply: %v", err)
		}
	})

	b := bot.New(token,
		bot.WithIntents(intents.Guilds),
		bot.WithRouter(router),
	)
	b.OnReady(func(ctx *bot.ReadyContext) {
		log.Printf("ready as %s", ctx.User.Username)
	})
	b.OnDisconnect(func() {
		log.Println("gateway disconnected")
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := b.RunContext(ctx); err != nil {
		log.Printf("bot stopped: %v", err)
	}
}
```

## Creating/Using

The program builds all state before starting the Gateway. In a service manager, the parent process can own the context and call `RunContext` directly. In a simple executable, `Run` is shorter and already listens for SIGINT and SIGTERM. Choose one lifecycle owner; do not call both `Run` and `Start` for the same bot.

## Common Patterns

- Read and validate environment configuration at the top of `main`.
- Create one router and attach it once.
- Register `OnError`, `OnReady`, and `OnDisconnect` before starting.
- Use `RunContext` when an HTTP server or service supervisor owns cancellation.
- Use `Start`, `Done`, and `Stop` when startup and shutdown need separate phases.

## Best Practices

- Keep `main` short and move feature registration into functions or packages.
- Give shutdown a deadline when active handlers can perform external work.
- Do not perform blocking initialization from `OnReady` without a bounded context.
- Treat reconnects as normal and make lifecycle observers idempotent.
- Return a non-zero process status for configuration and startup failures.

## Common Mistakes

### Incorrect

```go
b.Run()
b.RunContext(ctx)
```

### Correct

```go
if err := b.RunContext(ctx); err != nil {
	log.Printf("bot stopped: %v", err)
}
```

### Incorrect

```go
b.OnReady(func(ctx *bot.ReadyContext) {
	registerCommands(bot.NewRouter())
})
```

### Correct

```go
router := bot.NewRouter()
registerCommands(router)
b := bot.New(token, bot.WithRouter(router))
```

The router must be attached before the bot receives `READY`.

## API Walkthrough

- `bot.Run` owns OS signal handling for a simple executable.
- `bot.RunContext` runs until a supplied context is canceled.
- `bot.Start` starts asynchronously and returns after the connection is created.
- `bot.Done` closes when the current run terminates.
- `bot.Stop` cancels the Gateway and waits for active handlers.
- `bot.OnDisconnect` observes a terminated Gateway loop.

## Examples

- [Project Setup](project-setup.md) shows feature registration.
- [Gateway](../more-to-know/gateway.md) uses explicit subscriptions and shutdown.
- [Full Template](../advanced/full-template.md) demonstrates production-oriented wiring.

## Related Pages

- [Project Setup](project-setup.md)
- [Handling Events](handling-events.md)
- [Installation](../setup/installation.md)
- [Bot lifecycle source](../../../bot/lifecycle.go)
