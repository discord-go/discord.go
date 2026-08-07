# REST

## Overview

The `rest` package is the typed Discord API v10 client. Its `Client` builds
documented routes, serializes request parameters, applies authorization,
coordinates rate limits, retries 429 responses, protects against repeated
invalid requests, and decodes response models. Resource methods are grouped
by endpoint family rather than exposing route concatenation to callers.

## Architecture

`New(token, limiter, httpClient)` uses `Authorization: Bot`, a memory limiter,
and `http.NewClient` when the optional arguments are nil. It sets
`BaseURL` to `https://discord.com/api/v10`. The token is stored in an
unexported field and is only accessible via the internal request path.
`SetBearerToken` changes the mode to `Bearer`; `SetBotToken` restores `Bot`.
`SetToken` sets the token without changing the auth mode. `AuthNone` is used
internally by no-auth methods and prevents an authorization header.

`Request` and `RequestNoAuth` are the JSON escape hatches. Multipart methods
are described in [`uploads.md`](uploads.md). The client records 401, 403, and
429 responses and can return `ErrInvalidRequestLimitExceeded` before another
invalid request risks a Cloudflare IP ban. Successful non-empty response
bodies decode into the supplied target; a nil target discards the body.

## Quick Start

```go
package main

import (
	"context"
	"fmt"

	"github.com/discord-go/discord.go/rest"
)

func main() {
	c := rest.New("bot-token", nil, nil)
	fmt.Println(c.BaseURL, c.AuthMode == rest.AuthBot)
	c.SetBearerToken("oauth-token")
	fmt.Println(c.AuthMode == rest.AuthBearer)
	c.SetBotToken("bot-token")
	fmt.Println(c.AuthMode == rest.AuthBot, context.Background() != nil)
}
```

The example only changes local configuration. Endpoint calls require a
caller-owned context and valid credentials.

## Creating A REST Client

Supply a custom `ratelimit.Limiter` to share bucket state or a custom
`http.Client` to use a proxy, test server, or metrics transport. The custom
HTTP client must implement `Do`, `Get`, and `Post` from [`../http/`](../http/README.md).
Changing `BaseURL` is useful in tests but should not be used to bypass Discord
route validation in production.

## Using Endpoint Families

Applications and commands use `CreateGlobalApplicationCommand`, guild command
methods, application emoji and role-connection methods. Channels and messages
use `GetChannel`, `CreateMessage`, `CreateMessageComplex`, edit/delete,
reactions, pins, invites, and typing methods. Guilds and members use guild,
role, ban, prune, onboarding, integration, and member methods. Threads include
start, archive-list, join, leave, and member operations. Webhooks support
management, token execution, follow-ups, and webhook messages. Expressions and
events include emojis, stickers, soundboard, AutoMod, scheduled events, and
stage instances. Gateway, voice, OAuth-related, subscription, SKU, and
entitlement methods complete the resource surface. The exact family split and
method names are listed in [`endpoints.md`](endpoints.md).

## Common Patterns

Use typed parameter structs for writes. Use `rest.WithReason(ctx, reason)` for
requests that support an audit-log reason. Use `RequestNoAuth` for interaction
webhook URLs, not for ordinary bot endpoints. Check returned pointers before
use and use `errors.As` for `*APIError` or `*CaptchaError`.

## Best Practices

Pass a deadline-bearing context on every call. Share one client per token when
possible so the limiter and invalid-request guard have complete information.
Use idempotent payloads when a timeout makes a result uncertain. Never log
tokens, interaction tokens, or webhook tokens.

## Common Mistakes

`New` does not start a Gateway and does not validate the token. Bearer mode is
not automatic after an OAuth2 exchange. Omitted attachments are removed during
message edits. A 429 is not a normal success and may be retried after the
server-supplied delay; a context can still cancel that wait.

## API Walkthrough

The API is split into `Client` and auth constructors (`New`, `SetToken`,
`SetBearerToken`, `SetBotToken`), request/multipart helpers, typed endpoint
methods, response models, create/modify/query parameter structs, `APIError`
and `CaptchaError`, audit-reason helpers, attachment helpers, and
attachment-size constants. [`requests.md`](requests.md),
[`uploads.md`](uploads.md), and [`endpoints.md`](endpoints.md) document these
subsections separately so the full exported surface remains navigable.

## Examples

The Quick Start program is complete and runnable. The request and upload pages
contain local examples that exercise serialization without Discord credentials.

## Related APIs

- [`requests.md`](requests.md)
- [`ratelimits.md`](ratelimits.md)
- [`uploads.md`](uploads.md)
- [`endpoints.md`](endpoints.md)
