# REST Rate Limits

## Overview

Every REST call participates in `ratelimit.Limiter`. The client uses the path
without its query string as the initial route key, then adopts Discord's
returned bucket hash so routes sharing a server bucket coordinate. Global
resets and the process-wide request window are handled by the limiter.

## Architecture

The request sequence is: wait for the route and global limits, send the
request, parse `X-RateLimit-*` headers, and update the limiter. A 429 response
is special: REST reads JSON `retry_after` first, then the `Retry-After` header,
then `ResetAfter`, waits with the caller context, and repeats the request. A
custom `ratelimit.Store` can persist `BucketState` across client instances;
the built-in memory store is process-local.

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
	limiter.Update("GET:/channels/1/messages", info)
	state, ok := store.Get("bucket:" + info.Bucket)
	fmt.Println(ok, state.Remaining, info.ResetAfter)
}
```

## Using A Limiter

`NewLimiter(nil)` creates a memory-backed limiter. Use `Wait(ctx, bucket)` just
before a request and `Update(bucket, ratelimit.ParseHeaders(resp.Header))`
afterward. An unknown bucket proceeds immediately. Once a hash is returned,
the route is mapped to `bucket:<hash>` and subsequent calls use that state.

## Common Patterns

Share a limiter for all REST clients using one token. Use `ResetAfter` for
relative server timing and `Reset` when an absolute server time is available.
Pass a deadline-bearing context so a shutdown or user request can cancel a
wait. For distributed workers, implement `Store` over a shared atomic data
store and account for clock skew.

## Best Practices

Let REST own 429 waiting rather than adding an uncoordinated sleep around each
call. Keep route keys stable and never key on a full URL with access tokens.
Monitor remaining and reset values, but treat them as server hints rather than
permission to burst beyond a known global policy.

## Common Mistakes

The default memory limiter does not coordinate separate processes. A route key
with query parameters is normalized by REST requests but not by a direct
limiter caller. `Retry-After` may be fractional seconds. A context canceled
while waiting returns the context error and does not send the request.

## API Walkthrough

The REST-facing API is the `ratelimit.Limiter` interface plus `NewLimiter`;
the underlying [`../ratelimit/`](../ratelimit/README.md) package provides
`Store`, `MemoryStore`, `BucketState`, `Info`, `ParseHeaders`, and all storage
methods.

## Examples

The Quick Start program is complete and runnable and demonstrates bucket hash
adoption without HTTP.

## Related APIs

- [`../ratelimit/`](../ratelimit/README.md)
- [`requests.md`](requests.md)
- [`../http/`](../http/README.md)
