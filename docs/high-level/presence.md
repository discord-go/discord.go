# Presence And Latency

## Overview

Presence is the status and activity Discord displays for the bot. discord.go
supports an initial presence through `WithPresence` and runtime updates through
`SetPresence`, `SetStatus`, and `SetActivity`. It also exposes gateway heartbeat
latency and a lightweight REST latency measurement.

## Architecture

`PresenceUpdate` is serialized into the gateway OP 3 payload. A single bot sends
the payload through its gateway client; a sharded bot broadcasts it to every
active shard. The bot remembers the latest successful presence and reapplies it
after a fresh identify. `GatewayLatency` reads the heartbeat round-trip value,
while `APILatency` measures a `GetCurrentUser` REST request.

## Quick Start

This complete program configures an initial activity and changes the status
after READY.

```go
package main

import (
	"context"
	"log"
	"os"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}

	b := bot.New(token,
		bot.WithIntents(intents.Guilds),
		bot.WithPresence(bot.PresenceUpdate{
			Status:     "online",
			Activities: []bot.Activity{{Name: "discord.go", Type: 0}},
		}),
	)
	b.OnReady(func(ctx *bot.ReadyContext) {
		if err := ctx.Bot.SetStatus(context.Background(), "online"); err != nil {
			log.Printf("presence: %v", err)
		}
		log.Printf("ready as %s", ctx.User.Username)
	})
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

Presence updates require a running gateway. The activity `Type` is the integer
Discord activity type; `0` is the usual playing activity.

## Creating/Configuration

`bot.Activity` contains `Name`, `Type`, and optional `URL`.
`bot.PresenceUpdate` contains `Since`, `Activities`, `Status`, and `AFK`.
Pass a value to `WithPresence` for automatic application after READY. Use
`SetPresence` for multiple activities or to control all fields. Use
`SetStatus` to preserve the current activity list and `SetActivity` to replace
the activity list with one activity.

## Using

### Basic: set status

Call `SetStatus(ctx, "idle")`, `"dnd"`, or another Discord-supported status
after READY. The method preserves the remembered activities.

### Intermediate: update an activity

Call `SetActivity(ctx, bot.Activity{Name: "a queue", Type: 0})`. This replaces
the current activity slice with one item.

### Advanced: monitor latency

`GatewayLatency()` is local and normally cheap; `APILatency(ctx)` performs a
network request and returns both duration and error. Measure these separately
when diagnosing gateway versus REST problems.

## Common Patterns

- Configure a stable initial presence with `WithPresence` and change dynamic
  text only when it changes.
- Reapply a presence from `OnReady` only when the application needs a runtime
  value not known during construction.
- Do not use REST latency as a proxy for gateway health.
- In sharded deployments, rely on `SetPresence` so every shard receives the
  update.
- Treat a zero gateway latency as "not available yet", not as a real zero RTT.

## Best Practices

### Keep presence updates infrequent

Why: gateway status updates consume connection bandwidth and are visible UI.

Pros: less noise and fewer avoidable gateway operations.

Cons: highly dynamic counts may become slightly stale.

### Use an honest status

Why: presence is user-facing operational information.

Pros: users can understand whether the service is available.

Cons: an activity can expose deployment or workload details, so avoid secrets
and private identifiers.

### Measure both paths

Why: gateway heartbeat and REST API requests fail independently.

Pros: dashboards distinguish Discord gateway issues from HTTP/rate-limit issues.

Cons: `APILatency` adds a real request and should not be sampled too often.

## Common Mistakes

Incorrect: setting presence before the gateway starts.

```go
_ = b.SetStatus(context.Background(), "online")
_ = b.Start(context.Background())
```

Correct: configure it with `WithPresence` or wait for READY.

```go
b := bot.New(token, bot.WithPresence(bot.PresenceUpdate{Status: "online"}))
```

Incorrect: assuming `SetActivity` appends to existing activities.

```go
_ = b.SetActivity(ctx, bot.Activity{Name: "new", Type: 0})
```

Correct: use `SetPresence` when multiple activities must be preserved.

```go
_ = b.SetPresence(ctx, bot.PresenceUpdate{
	Status: "online",
	Activities: []bot.Activity{
		{Name: "one", Type: 0},
		{Name: "two", Type: 0},
	},
})
```

## API Walkthrough

- `Activity` has `Name`, `Type`, and `URL` fields.
- `PresenceUpdate` has `Since *int64`, `Activities []Activity`, `Status`, and
  `AFK` fields.
- `WithPresence(PresenceUpdate) bot.Option` stores an initial presence.
- `SetPresence(context.Context, PresenceUpdate) error` sends and remembers the
  complete presence; it returns an error when the gateway is not running.
- `SetStatus(context.Context, string) error` preserves activities while changing
  status.
- `SetActivity(context.Context, Activity) error` preserves status while replacing
  activities with the supplied activity.
- `GatewayLatency() time.Duration` returns the latest heartbeat ping or zero
  when no gateway heartbeater is available.
- `APILatency(context.Context) (time.Duration, error)` measures the current-user
  REST request.

## Examples

- [Basic client](../examples/setup/basic-client.md)
- [Full template](../examples/advanced/full-template.md)

## Related APIs

- [`lifecycle.md`](lifecycle.md) for READY and reconnect behavior.
- [`client.md`](client.md) for shard and gateway options.
- [`../low-level/gateway/heartbeat.md`](../low-level/gateway/heartbeat.md) for heartbeat details.
