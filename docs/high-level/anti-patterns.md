# Anti-Patterns

## Overview

This page documents common mistakes and anti-patterns when using discord.go.
Each entry explains the problem and the correct approach.

## Using VerifySignature Without Timestamp Validation

**Problem:** `interactions.VerifySignature` checks the Ed25519 signature but
does not validate timestamp freshness. An attacker who captures a valid signed
request can replay it indefinitely.

**Fix:** Use `interactions.VerifyRequest` which enforces both the signature
and a 5-minute timestamp skew check.

## Committing Tokens

**Problem:** Putting bot tokens, OAuth2 client secrets, or webhook tokens in
config files that get committed to version control.

**Fix:** Load tokens from environment variables (`ConfigFromEnv` or
`os.Getenv`). If using a config file, add it to `.gitignore` and never include
the token in the committed template.

## Using context.Background() for Unbounded Network Calls

**Problem:** Using `context.Background()` for REST calls means the request can
block indefinitely if the server hangs.

**Fix:** Use `context.WithTimeout` for all REST calls:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
client.Rest.CreateMessage(ctx, params)
```

## Sending a Second Initial Response After Defer

**Problem:** After calling `ctx.Defer()`, calling `ctx.Reply()` sends a second
initial response, which Discord rejects.

**Fix:** After deferral, use followups (`ctx.Followup`) to send additional
messages.

## Using Global Command Sync While Iterating

**Problem:** Global commands take up to an hour to propagate. Iterating on
command design with global sync wastes time waiting for propagation.

**Fix:** Use `bot.WithGuildCommandSync(guildID)` during development for
near-instant command updates.

## Not Handling Handler Panics

**Problem:** A panic in an event handler crashes the entire bot if not
recovered.

**Fix:** The bot's `invoke` method recovers panics by default and reports them
via the error handler. Ensure your error handler logs panics:

```go
client.OnError(func(err error) {
    var panicErr *bot.HandlerPanicError
    if errors.As(err, &panicErr) {
        log.Printf("panic in %s: %v", panicErr.Event, panicErr.Value)
    }
})
```

## Creating New REST Clients Per Request

**Problem:** Creating a new `rest.Client` for each REST call discards the rate
limiter state and connection pool.

**Fix:** Reuse the `rest.Client` from `bot.Rest` or a single `rest.New` call.

## Not Closing Response Bodies

**Problem:** Not closing HTTP response bodies causes connection leaks.

**Fix:** The library handles this in `rest/request.go`. If making raw HTTP calls,
always `defer resp.Body.Close()`.

## Ignoring Rate Limit Headers

**Problem:** Making REST calls without respecting rate limits causes 429 errors
and can lead to Cloudflare IP bans.

**Fix:** The library handles rate limits automatically. Do not bypass the rate
limiter by making raw HTTP calls to Discord's API.

## Using Embeds in Components V2 Messages

**Problem:** Setting `FlagIsComponentsV2` and using `embeds` in the same message
is rejected by Discord.

**Fix:** In V2 messages, use text components and media galleries instead of
embeds.

## Not Setting MaxHandlerConcurrency

**Problem:** Without a concurrency limit, a burst of gateway events can spawn
unbounded goroutines.

**Fix:** Set `bot.WithMaxHandlerConcurrency` to a reasonable limit:

```go
bot.New(token, bot.WithMaxHandlerConcurrency(100))
```

## Blocking the Event Handler

**Problem:** Long-running operations in event handlers block the dispatch
pipeline (or consume handler slots).

**Fix:** Defer the response, then do slow work in a goroutine:

```go
router.Command("search", "Search database", func(ctx *bot.InteractionContext) {
    ctx.Defer()
    go func() {
        results := expensiveDatabaseQuery(ctx)
        ctx.Followup(results)
    }()
})
```

## Not Using Context on Gateway Operations

**Problem:** `RequestGuildMembers` and `JoinVoiceChannel` use
`context.Background()` internally, so they cannot be cancelled.

**Fix:** Use the context-accepting variants `RequestGuildMembersContext` and
`JoinVoiceChannelContext`.
