# Creating Your App

## Overview

This section adapts the Discord.js guide's Creating Your App topics to discord.go. It explains how to organize a Go bot, define commands, route invocations, synchronize command schemas, and handle Gateway events. Start with [Project Setup](project-setup.md) if the application is empty.

## Architecture

A discord.go application normally has a small `main` package that loads configuration, creates one `bot.Router`, registers commands, creates one `bot.Bot`, registers event observers, and starts the lifecycle. The router handles slash commands and prefix commands. The bot handles Gateway events and exposes `Rest` for operations that need an explicit API call. Discord receives command definitions through REST, then sends command invocations back through the Gateway.

## Prerequisites

- Complete the [setup pages](../setup/README.md).
- Use Go `1.26.4` or newer.
- Have a bot token and an installed test application.
- Know the test guild ID if using fast guild synchronization.

## Quick Start

Save this complete program as `main.go` in a Go module and run it from the module root:

```bash
export DISCORD_TOKEN='replace-with-a-bot-token'
go run .
```

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
	router.Command("ping", "Check whether the bot is online", func(ctx *bot.InteractionContext) {
		if err := ctx.Reply("Pong."); err != nil {
			log.Printf("reply: %v", err)
		}
	})

	b := bot.New(token,
		bot.WithIntents(intents.Guilds),
		bot.WithRouter(router),
	)
	b.OnReady(func(ctx *bot.ReadyContext) {
		log.Printf("ready as %s", ctx.User.Username)
	})
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Creating/Using

Create definitions before starting the bot. `router.Command` registers a slash command and its handler. `bot.WithRouter` makes the bot synchronize those definitions after `READY`. For a larger application, keep command registration in feature functions and pass the same router to each function; do not create a new router per command.

## Common Patterns

- Keep `main` responsible for wiring and feature packages responsible for registration.
- Use `router.CommandE` or `router.MustCommand` to validate static definitions during startup.
- Use `bot.GuildOnly`, `bot.RequirePermissions`, and `bot.Cooldown` as route middleware.
- Defer an interaction before slow REST or database work.
- Use a guild sync target while developing and explicit REST deployment in CI or a release job.

## Best Practices

- Keep command names lower-case and descriptions meaningful.
- Treat options, custom IDs, and context-menu targets as untrusted input.
- Keep one deterministic command registry per deployment.
- Log command names and IDs, but do not log tokens or sensitive option values.
- Make handlers safe for concurrent execution and return all response errors to a logger.

## Common Mistakes

### Incorrect

```go
router := bot.NewRouter()
b := bot.New(token, bot.WithRouter(bot.NewRouter()))
router.Command("ping", "Check the bot", handler)
```

### Correct

```go
router := bot.NewRouter()
router.Command("ping", "Check the bot", handler)
b := bot.New(token, bot.WithRouter(router))
```

### Incorrect

```go
router.Command("Ping", "", handler)
```

### Correct

```go
router.Command("ping", "Check whether the bot is online", handler)
```

The router normalizes names to lower-case, but `router.Validate` still requires valid Discord names and non-empty chat-input descriptions.

## API Walkthrough

- `bot.NewRouter` creates the registry.
- `router.Command` registers a slash command.
- `router.Prefix` registers a message command when message content is enabled.
- `router.Use` applies middleware to slash commands.
- `router.Validate` checks the command registry without connecting.
- `bot.WithRouter` attaches the registry to the bot.
- `bot.WithCommandSync` or `bot.WithGuildCommandSync` selects synchronization behavior.

## Examples

- [Project Setup](project-setup.md) separates wiring from registration.
- [Creating Commands](creating-commands.md) defines options and handlers.
- [Handling Commands](handling-commands.md) covers acknowledgements and follow-ups.
- [Deploying Commands](deploying-commands.md) uses the REST API explicitly.
- [Handling Events](handling-events.md) observes Gateway lifecycle and guild events.
- [Slash Commands](../commands/slash-commands.md) contains a larger command catalog.

## Related Pages

- [Setup overview](../setup/README.md)
- [Main File](main-file.md)
- [Commands examples](../README.md)
- [Gateway](../more-to-know/gateway.md)
