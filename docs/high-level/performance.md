# Performance Guide

## Overview

discord.go is designed for low-latency gateway dispatch and efficient REST
operations. This guide covers allocation reduction, connection reuse, and
throughput tuning.

## Allocation Reduction

### Cached Voice Encryption

The voice `Client` caches the `cipher.AEAD` when the secret key is received,
avoiding per-packet AES cipher and GCM allocation. This eliminates 40+
allocations per second per voice connection at 20 packets/sec.

### JSON Handling

The REST client marshals request bodies once and reuses the bytes for retries.
Response bodies are read fully into memory before unmarshaling to support error
parsing and 429 retry logic.

### Byte Buffer Reuse

The `bytes.NewReader` used for request bodies is lightweight. For high-throughput
applications, consider passing a custom `http.Client` with a transport that pools
buffers.

## Connection Pooling

The default HTTP client clones `http.DefaultTransport` and sets
`MaxIdleConnsPerHost = 32` (default is 2). Since Discord REST requests go to a
single host (`discord.com`), the higher value prevents connection churn under
concurrent REST calls.

To tune further, provide a custom `http.Client`:

```go
transport := http.DefaultTransport.(*http.Transport).Clone()
transport.MaxIdleConnsPerHost = 64
transport.MaxConnsPerHost = 100
httpClient := &http.Client{Transport: transport}
restClient := rest.New(token, nil, httpClient)
```

## Gateway Throughput

### Event Dispatch

The bot dispatches events to handlers in goroutines, bounded by
`bot.WithMaxHandlerConcurrency`. Set this to limit goroutine creation:

```go
client := bot.New(token,
    bot.WithMaxHandlerConcurrency(100),
)
```

### Heartbeat Latency

Monitor gateway latency via `bot.GatewayLatency()`:

```go
latency := client.GatewayLatency()
if latency > 5*time.Second {
    log.Printf("high gateway latency: %v", latency)
}
```

### REST Latency

Measure REST latency with `bot.APILatency`:

```go
latency, err := client.APILatency(ctx)
```

## Cache Tuning

The default `cache.Memory` cache has no TTL or size limit. For long-running
bots, consider:

- Implementing a cache with TTL to bound memory usage.
- Implementing an LRU cache with a max-size eviction policy.
- Using the cache interfaces (`GuildCache`, `ChannelCache`, etc.) to cache only
  what your bot needs.

## Rate Limit Throughput

The rate limiter enforces:
- 50 requests/second global limit using `golang.org/x/time/rate.Limiter`
  (token bucket). This avoids global mutex contention under high concurrency
  — concurrent goroutines reserve tokens in parallel without serializing
  through a single lock.
- Per-bucket limits based on Discord's `X-RateLimit-Bucket` headers.
- Optimistic remaining count decrement to maximize throughput.

For bulk operations, use batch endpoints (e.g., `BulkOverwriteGlobalCommands`)
instead of individual requests.

## Compression

Enable gateway compression to reduce bandwidth:

```go
client := bot.New(token,
    bot.WithGatewayCompression(true),
)
```

This enables `zlib-stream` encoding on the gateway WebSocket.

## Common Patterns

- Set `MaxHandlerConcurrency` to bound goroutine creation.
- Monitor gateway and REST latency.
- Use batch REST endpoints for bulk operations.
- Enable compression for bandwidth-sensitive deployments.

## Best Practices

- Profile with `go test -bench` before optimizing.
- Use `context.Context` with deadlines on all REST calls.
- Avoid tight loops that make many REST requests without delays.
- Reuse `rest.Client` and `bot.Bot` instances; do not create new clients per
  request.

## Common Mistakes

- Not setting `MaxHandlerConcurrency`, allowing unbounded goroutine creation.
- Creating a new `rest.Client` per request instead of reusing one.
- Not monitoring gateway latency for connection health.
- Using global command sync in tight loops during development.
