# Cache Guide

## Overview

discord.go provides cache interfaces for guilds, channels, users, roles,
messages, and members. The default implementation is an in-memory cache. This
guide covers custom cache backends, invalidation, and sizing.

## Cache Interfaces

The cache package defines typed interfaces:

```go
type Cache interface {
    Get(key string) (any, bool)
    Set(key string, value any)
    Delete(key string)
    Clear()
}

type GuildCache interface {
    Cache
    GetGuild(id string) (any, bool)
    SetGuild(id string, guild any)
    DeleteGuild(id string)
}
```

Additional interfaces: `ChannelCache`, `UserCache`, `RoleCache`,
`MessageCache`, `MemberCache`.

## Default Memory Cache

```go
import "github.com/discord-go/discord.go/cache"

memCache := cache.NewMemoryCache()
client := bot.New(token,
    bot.WithCache(memCache),
)
```

The memory cache has no TTL or size limit. Memory grows unbounded for
long-running bots.

## Custom Cache Backend

Implement the cache interfaces for a custom backend (e.g., Redis):

```go
type RedisCache struct {
    client *redis.Client
    ttl    time.Duration
}

func (r *RedisCache) Get(key string) (any, bool) {
    val, err := r.client.Get(context.Background(), key).Result()
    if err != nil {
        return nil, false
    }
    return val, true
}

func (r *RedisCache) Set(key string, value any) {
    r.client.Set(context.Background(), key, value, r.ttl)
}

func (r *RedisCache) Delete(key string) {
    r.client.Del(context.Background(), key)
}

func (r *RedisCache) Clear() {
    r.client.FlushDB(context.Background())
}

func (r *RedisCache) GetGuild(id string) (any, bool) {
    return r.Get("guild:" + id)
}

func (r *RedisCache) SetGuild(id string, guild any) {
    r.Set("guild:"+id, guild)
}

func (r *RedisCache) DeleteGuild(id string) {
    r.Delete("guild:" + id)
}
```

Inject it via `bot.WithCache`:

```go
redisCache := &RedisCache{client: rdb, ttl: 5 * time.Minute}
client := bot.New(token, bot.WithCache(redisCache))
```

## Cache Invalidation

The gateway dispatcher updates the cache on events:
- `GUILD_CREATE` / `GUILD_UPDATE` / `GUILD_DELETE` update/remove guilds.
- `CHANNEL_CREATE` / `CHANNEL_UPDATE` / `CHANNEL_DELETE` update/remove channels.
- `MESSAGE_CREATE` / `MESSAGE_DELETE` add/remove messages.

For custom caches, ensure your `Set` and `Delete` methods handle these events
correctly. Treat cache as eventually consistent.

## Cache Sizing

For memory-constrained deployments:
- Implement a TTL to bound cache lifetime.
- Implement an LRU eviction policy to bound cache size.
- Cache only the entity types your bot needs (e.g., only guilds and channels,
  not messages).

## Common Patterns

- Use the memory cache for development and testing.
- Use Redis for multi-instance deployments.
- Set TTLs to prevent unbounded memory growth.
- Cache only what your bot reads frequently.

## Best Practices

- Do not rely on cache for critical data; always fall back to REST.
- Implement cache invalidation on gateway events.
- Monitor cache hit/miss ratios.
- Use typed cache interfaces for type safety.

## Common Mistakes

- Assuming cache is always populated; it may be empty after a restart.
- Not implementing all cache interface methods.
- Using `any` without type assertions in application code.
- Forgetting that cache is eventually consistent.
