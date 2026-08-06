# OAuth2 Authorization Code Flow

## Overview

OAuth2 lets a user authorize an application without giving the application a
bot token. The current `discord.go/oauth2` package provides typed helpers for
the authorization URL, authorization-code exchange, refresh, revocation, and
Bearer-authenticated Discord API calls.

This guide implements the web authorization-code flow:

1. Generate a cryptographically random `state` value.
2. Redirect the browser to `oauth2.Client.AuthorizationURL`.
3. Verify the returned `state` and exchange the one-time code with
   `Client.ExchangeCode`.
4. Call `CurrentUser` or `CurrentUserGuilds` with the returned access token.
5. Keep access and refresh tokens server-side, encrypted at rest, and out of
   logs and browser-visible URLs.

OAuth2 authentication and bot authentication are different. A user access
token is sent as `Authorization: Bearer <token>` by the package's typed
methods; it is not a replacement for the token passed to `bot.New`.

## Architecture

```text
browser
  |
  | GET /login
  v
Go HTTP server -- state cookie + server-side state record
  |
  | redirect to Discord
  v
Discord authorize endpoint
  |
  | GET /callback?code=...&state=...
  v
Go HTTP server -- verify state -- exchange code
  |
  v
oauth2.Client -- token endpoint -- access token
  |
  v
CurrentUser / CurrentUserGuilds / CurrentAuthorization
```

The Quick Start keeps pending state in memory and does not create a session
cookie containing a token. That is enough for one process and local testing.
For multiple replicas, put short-lived state records in a shared store or
route the callback to the same instance. For a signed-in session, store an
opaque session ID in a secure cookie and keep the token record on the server.

## Prerequisites

- Go `1.26.4` or a compatible newer toolchain, as declared by
  [`go.mod`](../../../go.mod).
- A Discord application with an OAuth2 redirect URI exactly matching
  `OAUTH2_REDIRECT_URI`.
- `OAUTH2_CLIENT_ID` and `OAUTH2_CLIENT_SECRET` from the Developer Portal.
- A local browser for `http://localhost:8080`, or HTTPS and a trusted proxy in
  a deployed environment.
- The `identify` scope for `CurrentUser`; add `guilds` for
  `CurrentUserGuilds`.

The example does not require a bot token or Gateway intents. OAuth2 client
secrets are confidential; public client IDs are not a substitute for secrets.

## Quick Start

The following is a complete runnable HTTP server. It uses the repository's
current `oauth2.Config`, `oauth2.New`, `AuthorizationURL`, `ExchangeCode`,
`CurrentUser`, and `CurrentUserGuilds` APIs. It prints only the authorized
user's public username and ID, never the token.

Configure the same callback URL in the Discord Developer Portal, then run:

```bash
export OAUTH2_CLIENT_ID='your-client-id'
export OAUTH2_CLIENT_SECRET='your-client-secret'
export OAUTH2_REDIRECT_URI='http://localhost:8080/callback'
go run oauth2-example.go
```

Open <http://localhost:8080/login> in a browser.

```go
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/discord-go/discord.go/oauth2"
)

const stateLifetime = 5 * time.Minute

type app struct {
	oauth  *oauth2.Client
	mu     sync.Mutex
	states map[string]time.Time
}

func newState() (string, error) {
	data := make([]byte, 32)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return hex.EncodeToString(data), nil
}

func (a *app) rememberState(state string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	for value, expiresAt := range a.states {
		if now.After(expiresAt) {
			delete(a.states, value)
		}
	}
	a.states[state] = now.Add(stateLifetime)
}

func (a *app) consumeState(state string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	expiresAt, ok := a.states[state]
	if !ok || time.Now().After(expiresAt) {
		delete(a.states, state)
		return false
	}
	delete(a.states, state)
	return true
}

func (a *app) home(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintln(w, `<a href="/login">Sign in with Discord</a>`)
}

func (a *app) login(w http.ResponseWriter, r *http.Request) {
	state, err := newState()
	if err != nil {
		http.Error(w, "could not create state", http.StatusInternalServerError)
		return
	}
	a.rememberState(state)
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   int(stateLifetime.Seconds()),
		HttpOnly: true,
		Secure:   os.Getenv("OAUTH_COOKIE_SECURE") == "1",
		SameSite: http.SameSiteLaxMode,
	})
	url := a.oauth.AuthorizationURL([]string{"identify", "guilds"}, state)
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (a *app) callback(w http.ResponseWriter, r *http.Request) {
	if providerError := r.URL.Query().Get("error"); providerError != "" {
		http.Error(w, "Discord authorization was declined", http.StatusBadRequest)
		return
	}
	queryState := r.URL.Query().Get("state")
	cookie, err := r.Cookie("oauth_state")
	if err != nil || queryState == "" || cookie.Value != queryState || !a.consumeState(queryState) {
		http.Error(w, "invalid OAuth2 state", http.StatusBadRequest)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "authorization code is missing", http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "oauth_state", Value: "", Path: "/", MaxAge: -1, HttpOnly: true})
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	token, err := a.oauth.ExchangeCode(ctx, code)
	if err != nil {
		http.Error(w, "code exchange failed", http.StatusBadGateway)
		log.Printf("oauth2 exchange: %v", err)
		return
	}
	user, err := a.oauth.CurrentUser(ctx, token.AccessToken)
	if err != nil {
		http.Error(w, "user lookup failed", http.StatusBadGateway)
		log.Printf("oauth2 user lookup: %v", err)
		return
	}
	guilds, err := a.oauth.CurrentUserGuilds(ctx, token.AccessToken)
	if err != nil {
		http.Error(w, "guild lookup failed", http.StatusBadGateway)
		log.Printf("oauth2 guild lookup: %v", err)
		return
	}

	// A real application stores an opaque session ID and encrypts token data on
	// the server. This example deliberately does not persist a token.
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "Signed in as %s (%s)\nGuilds returned: %d\n", user.Username, user.ID, len(guilds))
}

func main() {
	clientID := os.Getenv("OAUTH2_CLIENT_ID")
	clientSecret := os.Getenv("OAUTH2_CLIENT_SECRET")
	redirectURI := os.Getenv("OAUTH2_REDIRECT_URI")
	if redirectURI == "" {
		redirectURI = "http://localhost:8080/callback"
	}
	if clientID == "" || clientSecret == "" {
		log.Fatal("OAUTH2_CLIENT_ID and OAUTH2_CLIENT_SECRET are required")
	}

	application := &app{
		oauth: oauth2.New(oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURI:  redirectURI,
			HTTPClient:   &http.Client{Timeout: 10 * time.Second},
		}),
		states: make(map[string]time.Time),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", application.home)
	mux.HandleFunc("/login", application.login)
	mux.HandleFunc("/callback", application.callback)
	server := &http.Server{Addr: ":8080", Handler: mux}

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("OAuth2 server listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	select {
	case err := <-serverErrors:
		log.Fatal(err)
	case <-stop:
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}
}
```

## Practical Levels

### Basic: identity sign-in

Request only `identify`, call `ExchangeCode`, then `CurrentUser`. Keep the
access token in server memory only while testing. This is enough for a login
button that identifies an account but does not call guild APIs.

### Intermediate: guild-aware authorization

Add the `guilds` scope and call `CurrentUserGuilds`. Treat the returned list as
a snapshot: it can change after authorization, and it does not prove that a
user has a particular permission in your application. Check guild ownership or
permissions through the appropriate Discord data and your own authorization
rules.

Persist a server-side session record containing an encrypted refresh token,
the access-token expiry, user ID, granted scopes, and a session version. The
current `TokenResponse` exposes `ExpiresIn`; record the issue time yourself if
you need to refresh before expiry.

### Advanced: refresh, revoke, and multi-replica operation

Use `RefreshToken` when the access token expires and replace both token values
if Discord returns a new refresh token. Use `RevokeToken` on sign-out or when a
session is invalidated. Serialize refresh operations per session so two
requests do not race and overwrite a newer refresh token.

For multiple web replicas, use a shared `storage.Store` for one-time state and
sessions. Use a short TTL for state, encrypt token fields before `Set`, and
delete sessions on revocation. Never store raw tokens in a browser cookie.

## Best Practices

- Generate state with `crypto/rand`, bind it to the browser session, expire it,
  and consume it exactly once.
- Register the exact redirect URI in the Developer Portal and send the same
  value in `oauth2.Config.RedirectURI`.
- Use a bounded `http.Client` and request context, as the Quick Start does.
- Request the smallest scopes needed and handle denied or partially granted
  authorization.
- Keep client secrets, access tokens, and refresh tokens out of logs, metrics,
  URLs, source control, and client-side storage.
- Store an opaque session identifier in a `Secure`, `HttpOnly`, `SameSite=Lax`
  cookie in production. Set `OAUTH_COOKIE_SECURE=1` behind HTTPS.
- Encrypt refresh tokens at rest and restrict access to the encryption key.
- Refresh once per session at a time and rotate stored refresh tokens when a
  response supplies a replacement.
- Revoke tokens on explicit sign-out when the product's security model needs
  immediate invalidation.
- Validate redirect requests, use HTTPS outside localhost, and add CSRF,
  session fixation, and login-rate-limit protections.

## Common Mistakes

### Using a bot token in an OAuth2 Bearer request

Wrong:

```go
user, err := client.CurrentUser(ctx, os.Getenv("DISCORD_TOKEN"))
```

Correct:

```go
token, err := client.ExchangeCode(ctx, code)
if err != nil {
	return err
}
user, err := client.CurrentUser(ctx, token.AccessToken)
```

`bot.New` uses a bot token for Gateway authentication. `oauth2.Client` uses a
user access token for OAuth2 resource requests.

### Skipping or reusing state

Checking only that a `state` parameter exists does not prevent login CSRF. The
value must be unpredictable, tied to the initiating browser, stored briefly,
and deleted before or during callback handling.

### Logging the token response

Do not log `%+v` for `TokenResponse`; it contains access and refresh tokens.
Log a request ID, Discord status, and a redacted error instead.

### Calling a scoped endpoint without its scope

`CurrentUserGuilds` requires `guilds`. If a user authorized only `identify`, the
endpoint can fail even though the code exchange succeeded. Request and verify
scopes intentionally.

### Treating `ExpiresIn` as a durable expiry timestamp

`ExpiresIn` is a duration in seconds from token issuance. Store an issue time or
compute the expiry immediately; do not persist it as if it were an absolute
timestamp.

## API Walkthrough

- [`oauth2.Config`](../../../oauth2/client.go) contains `ClientID`,
  `ClientSecret`, `RedirectURI`, and an optional `HTTPClient`.
- [`oauth2.New`](../../../oauth2/client.go) defaults a nil HTTP client to
  `http.DefaultClient`.
- [`Client.AuthorizationURL`](../../../oauth2/client.go) creates the Discord
  authorization URL with scopes, redirect URI, and state.
- [`Client.ExchangeCode`](../../../oauth2/client.go) exchanges an authorization
  code for `TokenResponse`.
- [`Client.RefreshToken`](../../../oauth2/client.go) refreshes an access token.
- [`Client.RevokeToken`](../../../oauth2/client.go) revokes a token.
- [`Client.CurrentUser`](../../../oauth2/client.go) calls `/users/@me` with a
  Bearer token.
- [`Client.CurrentUserGuilds`](../../../oauth2/client.go) calls
  `/users/@me/guilds` and requires the `guilds` scope.
- [`Client.CurrentApplication`](../../../oauth2/client.go) and
  `Client.CurrentAuthorization` expose the application and authorization
  metadata endpoints.
- [`rest.Client.SetBearerToken`](../../../rest/client.go) configures the
  repository's general REST client for Bearer authentication when a resource
  call is not covered by the OAuth2 helper.

The package uses the current Discord API v10 endpoints and sends client ID,
client secret, code, and redirect URI as form data to the token endpoint. Use
the typed helper rather than duplicating this exchange in handlers.

## Examples

Refresh a server-side session with a bounded context and replace the stored
token record atomically in the session repository:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

next, err := client.RefreshToken(ctx, session.RefreshToken)
if err != nil {
	return err
}
session.AccessToken = next.AccessToken
session.RefreshToken = next.RefreshToken
session.ExpiresAt = time.Now().Add(time.Duration(next.ExpiresIn) * time.Second)
return saveSession(ctx, session)
```

Revoke before deleting a session when immediate Discord-side invalidation is
required:

```go
if err := client.RevokeToken(ctx, session.AccessToken); err != nil {
	return err
}
return deleteSession(ctx, session.ID)
```

## Related Links

- [`oauth2.Client`](../../../oauth2/client.go)
- [`oauth2` tests](../../../oauth2/client_test.go)
- [`rest.Client.SetBearerToken`](../../../rest/client.go)
- [`storage.Store`](../../../storage/store.go)
- [Sharding guide](sharding.md)
- [Discord OAuth2 documentation](https://discord.com/developers/docs/topics/oauth2)
- [Discord OAuth2 scopes](https://discord.com/developers/docs/topics/oauth2#shared-resources-oauth2-scopes)
- [Discord user API](https://discord.com/developers/docs/resources/user)
