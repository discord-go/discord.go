# Examples

## Overview

This directory is a set of tutorial-style paths through the current `discord.go` APIs. Each guide explains a complete bot shape, identifies the required Gateway intents and Discord permissions, and points to the matching repository source when an example exists under `examples/`.

The examples are deliberately small enough to run from a checkout, but the guides also call out the parts that need production hardening: secret management, REST timeouts, interaction response deadlines, command synchronization, authorization, and shutdown.

## Prerequisites

- Go `1.26.4` or a compatible newer Go toolchain, as declared by [`go.mod`](../../go.mod).
- A Discord application with a bot user and a token. Keep the token outside source control.
- A test guild where you can install the application and inspect commands.
- Gateway intents enabled in the Developer Portal for the intents selected by the example.
- The bot permissions required by the page. Start with the smallest permission set that works.

## Architecture

An example normally has four layers:

- `bot.Bot` owns the Gateway connection, event dispatch, REST client, lifecycle, and shutdown.
- `bot.Router` registers slash commands, prefix commands, middleware, and component routes.
- `bot.InteractionContext` acknowledges interactions and provides typed option, component, modal, and follow-up helpers.
- Package builders such as `components`, `messages`, and `rest` create the payloads sent through those contexts.

Discord sends events over the Gateway. The bot dispatches them to typed handlers or to `EventContext` for events that do not have a typed helper. REST calls are made through the bot's REST client and should always have a bounded context when application code performs a potentially slow operation.

## Quick Start

Run commands from the repository root:

```bash
export DISCORD_TOKEN='replace-with-a-bot-token'
go run ./docs/examples/code/ping
```

The first run of a slash-command example may register commands globally. Global command changes can take a while to appear. Use a test guild and `bot.WithGuildCommandSync` in application code when developing fast-changing commands.

## Complete Runnable Example

[`examples/ping/main.go`](code/ping/main.go) is the complete source for the smallest source-backed example. It includes `package main`, all imports, token loading, intents, slash and prefix routes, a direct message handler, a READY handler, and `b.Run()`.

Run it with:

```bash
go run ./docs/examples/code/ping
```

## Explanation

The source examples are ordinary Go programs, not code fragments loaded by a framework. Read the linked file before copying a small portion: imports, error handling, the selected intents, and the lifecycle call are part of a runnable program. The pages below use the same arrangement while focusing on one feature.

## Basic Usage

- Start with [Basic Client](setup/basic-client.md) for a connection, READY event, and simple replies.
- Use [Slash Commands](commands/slash-commands.md) for command options, middleware, deferrals, and follow-ups.
- Use [Gateway](more-to-know/gateway.md) when a typed high-level handler is not enough.

## Intermediate Usage

- Use [Buttons](interactions/buttons.md), [Modals](interactions/modals.md), and [Autocomplete](commands/autocomplete.md) for interaction-driven workflows.
- Use [Collectors](more-to-know/collectors.md) for one-shot, scoped waits with cancellation.
- Use [Components V2](interactions/components-v2.md) for typed message layouts, files, and select menus.

## Advanced Usage

- Use [Moderation](commands/moderation.md) for permission middleware, audit-log reasons, REST operations, and deferred responses.
- Use [Voice](voice/README.md) for the main-Gateway voice state flow and the separate voice transport.
- Use [Full Template](advanced/full-template.md) for configuration, presence, triggers, command organization, and sharding options.

## Common Patterns

- Read `TOKEN` or `DISCORD_TOKEN` from the environment and fail closed when it is missing.
- Acknowledge an interaction immediately with `Reply`, `ReplyEphemeral`, `Defer`, `Update`, or `ShowModalBuilder`.
- Attach `bot.GuildOnly`, `bot.RequirePermissions`, or `bot.RequireBotPermissions` before business logic.
- Use custom IDs as stable routing keys and include a resource identifier only after validating it.
- Use `context.WithTimeout` for collectors and REST work; call every returned cancel function.
- Call `Stop` or `RunContext` during shutdown so Gateway connections, collectors, scheduled jobs, and handlers can finish cleanly.

## Best Practices

- Never log, commit, or put a token in a command line that can be captured by shell history.
- Enable only the intents an example actually consumes, and request privileged intents explicitly in the Portal.
- Prefer guild command synchronization during development and global synchronization for released commands.
- Treat all user, option, custom ID, and modal values as untrusted input.
- Log errors from replies and REST calls, but redact tokens and sensitive user-provided content.
- Give external REST calls a deadline shorter than the surrounding operation and use `rest.WithReason` for moderation actions.
- Make cleanup idempotent: unsubscribe handlers, cancel collectors, disconnect voice sessions, and stop the bot exactly once.

## Common Mistakes with wrong/correct examples

### Wrong

```go
const token = "paste-a-real-token-here"
```

### Correct

```go
token := os.Getenv("DISCORD_TOKEN")
if token == "" {
    log.Fatal("DISCORD_TOKEN is required")
}
```

### Wrong

```go
router.Button("approve", func(ctx *bot.InteractionContext) {
    time.Sleep(5 * time.Second)
    _ = ctx.UpdateContent("Approved")
})
```

### Correct

```go
router.Button("approve", func(ctx *bot.InteractionContext) {
    if err := ctx.DeferUpdate(); err != nil {
        return
    }
    // Do slow work, then edit the original response.
    _, _ = ctx.EditReply("Approved")
})
```

The second fragment illustrates the rule, but a runnable program still needs imports and a `main` function. The feature pages provide those complete programs or link to the repository source.

## Expected Result

Every guide should produce a bot that connects, reports READY, handles the documented test command or event, and exits without leaking a Gateway or voice connection when the process receives SIGINT or SIGTERM.

## Guide Sections

- [`setup/`](setup/README.md) covers installation, project layout, the application setup, and the main file.
- [`commands/`](commands/README.md) mirrors the command-building path from first command through options, permissions, cooldowns, responses, context menus, and deployment.
- [`interactions/`](interactions/README.md) covers buttons, action rows, select menus, modals, interactions, and Components V2.
- [`more-to-know/`](more-to-know/README.md) covers audit logs, collectors, formatting, intents, embeds, cache/partials, permissions, reactions, threads, webhooks, and image/canvas alternatives.
- [`persistence/`](persistence/README.md) explains how Keyv- and Sequelize-style application persistence maps to `storage.Store`.
- [`advanced/`](advanced/README.md) covers OAuth2 and sharding.

## Runnable Source Layout

- [`code/`](code/) contains the consolidated runnable Go examples that used to
  live in the separate root `example` and `examples` directories.
- [`templates/`](templates/) contains larger audit and music bot templates with
  their own Go modules and dependencies.

The documentation pages explain the source examples; the source trees are kept
under this single documentation-owned location so there is no ambiguity about
which example directory is current.

## Related Pages

- [Basic Client](setup/basic-client.md)
- [Gateway](more-to-know/gateway.md)
- [Slash Commands](commands/slash-commands.md)
- [Buttons](interactions/buttons.md)
- [Modals](interactions/modals.md)
- [Autocomplete](commands/autocomplete.md)
- [Collectors](more-to-know/collectors.md)
- [Components V2](interactions/components-v2.md)
- [Moderation](commands/moderation.md)
- [Voice](voice/README.md)
- [Full Template](advanced/full-template.md)
