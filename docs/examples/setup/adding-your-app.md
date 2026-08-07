# Adding Your App

## Overview

Creating an application does not put it in a guild. This page adapts Discord.js's Adding Your App topic: create an OAuth2 installation URL, select `bot` and `applications.commands`, grant only the permissions needed by the example, and test a guild-scoped command.

## Architecture

OAuth2 installs the application and bot user in a guild. Gateway intents describe which events the bot may receive; the OAuth2 `permissions` value describes what the bot may do in channels. After the bot connects, `bot.WithGuildCommandSync` sends router definitions to the selected guild through REST, avoiding the propagation delay of global commands.

## Prerequisites

- A bot token from [App Setup](app-setup.md).
- The target guild ID, copied with Discord Developer Mode enabled.
- Permission to manage that guild or an administrator who can install the app.
- `Guilds` enabled in the code and in the application's Gateway configuration.

## Quick Start

In the Developer Portal's OAuth2 URL Generator, select the `bot` and `applications.commands` scopes and the `View Channel` and `Send Messages` permissions. Install the generated URL in the test guild. Then save this complete program as `main.go`, replacing `DISCORD_GUILD_ID` with the guild ID in the environment:

```bash
export DISCORD_TOKEN='replace-with-a-bot-token'
export DISCORD_GUILD_ID='123456789012345678'
go run .
```

```go
package main

import (
	"log"
	"os"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/snowflake"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}
	guildID, err := snowflake.Parse(os.Getenv("DISCORD_GUILD_ID"))
	if err != nil || guildID == 0 {
		log.Fatal("DISCORD_GUILD_ID must be a valid snowflake")
	}

	router := bot.NewRouter()
	router.Command("hello", "Say hello in the test guild", func(ctx *bot.InteractionContext) {
		if err := ctx.Reply("The app is installed and working."); err != nil {
			log.Printf("reply: %v", err)
		}
	})

	b := bot.New(token,
		bot.WithIntents(intents.Guilds),
		bot.WithRouter(router),
		bot.WithGuildCommandSync(guildID),
	)
	b.OnReady(func(ctx *bot.ReadyContext) {
		log.Printf("ready as %s; /hello is synced to guild %s", ctx.User.Username, guildID)
	})
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Creating/Using

The generated installation URL is the bridge between the Portal and a guild. `applications.commands` makes slash commands installable, while `bot` adds the bot user. The `permissions` query parameter affects guild permissions, not Gateway intents. Use guild synchronization while developing, then remove `bot.WithGuildCommandSync` for global release synchronization.

## Common Patterns

- Copy IDs, not names; Discord API resources are addressed by snowflakes.
- Validate IDs with `snowflake.Parse` before using them.
- Use guild command sync for fast iteration and global sync for released commands.
- Grant `Embed Links`, `Read Message History`, or moderation permissions only when a command needs them.
- Apply `bot.RequireBotPermissions` to commands that need action-specific permissions.

## Best Practices

- Install the app only in test guilds until command schemas are stable.
- Do not use administrator permission as a shortcut for missing design decisions.
- Keep the guild ID in deployment configuration, not in a source constant.
- Remember that a bot can receive an interaction and still lack permission to complete the requested REST action.

## Common Mistakes

### Incorrect

```go
bot.WithGuildCommandSync(snowflake.ID(0))
```

### Correct

```go
if err != nil || guildID == 0 {
	log.Fatal("DISCORD_GUILD_ID must be a valid snowflake")
}
bot.WithGuildCommandSync(guildID)
```

### Incorrect

```text
Scopes: bot
Permissions: Administrator
```

### Correct

```text
Scopes: bot, applications.commands
Permissions: View Channel, Send Messages
```

The correct installation includes the command scope and starts with least privilege.

## API Walkthrough

- `snowflake.Parse` converts a Discord ID string into `snowflake.ID`.
- `bot.WithGuildCommandSync` configures fast guild-scoped synchronization.
- `bot.WithRouter` connects command definitions to the bot lifecycle.
- `router.Command` defines the command sent during synchronization.
- `bot.GuildOnly` and permission middleware enforce runtime scope separately from installation permissions.

## Examples

- [Slash Commands](../commands/slash-commands.md) demonstrates guild-only and permission middleware.
- [Moderation](../commands/moderation.md) maps permissions to REST actions.
- [Deploying Commands](../commands/deploying-commands.md) shows explicit REST synchronization.

## Related Pages

- [App Setup](app-setup.md)
- [Installation](installation.md)
- [Project Setup](../commands/project-setup.md)
- [Discord OAuth2 documentation](https://discord.com/developers/docs/topics/oauth2)
