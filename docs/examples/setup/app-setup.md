# App Setup

## Overview

An application is the Discord identity that owns commands, permissions, and a bot user. This page adapts the Discord.js App Setup topic to discord.go. The Portal steps are language-independent; the Go code below proves that the resulting bot token can authenticate and receive `READY`.

## Architecture

The Developer Portal stores application metadata and bot configuration. Discord issues a bot token for the Gateway and REST APIs. `bot.New` keeps that token private and creates the two clients used by the high-level API. `OnReady` runs after Discord accepts the token and sends the initial Gateway state.

## Prerequisites

- A Discord account allowed to create applications.
- Go `1.26.4` or newer.
- A shell with `DISCORD_TOKEN` available.
- A test guild for the later [Adding Your App](adding-your-app.md) step.

## Quick Start

In the Developer Portal, choose **New Application**, open **Bot**, choose **Reset Token** only when necessary, and copy the new token to a secret store. Enable the `Guilds` intent in the code below. Save it as `main.go` in a Go module and run it:

```bash
export DISCORD_TOKEN='replace-with-a-bot-token'
go mod init example.com/app-setup
go mod edit -replace discord.go=/path/to/discord.go
go get discord.go
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

	b := bot.New(token, bot.WithIntents(intents.Guilds))
	b.OnReady(func(ctx *bot.ReadyContext) {
		log.Printf("authenticated as %s (application ID %s)", ctx.User.Username, ctx.User.ID)
	})
	b.OnError(func(err error) {
		log.Printf("discord.go error: %v", err)
	})
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Creating/Using

1. Create an application and record its application ID for later REST work.
2. Create a bot user from the Bot page.
3. Enable only the intents required by the planned features. `Guilds` is enough for the program above.
4. Never paste the token into source, a public issue, or a command that will be stored in shell history.
5. Use OAuth2 installation after the bot exists; the token alone does not install the application in a guild.

The code does not need the application ID for Gateway startup. `b.AppID()` becomes available after `READY` when code needs it at runtime.

## Common Patterns

- Make `DISCORD_TOKEN` mandatory and fail closed.
- Configure privileged intents in both the Portal and `bot.WithIntents`.
- Use `bot.WithErrorHandler` or `OnError` to capture startup and handler failures.
- Use a dedicated test bot so development commands cannot affect production.

## Best Practices

- Reset a leaked token immediately in the Portal; changing an environment variable is not enough.
- Treat the application ID as public metadata but the bot token as a credential.
- Start with `Guilds`; add `MessageContent` only when content inspection is unavoidable.
- Give the bot the smallest guild permissions required by its commands.

## Common Mistakes

### Incorrect

```go
b := bot.New("YOUR_BOT_TOKEN")
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
bot.WithIntents(intents.Guilds | intents.MessageContent)
```

### Correct

```go
bot.WithIntents(intents.Guilds)
```

Add `MessageContent` only when the bot reads message text, and enable the privileged intent in the Portal first.

## API Walkthrough

- `bot.New` accepts the token and functional options.
- `bot.WithIntents(intents.Guilds)` selects a non-privileged Gateway intent.
- `b.OnReady` receives a typed `ReadyContext` with the authenticated user.
- `b.AppID()` exposes the application/user ID after `READY`.
- `b.Run` owns signal handling and graceful shutdown.
- `b.OnError` receives Gateway, synchronization, and handler errors.

## Examples

- [Basic Client](../basic-client.md) connects and responds to `/ping`.
- [Gateway](../gateway.md) explains typed and generic dispatch handlers.
- [Full Template](../full-template.md) adds presence and configuration.

## Related Pages

- [Adding Your App](adding-your-app.md)
- [Installation](installation.md)
- [Commands project setup](../commands/project-setup.md)
- [Discord application documentation](https://discord.com/developers/docs/quick-start/getting-started)
