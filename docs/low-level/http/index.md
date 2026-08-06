# HTTP

## Overview

The `http` package provides the transport boundary used by REST. `Client` is
the small interface needed by `rest.Client`; `DefaultClient` wraps
`net/http.Client`, adds a user agent, and retries transient failures through a
`Retrier`. `LoggingTransport` is a pass-through `http.RoundTripper` that can be
extended with logging or metrics.

## Architecture

`NewClient(userAgent)` creates a default client with a logging transport around
`http.DefaultTransport` and a `NewDefaultRetrier(3)`. `DefaultClient.Do` adds
the user agent when the request has none, executes up to the configured retry
count, recreates a request body through `Request.GetBody`, and honors context
cancellation while waiting. `Get` and `Post` create standard requests and
delegate to `Do`.

`DefaultRetrier.ShouldRetry` retries transport errors except context canceled
and deadline exceeded, plus HTTP 408, 425, 429, 500, 502, 503, and 504. Its
`Backoff(attempt)` is `2^attempt * 10ms`; a positive `Retry-After` header takes
precedence in `DefaultClient.Do`. A body without `GetBody` cannot be retried.

## Quick Start

```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	discordhttp "github.com/discord-go/discord.go/http"
)

func main() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, r.UserAgent())
	}))
	defer server.Close()

	client := discordhttp.NewClient("example/1.0")
	response, err := client.Get(server.URL)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	fmt.Println(response.StatusCode)
}
```

## Creating A Client

Use `NewClient` for the standard defaults, or construct `DefaultClient` when
you need a custom `HTTPClient`, `UserAgent`, or `Retrier`. `DefaultRetrier` is
created with a maximum retry count; a nil retrier disables retry logic.

## Using Transports

`LoggingTransport.RoundTrip` delegates to `Base`, and uses
`http.DefaultTransport` when `Base` is nil. Wrap it with a custom transport to
record timing, status, or request IDs, but do not log Authorization headers or
request bodies containing tokens.

## Common Patterns

Use `http.NewRequestWithContext` and set `GetBody` automatically by using
`bytes.NewReader` or `strings.NewReader` for retryable bodies. Keep retryable
operations idempotent. Let REST's rate limiter handle Discord bucket timing;
this package's retry policy is transport/status oriented.

## Best Practices

Bound every request with a context. Close response bodies before retrying or
returning. Tune retry counts for the service's latency budget and make sure a
custom retrier does not retry authentication or validation failures.

## Common Mistakes

A POST body supplied by a reader without `GetBody` is not safely replayed.
`Retry-After` is interpreted as seconds. Context cancellation is not retried.
The package's logging transport currently only passes through; it does not
automatically emit logs.

## API Walkthrough

The public API is `Client`, `DefaultClient`, `NewClient`, `DefaultClient.Do`,
`Get`, and `Post`; `Retrier`, `DefaultRetrier`, `NewDefaultRetrier`,
`ShouldRetry`, `Backoff`, and `MaxRetries`; and `LoggingTransport.RoundTrip`.

## Examples

The Quick Start program is complete and runnable with an in-process HTTP
server. REST constructs its default client through this package.

## Related APIs

- [`../rest/`](../rest/README.md)
- [`../ratelimit/`](../ratelimit/README.md)
- [`../oauth2/`](../oauth2/README.md)
