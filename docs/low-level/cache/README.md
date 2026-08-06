# Cache

## Overview

The `cache` package provides a deliberately small `Cache` interface and a
concurrency-safe `MemoryCache`. Values are `any`, so the cache does not impose
a model package or copy values on the way in or out. The typed interfaces
(`GuildCache`, `ChannelCache`, `UserCache`, `RoleCache`, `MessageCache`, and
`MemberCache`) are optional contracts for adapters that want named methods.

## Architecture

The generic methods use the key supplied by the caller. `MemoryCache`'s typed
methods namespace keys as `guild:`, `channel:`, `user:`, `role:`, `message:`,
and `member:guildID:userID`. This means a generic `Get("guild:42")` can see a
value written by `SetGuild("42", value)`. Entries are guarded by a mutex and
are safe for concurrent readers and writers.

## Quick Start

`NewMemoryCache(opts ...Option)` starts with `DefaultOptions`: `TTL == 0`
(entries do not expire) and `MaxSize == 0` (unlimited). `WithTTL` and
`WithMaxSize` mutate those options. A positive TTL is evaluated on `Get` and
by `CleanUp`; there is no background janitor goroutine.

```go
package main

import (
	"fmt"
	"time"

	"github.com/discord-go/discord.go/cache"
)

func main() {
	c := cache.NewMemoryCache(cache.WithTTL(20*time.Millisecond), cache.WithMaxSize(2))
	c.SetUser("100", "Ada")
	value, ok := c.GetUser("100")
	fmt.Println(value, ok)
	time.Sleep(30 * time.Millisecond)
	_, ok = c.GetUser("100")
	fmt.Println(ok)
}
```

## Creating A Cache

Use `NewMemoryCache` rather than a zero `MemoryCache`; the constructor
initializes the map and applies options. `Option` is `func(*Options)`. The
exported `Options` fields are `TTL` and `MaxSize`, and `DefaultOptions` returns
a fresh default value. A negative TTL behaves like no expiration because only
positive values create expiration timestamps. A negative max size also behaves
as unlimited because the size check is only made for values greater than zero.

## Using Typed Entries

`Get`, `Set`, `Delete`, and `Clear` implement `Cache`. `CleanUp` removes all
expired entries. Each typed interface embeds `Cache` and adds its three
operations. `Set` replaces an existing value. If `MaxSize` is reached and the
key is new, `MemoryCache` deletes one arbitrary map entry before inserting; it
is not LRU and callers must not depend on which key is evicted.

## Common Patterns

Store pointers when the application wants later mutations to be visible, or
store immutable snapshots when event handlers should not race with consumers.
Use IDs converted with `snowflake.ID.String()` as stable keys. Call `CleanUp`
from an application maintenance loop for long-lived caches with TTLs.

## Best Practices

Treat a cache miss as normal and fall back to REST. Hydrate only from events
your Gateway intents actually deliver. Bound `MaxSize` when values can grow
without limit, and use a durable or distributed implementation for state that
must survive restarts. The interface stores `any`, so document the concrete
type associated with each key namespace.

## Common Mistakes

Do not assume expiration is proactive: an expired item can remain in the map
until accessed or cleaned. Do not use the generic key space without a naming
convention. Do not mistake `MemoryCache` for a persistent store; use
[`../storage/`](../storage/README.md) for JSON-oriented application state.

## API Walkthrough

The exported constructors and options are `NewMemoryCache`, `DefaultOptions`,
`WithTTL`, and `WithMaxSize`. The contracts are `Cache`, `GuildCache`,
`ChannelCache`, `UserCache`, `RoleCache`, `MessageCache`, and `MemberCache`.
`MemoryCache` implements every contract and exposes `Get*`, `Set*`, and
`Delete*` methods for all six entity kinds, plus `CleanUp` and `Clear`.

## Examples

The Quick Start program is runnable as-is. Replace the string value with a
`users.User`, `messages.Message`, or other model when integrating with Gateway
hydration.

## Related APIs

- [`../storage/`](../storage/README.md)
- [`../gateway/`](../gateway/README.md)
- [`../models/`](../models/README.md)
