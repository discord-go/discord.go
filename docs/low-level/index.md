# Low-Level API

## Overview

The low-level API is the typed protocol layer of discord.go. It does not try to
hide Discord's payload shapes or lifecycle rules. You construct resource
values, serialize them, send them through REST, and consume Gateway dispatches
as separate operations. This makes the packages useful for libraries,
services that already have their own event loop, and tests that need a precise
wire representation.

The package pages are organized by responsibility. Start with [`client/`](client/README.md)
if you want one object that composes REST, Gateway sharding, intents, and an
optional cache. Start with [`rest/`](rest/README.md) or [`gateway/`](gateway/README.md)
when you need to own those lifecycles directly.

## Architecture

Resource models are deliberately independent of transport. [`models/`](models/README.md)
describes the object packages, while [`messages/`](messages/README.md),
[`components/`](components/README.md), and [`interactions/`](interactions/README.md)
describe payloads that cross the REST and Gateway boundaries. [`snowflake/`](snowflake/README.md)
is used by most models to preserve Discord IDs. [`json/`](json/README.md) is the
small serialization seam used by protocol code.

REST calls pass through [`ratelimit/`](ratelimit/) and [`http/`](http/).
Gateway sessions use [`events/`](events/README.md), [`gateway/heartbeat.md`](gateway/heartbeat.md),
and [`gateway/shards.md`](gateway/shards.md). Voice is a separate connection
that is coordinated with Gateway voice-state events; see [`voice/`](voice/README.md).

## Quick Start

The example below only builds local values, so it is safe to run without a
Discord token. A real application should keep the token out of source control,
derive a cancellable context for each operation, and call `c.Gateway.Start`
only after configuring the required Gateway intents.

```go
package main

import (
	"fmt"

	"github.com/discord-go/discord.go/cache"
	"github.com/discord-go/discord.go/client"
	"github.com/discord-go/discord.go/intents"
)

func main() {
	store := cache.NewMemoryCache(cache.WithTTL(5 * 60 * 1e9))
	c := client.New("token-is-not-used-by-this-example",
		client.WithIntents(intents.Guilds|intents.GuildMessages),
		client.WithCache(store),
	)

	fmt.Println(c.Rest != nil, c.Gateway.NumShards(), c.Cache != nil)
}
```

## Common Patterns

Use pointers in edit parameter structs when Discord distinguishes "unset" from
an explicit zero, false, empty string, or null. Use `snowflake.ID` rather than
`int64` for IDs, and use the package's custom marshalers for ID arrays. Keep
network contexts caller-owned and let REST return its typed errors. For event
payloads not represented by a typed model, retain the original raw `d` value
and decode it later rather than discarding it.

## Best Practices

Request only the intents the application needs. Share one cache and one rate
limit strategy per process, but use a shared store when multiple processes
make REST requests. Treat Gateway reconnects and REST retries as normal
control flow. Validate embeds and attachment sizes before making a request,
and make writes repeatable when a timeout leaves the result uncertain.

## Common Mistakes

Do not use a cancelled context for a retry or a follow-up call. Do not assume a
missing JSON field means false when the API uses a pointer field to represent
absence. Do not send a bot token to an interaction webhook URL that expects
`RequestNoAuth`.

## Examples

The package guides linked above contain runnable examples for each subsystem:
serialization, model construction, transport setup, Gateway dispatch, and
multipart requests. They intentionally avoid real network calls unless a page
is specifically explaining a transport boundary.

## Related APIs

- [`client/`](client/README.md) for composition.
- [`rest/`](rest/README.md) for Discord HTTP resources.
- [`gateway/`](gateway/README.md) for the WebSocket lifecycle.
- [`models/`](models/README.md) for model-package coverage.
