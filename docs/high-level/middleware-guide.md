# Middleware Guide

## Overview

Middleware wraps command handlers to add cross-cutting concerns like
permission checks, validation, cooldowns, and logging. discord.go provides
built-in middleware and a simple interface for custom middleware.

## Middleware Chain

Middleware is a function that wraps a `CommandHandler`:

```go
type Middleware func(CommandHandler) CommandHandler
type CommandHandler func(*InteractionContext)
```

Middleware is applied in registration order. Each middleware wraps the next
handler, forming a chain. The last middleware registered runs first (outermost
wrapper).

## Built-in Middleware

### Validate

Runs custom validation before the command handler. A non-nil error is returned
to the user as an ephemeral message:

```go
router.CommandWithMiddleware("ban", "Ban a user",
    handler,
    bot.Validate(func(ctx *bot.InteractionContext) error {
        if ctx.Member == nil {
            return errors.New("this command must be used in a server")
        }
        return nil
    }),
)
```

### RequirePermissions

Checks that the invoking member has all specified Discord permissions:

```go
router.CommandWithMiddleware("kick", "Kick a member",
    kickHandler,
    bot.RequirePermissions(permissions.KickMembers),
)
```

### RequireAnyPermissions

Allows the command when the member has at least one of the supplied permission
bits:

```go
router.CommandWithMiddleware("moderate", "Moderation tools",
    modHandler,
    bot.RequireAnyPermissions(permissions.KickMembers | permissions.BanMembers),
)
```

### RequireBotPermissions

Checks the permissions Discord granted to the bot:

```go
router.CommandWithMiddleware("purge", "Bulk delete messages",
    purgeHandler,
    bot.RequireBotPermissions(permissions.ManageMessages),
)
```

### Cooldown

Rate-limits command usage per user:

```go
router.CommandWithMiddleware("daily", "Claim daily reward",
    dailyHandler,
    bot.Cooldown(24 * time.Hour),
)
```

## Custom Middleware

Write custom middleware by implementing `func(CommandHandler) CommandHandler`:

```go
func LoggingMiddleware(next bot.CommandHandler) bot.CommandHandler {
    return func(ctx *bot.InteractionContext) {
        start := time.Now()
        next(ctx)
        log.Printf("command %s took %v", ctx.CommandName(), time.Since(start))
    }
}
```

Apply it to a specific command:

```go
router.CommandWithMiddleware("ping", "Check bot status",
    pingHandler,
    LoggingMiddleware,
)
```

Or apply to all commands on a router:

```go
router.Use(LoggingMiddleware)
```

## Middleware Ordering

Middleware wraps the handler. When multiple middleware are applied, the last
registered runs first (outermost):

```go
router.CommandWithMiddleware("ban", "Ban a user",
    banHandler,
    bot.RequirePermissions(permissions.BanMembers), // runs second
    LoggingMiddleware,                              // runs first
)
```

Execution order: LoggingMiddleware -> RequirePermissions -> banHandler.

## Combining with Router

- `router.Use(middleware)` applies middleware to all commands on the router.
- `router.CommandWithMiddleware(name, desc, handler, middleware...)` applies
  middleware to a specific command.

Router-level middleware runs before command-level middleware.

## Common Patterns

- Use `RequirePermissions` for moderation commands.
- Use `Validate` for input validation before the handler.
- Use `Cooldown` for commands that should be rate-limited per user.
- Combine multiple middleware for layered checks.

## Best Practices

- Put permission checks before expensive validation.
- Return ephemeral errors for user-facing messages.
- Keep middleware fast; avoid network calls in middleware when possible.
- Use `router.Use` for global concerns like logging.

## Common Mistakes

- Forgetting that middleware order matters (last registered runs first).
- Not returning after sending an ephemeral error in middleware.
- Using `router.Command` instead of `router.CommandWithMiddleware` when
  middleware is needed.
