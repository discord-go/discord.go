# OAuth2 Tutorial

## Overview

This tutorial walks through a complete OAuth2 flow with a web server, including
authorization URL generation, code exchange, token refresh, and API calls.

## Prerequisites

- A Discord application with a client ID and client secret.
- A redirect URI configured in the Discord Developer Portal.
- An HTTP server (e.g., `net/http`).

## Setup

```go
import "github.com/discord-go/discord.go/oauth2"

oauth2Client := oauth2.New(oauth2.Config{
    ClientID:     os.Getenv("DISCORD_CLIENT_ID"),
    ClientSecret: os.Getenv("DISCORD_CLIENT_SECRET"),
    RedirectURI:  "http://localhost:8080/callback",
})
```

## Step 1: Generate Authorization URL

```go
func loginHandler(w http.ResponseWriter, r *http.Request) {
    state := generateState() // cryptographically random
    // Store state in session
    session.Set(r, "oauth_state", state)

    url := oauth2Client.AuthorizationURL([]string{"identify", "guilds"}, state)
    http.Redirect(w, r, url, http.StatusFound)
}
```

## Step 2: Handle Callback

```go
func callbackHandler(w http.ResponseWriter, r *http.Request) {
    // Verify state to prevent CSRF
    expectedState := session.Get(r, "oauth_state")
    receivedState := r.URL.Query().Get("state")
    if expectedState != receivedState {
        http.Error(w, "invalid state", http.StatusBadRequest)
        return
    }

    code := r.URL.Query().Get("code")
    if code == "" {
        http.Error(w, "missing code", http.StatusBadRequest)
        return
    }

    // Exchange code for token
    ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
    defer cancel()

    token, err := oauth2Client.ExchangeCode(ctx, code)
    if err != nil {
        http.Error(w, "token exchange failed", http.StatusInternalServerError)
        return
    }

    // Store token (access_token, refresh_token, expires_in)
    session.Set(r, "access_token", token.AccessToken)
    session.Set(r, "refresh_token", token.RefreshToken)
    http.Redirect(w, r, "/dashboard", http.StatusFound)
}
```

## Step 3: Use the Access Token

```go
func dashboardHandler(w http.ResponseWriter, r *http.Request) {
    accessToken := session.Get(r, "access_token")

    ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
    defer cancel()

    user, err := oauth2Client.CurrentUser(ctx, accessToken)
    if err != nil {
        http.Error(w, "failed to get user", http.StatusUnauthorized)
        return
    }

    fmt.Fprintf(w, "Hello, %s!", user.Username)
}
```

## Step 4: Refresh Token

```go
func refreshIfNeeded(refreshToken string) (*oauth2.TokenResponse, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    return oauth2Client.RefreshToken(ctx, refreshToken)
}
```

## Step 5: Revoke Token

```go
func logoutHandler(w http.ResponseWriter, r *http.Request) {
    token := session.Get(r, "access_token")

    ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
    defer cancel()

    oauth2Client.RevokeToken(ctx, token)
    session.Clear(r)
    http.Redirect(w, r, "/", http.StatusFound)
}
```

## Security Notes

- Always generate and verify the `state` parameter to prevent CSRF.
- Use HTTPS for the redirect URI in production.
- Store the client secret on the server, never in client-side code.
- Set timeouts on all OAuth2 HTTP requests.
- Store refresh tokens encrypted at rest.

## Common Patterns

- Use `identify` scope to get user info.
- Use `guilds` scope to list user guilds.
- Combine scopes: `[]string{"identify", "guilds", "email"}`.
- Check `expires_in` and refresh before expiry.

## Best Practices

- Use short-lived sessions and refresh tokens.
- Revoke tokens on logout.
- Use the minimum required scopes.
- Monitor for 401 responses indicating token expiry.

## Common Mistakes

- Not verifying the state parameter (CSRF vulnerability).
- Exposing the client secret in client-side code.
- Not setting timeouts on OAuth2 requests.
- Storing access tokens in localStorage (use httpOnly cookies).
