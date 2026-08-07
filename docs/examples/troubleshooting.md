# Troubleshooting

## Overview

This page covers common errors and their solutions when using discord.go.

## Bot Connects But Commands Don't Appear

**Cause:** Global commands take up to one hour to propagate.

**Fix:** Use guild command sync during development:

```go
bot.WithGuildCommandSync(guildID)
```

## 401 Unauthorized on REST Calls

**Cause:** Invalid or expired token, or wrong auth mode.

**Fix:**
- Verify the token is correct and not expired.
- Ensure `rest.New` defaults to `AuthBot` mode.
- Use `client.SetBearerToken` for OAuth2 bearer tokens.

## 403 Forbidden on Interaction Responses

**Cause:** Interaction tokens expire after 15 minutes, or the initial response
deadline (3 seconds) was missed.

**Fix:**
- Respond within 3 seconds using `ctx.Reply` or `ctx.Defer`.
- After deferral, use `ctx.Followup` for additional messages.
- Do not send a second initial response after `ctx.Defer`.

## 429 Too Many Requests

**Cause:** Rate limit exceeded.

**Fix:** The library handles rate limits automatically. If you still see 429s:
- Reduce concurrency on batch operations.
- Add delays between REST calls in tight loops.
- Check if you are bypassing the rate limiter with raw HTTP calls.

## Gateway Disconnects with Code 4004

**Cause:** Authentication failed.

**Fix:** Verify the bot token is valid and not expired. Regenerate the token in
the Discord Developer Portal if needed.

## Gateway Disconnects with Code 4014

**Cause:** Disallowed intents.

**Fix:** Enable the required privileged intents in the Discord Developer Portal
under Bot > Privileged Gateway Intents. Common privileged intents:
- `MessageContent` (for prefix commands)
- `GuildMembers` (for member lists)
- `GuildPresences` (for presence data)

## Gateway Disconnects with Code 4009

**Cause:** Session timed out.

**Fix:** The library automatically reconnects and resumes. If this recurs
frequently, check network stability.

## Bot Token in Error Logs

**Cause:** An error message contains the token (e.g., from an HTTP response
echoing the Authorization header).

**Fix:** The default error handler redacts tokens automatically. If using a
custom error handler, use `bot.redactToken` or avoid logging raw error strings.

## Commands Not Being Synced

**Cause:** `CommandSyncDisabled` is set, or the application ID is not yet
known.

**Fix:**
- Ensure `commandSync.Mode` is not `CommandSyncDisabled`.
- Wait for the `READY` event before syncing (the library does this
  automatically).
- Check the error handler for sync errors.

## Voice Connection Fails

**Cause:** Missing `GuildVoiceStates` intent, or voice server update not
received.

**Fix:**
- Add `intents.GuildVoiceStates` to the bot intents.
- Ensure you handle `VOICE_SERVER_UPDATE` events.
- Check that the bot has permission to join the voice channel.

## High Memory Usage

**Cause:** Cache grows unbounded without TTL or size limits.

**Fix:**
- Implement a cache with TTL.
- Use an LRU cache with max-size eviction.
- Monitor cache size in production.

## Handler Panics

**Cause:** A panic in an event handler.

**Fix:** The bot recovers panics by default and reports them via the error
handler. Check the error handler output for `HandlerPanicError`:

```go
client.OnError(func(err error) {
    var panicErr *bot.HandlerPanicError
    if errors.As(err, &panicErr) {
        log.Printf("panic in %s: %v", panicErr.Event, panicErr.Value)
    }
})
```

## OAuth2 State Mismatch

**Cause:** The state parameter returned by Discord does not match the stored
value.

**Fix:** Ensure the state is stored correctly in the user session before
redirecting to Discord. Use a cryptographically random state per request.
Use `oauth2.GenerateState()` to generate a random 16-byte hex-encoded state
string.

## ErrInvalidToken on Start

**Cause:** The token does not have the expected three-segment format
(separated by dots).

**Fix:** Verify the token is copied correctly from the Discord Developer
Portal. Bot tokens have three dot-separated segments. Check for leading/trailing
whitespace or missing characters. Use `os.Getenv("DISCORD_TOKEN")` rather than
hardcoding the token.
