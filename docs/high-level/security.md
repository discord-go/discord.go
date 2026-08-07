# Security

## Overview

discord.go handles authentication credentials, cryptographic signatures, and
network connections to Discord. This page covers the security practices every
application should follow when using the library.

## Token Handling

Bot tokens, OAuth2 client secrets, and webhook tokens are credentials. Never
commit them to version control.

Load tokens from environment variables:

```go
token := os.Getenv("DISCORD_TOKEN")
if token == "" {
    log.Fatal("DISCORD_TOKEN is required")
}
client := bot.New(token, bot.WithIntents(intents.Guilds))
```

Or use `ConfigFromEnv`:

```go
config := bot.ConfigFromEnv()
client := bot.NewFromConfig(config)
```

The `Config.Token` field has a JSON tag for config-file loading, but putting
tokens in config files risks accidental commit. Prefer environment variables or
a secrets manager. If you must use a config file, add it to `.gitignore`.

The bot token is stored in unexported fields on `rest.Client` and
`gateway.Client`. It is set via `rest.New`, `SetToken`, `SetBearerToken`, or
`SetBotToken` on the REST client, and via `SetToken` on the gateway client.
No code with a reference to these clients can read the token directly.

Token format is validated on `bot.Start`. A token that does not have three
dot-separated segments returns `ErrInvalidToken` immediately, giving a clear
error instead of an opaque identify failure.

The default error handler redacts the bot token from log output by replacing
occurrences with `[REDACTED]`.

## Interaction Signature Verification

Discord signs interaction webhook requests with Ed25519. Always verify the
signature before processing the request body.

Use `interactions.VerifyRequest` (not `VerifySignature`) to prevent replay
attacks. `VerifyRequest` checks both the cryptographic signature and the
freshness of the timestamp:

```go
package main

import (
    "io"
    "log"
    "net/http"
    "time"

    "github.com/discord-go/discord.go/interactions"
)

var publicKey = "your-app-public-key-hex"

func interactionHandler(w http.ResponseWriter, r *http.Request) {
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "bad request", http.StatusBadRequest)
        return
    }

    timestamp := r.Header.Get("X-Signature-Timestamp")
    signature := r.Header.Get("X-Signature-Ed25519")

    if !interactions.VerifyRequest(publicKey, timestamp, signature, body, time.Now()) {
        http.Error(w, "invalid signature", http.StatusUnauthorized)
        return
    }

    // Signature is valid and timestamp is fresh. Process the interaction.
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"type":1}`)) // Pong
}

func main() {
    http.HandleFunc("/interactions", interactionHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

`VerifySignature` verifies only the Ed25519 signature. It does not check
timestamp freshness. If you use it directly, you must validate the timestamp
yourself. Prefer `VerifyRequest`.

## OAuth2 Security

### CSRF Protection with State Parameter

The `AuthorizationURL` method accepts a `state` parameter. Generate a
cryptographically random state per request, store it in the user's session,
and verify it when Discord redirects back:

```go
import "crypto/rand"

func generateState() string {
    b := make([]byte, 16)
    rand.Read(b)
    return hex.EncodeToString(b)
}

// Redirect to Discord OAuth2
state := generateState()
// Store state in user session
url := oauth2Client.AuthorizationURL([]string{"identify", "guilds"}, state)
http.Redirect(w, r, url, http.StatusFound)
```

When Discord redirects back, verify the state matches the stored value before
exchanging the code.

### Client Secret Handling

The OAuth2 client secret is sent in the request body when exchanging codes and
revoking tokens. Never expose the client secret to the browser or client-side
code. Keep OAuth2 token exchange on the server.

## Webhook Security

Discord does not sign incoming webhook payloads with Ed25519 like it does for
interactions. To secure webhook endpoints:

- Use HTTPS for the webhook URL.
- Validate the webhook token in the URL path.
- Consider rate-limiting or IP-restricting the endpoint.
- Do not trust webhook payloads without verifying the source.

Interaction webhooks (for slash commands) are signed and should use
`interactions.VerifyRequest` as shown above.

## Rate Limiting

The library handles Discord REST rate limits automatically with bucket hashing,
optimistic remaining counts, and 429 retry. Applications should still:

- Use `context.Context` with deadlines on REST calls to avoid unbounded waits.
- Avoid tight loops that make many REST requests without delays.
- Use bulk operations (e.g., `BulkOverwriteGlobalCommands`) instead of
  individual create/delete calls.

For incoming interaction endpoints, the application should implement its own
rate limiting to prevent flooding.

## Token Redaction

The `bot` package redacts the bot token from error messages before logging.
The default error handler uses `redactToken` to replace the token with
`[REDACTED]`. Custom error handlers should also avoid logging raw errors that
may contain tokens.

## Common Patterns

- Load tokens from environment variables, not config files.
- Use `VerifyRequest` for interaction verification.
- Generate and verify OAuth2 state parameters.
- Set deadlines on REST contexts.
- Use HTTPS for all webhook endpoints.

## Best Practices

- Rotate tokens if they may have been exposed.
- Use the minimum required OAuth2 scopes.
- Store OAuth2 refresh tokens encrypted at rest.
- Monitor for 401/403 responses that may indicate token expiry.

## Common Mistakes

- Using `VerifySignature` without timestamp validation, allowing replay
  attacks.
- Putting tokens in config files that get committed to git.
- Exposing OAuth2 client secrets in client-side code.
- Not verifying the OAuth2 state parameter on redirect.
