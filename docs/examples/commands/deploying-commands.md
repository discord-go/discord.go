# Deploying Commands

## Overview

Deploying Commands means sending application-command definitions to Discord. discord.go can synchronize a router automatically after `READY`, but a release pipeline may prefer an explicit REST deployment. This page shows guild bulk overwrite, the fast development path, and the distinction between guild and global commands.

## Architecture

Command deployment is a REST operation separate from command handling. `rest.BulkOverwriteGuildCommands` replaces the complete command set for one guild. `rest.BulkOverwriteGlobalCommands` replaces the global set. Automatic router sync performs the same kind of operation after the bot connects; a deployment job can use a REST client without opening a Gateway session.

## Prerequisites

- A bot token with access to the application.
- A valid guild ID for development deployment.
- `applications.commands` installed in the guild.
- Go `1.26.4` or newer.

## Quick Start

Save this complete deployment program as `main.go` in a deployment module. It overwrites the commands in `DISCORD_GUILD_ID`:

```bash
export DISCORD_TOKEN='replace-with-a-bot-token'
export DISCORD_GUILD_ID='123456789012345678'
```

```go
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/rest"
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

	typ := interactions.ApplicationCommandTypeChatInput
	commands := []rest.CreateCommandParams{
		{Name: "ping", Description: "Check whether the bot is online", Type: &typ},
	}
	client := rest.New(token, nil, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	application, err := client.GetCurrentApplication(ctx)
	if err != nil {
		log.Fatalf("get application: %v", err)
	}
	deployed, err := client.BulkOverwriteGuildCommands(ctx, application.ID, guildID, commands)
	if err != nil {
		log.Fatalf("deploy commands: %v", err)
	}
	log.Printf("deployed %d commands to guild %s", len(deployed), guildID)
}
```

## Creating/Using

For a bot process, attach a router and let `bot.WithGuildCommandSync` synchronize during development. For a release job, build a deterministic slice of `rest.CreateCommandParams`, resolve the application ID with `GetCurrentApplication`, and call one bulk-overwrite method. Bulk overwrite removes commands not present in the submitted slice, which makes deletion intentional and reviewable.

Guild commands appear quickly and are ideal for testing. Global commands are visible across installed guilds but changes can take up to an hour to propagate. Do not use a production global application as a test registry.

## Common Patterns

- Use `bot.WithGuildCommandSync` for local development.
- Use `bot.WithCommandSyncDisabled` when an external deploy job owns registration.
- Use `rest.BulkOverwriteGuildCommands` for a complete test-guild registry.
- Use `rest.BulkOverwriteGlobalCommands` for a released application registry.
- Use `GetGuildApplicationCommands` or `GetGlobalApplicationCommands` to inspect deployed state.
- Use a context timeout for every deployment request.

## Best Practices

- Deploy the complete intended set, not a partial list, with bulk overwrite.
- Keep deployment and runtime credentials protected in CI.
- Verify application and guild IDs before making destructive synchronization calls.
- Review command schema changes as carefully as code changes.
- Use separate applications for staging and production.

## Common Mistakes

### Incorrect

```go
client.BulkOverwriteGlobalCommands(context.Background(), appID, commands)
```

### Correct

```go
ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
defer cancel()
_, err := client.BulkOverwriteGlobalCommands(ctx, appID, commands)
if err != nil {
	log.Printf("deploy commands: %v", err)
}
```

### Incorrect

```go
commands := []rest.CreateCommandParams{newCommand}
client.BulkOverwriteGuildCommands(ctx, appID, guildID, commands)
```

### Correct

```go
commands := allCommandsForThisRelease()
deployed, err := client.BulkOverwriteGuildCommands(ctx, appID, guildID, commands)
```

Bulk overwrite treats the slice as the source of truth; omitting a command deletes it from that scope.

## API Walkthrough

- `rest.New` creates a bot-authenticated REST client.
- `GetCurrentApplication` resolves the application ID through `/applications/@me`.
- `rest.CreateCommandParams` is the wire shape for a command definition.
- `BulkOverwriteGuildCommands` replaces a guild command set.
- `BulkOverwriteGlobalCommands` replaces a global command set.
- `GetGuildApplicationCommands` and `GetGlobalApplicationCommands` inspect deployed commands.
- `bot.WithCommandSyncDisabled` prevents automatic synchronization in a runtime bot.
- `bot.WithGuildCommandSync` selects fast automatic guild synchronization.

## Examples

- [Adding Your App](../setup/adding-your-app.md) uses automatic guild sync.
- [Creating Commands](creating-commands.md) builds command options.
- [Slash Commands](../slash-commands.md) explains global propagation.
- [REST command source](../../../rest/commands.go) lists all deployment methods.

## Related Pages

- [Project Setup](project-setup.md)
- [Handling Commands](handling-commands.md)
- [Adding Your App](../setup/adding-your-app.md)
- [REST documentation](../../low-level/rest/README.md)
