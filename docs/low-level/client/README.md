# Low-Level Client

## Overview

`client.Client` is a small composition object. It owns a `*rest.Client`, a
`*gateway.ShardManager`, and the optional `cache.Cache` supplied at creation.
It is intentionally not a command router and it does not expose high-level
event contexts. Use it when your service wants the library's transport and
models but wants to own dispatch, persistence, and application policy.

## Architecture

`New(token string, opts ...Option)` first creates a `Config` containing the
token, applies each option in order, creates a one-shard Gateway manager, and
creates a REST client with the default HTTP client and in-memory rate limiter.
The default Gateway URL is `wss://gateway.discord.gg/?v=10&encoding=json`.
The default client does not enable intents and does not attach a cache. The
Gateway manager can later be reconfigured with the methods documented in
[`../gateway/shards.md`](../gateway/shards.md).

## Quick Start

`WithIntents` replaces the configured intent bitfield; it does not add to a
previous option. `WithCache` stores the exact interface value and the same
value is passed to the shard manager. The REST token is initialized from the
argument and uses `Authorization: Bot` by default.

```go
package main

import (
	"fmt"

	"github.com/discord-go/discord.go/cache"
	"github.com/discord-go/discord.go/client"
	"github.com/discord-go/discord.go/intents"
)

func main() {
	cacheStore := cache.NewMemoryCache(cache.WithMaxSize(1000))
	c := client.New("example-token",
		client.WithIntents(intents.Guilds|intents.GuildMessages),
		client.WithCache(cacheStore),
	)

	fmt.Printf("rest=%t shards=%d cache=%T\n", c.Rest != nil, c.Gateway.NumShards(), c.Cache)
}
```

## Creating A Client

The exported API is `Client`, `Config`, `Option`, `New`, `WithCache`, and
`WithIntents`. `Config` has exported `Token`, `Cache`, and `Intents` fields and
is the target of an `Option`; applications normally use the option functions
instead of constructing it directly. `Client` fields are exported so a caller
can replace or further configure the composed REST and Gateway objects.

## Using The Client

Start Gateway traffic with `c.Gateway.Start(ctx)` and perform HTTP operations
with `c.Rest`. The client constructor does not start goroutines or make a
network request. `Gateway.Shutdown(ctx)` should be used during service
shutdown. REST methods accept a context on each call; there is no client-wide
context to cancel.

## Common Patterns

Use `client.New` for production wiring, then replace the Gateway connection
factory before `Start` when a proxy, test server, or custom WebSocket wrapper
is needed. Add a cache only when the selected intents deliver enough events to
keep the cached records meaningful. For a multi-shard application, configure
the shard count and connection factory on `c.Gateway` rather than trying to
create multiple `client.Client` values with the same token.

## Best Practices

Keep the token in an environment variable or secret manager. Configure
privileged intents both in code and in the Developer Portal. Give REST and
Gateway operations independent contexts with bounded shutdown deadlines.
Install event handlers before starting the manager so the first dispatch is
not missed.

## Common Mistakes

`client.New` creates one shard by default; it does not query Discord for a
recommended count. It also does not call `Start`. A nil cache is valid, but
code must not type-assert `c.Cache` without checking it. Options are applied
in sequence, so two `WithIntents` options leave only the second bitfield.

## API Walkthrough

The package has no constructors besides `New`. `WithCache(c)` and
`WithIntents(i)` return `Option` closures. `Client.Rest`, `Client.Gateway`, and
`Client.Cache` are the integration points; their methods are documented in
[`../rest/`](../rest/README.md), [`../gateway/`](../gateway/README.md), and
[`../cache/`](../cache/README.md).

## Examples

The Quick Start program is complete and runnable. It does not contact Discord,
which makes it suitable for checking option wiring in CI.

## Related APIs

- [`../rest/`](../rest/README.md)
- [`../gateway/`](../gateway/README.md)
- [`../cache/`](../cache/README.md)
