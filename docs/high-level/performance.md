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

For a repeatable local measurement of string-form snowflake-list decoding,
run `go test ./snowflake -run '^$' -bench '^BenchmarkIDsDecode$' -benchmem`.
The benchmark does not measure network latency or full model decoding.

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

The default limiter provides:

- A 50 requests/second token bucket per limiter using `golang.org/x/time/rate.Limiter`
  (token bucket). This avoids global mutex contention under high concurrency
  — concurrent goroutines reserve tokens in parallel without serializing
  through a single lock.
- Per-bucket limits based on Discord's `X-RateLimit-Bucket` headers and major resources.
- Optimistic remaining count decrement to maximize throughput.

For bulk operations, use batch endpoints (e.g., `BulkOverwriteGlobalCommands`)
instead of individual requests. Share a limiter between clients using the same
token; independent limiters and processes do not coordinate automatically.
For coordinated processes, see the [Redis limiter](../low-level/ratelimit/distributed.md).

For local limiter overhead measurements, run:

```sh
go test ./ratelimit -run '^$' -bench 'BenchmarkLimiter_(Wait_Overhead|Acquire_)' -benchmem -count=3
```

These benchmarks disable the global wait only within the benchmark process.
They measure local coordination overhead; they do not measure Discord network
latency or the throughput of the default global rate limit.

## Local benchmark examples

The following microbenchmarks exercise cache reads, generic event lookups, and
REST processing of an empty 204 response. The REST benchmark uses a mock
transport and limiter, so it does not measure network I/O or Discord limits.

```sh
go test ./cache -run '^$' -bench 'BenchmarkMemoryCache_(GetParallel|MixedParallel)$' -benchmem -count=3
go test ./bot -run '^$' -bench 'Benchmark(DispatchEventUnsubscribed|MessageCreateDispatch)$' -benchmem -count=3
go test ./rest -run '^$' -bench '^BenchmarkRequestNoContent$' -benchmem -count=3
go test ./ratelimit -run '^$' -bench '^BenchmarkLimiter_UpdateRoute$' -benchmem -count=3
```

In separate 2026-11-27 runs on an AMD Ryzen 7 7445HS with Go 1.22.0 and
`GOMAXPROCS=12`, median results for the parallel cache-hit, mixed cache
(95% reads, TTL enabled), and unsubscribed generic-event lookup benchmarks
decreased from roughly 71 to 32 ns/op, 153 to 114 ns/op, and 90 to 26 ns/op,
respectively. The mocked empty REST response went from roughly 1,548 ns and
1,584 bytes per operation to 1,362 ns and 1,072 bytes. These are synthetic
throughput measurements from separate short runs, not network latency or a
guarantee for a particular bot. Full single-threaded message dispatch did not
show a clear improvement in the measured runs. A separate route-update
benchmark on the same machine fell from a median of 615 ns and 304 bytes
(10 allocations) to 466 ns and 208 bytes (8 allocations) after removing
temporary path-parsing allocations. It measures local limiter bookkeeping,
not the throughput permitted by Discord's rate limits.

### Live REST samples

On 2026-11-27, one test bot made read-only requests to Discord over a reused
connection. After a warm-up round, five sequential samples per operation gave
the following median end-to-end times: current user 465 ms, guild 483 ms,
guild channels 471 ms, guild roles 466 ms, guild member permissions 568 ms,
and current application 672 ms. Sample ranges included occasional requests
above one second. Calls used the default limiter, with 250 ms between calls.

A separate test rotated the order of discord.go (local checkout), DiscordGo
v0.29.0, DisGo v0.20.0, and Arikawa v3.6.0 for the same read-only user lookup.
With one warm-up round, seven samples per library, and 300 ms between calls,
their respective medians were 491, 504, 519, and 519 ms. These samples include
network and Discord latency and use each library's own limiter; they do not
establish a reliable performance ranking. The test was run with Go 1.26.4.

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
