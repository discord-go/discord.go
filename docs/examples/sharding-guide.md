# Sharding Guide

## Overview

Sharding splits gateway connections across multiple WebSocket connections.
Discord requires sharding for bots in more than 1,000 guilds. This guide covers
basic sharding, large bot sharding, and shard health monitoring.

## Basic Sharding

Enable sharding with a fixed shard count:

```go
client := bot.New(token,
    bot.WithIntents(intents.Guilds),
    bot.WithShards(4),
)
```

Or let the library detect the recommended shard count automatically:

```go
client := bot.New(token,
    bot.WithIntents(intents.Guilds),
    bot.WithShards(0), // 0 = auto-detect via gateway/bot
)
```

When `numShards` is 0, the `ShardManager` calls the `gateway/bot` endpoint to
get Discord's recommended shard count and `max_concurrency`.

## Shard Startup

Shards start in concurrency buckets based on Discord's `max_concurrency` limit.
Shards in the same bucket start concurrently. Buckets start sequentially with a
5-second delay (`ShardDelay`) between them.

## Large Bot Sharding

For very large bots, Discord may assign a shard count that is a multiple of the
recommended count. Use `gateway.CalculateShards` to compute the total:

```go
recommended := 16 // from gateway/bot
multiple := 2     // Discord-assigned multiple
totalShards := gateway.CalculateShards(recommended, multiple)
// totalShards = 32
```

## Per-Shard Identify Rate Limits

Discord enforces identify rate limits per shard. The `IdentifyTracker` manages
the 1-identify-per-5-seconds limit across all shards in the same process. The
`ShardManager` respects `max_concurrency` when starting shards.

## Shard Health Monitoring

Monitor per-shard health via the `ShardManager`:

```go
for i := 0; i < manager.NumShards(); i++ {
    shard := manager.Shard(i)
    if shard == nil {
        continue
    }
    latency := shard.Heartbeater.Ping()
    if latency > 5*time.Second {
        log.Printf("shard %d high latency: %v", i, latency)
    }
}
```

## Dynamic Shard Scaling

To change shard count, stop the bot and restart with a new count:

```go
// Stop current bot
client.Stop(context.Background())

// Restart with more shards
client = bot.New(token,
    bot.WithIntents(intents.Guilds),
    bot.WithShards(newShardCount),
)
client.Run()
```

Shard count cannot be changed without restarting. Plan for headroom.

## Shard ID Calculation

Guilds are assigned to shards by:

```go
shardID := (guildID >> 22) % numShards
```

Use `gateway.CalculateShardID` to compute this:

```go
shardID := gateway.CalculateShardID(guildID, numShards)
```

## Common Patterns

- Start with 1 shard for small bots.
- Use auto-detection (`WithShards(0)`) for growing bots.
- Monitor per-shard latency.
- Use `ShardManager.Broadcast` to send a payload to all shards.

## Best Practices

- Use `max_concurrency` from the `gateway/bot` response.
- Do not exceed Discord's session start limits.
- Monitor `SessionStartLimit.Remaining` before restarting.
- Keep shard count stable; changing it requires a full restart.

## Common Mistakes

- Using too few shards for a large bot (causes rate limits).
- Not respecting `max_concurrency` (causes identify failures).
- Assuming shard count can change without restart.
- Not monitoring per-shard health.
