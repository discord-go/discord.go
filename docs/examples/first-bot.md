# Building Your First Bot

## Overview

This tutorial walks through creating a Discord bot from scratch using discord.go.
By the end, you will have a running bot with slash commands.

## Prerequisites

- Go 1.26 or newer.
- A Discord application with a bot token (see [App Setup](setup/app-setup)).

## Step 1: Create a Project

```bash
mkdir mybot && cd mybot
go mod init mybot
go get github.com/discord-go/discord.go
```

## Step 2: Write the Main File

Create `main.go`:

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

    // Register a slash command
    router.Command("ping", "Check bot status", func(ctx *bot.InteractionContext) {
        if err := ctx.Reply("Pong!"); err != nil {
            log.Printf("reply: %v", err)
        }
    })

    // Register a second command
    router.Command("hello", "Say hello", func(ctx *bot.InteractionContext) {
        if err := ctx.Reply("Hello, world!"); err != nil {
            log.Printf("reply: %v", err)
        }
    })

    client := bot.New(token,
        bot.WithIntents(intents.Guilds),
        bot.WithRouter(router),
        bot.WithGuildCommandSync(os.Getenv("GUILD_ID")), // fast dev sync
    )

    client.OnReady(func(ctx *bot.ReadyContext) {
        log.Printf("logged in as %s", ctx.User.Username)
    })

    if err := client.Run(); err != nil {
        log.Fatal(err)
    }
}
```

## Step 3: Run the Bot

```bash
export DISCORD_TOKEN='your-bot-token'
export GUILD_ID='your-test-guild-id'
go run main.go
```

You should see:

```
logged in as MyBot
registered 2 slash commands: /hello, /ping
```

## Step 4: Add Options

Add a command with options. Options are passed as variadic
`interactions.ApplicationCommandOption` values:

```go
router.Command("greet", "Greet someone", func(ctx *bot.InteractionContext) {
    name := ctx.GetStringOption("name")
    if name == "" {
        name = "friend"
    }
    ctx.Reply("Hello, " + name + "!")
}, interactions.ApplicationCommandOption{
    Type:        interactions.ApplicationCommandOptionTypeString,
    Name:        "name",
    Description: "Who to greet",
    Required:    false,
})
```

> **Note:** The `AddStringOption`, `AddUserOption`, and other `Add*Option`
> methods exist only on `interactions.SlashCommandBuilder` (the low-level
> builder for direct REST registration). They are **not** available on the
> `*bot.Command` returned by `router.Command`. For the high-level router,
> pass options as variadic arguments.

## Step 5: Add a Button

```go
router.Command("click", "Click the button", func(ctx *bot.InteractionContext) {
    ctx.ReplyWithComponents("Click me!", []components.Component{
        components.ActionRow{
            Components: []components.Component{
                components.Button{
                    CustomID: "click_me",
                    Label:    "Click!",
                    Style:    components.ButtonStylePrimary,
                },
            },
        },
    })
})

// Handle the button click
router.Button("click_me", func(ctx *bot.InteractionContext) {
    ctx.Reply("You clicked the button!")
})
```

## Step 6: Add Error Handling

```go
client.OnError(func(err error) {
    log.Printf("bot error: %v", err)
})
```

## Step 7: Add Graceful Shutdown

`client.Run()` already handles SIGINT and SIGTERM. For custom shutdown:

```go
ctx, cancel := context.WithCancel(context.Background())
go func() {
    // Wait for your shutdown signal
    <-shutdownChan
    cancel()
}()
client.RunContext(ctx)
```

## Next Steps

- Read the [middleware guide](../high-level/middleware-guide) for permission checks
  and cooldowns.
- Read the [components V2 guide](../high-level/components-v2-guide) for modern
  message layouts.
- Read the [error handling guide](../high-level/error-handling) for error
  classification and retry.
- Read the [security guide](../high-level/security) for token handling and
  interaction verification.
