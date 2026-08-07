# Creating Commands

## Overview

Creating Commands covers slash command definitions, options, handlers, and context-menu commands. discord.go keeps definitions in `bot.Router` and represents option schemas with `interactions.ApplicationCommandOption`. The router validates the schema before automatic synchronization.

## Architecture

The command definition is sent to Discord as an application command. When a user invokes it, Discord sends an interaction containing the command name and option values. The router selects the registered handler, applies middleware, and passes an `InteractionContext`. Option helpers convert values into Go types; they do not replace application validation.

## Prerequisites

- A bot application installed with `applications.commands`.
- `DISCORD_TOKEN` set.
- A test guild where commands can be checked.
- Go `1.26.4` or newer.

## Quick Start

This complete program creates `/greet` with string, integer, and boolean options:

```bash
export DISCORD_TOKEN='replace-with-a-bot-token'
```

```go
package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/interactions"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}

	router := bot.NewRouter()
	router.MustCommand("greet", "Greet a person", func(ctx *bot.InteractionContext) {
		name := strings.TrimSpace(ctx.GetStringOption("name"))
		if name == "" {
			name = "friend"
		}
		count := ctx.GetIntOption("count")
		if count < 1 {
			count = 1
		}
		if count > 3 {
			count = 3
		}
		message := fmt.Sprintf("Hello, %s!", name)
		if ctx.GetBoolOption("shout") {
			message = strings.ToUpper(message)
		}
		for i := int64(1); i < count; i++ {
			message += "\n" + fmt.Sprintf("Hello, %s!", name)
		}
		if err := ctx.Reply(message); err != nil {
			log.Printf("greet reply: %v", err)
		}
	},
		interactions.ApplicationCommandOption{Type: interactions.ApplicationCommandOptionTypeString, Name: "name", Description: "Person to greet"},
		interactions.ApplicationCommandOption{Type: interactions.ApplicationCommandOptionTypeInteger, Name: "count", Description: "Number of greetings"},
		interactions.ApplicationCommandOption{Type: interactions.ApplicationCommandOptionTypeBoolean, Name: "shout", Description: "Use uppercase text"},
	)

	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Creating/Using

Pass a lower-case name, a non-empty description, a handler, and zero or more options to `router.Command`. Set `Required: true` for values Discord must collect. For application-owned definitions that should fail startup on invalid input, use `MustCommand` or check the error from `CommandE`. `UserCommand` and `MessageCommand` create context-menu commands and do not accept options.

## Common Patterns

- Read optional strings with `GetStringOption` and apply a deliberate default.
- Read IDs with `GetUserID`, `GetRoleID`, or `GetChannelID` and verify they are non-zero.
- Clamp numeric options after reading them, even when the UI has limits.
- Use choices for small, fixed sets of string values.
- Use `InCategory` to label commands for help menus and `Cooldown` for per-user throttling.

## Best Practices

- Keep names and descriptions within Discord's limits.
- Validate lengths, ranges, relationships, and authorization in the handler or middleware.
- Never trust the option description or UI constraint as server-side validation.
- Keep command schemas stable; schema changes can take time to reach global commands.
- Use a distinct command name for incompatible behavior instead of silently changing semantics.

## Common Mistakes

### Incorrect

```go
router.Command("Greet", "", handler)
```

### Correct

```go
router.Command("greet", "Greet a person", handler)
```

### Incorrect

```go
count := ctx.GetIntOption("count")
for i := int64(0); i < count; i++ {
	// potentially unbounded user-controlled work
}
```

### Correct

```go
count := ctx.GetIntOption("count")
if count < 1 {
	count = 1
}
if count > 3 {
	count = 3
}
```

The corrected code bounds application work even if a command payload is malformed or changed outside the expected UI.

## API Walkthrough

- `router.Command` registers a chat-input command.
- `router.CommandE` returns validation errors immediately.
- `router.MustCommand` is a startup-time validation helper.
- `interactions.ApplicationCommandOption` describes option type, name, description, and required state.
- `ctx.GetStringOption`, `GetIntOption`, and `GetBoolOption` read typed values.
- `ctx.GetUserID`, `GetRoleID`, and `GetChannelID` parse snowflake-valued options.
- `ctx.CommandName`, `Options`, `Subcommand`, and `TargetID` expose interaction metadata.

## Examples

- [Slash Commands](../commands/slash-commands.md) combines options with permission middleware.
- [Autocomplete](../commands/autocomplete.md) handles focused option queries.
- [Moderation](../commands/moderation.md) validates user targets before REST actions.
- [Context-menu commands](../../low-level/interactions/README.md) explains the underlying interaction model.

## Related Pages

- [Project Setup](project-setup.md)
- [Handling Commands](handling-commands.md)
- [Deploying Commands](deploying-commands.md)
- [Interactions package](../../../interactions/command.go)
