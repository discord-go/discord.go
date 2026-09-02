# Commands And Routing

## Overview

`bot.Router` turns Discord commands into application handlers. It supports chat
input slash commands, user and message context-menu commands, prefix commands,
aliases, global and local middleware, and interaction routes for buttons,
selects, modals, and autocomplete.

## Architecture

The router stores commands by normalized name. When an interaction arrives, it
first identifies the command or custom-ID route, then wraps the handler with
command middleware and global middleware. Prefix messages are tokenized with
quoted values preserved, validated, and passed to a `PrefixHandler`.

Attaching a router with `bot.WithRouter` enables automatic command sync on READY.
The default is global sync; `bot.WithGuildCommandSync` is faster for development.
The router only registers definitions. Discord still decides whether an
application is installed and whether a user can see a command.

## Quick Start

This program registers a slash command with a string option and a prefix alias.

```go
package main

import (
	"log"
	"os"

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
	router.Command("hello", "Greet a person", func(ctx *bot.InteractionContext) {
		name := ctx.GetStringOption("name")
		if name == "" {
			name = "friend"
		}
		if err := ctx.Reply("Hello, " + name); err != nil {
			log.Printf("reply: %v", err)
		}
	}, interactions.ApplicationCommandOption{
		Type:        interactions.ApplicationCommandOptionTypeString,
		Name:        "name",
		Description: "Name to greet",
	})
	router.Prefix("hello", func(ctx *bot.MessageContext, args []string) {
		_, err := ctx.Reply("Hello from a prefix command")
		if err != nil {
			log.Printf("reply: %v", err)
		}
	}).Aliases("hi")

	b := bot.New(token,
		bot.WithIntents(intents.Guilds|intents.GuildMessages|intents.MessageContent),
		bot.WithRouter(router),
	)
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

Prefix dispatch requires `GuildMessages` and `MessageContent`, plus the
corresponding portal configuration. Use a test guild while developing command
definitions.

## Creating/Configuration

Create a router with `bot.NewRouter() *bot.Router`. Register chat input commands
with `Command(name, description string, handler CommandHandler, opts ...interactions.ApplicationCommandOption)`.
Use `ContextCommand`, `UserCommand`, or `MessageCommand` for context menus.
Use `Prefix(name string, handler PrefixHandler)` for text commands.

`CommandE` returns `(*Command, error)` and validates immediately. `MustCommand`
panics on invalid static setup. `Router.Validate()` can validate the whole
registry before startup. `WithCommandSync` controls whether READY automatically
calls `BulkOverwriteGlobalCommands` or `BulkOverwriteGuildCommands`.

### Subcommands (chat input)

Nested `options` on an option of type `SUB_COMMAND` (or `SUB_COMMAND_GROUP`)
declare a subcommand tree. Discord then treats the subcommand options as
required only for the chosen subcommand — no more flat `action:create`
workarounds.

```go
router.Command("giveaway", "Server giveaways", func(ctx *bot.InteractionContext) {
	switch ctx.Subcommand() {
	case "create":
		prize := ctx.GetStringOption("prize") // nested options resolve automatically
		_ = prize
	}
}, interactions.ApplicationCommandOption{
	Type:        interactions.ApplicationCommandOptionTypeSubCommand,
	Name:        "create",
	Description: "Start a new giveaway",
	Options: []interactions.ApplicationCommandOption{
		{Type: interactions.ApplicationCommandOptionTypeString, Name: "prize", Description: "The prize", Required: true},
		{Type: interactions.ApplicationCommandOptionTypeInteger, Name: "winners", Description: "Winner count", Required: true},
	},
}, interactions.ApplicationCommandOption{
	Type:        interactions.ApplicationCommandOptionTypeSubCommand,
	Name:        "end",
	Description: "End a giveaway",
	Options: []interactions.ApplicationCommandOption{
		{Type: interactions.ApplicationCommandOptionTypeString, Name: "id", Description: "Giveaway ID", Required: true},
	},
})
```

Rules enforced by the router's validator:

* subcommands (and groups) must provide nested options and are never required;
* subcommand groups may only contain subcommands (one level of grouping);
* subcommands cannot nest inside other subcommands;
* leaf options may not carry nested options.

On the handler side, `ctx.Subcommand()` returns the selected subcommand name and
the typed getters (`ctx.GetStringOption`, `ctx.GetUserID`, …) search nested
subcommand options, so each subcommand can carry equally named options.

The `SlashCommandBuilder` equivalents are `AddSubcommand` and
`AddSubcommandGroup`.

## Using

### Basic: slash and context commands

Options are `interactions.ApplicationCommandOption` values. Names and
descriptions must satisfy Discord's limits. In a context command, use
`ctx.TargetID()` to get the selected user or message ID.

### Intermediate: middleware and prefix commands

Call `router.Use(middleware)` for every slash command or `command.Use(...)` for
one command. Prefix commands have their own `PrefixMiddleware` type. Configure
`MinArgs`, `Usage`, `Validate`, `Description`, and `Aliases` fluently.

### Advanced: routes and registry introspection

Use `Button`, `ButtonPrefix`, `Select`, `SelectPrefix`, `Modal`, and
`Autocomplete` for non-command interactions. `Commands`, `Lookup`,
`CommandCount`, `HasCommand`, and `RangeCommands` support help output and
startup checks. `RemoveCommand` and `RemovePrefix` support dynamic registries.

## Common Patterns

- Use `CommandE` in tests to catch duplicate and malformed definitions.
- Prefer a stable custom-ID prefix for routes that contain IDs.
- Use guild sync for development and disable automatic sync when another deploy
  system owns command registration.
- Keep handlers thin and move business logic into functions that accept a
  context and return an error.
- Use `Cooldown`, `GuildOnly`, and permission middleware before expensive work.

## Best Practices

### Validate at registration time

Why: malformed names and options otherwise fail during READY synchronization.

Pros: failures are local, deterministic, and easy to test.

Cons: `Command` itself replaces duplicates and reports them through the bot
error handler, so `CommandE` is preferable when setup must fail immediately.

### Separate command types

Why: prefix handlers receive message arguments while interaction handlers receive
structured options and a three-second initial response window.

Pros: each handler has the correct data and response semantics.

Cons: shared behavior needs a small application service rather than one handler
being reused blindly.

### Make custom IDs data-safe

Why: component routes match exact IDs or prefixes, not arbitrary application
state.

Pros: routes are simple and stateless.

Cons: long or user-controlled IDs can exceed Discord limits or create collisions;
validate and encode values before constructing them.

## Common Mistakes

Incorrect: using an uppercase slash command name.

```go
router.Command("Hello World", "Greeting", handler)
```

Correct: use lowercase letters, digits, hyphens, or underscores and a valid
description.

```go
router.Command("hello-world", "Send a greeting", handler)
```

Incorrect: assuming a prefix command receives quoted words separately.

```go
// !say "hello world" is expected to produce two arguments.
```

Correct: quoted values remain one argument.

```go
router.Prefix("say", func(ctx *bot.MessageContext, args []string) {
	// !say "hello world" produces []string{"hello world"}.
})
```

## API Walkthrough

- `CommandHandler` is `func(*InteractionContext)`; `PrefixHandler` is
  `func(*MessageContext, []string)`.
- `Middleware`, `PrefixMiddleware`, and `PrefixValidation` wrap or validate
  handlers.
- `Router.Use`, `Command.Use`, and `InteractionRoute.Use` add middleware.
- `Command` has `Name`, `Description`, `Type`, `Options`, `Handler`, and
  `Category`; `InCategory`, `Cooldown`, `RequirePermissions`, and
  `RequireBotPermissions` return the command for chaining.
- `PrefixCommand.Use`, `Description`, `Usage`, `MinArgs`, `Validate`, and
  `Aliases` configure text commands.
- `Command`, `CommandE`, `MustCommand`, `ContextCommand`, `UserCommand`,
  `MessageCommand`, `Prefix`, `RemoveCommand`, `RemovePrefix`, `Lookup`,
  `HasCommand`, `CommandCount`, `Commands`, `RangeCommands`, and `Validate` are
  the command registry methods.
- `Button`, `ButtonPrefix`, `Select`, `SelectPrefix`, `Modal`, and
  `Autocomplete` register `InteractionRoute` values. `InteractionRoute.ID` and
  `Handler` identify the route.
- `interactions.ApplicationCommandOption` describes `Type`, `Name`,
  `Description`, `Required`, `Choices`, `Autocomplete`, value bounds,
  localization, string lengths, and channel types.
- Option types include `SubCommand`, `SubCommandGroup`, `String`, `Integer`,
  `Boolean`, `User`, `Channel`, `Role`, `Mentionable`, `Number`, and
  `Attachment`.
- `interactions.NewSlashCommandBuilder` and its `SetName`, `SetDescription`,
  `AddStringOption`, `AddStringOptionWithChoices`, `AddIntegerOption`,
  `AddBooleanOption`, `AddUserOption`, `AddChannelOption`, `AddRoleOption`,
  `AddMentionableOption`, `AddOption`, `SetIntegrationTypes`, `SetContexts`,
  and `Build` methods create a low-level command value for direct REST use.
  These `Add*Option` methods are **not** available on the `*bot.Command`
  returned by `router.Command`; for the high-level router, pass
  `interactions.ApplicationCommandOption` values as variadic arguments.

## Examples

- [Slash commands](../examples/commands/slash-commands.md)
- [Autocomplete](../examples/commands/autocomplete.md)
- [Moderation](../examples/commands/moderation.md)

## Related APIs

- [`interactions.md`](interactions.md) for reading options and responding.
- [`buttons.md`](buttons.md) and [`modals.md`](modals.md) for component routes.
- [`permissions.md`](permissions.md) for command middleware.
- [`../low-level/rest/endpoints.md`](../low-level/rest/endpoints.md) for manual command sync.
