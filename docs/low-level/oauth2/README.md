# OAuth2

## Overview

The `oauth2` package implements Discord's OAuth2 authorization-code helpers.
It builds authorization URLs, exchanges and refreshes codes, revokes tokens,
and reads the current user, user guilds, current application, and current
authorization through bearer requests. OAuth2 credentials are separate from a
bot token and should normally be handled by a server.

## Architecture

`Config` contains `ClientID`, `ClientSecret`, `RedirectURI`, and an optional
`*http.Client`. `New(config)` installs `http.DefaultClient` when the client is
nil and returns `*oauth2.Client`, which embeds the config. The package constants
are `AuthorizeURL`, `TokenURL`, and `RevokeURL`.

`AuthorizationURL` encodes client ID, response type `code`, space-separated
scopes, optional redirect URI, and optional state. `ExchangeCode` and
`RefreshToken` POST form data to the token endpoint and decode `TokenResponse`.
`RevokeToken` posts the client credentials and token. Bearer resource methods
set `Authorization: Bearer <accessToken>` and return typed models or an
`AuthorizationInfo` value. HTTP status 400 and above become formatted errors;
the package does not expose a special error type.

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/discord-go/discord.go/oauth2"
)

func main() {
	client := oauth2.New(oauth2.Config{
		ClientID: "1234567890",
		RedirectURI: "https://example.test/oauth/callback",
	})
	url := client.AuthorizationURL([]string{"identify", "guilds"}, "csrf-state")
	fmt.Println(url)
}
```

This program only builds a URL. Token exchange and bearer methods require a
real HTTP client and credentials; inject an HTTP client in `Config` when
testing so calls do not reach Discord.

## Creating Authorization URLs

Pass scopes as individual strings; the implementation joins them with spaces
and URL-encodes the result. State is omitted when empty, but applications
should use a cryptographically random state value and verify it on callback.
The redirect URI is omitted when empty and otherwise must match the registered
Discord redirect URI.

## Using Tokens

`TokenResponse` includes access token, token type, expiry seconds, refresh
token, and scope string. Store refresh tokens securely and account for token
rotation. `CurrentUser` returns `*users.User`, `CurrentUserGuilds` returns
`[]guilds.Guild`, `CurrentApplication` returns `*application.Application`, and
`CurrentAuthorization` returns `*AuthorizationInfo`. These methods do not
refresh tokens automatically.

## Common Patterns

Keep the authorization callback on the server, exchange the short-lived code
there, and issue an application session rather than returning the client
secret. Use `RefreshToken` before expiry and replace stored credentials with
the response values. Use `RevokeToken` during disconnect or account removal.

## Best Practices

Never put `ClientSecret` in a bot binary, browser bundle, or log. Use context
deadlines on every network method. Validate state, redirect URI, scopes, and
the returned token type before granting access. Treat guild lists as snapshots
and refresh them when authorization changes.

## Common Mistakes

OAuth2 bearer tokens are not bot tokens and must not be passed to a bot-token
REST client without calling `SetBearerToken`. An empty `state` disables the
CSRF correlation value. The package does not cache, revoke automatically, or
interpret `ExpiresIn` into a timer.

## API Walkthrough

The complete exported API is `AuthorizeURL`, `TokenURL`, `RevokeURL`, `Config`,
`Client`, `New`, `TokenResponse`, `AuthorizationInfo`,
`AuthorizationURL`, `ExchangeCode`, `RefreshToken`, `RevokeToken`,
`CurrentUser`, `CurrentUserGuilds`, `CurrentApplication`, and
`CurrentAuthorization`.

## Examples

The Quick Start program is complete and runnable without network access. REST
bearer authentication is described in [`../rest/README.md`](../rest/README.md).

## Related APIs

- [`../rest/`](../rest/README.md)
- [`../users/`](../users/README.md)
- [`../application/`](../application/README.md)
