# REST Requests

## Overview

`Client.Request(ctx, method, path, body, target)` is the low-level JSON request
boundary. It combines the client's `BaseURL`, authorization mode, JSON
serialization, audit-log reason, rate limiter, invalid-request guard, response
decoding, and error conversion. `RequestNoAuth` performs the same work without
an Authorization header.

## Architecture

The `path` is appended to `BaseURL`, and its query string is removed when the
route bucket key is selected. A non-nil body is marshaled and sends
`Content-Type: application/json`. `WithReason(ctx, reason)` attaches a value
that becomes the URL-escaped `X-Audit-Log-Reason` header;
`ReasonFromContext` reads it back.

After `Limiter.Wait`, the request is sent. Headers update the limiter. A 429
uses the JSON `retry_after`, then the `Retry-After` header, then reset timing,
and retries while the context remains active. Responses at 400 or above become
`*APIError`, except CAPTCHA-shaped bodies which become `*CaptchaError`. Both
carry HTTP status, API code/message, and structured fields where available.
The invalid-request guard records 401, 403, and 429 responses and can return
`ErrInvalidRequestLimitExceeded` before another request is sent.

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/discord-go/discord.go/rest"
)

type localHTTP struct{}

func (localHTTP) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"ok":true}`))}, nil
}
func (localHTTP) Get(string) (*http.Response, error) { return nil, nil }
func (localHTTP) Post(string, string, io.Reader) (*http.Response, error) { return nil, nil }

func main() {
	c := rest.New("token", nil, localHTTP{})
	var result struct{ OK bool `json:"ok"` }
	ctx := rest.WithReason(context.Background(), "maintenance")
	if err := c.Request(ctx, http.MethodGet, "/example", nil, &result); err != nil {
		panic(err)
	}
	reason, _ := rest.ReasonFromContext(ctx)
	fmt.Println(result.OK, reason)
}
```

## Using Authentication

`Request` uses `AuthBot` or `AuthBearer` when the client token is non-empty.
`RequestNoAuth` omits it regardless of the client's mode and is intended for
interaction and webhook URLs whose token is already in the path. Neither
method signs arbitrary external URLs; `path` remains relative to `BaseURL`.

## Common Patterns

Pass `nil` for body on GET and DELETE requests. Pass `nil` for target when the
endpoint returns no useful body. Use a typed response target for JSON and let
model custom unmarshalers handle components and snowflakes. Use
`errors.As(err, &apiErr)` to inspect `Code`, `Message`, `Errors`, and
`HTTPStatus`.

## Best Practices

Do not reuse a canceled context while REST is sleeping on a 429. Keep route
paths stable so the limiter can map returned bucket hashes. Make writes
repeatable before enabling application-level retries around a call that already
retries 429s.

## Common Mistakes

The request body is JSON, not form data; use multipart helpers for files. The
reason is URL-escaped by the client, so do not pre-escape it. A nil response
target does not mean the server returned no bytes. Do not use `RequestNoAuth`
to avoid authenticating ordinary bot operations.

## API Walkthrough

The exported request API is `Client.Request`, `Client.RequestNoAuth`,
`WithReason`, `ReasonFromContext`, `APIError.Error`, `CaptchaError.Error`, and
`ErrInvalidRequestLimitExceeded`. Client construction and authentication are
covered in [`README.md`](README.md); rate-limit behavior is in [`ratelimits.md`](ratelimits.md).

## Examples

The Quick Start program is complete and runnable using an in-memory HTTP
implementation. It demonstrates JSON decoding and reason propagation without
contacting Discord.

## Related APIs

- [`README.md`](README.md)
- [`ratelimits.md`](ratelimits.md)
- [`uploads.md`](uploads.md)
