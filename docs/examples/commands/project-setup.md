# Project Setup

## Overview

Project Setup is the first application-creation topic. A Go bot does not need a framework-generated directory tree, but a stable shape prevents command registration, configuration, and lifecycle code from becoming one untestable function. This page starts with one package and shows the boundary to use as the application grows.

## Architecture

`main` owns process concerns: token loading, router creation, bot options, event registration, and shutdown. Feature functions receive a router and register commands. Handlers receive `*bot.InteractionContext`, which keeps Discord transport details at the edge of the application. This arrangement can later become `cmd/bot`, `internal/commands`, and `internal/events` without changing the Discord API usage.

## Prerequisites

- A Go module with `discord.go` available; see [Installation](../setup/installation.md).
- A token in `DISCORD_TOKEN`.
- An installed application with `applications.commands`.
- A text editor and the Go formatter.

## Quick Start

Create `main.go` with this complete runnable program:

```bash
gofmt -w main.go
export DISCORD_TOKEN='replace-with-a-bot-token'
```

```go
package main

import (
	"log"
	"os"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
)

func registerCommands(router *bot.Router) {
	router.Command("hello", "Greet the person who used the command", func(ctx *bot.InteractionContext) {
		if err := ctx.Reply("Hello from a feature registration function."); err != nil {
			log.Printf("hello reply: %v", err)
		}
	})
}

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}

	router := bot.NewRouter()
	registerCommands(router)
	if err := router.Validate(); err != nil {
		log.Fatal(err)
	}

	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	b.OnError(func(err error) {
		log.Printf("runtime error: %v", err)
	})
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Creating/Using

Start with one package while learning. Once there are several features, move `registerCommands` and its handler into feature packages. Keep one router instance. Use dependency parameters for databases and services rather than package globals, and keep the `main` package as the composition root.

## Common Patterns

- Name feature registration functions `registerXCommands` and call them from `main`.
- Keep command definitions near their handler or domain service.
- Use a shared error logger through `bot.WithErrorHandler` or `OnError`.
- Test registration with `router.CommandCount`, `router.Lookup`, and `router.Validate`.
- Use `bot.WithMaxHandlerConcurrency` when downstream services have a fixed capacity.

## Best Practices

- Do not register commands from `OnReady`; reconnects can duplicate side effects.
- Keep configuration and secrets out of command packages.
- Make command registration deterministic so every process exposes the same schema.
- Prefer explicit interfaces for database and external API dependencies.
- Keep `go test ./...` independent of Discord network access.

## Common Mistakes

### Incorrect

```go
func registerCommands() {
	router := bot.NewRouter()
	router.Command("hello", "Say hello", handler)
}
```

### Correct

```go
func registerCommands(router *bot.Router) {
	router.Command("hello", "Say hello", handler)
}
```

### Incorrect

```go
router.Command("health", "Health", nil)
```

### Correct

```go
router.MustCommand("health", "Health check", func(ctx *bot.InteractionContext) {
	_ = ctx.Reply("ok")
})
```

Use `MustCommand` only for static startup definitions where a programming error should stop the process.

## API Walkthrough

- `*bot.Router` is safe to pass to registration helpers.
- `router.Command` returns `*bot.Command` for middleware chaining.
- `router.MustCommand` validates and panics with a useful error for invalid static definitions.
- `router.CommandCount` and `router.Lookup` support startup checks and help systems.
- `router.Validate` returns combined validation errors.

## Examples

- [Main File](main-file.md) focuses on lifecycle composition.
- [Creating Commands](creating-commands.md) adds typed options.
- [Full Template](../full-template.md) shows a multi-file repository example.
- [Template source](../code/v2_template/main.go) demonstrates split registration files.

## Related Pages

- [Installation](../setup/installation.md)
- [Main File](main-file.md)
- [Handling Commands](handling-commands.md)
- [Linter](../setup/linter.md)
