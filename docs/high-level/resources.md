# Resource Access

## Overview

The `Bot` resource helpers provide a consistent cache-first or REST-backed way
to retrieve common Discord objects: guilds, channels, users, guild members, and
messages. They are convenience methods, not a replacement for the complete REST
surface exposed by `Bot.Rest`.

## Architecture

`Cached*` methods perform local typed lookups and never make a network request.
`Fetch*` methods call the corresponding REST endpoint with the supplied context,
return the typed object, and hydrate the configured typed cache on success. The
bot returns copies for the current bot user through `User`, but fetched resource
objects are the values decoded by REST.

## Quick Start

This complete program fetches the invoking user and replies with its username.

```go
package main

import (
	"log"
	"os"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/snowflake"
)

func userID(ctx *bot.InteractionContext) snowflake.ID {
	if ctx.User != nil {
		return ctx.User.ID
	}
	if ctx.Member != nil && ctx.Member.User != nil {
		return ctx.Member.User.ID
	}
	return 0
}

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}
	router := bot.NewRouter()
	router.Command("whoami", "Fetch the invoking user", func(ctx *bot.InteractionContext) {
		id := userID(ctx)
		if id == 0 {
			_ = ctx.Reply("No user ID was present")
			return
		}
		user, err := ctx.Bot.FetchUser(ctx.Context(), id)
		if err != nil {
			log.Printf("fetch user: %v", err)
			_ = ctx.Reply("The user could not be fetched")
			return
		}
		_ = ctx.Reply("You are " + user.Username)
	})
	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

For normal application code, attach a cache as shown in

## Creating/Configuration

Configure a cache with `WithCache` if local lookups are useful. The helpers use
the bot's `Rest` client, so `WithRESTClient` can provide a custom HTTP client or
rate-limit store. All fetch methods accept a `context.Context`; use a deadline
for command work that depends on external REST calls.

## Using

### Basic: fetch one resource

Call `FetchGuild`, `FetchChannel`, `FetchUser`, `FetchMember`, or `FetchMessage`
with the IDs required by the REST endpoint. Check the returned error before
reading the pointer.

### Intermediate: cache-first

Try the matching `Cached*` method first, then fetch on a miss. A successful fetch
will populate the configured typed cache.

### Advanced: use `Bot.Rest`

When the helper does not cover an operation, use `Bot.Rest` directly. Examples
include listing messages, editing channels, managing roles, fetching audit logs,
or calling application command endpoints. Use the [REST endpoint guide](../low-level/rest/endpoints.md)
to choose the exact signature.

## Common Patterns

- Use `ctx.Context()` from an event context so cancellation follows the bot run.
- Pass snowflake IDs rather than string paths to high-level helpers.
- Fetch a message by both `channelID` and `messageID`.
- Use `WithReason` on REST contexts for supported audit-log operations.
- Cache only objects the application can usefully reuse.

## Best Practices

### Use the narrowest helper

Why: typed helpers make required IDs and returned resources explicit.

Pros: less endpoint construction and fewer serialization mistakes.

Cons: a new Discord endpoint may not have a high-level wrapper yet.

### Use context deadlines

Why: REST calls can stall independently of gateway dispatch.

Pros: handlers remain responsive and shutdown can complete.

Cons: a short deadline can turn a temporarily slow API into a user-visible
failure.

### Distinguish cache freshness from REST success

Why: a cache hit may be stale while a fetch failure does not make the old value
invalid automatically.

Pros: consistency decisions remain in application code.

Cons: callers must define when stale data is acceptable.

## Common Mistakes

Incorrect: passing a message ID where a channel ID is required.

```go
message, _ := b.FetchMessage(ctx, messageID, channelID)
```

Correct:

```go
message, err := b.FetchMessage(ctx, channelID, messageID)
if err != nil {
	return err
}
```

Incorrect: dereferencing a failed lookup.

```go
guild, _ := b.CachedGuild(id)
fmt.Println(guild.Name)
```

Correct:

```go
guild, ok := b.CachedGuild(id)
if !ok {
	guild, err = b.FetchGuild(ctx, id)
	if err != nil {
		return err
	}
}
fmt.Println(guild.Name)
```

## API Walkthrough

- `CachedGuild(snowflake.ID) (*guilds.Guild, bool)` and `FetchGuild(context.Context,
  snowflake.ID) (*guilds.Guild, error)` access guilds.
- `CachedChannel(snowflake.ID) (*channels.Channel, bool)` and `FetchChannel`
  access channels.
- `CachedUser(snowflake.ID) (*users.User, bool)` and `FetchUser` access users.
- `CachedMember(guildID, userID snowflake.ID) (*users.Member, bool)` and
  `FetchMember(context.Context, guildID, userID snowflake.ID) (*users.Member, error)`
  access guild members.
- `CachedMessage(snowflake.ID) (*messages.Message, bool)` and
  `FetchMessage(context.Context, channelID, messageID snowflake.ID) (*messages.Message, error)`
  access messages.
- `User() *users.User`, `AppID() snowflake.ID`, `ReadyAt() time.Time`, and
  `Uptime() time.Duration` expose bot identity and readiness state.
- `Bot.Rest *rest.Client` exposes direct REST methods such as
  `GetChannelMessage`, `GetGuild`, `GetUser`, `GetGuildMember`, and the many
  channel, guild, command, webhook, and interaction operations.
- `rest.WithReason(context.Context, string) context.Context` attaches an audit
  log reason; `ReasonFromContext` retrieves it.

## Examples

- [Basic client](../examples/setup/basic-client.md)
- [Moderation](../examples/commands/moderation.md)
- [REST endpoints](../low-level/rest/endpoints.md)

## Related APIs

- [`caching.md`](caching.md) for cache implementations and TTLs.
- [`client.md`](client.md) for `Bot.Rest` configuration.
- [`../low-level/models/README.md`](../low-level/models/README.md) for resource models.
