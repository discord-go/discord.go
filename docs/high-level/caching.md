# Caching

## Overview

Caching is optional application infrastructure for avoiding repeated Discord
lookups. Attach a cache with `bot.WithCache`; the bot can hydrate it from gateway
events and the high-level `Cached*` helpers can read typed resources. Matching
`Fetch*` helpers make an authenticated REST request and populate the cache after
a successful response.

## Architecture

The `cache.Cache` interface stores `any` values by string key. Typed subinterfaces
such as `GuildCache`, `MemberCache`, and `MessageCache` add resource-specific
methods. `cache.MemoryCache` implements all built-in typed interfaces and can be
configured with TTL and maximum size options.

The bot performs type assertions before each typed lookup. If no cache is
configured, or the configured implementation does not implement the requested
typed interface, `Cached*` returns `(nil, false)`. Fetch methods always use
`Bot.Rest`; cache hydration is a successful-result side effect, not a replacement
for error handling.

## Quick Start

This complete program uses a bounded memory cache and reports whether the
invoking user's fresh resource was already cached.

```go
package main

import (
	"log"
	"os"
	"time"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/cache"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/snowflake"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}

	router := bot.NewRouter()
	router.Command("user-cache", "Inspect the invoking user cache", func(ctx *bot.InteractionContext) {
		var id snowflake.ID
		if ctx.User != nil {
			id = ctx.User.ID
		} else if ctx.Member != nil && ctx.Member.User != nil {
			id = ctx.Member.User.ID
		}
		if id == 0 {
			_ = ctx.Reply("No invoking user was present")
			return
		}
		if _, ok := ctx.Bot.CachedUser(id); ok {
			_ = ctx.Reply("The user was in cache")
			return
		}
		user, err := ctx.Bot.FetchUser(ctx.Context(), id)
		if err != nil {
			_ = ctx.Reply("Could not fetch the user")
			return
		}
		_ = ctx.Reply("Fetched " + user.Username)
	})

	store := cache.NewMemoryCache(cache.WithTTL(10*time.Minute), cache.WithMaxSize(10000))
	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithCache(store), bot.WithRouter(router))
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Creating/Configuration

`cache.NewMemoryCache(opts ...cache.Option) *cache.MemoryCache` creates the
built-in store. `cache.WithTTL` sets entry expiration and `WithMaxSize` limits
the number of keys; zero means no limit. Call `WithCache(store)` when creating
the bot.

For a database or distributed cache, implement `cache.Cache` plus the typed
interfaces needed by the high-level helpers. Values may be pointers or values;
the bot accepts either when the concrete type matches.

## Using

### Basic: cached read

Call `CachedGuild`, `CachedChannel`, `CachedUser`, `CachedMember`, or
`CachedMessage`. Each returns a typed pointer and a boolean hit indicator.

### Intermediate: cache-aside fetch

Try a cache lookup, then call the matching `Fetch*` method on a miss. Successful
fetches call the typed cache setter automatically.

### Advanced: direct cache maintenance

Use the cache interface's `Delete`, `Clear`, and typed delete methods when an
application mutation makes a value stale. Use `MemoryCache.CleanUp` if a
long-running process wants to remove expired entries proactively.

## Common Patterns

- Use cache keys only through typed methods where possible.
- Treat a cache hit as a snapshot and a miss as normal control flow.
- Fetch after a mutation when the fresh representation is needed immediately.
- Set TTLs for users, channels, and messages whose values change frequently.
- Keep cache size and gateway intents aligned; caching events you did not request
  is impossible.

## Best Practices

### Use cache-aside reads

Why: the application decides when stale data is acceptable.

Pros: simple, explicit consistency policy and fewer REST calls.

Cons: the first request is slower and concurrent misses can stampede the API.

### Bound memory

Why: gateway-driven caches can grow with guild and message volume.

Pros: predictable resource use.

Cons: eviction or expiry causes more REST fetches and can reduce hit rate.

### Invalidate after writes

Why: a cached resource can immediately become outdated after a REST mutation.

Pros: subsequent reads do not return a known stale value.

Cons: manual invalidation is easy to forget; prefer fetching the returned object
when the endpoint provides one.

## Common Mistakes

Incorrect: assuming a cache exists because `CachedUser` has a simple signature.

```go
user, _ := b.CachedUser(id)
fmt.Println(user.Username)
```

Correct: check the boolean and fetch on a miss.

```go
user, ok := b.CachedUser(id)
if !ok {
	var err error
	user, err = b.FetchUser(ctx, id)
	if err != nil {
		return err
	}
}
fmt.Println(user.Username)
```

Incorrect: using `WithMaxSize(0)` as a safety limit.

```go
cache.NewMemoryCache(cache.WithMaxSize(0))
```

Correct: supply an explicit bound for high-volume processes.

```go
cache.NewMemoryCache(cache.WithMaxSize(10000))
```

## API Walkthrough

- `cache.Cache` requires `Get`, `Set`, `Delete`, and `Clear`.
- `GuildCache`, `ChannelCache`, `UserCache`, `RoleCache`, `MessageCache`, and
  `MemberCache` add typed-key methods to `Cache`.
- `Options` has `TTL` and `MaxSize`; `Option` is `func(*Options)`.
- `DefaultOptions`, `WithTTL`, and `WithMaxSize` configure caches.
- `NewMemoryCache(opts ...Option) *MemoryCache` creates a thread-safe in-memory
  cache. `MemoryCache.Get`, `Set`, `Delete`, `Clear`, and `CleanUp` manage it;
  typed `GetGuild`, `SetGuild`, and matching methods exist for every supported
  resource.
- `WithCache(cache.Cache) bot.Option` attaches a cache and enables gateway
  hydration.
- `CachedGuild`, `CachedChannel`, `CachedUser`, `CachedMember`, and
  `CachedMessage` return typed pointers and hit booleans.
- `FetchGuild`, `FetchChannel`, `FetchUser`, `FetchMember`, and `FetchMessage`
  accept `context.Context` and IDs, return typed pointers and errors, and set the
  corresponding cache entry after a successful REST response.

## Examples

- [Moderation](../examples/commands/moderation.md)
- [Basic client](../examples/setup/basic-client.md)
- [Cache low-level guide](../low-level/cache/README.md)

## Related APIs

- [`resources.md`](resources.md) for cache and REST resource helpers.
- [`permissions.md`](permissions.md) for prefix permission checks using members.
- [`../low-level/cache/README.md`](../low-level/cache/README.md) for cache interfaces.
