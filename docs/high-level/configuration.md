# Configuration

## Overview

The `bot` package supports application-owned startup configuration from a JSON
file or conventional environment variables. `Config` contains the token and
basic gateway and command settings; functional options remain available for
transports, logging, caches, stores, and test infrastructure.

## Architecture

`LoadConfig(path)` reads and unmarshals one JSON file. `ConfigFromEnv()` reads
environment variables and returns a value without an error channel, so malformed
optional values are ignored and remain at their zero value. `NewFromConfig` maps
non-zero configuration fields to `bot.Option` values and calls `New`.

The token is application secret material. Configuration files should be local or
secret-managed, and environment values should not be printed in logs.

## Quick Start

This complete program uses environment configuration and adds a router option.

```go
package main

import (
	"log"

	"github.com/discord-go/discord.go/bot"
)

func main() {
	config := bot.ConfigFromEnv()
	if config.Token == "" {
		log.Fatal("set DISCORD_TOKEN or TOKEN")
	}
	router := bot.NewRouter()
	router.Command("ping", "Check configuration", func(ctx *bot.InteractionContext) {
		if err := ctx.Reply("Configured"); err != nil {
			log.Printf("reply: %v", err)
		}
	})
	b := bot.NewFromConfig(config, bot.WithRouter(router))
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

Run with `DISCORD_TOKEN=... BOT_PREFIX=! go run .`. An unset `Intents`
keeps the constructor defaults (`bot.DefaultIntents`); a non-zero value
replaces them entirely, and `NewFromConfig` logs a warning when the
replacement drops a privileged intent such as `MessageContent`.

## Creating/Configuration

`bot.Config` fields are `Token`, `Prefix`, `BotName`, `MentionTriggers`,
`Intents`, `Shards`, `AutomaticShards`, `Compression`, `Presence`, and
`CommandSync`.

JSON uses `token`, `prefix`, `bot_name`, `mention_triggers`, `intents`,
`shards`, `automatic_shards`, `compression`, `presence`, and `command_sync`.
`CommandSyncConfig.Mode` is numeric JSON because it is a Go integer enum; use
`0` for global, `1` for guild, and `2` for disabled.

Environment names are `TOKEN` or `DISCORD_TOKEN`, `BOT_PREFIX`, `BOT_NAME`,
`BOT_MENTION_TRIGGERS`, `BOT_INTENTS`, `BOT_SHARDS`, `BOT_AUTOMATIC_SHARDS`, and
`BOT_GATEWAY_COMPRESSION`. Boolean values use Go's `strconv.ParseBool` forms;
intents are decimal bitfields.

## Using

### Basic: environment

Call `ConfigFromEnv`, check `Token`, optionally apply programmatic defaults, and
pass the result to `NewFromConfig`.

### Intermediate: JSON

Write a JSON file without secrets in source control, call `LoadConfig`, check the
returned error, and pass the result to `NewFromConfig`.

### Advanced: layered configuration

Load a base JSON config, then override fields from a deployment environment or
append functional options for infrastructure concerns. Avoid silently replacing
a secure token with an empty environment variable.

## Intents

`bot.New` starts from `bot.DefaultIntents()`: `Guilds | GuildMessages |
MessageContent`. Two options control the final set:

- `WithIntents(intents.Intent)` adds to the defaults. Calling it multiple
  times unions the sets, so a bot that requests more events never loses the
  defaults by accident.
- `WithIntentsExclusive(intents.Intent)` replaces the defaults entirely. Use
  it to opt out of `MessageContent` or to pin an exact set. Every privileged
  intent the replacement drops (`GuildMembers`, `GuildPresences`,
  `MessageContent`) is logged, because the gateway then stops delivering its
  events without returning an error; message content arrives empty without
  `MessageContent`.

A non-zero `Config.Intents` (JSON `intents`, env `BOT_INTENTS`) replaces the
defaults the same way `WithIntentsExclusive` does.

```go
// Defaults plus voice tracking and member events.
b := bot.New(token, bot.WithIntents(intents.GuildVoiceStates|intents.GuildMembers))

// Exactly these intents, nothing else (logs when MessageContent is lost).
b = bot.New(token, bot.WithIntentsExclusive(intents.Guilds))
```

## Common Patterns

- Validate the token explicitly before `NewFromConfig` or `Run`. `Start`
  returns `ErrInvalidToken` if the token does not have three dot-separated
  segments.
- Use guild command sync in a development config and global sync in production.
- Keep intent declarations close to the feature that needs them; extend the
  defaults with `WithIntents` and replace them only with
  `WithIntentsExclusive`.
- Use `WithRESTClient`, `WithLogger`, and `WithErrorHandler` as options after
  loading application config.
- Document which configuration source wins when multiple sources are used.

## Best Practices

### Keep secrets out of JSON committed to Git

Why: repository history is difficult to clean after a token leak.

Pros: safer review and simpler rotation.

Cons: deployment requires a secret manager or environment injection.

### Validate configuration at startup

Why: `ConfigFromEnv` intentionally ignores malformed optional values.

Pros: failures happen before a partial gateway run.

Cons: the application owns validation rules for IDs, intents, and deployment
policy.

### Use explicit environment names

Why: `ConfigFromEnv` checks both `TOKEN` and `DISCORD_TOKEN` and can hide a typo
if an old variable remains present.

Pros: predictable deployment behavior.

Cons: a small amount of environment documentation is required.

## Common Mistakes

Incorrect: assuming `BOT_INTENTS` accepts names.

```sh
BOT_INTENTS=Guilds go run .
```

Correct: provide the decimal bitfield or set intents in Go.

```sh
BOT_INTENTS=513 go run .
```

Incorrect: ignoring a JSON load error.

```go
config, _ := bot.LoadConfig("bot.json")
```

Correct:

```go
config, err := bot.LoadConfig("bot.json")
if err != nil {
	return err
}
```

## API Walkthrough

- `Config` is the JSON/environment value object with token, trigger, intent,
  shard, compression, presence, and command-sync fields.
- `LoadConfig(path string) (Config, error)` reads a JSON file.
- `ConfigFromEnv() Config` reads the documented environment variables and returns
  zero values for missing or unparsable optional fields.
- `NewFromConfig(config Config, opts ...Option) *Bot` converts configuration to
  a bot and applies additional options.
- `CommandSyncConfig` contains `Mode`, `GuildID`, and `Timeout`.
- `CommandSyncGlobal`, `CommandSyncGuild`, and `CommandSyncDisabled` select
  automatic global, guild, or no synchronization.
- `WithPrefix`, `WithBotName`, `WithMentionTriggers`, `WithShards`,
  `WithGatewayCompression`, `WithPresence`, and `WithCommandSync` are the
  equivalent direct options. `WithIntents` adds to the default intents and
  `WithIntentsExclusive` replaces them (see [Intents](#intents)).

## Examples

- [Basic client](../examples/setup/basic-client.md)
- [Full template](../examples/advanced/full-template.md)
- [Configuration low-level context](../README.md)

## Related APIs

- [`client.md`](client.md) for functional options.
- [`commands.md`](commands.md) for command sync behavior.
- [`presence.md`](presence.md) for presence configuration.
