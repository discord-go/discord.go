# REST Rate Limits

## Overview

Every REST call participates in `ratelimit.Limiter`. The default limiter uses
the HTTP method and path without its query string, then associates Discord's
returned bucket hash with the route and its major resource. Clients that share
one limiter coordinate their route and global limits in the same process.

## Architecture

The request sequence is: wait for the route and global limits, send the
request, parse `X-RateLimit-*` headers, and update the limiter. A 429 response
is special: REST reads JSON `retry_after` first, then the `Retry-After` header,
then `ResetAfter`, waits with the caller context, and repeats the request. A
custom `ratelimit.Store` can persist `BucketState`; the built-in memory store
is process-local. A shared store alone cannot coordinate multiple processes:
its interface does not provide atomic reservations or a shared global limiter.
The default limiter conservatively applies its token bucket to all REST paths,
including interaction endpoints that Discord exempts from the bot global limit.
Concurrent requests can still encounter 429s, especially before a route's
limits are known; the limiter cannot guarantee that Discord will never reject a
request.

The default limiter holds concurrent requests for a route while its first
request discovers the bucket, and allows one discovery request after a known
window expires. The REST client releases this reservation on a response or
transport error. Custom limiters can implement `ratelimit.RouteReserver` to
coordinate the same request lifecycle; `Limiter` and `RouteLimiter` remain
supported.

`Client.MaxRetries` bounds the 429 retry loop. When set to a positive value,
the client gives up after that many retries and returns `*RateLimitError`
instead of retrying forever; zero (the default) keeps the historical unbounded
behavior. Every 429 honors this budget, including responses without a usable
`retry_after` hint (for example a Cloudflare-fronted HTML 429); in that case
the client backs off one second between attempts. `Client.BucketState(route)`
exposes the limiter's current state for a route so callers can inspect
remaining budget and reset time without guessing.

## Quick Start

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/discord-go/discord.go/ratelimit"
)

func main() {
	header := make(http.Header)
	header.Set("X-RateLimit-Bucket", "abc")
	header.Set("X-RateLimit-Remaining", "4")
	header.Set("X-RateLimit-Reset-After", "1.5")
	info := ratelimit.ParseHeaders(header)
	store := ratelimit.NewMemoryStore()
	limiter := ratelimit.NewLimiter(store)
	limiter.Update("/channels/1/messages", info)
	state, ok := limiter.(ratelimit.StateReader).State("/channels/1/messages")
	fmt.Println(ok, state.Remaining, info.ResetAfter)
}
```

## Using A Limiter

`NewLimiter(nil)` creates a memory-backed limiter. Use `Wait(ctx, bucket)` just
before a request and `Update(bucket, ratelimit.ParseHeaders(resp.Header))`
afterward. An unknown bucket proceeds immediately. Once a hash is returned,
the limiter associates it with the route's major resource. The REST client
uses the optional `RouteLimiter` interface to distinguish methods; custom
limiters implementing only `Limiter` continue to receive the path.

## Common Patterns

Share a limiter for all REST clients using one token. Use `ResetAfter` for
relative server timing and `Reset` when an absolute server time is available.
Pass a deadline-bearing context so a shutdown or user request can cancel a
wait. Distributed workers need a limiter that coordinates atomic reservations
and global limits across processes, with appropriate handling of clock skew.

## Best Practices

Let REST own 429 waiting rather than adding an uncoordinated sleep around each
call. Keep route keys stable and avoid logging them: webhook paths can contain
access tokens, including in limiter state stored by a custom store.
Monitor remaining and reset values, but treat them as server hints rather than
permission to burst beyond a known global policy.

## Common Mistakes

The default memory limiter does not coordinate separate processes. A route key
with query parameters is normalized by REST requests but not by a direct
limiter caller. `Retry-After` may be fractional seconds. A context canceled
while waiting returns the context error and does not send the request.

## API Walkthrough

The REST-facing API is the `ratelimit.Limiter` interface plus `NewLimiter`,
`Client.MaxRetries`, `Client.BucketState`, and `RateLimitError`; the
underlying [`../ratelimit/`](../ratelimit/README.md) package provides
`Store`, `MemoryStore`, `BucketState`, `Info`, `ParseHeaders`, and all storage
methods.

## Examples

The Quick Start program is complete and runnable and demonstrates bucket hash
adoption without HTTP.

## Related APIs

- [`../ratelimit/`](../ratelimit/README.md)
- [`requests.md`](requests.md)
- [`../http/`](../http/README.md)
