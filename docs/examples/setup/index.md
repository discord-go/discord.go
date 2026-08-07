# Getting Started

## Overview

This section is the discord.go adaptation of the Discord.js guide's Getting Started topics. It takes you from an empty Go directory to an application that can connect to Discord, receive `READY`, and answer a slash command. Discord.js calls its entry point a client; discord.go uses `bot.Bot`, a router, and typed contexts.

Read the pages in this order:

1. [App Setup](app-setup.md) creates the application and bot token.
2. [Adding Your App](adding-your-app.md) installs the application in a test guild.
3. [Installation](installation.md) creates the Go module and adds discord.go.
4. [Linter](linter.md) establishes formatting and static checks.

## Architecture

Discord has two separate control planes. The Gateway delivers events such as `READY` and interactions over a WebSocket. The REST API sends replies, registers commands, and changes guild resources. A `bot.Bot` owns both connections. `bot.Router` turns application-command definitions into interaction handlers, while `InteractionContext` provides typed option access and response methods.

## Prerequisites

- A Linux, macOS, or Windows machine with Go `1.26.4` or newer.
- A Discord account that can create an application and install it in a test guild.
- A private test guild and a channel where the bot can send messages.
- A shell environment where `DISCORD_TOKEN` can be set without committing it.

## Quick Start

Create an empty module, add the local module as a dependency when working from a checkout, and run this complete program as `main.go`:

```bash
mkdir discord-bot
cd discord-bot
go mod init example.com/discord-bot
go mod edit -replace discord.go=/path/to/discord.go
go get discord.go
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
	router.Command("hello", "Say hello", func(ctx *bot.InteractionContext) {
		if err := ctx.Reply("Hello from discord.go."); err != nil {
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

Create the application in the Developer Portal, add a bot user, copy the token once, and then create a Go module. Install the application with the `bot` scope and the `applications.commands` scope. In code, construct a router before `bot.New`, attach it with `bot.WithRouter`, and finish with `b.Run()`.

## Common Patterns

- Use a separate test application and test guild while learning.
- Use `intents.Guilds` for slash-command-only bots.
- Add `intents.GuildMessages` and `intents.MessageContent` only for message content or prefix commands.
- Use `bot.WithGuildCommandSync` during development so command edits appear quickly.
- Put startup logging in `OnReady` and runtime diagnostics in `OnError`.

## Best Practices

- Store tokens in environment variables or a secret manager.
- Request only the Gateway intents and Discord permissions the features need.
- Keep production and development applications separate.
- Validate the router before startup with `router.Validate()` when commands are assembled dynamically.
- Treat `READY` as a lifecycle event, not a one-time configuration hook; reconnects can produce another ready cycle.

## Common Mistakes

### Incorrect

```go
const token = "paste-the-token-here"
b := bot.New(token)
```

### Correct

```go
token := os.Getenv("DISCORD_TOKEN")
if token == "" {
	log.Fatal("DISCORD_TOKEN is required")
}
b := bot.New(token, bot.WithIntents(intents.Guilds))
```

### Incorrect

```go
b.OnReady(func(ctx *bot.ReadyContext) {
	router.Command("hello", "Say hello", handler)
})
```

### Correct

```go
router.Command("hello", "Say hello", handler)
b := bot.New(token, bot.WithRouter(router))
```

Register static routes before `Run`; the bot's command synchronizer reads the router when `READY` is handled.

## API Walkthrough

- `bot.New(token, opts...)` creates the Gateway and REST-backed bot.
- `bot.WithIntents` selects the Gateway subscription.
- `bot.NewRouter` creates command and interaction routing tables.
- `router.Command` registers a chat-input command.
- `bot.WithRouter` attaches the router and enables automatic command synchronization.
- `b.OnReady` observes successful Gateway identification.
- `b.Run` starts the bot and handles SIGINT and SIGTERM graceful shutdown.

## Examples

- [Basic Client](../setup/basic-client.md) adds prefix and direct message handling.
- [Slash Commands](../commands/slash-commands.md) expands options, middleware, and deferred responses.
- [Full Template](../advanced/full-template.md) shows configuration and a multi-file layout.
- [Complete source example](../code/ping/main.go) is a runnable package in this repository.

## Related Pages

- [App Setup](app-setup.md)
- [Adding Your App](adding-your-app.md)
- [Installation](installation.md)
- [Linter](linter.md)
- [Commands overview](../commands/README.md)
- [Examples overview](../README.md)
