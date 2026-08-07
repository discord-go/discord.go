# Best Practices

## Overview

This page consolidates best practices for building production Discord bots with
discord.go.

## Token Security

- Load tokens from environment variables, not config files.
- Never commit tokens to version control.
- Use `ConfigFromEnv` or `os.Getenv("DISCORD_TOKEN")`.
- Rotate tokens if exposed.

## Context Usage

- Use `context.WithTimeout` for all REST calls.
- Use `context.WithCancel` for long-running operations.
- Never use `context.Background()` for unbounded network calls.
- Use the context-accepting variants (`RequestGuildMembersContext`,
  `JoinVoiceChannelContext`) for gateway operations.

## Command Design

- Use guild command sync during development for fast iteration.
- Switch to global command sync for production.
- Add descriptions to all commands and options.
- Use middleware for permission checks and validation.
- Set cooldowns on commands that should be rate-limited.

## Event Handling

- Keep handlers fast; defer slow work to goroutines.
- Set `WithMaxHandlerConcurrency` to bound goroutine creation.
- Handle panics via the error handler.
- Use collectors for button/select/menu flows with timeouts.

## REST Usage

- Reuse a single `rest.Client`; do not create new clients per request.
- Use bulk endpoints instead of individual calls.
- Set audit log reasons with `rest.WithReason`.
- Check for `*rest.APIError` to handle API-specific errors.
- Do not bypass the rate limiter with raw HTTP calls.

## Caching

- Use the memory cache for development.
- Implement a TTL or LRU cache for long-running bots.
- Do not rely on cache for critical data; fall back to REST.
- Treat cache as eventually consistent.

## Sharding

- Use auto-detection (`WithShards(0)`) for growing bots.
- Monitor per-shard latency.
- Do not change shard count without restarting.
- Respect `max_concurrency` from the `gateway/bot` response.

## Voice

- Join via the main gateway first, then create the voice client.
- Use DAVE end-to-end encryption when available.
- Send Opus frames at 20ms intervals.
- Close the voice client on shutdown.

## Error Handling

- Use `errors.As` to check for `APIError` and `CaptchaError`.
- Classify errors as retryable or fatal before retrying.
- Log structured fields for observability.
- Do not retry on 401/403; check credentials instead.

## Security

- Use `interactions.VerifyRequest` for interaction verification.
- Use `interactions.Server` as an `http.Handler` that verifies signatures and
  timestamps automatically.
- Generate and verify OAuth2 state parameters using `oauth2.GenerateState()`.
- Use HTTPS for all webhook endpoints.
- Rate-limit incoming interaction endpoints.
- Store bot tokens in environment variables, not config files.
- Use `SetToken`, `SetBearerToken`, or `SetBotToken` to configure credentials.
  The token is stored in an unexported field and cannot be read by external code.
- `bot.Start` validates token format and returns `ErrInvalidToken` for malformed
  tokens.

## Deployment

- Use `client.RunContext(ctx)` for service-managed lifecycle.
- Monitor gateway latency and REST latency.
- Set `GOMAXPROCS` appropriately for your container.
- Use graceful shutdown to drain event handlers.
