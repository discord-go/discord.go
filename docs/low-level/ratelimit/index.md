# Rate Limits

## Overview

The `ratelimit` package coordinates REST requests across Discord route buckets
and the global limit. `Limiter` blocks until a request can proceed, while
`Store` persists bucket state. `MemoryStore` is concurrency-safe and is the
default when `NewLimiter(nil)` is used.

## Architecture

`Info` is parsed from response headers: bucket hash, remaining requests, reset
time, reset-after duration, global status, and scope. `ParseHeaders` reads
`X-RateLimit-Bucket`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`,
`X-RateLimit-Reset-After`, and `X-RateLimit-Global`. `BucketState` stores only
remaining and reset time.

`NewLimiter(store)` creates a limiter. It applies a 50-request-per-second
global limit per limiter using `golang.org/x/time/rate.Limiter` (a token
bucket) and tracks route locks. The token bucket allows concurrent goroutines
to reserve tokens in parallel without serializing through a global mutex.
Once a response supplies a bucket hash, routes with the same hash and major
resource share state. `Wait` honors context cancellation while sleeping for
global or bucket resets; unknown buckets proceed immediately. `Update`
records the parsed response and global reset.

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/discord-go/discord.go/ratelimit"
)

func main() {
	header := make(http.Header)
	header.Set("X-RateLimit-Bucket", "messages")
	header.Set("X-RateLimit-Remaining", "2")
	header.Set("X-RateLimit-Reset-After", "0.25")
	info := ratelimit.ParseHeaders(header)

	limiter := ratelimit.NewLimiter(ratelimit.NewMemoryStore())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := limiter.Wait(ctx, "/channels/1/messages"); err != nil {
		panic(err)
	}
	limiter.Update("/channels/1/messages", info)
	fmt.Println(info.Bucket, info.Remaining)
}
```

## Creating A Limiter

`Store` has `Get(bucketID) (BucketState, bool)` and `Put(bucketID, state)`.
The in-memory store is process-local. `NewLimiter` accepts nil and creates an
in-memory store. A shared store by itself does not coordinate multiple
processes: `Get` and `Put` do not provide atomic reservations, and the global
limiter is local. Use a limiter designed for distributed coordination instead.
The opt-in [`redislimiter` package](distributed.md) provides Redis-backed
coordination when its operational requirements are met.

## Using Buckets

Call `Wait` immediately before the HTTP request and `Update` after parsing its
headers. Use a stable route key, usually the REST path without query
parameters. The REST client does this itself. A 429 response should supply
`Retry-After` or a JSON retry interval at the transport layer; the REST client
applies that interval to subsequent calls as a cooldown.

## Common Patterns

Use `ParseHeaders(resp.Header)` and pass the result to `Update`. Store reset
times using the server values when available and use `ResetAfter` for relative
responses. Pass a context with a deadline to `Wait` so shutdown cannot leave a
goroutine asleep until a long reset.

## Best Practices

Share one limiter among REST clients that share a token and process. Keep route
keys consistent. Treat the memory limiter as coordination, not a guarantee
against a server-side limit; Discord can change buckets and scopes.

The in-memory limiter uses a fixed number of shared locks, but retains route
mappings and bucket state for its lifetime. Long-running clients that encounter
many distinct paths should monitor memory use; automatic eviction of those
entries is not currently implemented.

## Common Mistakes

Do not call `Wait` with an empty key and expect per-route coordination; the
global window still applies, but no bucket state is stored. Do not use a route
path that includes volatile query parameters as a permanent bucket key. Do not
ignore `context.Canceled` from a blocked wait.

## API Walkthrough

The public API is `Limiter`, `Store`, `BucketState`, `Info`, `ParseHeaders`,
`NewLimiter`, `MemoryStore`, `NewMemoryStore`, `MemoryStore.Get`, and
`MemoryStore.Put`.

## Examples

The Quick Start program is complete and runnable without a Discord request.
REST integration details are in [`../rest/ratelimits.md`](../rest/ratelimits.md).

## Related APIs

- [`../rest/ratelimits.md`](../rest/ratelimits.md)
- [`../rest/requests.md`](../rest/requests.md)
- [`../http/`](../http/README.md)
