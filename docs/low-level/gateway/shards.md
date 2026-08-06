# Shards

## Overview

`gateway.ShardManager` owns one Gateway client per shard and shares a
dispatcher, cache, connection policy, and Gateway URL. It is the low-level
choice for applications that need to scale beyond one connection while
retaining control over handlers and shutdown.

## Architecture

`NewShardManager(token, numShards, gatewayIntents)` stores the token and intent
configuration. If `numShards` is zero, `Start` fetches Discord's recommended
count directly from `/api/v10/gateway/bot` using the bot token.
`SetConnectionFactory` gets a
`ShardID` and is suitable for a normal initial connection. Use
`SetConnectionURLFactory` when reconnects must receive the session's resume
URL. `SetGatewayURL`, `SetCache`, and `SetCompression` configure all clients.

`Start(ctx)` groups shards by Discord's advertised identify concurrency and
starts each group with the package's `ShardDelay` between groups. It returns
the first startup error. `Broadcast` sends a
`GatewayPayload` to every active shard. `JoinVoiceChannel` calculates the
target shard and forwards the voice-state update. `Shard(index)` returns nil
for an invalid index, and `NumShards` reports the configured count.

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/discord-go/discord.go/gateway"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/snowflake"
)

func main() {
	manager := gateway.NewShardManager("example-token", 2, intents.Guilds)
	manager.SetConnectionFactory(func(id gateway.ShardID) (gateway.Connection, error) {
		return nil, fmt.Errorf("dial shard %s only in production", id.String())
	})
	shard := gateway.CalculateShardID(snowflake.ID(123), manager.NumShards())
	fmt.Println(manager.NumShards(), shard)
}
```

The example does not call `Start` because its factory intentionally does not
dial. A production factory must return a working `gateway.Connection`.

## Creating A Shard Manager

Use a positive shard count when Discord or deployment policy has already
chosen it. `CalculateShards(recommended, multiple)` returns a usable multiple
of Discord's recommendation for large deployments; non-positive inputs are
clamped to one. `CalculateShardID` uses Discord's snowflake formula and also
clamps a non-positive shard count to one.

## Using Factories And Shutdown

The connection factory is called for each shard and again on reconnect. The
URL-aware factory receives either the initial Gateway URL or the saved resume
URL. Register handlers on `manager.Dispatcher` before `Start`. On shutdown,
cancel the parent context and call `Shutdown(ctx)` to close all shard clients
and wait for them; the shutdown context should have a finite deadline.

## Common Patterns

Share a cache only when event coverage and shard ownership are understood.
Broadcast presence or other control payloads through the manager instead of
iterating over clients. Route guild-specific actions with the formula rather
than guessing from guild ID ranges.

## Best Practices

Respect `ShardDelay` and identify concurrency limits. Use the URL-aware factory
when resume correctness matters. Treat shard sequences as independent and do
not assume dispatch order across shards.

## Common Mistakes

`numShards == 0` is not a zero-connection mode; `Start` resolves a recommended
count. `Shard(index)` does not create a client. A factory returning nil without
an error will fail later during startup. Do not call `Start` twice on one
manager.

## API Walkthrough

The public API is `ShardManager`, `NewShardManager`, `Start`, `Shutdown`,
`Broadcast`, `JoinVoiceChannel`, `Shard`, `NumShards`, all configuration
setters, `CalculateShardID`, `CalculateShards`, `ShardID`, and `ShardDelay`.

## Examples

The Quick Start program is complete and runnable without dialing Discord.

## Related APIs

- [`README.md`](README.md)
- [`../intents/`](../intents/README.md)
- [`../cache/`](../cache/README.md)
