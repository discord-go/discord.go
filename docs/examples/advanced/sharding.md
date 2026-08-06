# Gateway Sharding

## Overview

Gateway sharding splits one bot's guild event stream across multiple Gateway
connections. Discord assigns each connection a shard ID and total shard count;
guilds are routed using the Discord snowflake formula. Sharding is a Gateway
scaling concern, not a replacement for a shared application database or cache.

This repository exposes two levels:

- `bot.WithShards(count)` enables the high-level bot lifecycle and is the right
  starting point for ordinary applications;
- `gateway.NewShardManager(token, count, intents)` provides per-shard
  connection factories, dispatching, cache configuration, broadcasts, voice
  routing, and explicit shutdown.

`count == 0` means "ask Discord for the recommended count" when the manager
starts. It does not mean "run without sharding". Use a normal `bot.New` without
`WithShards` for one connection.

## Architecture

```text
bot.Bot
  |
  +-- gateway.ShardManager
        |
        +-- shard 0: gateway.Client + session + connection
        +-- shard 1: gateway.Client + session + connection
        +-- shard N: gateway.Client + session + connection
        |
        +-- shared Dispatcher
        +-- optional shared cache.Cache
        +-- identify concurrency tracker
```

On startup, `ShardManager.Start` resolves a zero count through
`/api/v10/gateway/bot`, groups identifies by Discord's advertised maximum
concurrency, and waits `gateway.ShardDelay` between groups. Each shard has its
own session and resume state. The manager dispatches events through one
dispatcher, but event ordering across shards is not guaranteed.

The high-level bot configures a URL-aware connection factory so reconnects can
use a shard's resume URL. The low-level manager also lets an application
inspect `ShardID`, choose a connection transport, call `Broadcast`, and route
voice state with `JoinVoiceChannel`.

## Prerequisites

- Go `1.26.4` or a compatible newer toolchain, as declared by
  [`go.mod`](../../../go.mod).
- A Discord bot token and an application allowed to use the selected shard
  count and intents.
- The `Guilds` intent for the Quick Start; add other intents only for events
  the application consumes and enable privileged intents in the Portal.
- A deployment plan for one shard owner per shard. Multiple independent
  processes must not all identify the same shard set.
- A shared durable store or cache for state that must be visible across
  replicas. A process-local `storage.MemoryStore` is not cross-shard storage.

Discord may require sharding for large applications and may impose session
start limits. Read the `session_start_limit` returned by `/gateway/bot` before
planning a rollout.

## Quick Start

This complete program uses the current high-level lifecycle. With no
`SHARD_COUNT`, `bot.WithShards(0)` asks Discord for the recommended count. Set a
positive value only when Discord and the deployment plan have selected it.
Automatic command synchronization is disabled here so every shard does not
try to register the same command during startup.

```bash
export DISCORD_TOKEN='replace-with-a-bot-token'
# Optional: export SHARD_COUNT='4'
go run sharding-example.go
```

```go
package main

import (
	"log"
	"os"
	"strconv"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
)

func configuredShardCount() int {
	raw := os.Getenv("SHARD_COUNT")
	if raw == "" {
		return 0
	}
	count, err := strconv.Atoi(raw)
	if err != nil || count < 0 {
		log.Fatalf("SHARD_COUNT must be zero or a positive integer")
	}
	return count
}

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}

	client := bot.New(token,
		bot.WithIntents(intents.Guilds),
		bot.WithShards(configuredShardCount()),
		bot.WithCommandSyncDisabled(),
	)
	client.OnReady(func(ctx *bot.ReadyContext) {
		log.Printf("READY as %s; configured shards=%d", ctx.User.Username, configuredShardCount())
	})
	client.OnError(func(err error) {
		log.Printf("runtime: %v", err)
	})

	if err := client.Run(); err != nil {
		log.Fatal(err)
	}
}
```

The example connects and logs one READY event per shard. It does not register
commands because synchronization is intentionally disabled. Register commands
from one controlled deployment step, or use `bot.WithGuildCommandSync` or the
default global sync only when repeated READY-time synchronization is acceptable
for the application.

## Practical Levels

### Basic: high-level sharding

Use `bot.WithShards(0)` and let the repository query the recommended count.
Keep handlers shard-agnostic and use `bot.Run` or `bot.RunContext` for
shutdown. This is enough for a single process that owns all shards.

If Discord or operations specifies a fixed count, use a positive count:

```go
client := bot.New(token,
	bot.WithIntents(intents.Guilds),
	bot.WithShards(4),
)
```

Do not choose a count from guild count alone. Respect Discord's recommendation,
session start limits, deployment capacity, and any approval requirements.

### Intermediate: shard-aware routing

At the low-level API, calculate a guild's shard with
`gateway.CalculateShardID(guildID, manager.NumShards())`. Use
`ShardManager.JoinVoiceChannel` for voice state routing instead of selecting a
client by guesswork. Use `Broadcast` for control payloads that must reach every
active shard.

```go
shardID := gateway.CalculateShardID(guildID, manager.NumShards())
log.Printf("guild %s belongs to shard %d", guildID, shardID)
```

The high-level `ReadyContext` does not expose a shard ID. If a feature needs
per-shard identity or a custom transport, use `gateway.ShardManager` directly
or add that information at the application boundary rather than deriving it
from event order.

### Advanced: multi-process ownership

Partition shard IDs across workers only when the deployment has an explicit
ownership protocol. A worker needs a lease, fencing, and a shutdown handoff so
two workers cannot identify the same shard simultaneously. Keep application
state in a shared backend and make event processing idempotent because reconnect
and ownership transitions can produce repeated work.

For very large fleets, read `rest.Client.GetGatewayBot` during planning and
monitor `gateway.SessionStartLimit`. Roll out in identify-concurrency buckets,
not by opening all connections at once.

## Best Practices

- Respect `gateway.ShardDelay`, Discord's identify concurrency, and session
  start remaining/reset values.
- Use `WithShards(0)` only when automatic recommendation lookup is desired;
  otherwise record and review the fixed count.
- Make scheduled jobs shard-safe. A job that runs once per process can duplicate
  work when every shard or replica starts it.
- Use shared durable state for cross-shard coordination and per-guild keys for
  storage. Do not assume a process-local cache is shared.
- Treat each shard's event sequence as independent; do not infer a global
  ordering across shards.
- Register handlers before `Start` and use a finite shutdown context.
- Use `SetConnectionURLFactory` when resume URLs must be preserved across
  reconnects; the high-level bot already configures a URL-aware factory.
- Keep intents minimal and enable privileged intents explicitly in the Portal.
- Make command registration deterministic and coordinate it separately from
  shard ownership when running multiple workers.
- Monitor connection count, READY time, reconnects, identify failures, event
  lag, REST rate limits, and shutdown duration.

## Common Mistakes

### Treating zero as no shards

Wrong assumption:

```go
bot.WithShards(0) // "disable sharding"
```

Correct: omit `WithShards` for one Gateway connection. `WithShards(0)` asks
Discord for a recommended count during `ShardManager.Start`.

### Starting every shard concurrently

Do not replace `ShardManager.Start` with a loop that opens all connections at
once. The manager groups starts by `MaxConcurrency` and applies
`gateway.ShardDelay` between groups to respect Gateway limits.

### Reading a shard before startup

`manager.Shard(index)` returns `nil` before that shard exists and after shutdown.
Check the result and do not treat it as a constructor.

### Running the same shard set in two processes

Two replicas both using `WithShards(0)` do not form a shared cluster; both try
to identify every shard. Use one owner for the set, or implement explicit shard
leases and partitioning.

### Assuming events are globally ordered

Each shard has its own sequence. A guild event on shard 1 may be observed before
an unrelated event on shard 0 even if the events were generated in another
order. Use timestamps, versions, or domain reconciliation where ordering
matters.

### Calling `Start` twice

One `ShardManager` owns its sessions and connection entries. Create a new
manager for a new lifecycle; cancel and `Shutdown` the old one first.

## API Walkthrough

- [`bot.WithShards`](../../../bot/bot.go) enables high-level sharding; zero
  requests Discord's recommended count.
- [`bot.RunContext`](../../../bot/lifecycle.go) and `bot.Stop` provide bounded
  lifecycle control.
- [`gateway.NewShardManager`](../../../gateway/shard.go) creates the low-level
  manager. A negative count is clamped to one; zero is resolved at `Start`.
- [`ShardManager.SetConnectionFactory`](../../../gateway/shard.go) installs a
  factory receiving a `gateway.ShardID`.
- [`ShardManager.SetConnectionURLFactory`](../../../gateway/shard.go) receives
  an initial or resume URL and a `ShardID`.
- [`ShardManager.SetCache`](../../../gateway/shard.go) shares a cache instance
  with all clients in that manager.
- [`ShardManager.Start`](../../../gateway/shard.go) resolves recommendations,
  groups identifies, and starts shard clients.
- [`ShardManager.Shutdown`](../../../gateway/shard.go) cancels and closes all
  active shards and waits with the caller's context.
- [`ShardManager.Broadcast`](../../../gateway/shard.go) sends a raw
  `gateway.GatewayPayload` to every active shard.
- [`ShardManager.JoinVoiceChannel`](../../../gateway/shard.go) routes a voice
  state update to the calculated guild shard.
- [`gateway.CalculateShardID`](../../../gateway/shard.go) implements the
  snowflake routing formula.
- [`gateway.CalculateShards`](../../../gateway/shard.go) returns a multiple of
  a recommended count for deployments that have an explicit scaling policy.
- [`gateway.ShardID`](../../../gateway/shard.go) contains the zero-based shard
  ID and total count; `ToIdentifyShard` produces the identify payload field.
- [`gateway.ShardDelay`](../../../gateway/shard.go) is the repository's five
  second minimum delay between startup buckets.
- [`rest.Client.GetGatewayBot`](../../../rest/gateway_info.go) exposes gateway
  and session-start information through the REST client.

For a custom low-level transport, the factory returns the repository's
`gateway.Connection` interface: `Read()`, `Write([]byte)`, and `Close()`. The
default high-level bot factory uses Gorilla WebSocket and handles resume URL
selection through its lifecycle.

## Examples

A low-level manager with a connection factory has this shape:

```go
manager := gateway.NewShardManager(token, 0, intents.Guilds)
manager.Dispatcher.AddHandler(func(event *events.Ready) {
	log.Printf("one shard is ready as %s", event.User.Username)
})
manager.SetConnectionURLFactory(func(url string, id gateway.ShardID) (gateway.Connection, error) {
	return dialDiscordWebSocket(url, id)
})

ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer cancel()
if err := manager.Start(ctx); err != nil {
	return err
}
<-ctx.Done()
shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
defer shutdownCancel()
return manager.Shutdown(shutdownCtx)
```

The snippet assumes an application-defined `dialDiscordWebSocket` that returns
the current `gateway.Connection` interface. The high-level Quick Start is
complete because `bot.New` supplies that transport internally.

For a control payload, use `Broadcast` with a fully formed
`gateway.GatewayPayload`; do not write JSON directly to a client connection.
For guild voice, call `JoinVoiceChannel(guildID, channelID, mute, deaf)` and let
the manager calculate the target shard.

## Related Links

- [`gateway.ShardManager`](../../../gateway/shard.go)
- [`bot.WithShards`](../../../bot/bot.go)
- [`bot` lifecycle](../../../bot/lifecycle.go)
- [`rest.GetGatewayBot`](../../../rest/gateway_info.go)
- [Low-level shard guide](../../low-level/gateway/shards.md)
- [Examples overview](../README.md)
- [Persistence guide](../persistence/keyv.md)
- [Discord Gateway sharding](https://discord.com/developers/docs/topics/gateway#sharding)
- [Discord Gateway session start limit](https://discord.com/developers/docs/topics/gateway#session-start-limit-object)
