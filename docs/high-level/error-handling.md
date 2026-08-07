# Error Handling Guide

## Overview

discord.go provides typed errors for REST API failures, captcha challenges, and
gateway close codes. This guide covers error classification, retry strategies,
and structured error logging.

## Error Types

### APIError

Returned when the Discord API responds with a 4xx or 5xx status:

```go
var apiErr *rest.APIError
if errors.As(err, &apiErr) {
    log.Printf("API error %d (HTTP %d): %s", apiErr.Code, apiErr.HTTPStatus, apiErr.Message)
}
```

Fields: `Code` (Discord error code), `Message`, `Errors` (detailed validation
errors), `HTTPStatus`.

### CaptchaError

Returned when Discord requires a CAPTCHA challenge:

```go
var captchaErr *rest.CaptchaError
if errors.As(err, &captchaErr) {
    log.Printf("CAPTCHA required: service=%s sitekey=%s",
        captchaErr.CaptchaService, captchaErr.CaptchaSitekey)
}
```

### Gateway Close Codes

Gateway disconnections carry close codes defined in `gateway/closecodes.go`:

| Code | Meaning | Recoverable |
|------|---------|:---:|
| 4000 | Unknown error | Yes |
| 4001 | Unknown opcode | Yes |
| 4002 | Decode error | Yes |
| 4003 | Not authenticated | Yes |
| 4004 | Authentication failed | No |
| 4007 | Invalid sequence | Yes |
| 4008 | Rate limited | Yes |
| 4009 | Session timed out | Yes |
| 4010 | Invalid shard | No |
| 4011 | Sharding required | No |
| 4013 | Invalid intents | No |
| 4014 | Disallowed intents | No |

The gateway client automatically reconnects for recoverable codes. Fatal codes
(4004, 4010, 4011, 4013, 4014) stop the connection.

## Error Classification

Classify errors to decide whether to retry:

```go
func classifyError(err error) string {
    var apiErr *rest.APIError
    if errors.As(err, &apiErr) {
        switch {
        case apiErr.HTTPStatus == 429:
            return "rate-limited" // library retries automatically
        case apiErr.HTTPStatus >= 500:
            return "server-error" // retryable
        case apiErr.HTTPStatus == 401 || apiErr.HTTPStatus == 403:
            return "auth-error" // not retryable, check token
        default:
            return "client-error" // not retryable
        }
    }
    if errors.Is(err, context.DeadlineExceeded) {
        return "timeout" // retryable
    }
    return "unknown"
}
```

## Retry Strategies

The library handles retries for:
- HTTP 429 (rate limited) with `Retry-After` header.
- Transient errors via the configured `Retrier` (default 3 retries).

For application-level retries, use exponential backoff:

```go
func withRetry(ctx context.Context, fn func() error) error {
    for attempt := 0; attempt < 3; attempt++ {
        err := fn()
        if err == nil {
            return nil
        }
        if isRetryable(err) {
            time.Sleep(backoff(attempt))
            continue
        }
        return err
    }
    return errors.New("max retries exceeded")
}
```

## Idempotent Operations

REST operations like `BulkOverwriteGlobalCommands` are idempotent: calling them
with the same data produces the same result. Safe to retry.

Non-idempotent operations like `CreateMessage` may produce duplicates on retry.
Use idempotency keys or check before retrying.

## Structured Error Logging

Use the bot's error handler for structured logging:

```go
client.OnError(func(err error) {
    var apiErr *rest.APIError
    if errors.As(err, &apiErr) {
        log.Printf("api_error code=%d http=%d msg=%s",
            apiErr.Code, apiErr.HTTPStatus, apiErr.Message)
        return
    }
    log.Printf("error: %v", err)
})
```

The default error handler redacts the bot token from error messages before
logging.

## Common Patterns

- Use `errors.As` to check for `APIError` and `CaptchaError`.
- Classify errors as retryable or fatal before retrying.
- Log error codes and HTTP status for debugging.
- Use context deadlines to prevent unbounded waits.

## Best Practices

- Don't retry on 401/403; check credentials instead.
- Don't retry on 400; fix the request.
- Retry on 429 (handled by library), 500, 502, 503.
- Use exponential backoff for application-level retries.
- Log structured fields for observability.

## Common Mistakes

- Retrying non-idempotent operations without deduplication.
- Not checking `errors.As` before accessing `APIError` fields.
- Ignoring `context.DeadlineExceeded` which indicates a timeout.
- Logging raw errors that may contain tokens.
