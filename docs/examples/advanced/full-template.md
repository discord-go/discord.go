# Full Template

## Overview

The V2 template is a multi-file application skeleton rather than a single feature demo. It loads application configuration, registers slash and prefix commands, enables mention and bot-name triggers, configures presence, subscribes to READY and error events, and optionally enables sharding. It is a practical starting point for organizing a production bot without dynamic module loading.

## Prerequisites

- Go `1.26.4` or newer.
- `TOKEN` or `DISCORD_TOKEN` set to a bot token.
- A test guild and an installed application.
- The intents requested by the template enabled in the Portal: guilds, members, messages, reactions, direct messages, message content, and voice states.
- Optional configuration from [`examples/v2_template/config.json`](../code/v2_template/config.json), or a custom path in `V2_CONFIG`.

## Architecture

`main.go` loads `templateConfig`, creates a router, applies `GuildOnly`, and delegates command registration to `slash_commands.go` and `message_commands.go`. `config.go` loads JSON with safe defaults and provides `accentColor`. `events.go` contains lifecycle logging. The bot receives all required options through `bot.New`, including prefix, bot name, mention triggers, presence, router, and shard configuration.

## Quick Start

Run from the repository root:

```bash
export DISCORD_TOKEN='replace-with-a-bot-token'
go run ./docs/examples/code/v2_template
```

The loader first checks `config.json` in the current directory and then `docs/examples/code/v2_template/config.json`. Set `V2_CONFIG` to an explicit file when deploying from another working directory.

## Complete Runnable Example

The complete runnable program is the set of linked files, not only `main.go`:

- [`examples/v2_template/main.go`](../code/v2_template/main.go)
- [`examples/v2_template/config.go`](../code/v2_template/config.go)
- [`examples/v2_template/events.go`](../code/v2_template/events.go)
- [`examples/v2_template/slash_commands.go`](../code/v2_template/slash_commands.go)
- [`examples/v2_template/message_commands.go`](../code/v2_template/message_commands.go)
- [`examples/v2_template/config.json`](../code/v2_template/config.json)

Run the complete package with:

```bash
go run ./docs/examples/code/v2_template
```

Every source file has its package declaration and imports; copying only a function from the template is not a runnable example.

## Explanation

The template uses static Go registration so command names, options, component layouts, and middleware remain visible to the compiler and code review. `bot.WithPresence` reapplies the configured presence after a fresh identify. `WithMentionTriggers(true)` enables `<@bot-id>` style triggers, while `WithBotName` enables a configured name trigger. `WithShards(0)` asks Discord for a recommended shard count when sharding is enabled.

The template's broad intents are intentional for its feature catalog, not a universal default. Remove unused intents and privileged Portal settings when reducing the application.

## Basic Usage

- Set `TOKEN` or `DISCORD_TOKEN`.
- Change `prefix` and `botName` in the JSON configuration.
- Add a slash command in `registerSlashCommands`.
- Add a prefix command in `registerMessageCommands`.
- Run `go run ./docs/examples/code/v2_template` from the repository root.

## Intermediate Usage

- Split registrations by feature or domain while keeping one router.
- Add global middleware for guild scope, logging, cooldowns, and shared authorization.
- Use `bot.ConfigFromEnv` or `bot.LoadConfig` when the application needs the framework's generic configuration rather than the template's display configuration.
- Use guild command synchronization while testing new command definitions.
- Add `OnError`, reconnect, resume, and disconnect metrics to the event registration file.

## Advanced Usage

- Supply `bot.WithStore` for application-owned durable state and define migrations before deployment.
- Enable `WithGatewayCompression` only after measuring CPU, memory, and bandwidth tradeoffs.
- Configure shard count deliberately and ensure scheduled work is shard-safe.
- Use an explicit `RunContext` and shutdown deadline under a service manager.
- Keep command registration deterministic across replicas so every instance exposes the same schema.

## Common Patterns

- Keep configuration parsing separate from handler registration.
- Use helper functions for repeated Components V2 containers and status responses.
- Centralize bot permissions in a named constant and apply `RequireBotPermissions` to commands that need it.
- Use `config.accentColor()` only after parsing and validate configuration at startup.
- Keep prefix and mention triggers as compatibility paths while promoting slash commands for discoverability.

## Best Practices

- Do not store the bot token in `config.json`; the template's JSON contains display settings only.
- Pin the working directory or use `V2_CONFIG` so configuration resolution is deterministic in production.
- Reduce the template's privileged intents when the deployed feature set does not need them.
- Do not enable automatic sharding with multiple replicas unless ownership and coordination are designed explicitly.
- Bound REST and database operations, meet interaction deadlines, and make shutdown cancel scheduled jobs and voice sessions.

## Common Mistakes with wrong/correct examples

### Wrong

```go
token := os.Getenv("TOKEN")
if token == "" {
	log.Fatal("missing TOKEN")
}
```

### Correct

```go
token := os.Getenv("TOKEN")
if token == "" {
	token = os.Getenv("DISCORD_TOKEN")
}
if token == "" {
	log.Fatal("TOKEN or DISCORD_TOKEN is required")
}
```

### Wrong

```go
bot.WithIntents(intents.Guilds | intents.GuildMembers | intents.MessageContent)
```

### Correct

```go
bot.WithIntents(intents.Guilds)
```

Use the smallest set that matches the actual handlers. Add privileged intents only when the feature requires them and the Portal has approved them.

### Wrong

```go
router.Command("slow", "Do slow work", func(ctx *bot.InteractionContext) {
	doSlowWork()
	_ = ctx.Reply("done")
})
```

### Correct

```go
router.Command("slow", "Do slow work", func(ctx *bot.InteractionContext) {
	if err := ctx.Defer(); err != nil {
		return
	}
	doSlowWork()
	_, _ = ctx.Followup("done")
})
```

These are excerpts; the linked package is the complete runnable template.

## Expected Result

The template logs READY and errors, registers the Components V2 command catalog, responds to the configured prefix and aliases, recognizes the configured bot name and mentions, applies presence, and starts the configured Gateway lifecycle. Commands are statically discoverable in Go source.

## Related Pages

- [Examples Overview](README.md)
- [Basic Client](../setup/basic-client.md)
- [Slash Commands](../commands/slash-commands.md)
- [Components V2](../interactions/components-v2.md)
- [Moderation](../commands/moderation.md)
- [Complete template source](../code/v2_template/main.go)
