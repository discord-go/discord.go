# Distributed rate limiting

`ratelimit/redislimiter` coordinates REST requests from multiple processes using
atomic Redis scripts. Workers using the same bot must use the same namespace
and Redis 7 or newer database. Choose a different namespace for each bot; use a stable
application identifier, not the bot token. Namespace and route values are
hashed before they are stored in Redis.

The package accepts an `EvalFunc`, so applications can use their existing Redis
client without adding a Redis client dependency to discord.go. For example,
with `github.com/redis/go-redis/v9`:

```go
eval := func(ctx context.Context, script string, keys []string, args ...any) (any, error) {
    return redisClient.Eval(ctx, script, keys, args...).Result()
}
limiter, err := redislimiter.New(eval, redislimiter.Options{
    Namespace: "my-bot-production",
    ProbeLease: 5 * time.Minute,
})
if err != nil {
    return err
}
client := rest.New(botToken, limiter, nil)
requestCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()
_, err = client.GetCurrentUser(requestCtx)
return err
```

Import `context`, `time`, `github.com/discord-go/discord.go/rest`, and
`github.com/discord-go/discord.go/ratelimit/redislimiter`. Construct and
configure `redisClient` using the selected client's connection, authentication,
and TLS options. The evaluator must return Redis command errors as Go errors.

The limiter uses Redis server time, a shared budget of 50 requests in any
rolling second, and server-reported route buckets and cooldowns. Unknown and
expired buckets permit one discovery request at a time. Workers poll while
waiting; cancellation and Redis failures return errors. Idle route metadata is
removed opportunistically as requests arrive.
`GlobalLimit` can be raised if Discord has granted the bot a higher global
limit; use the same value in every worker sharing the namespace. The namespace
records the first configured value and rejects a mismatch. To change it, stop
the old workers, wait for their one-second budget and any Discord-reported
cooldowns to expire, and then start the new workers with a new namespace.

## Operational requirements

- Give every REST request a deadline shorter than `ProbeLease`; admission
  returns `ErrRequestDeadline` if it is missing or too far away. A lease expires
  after a worker crash; a request that continues beyond it can overlap a new
  discovery request. The default lease is five minutes.
- Use a Redis deployment that preserves its data across failover and does not
  evict limiter keys. A restart, failover with lost writes, or key eviction can
  reset budgets. Redis availability is required for request admission.
- A backend error while recording a response is returned to the caller. The
  Discord request may already have succeeded, so treat retries of non-idempotent
  requests accordingly. `LastError` also reports errors from the legacy update
  interface and failed reservation releases.
- Route aliases are shared after Discord supplies their bucket identifier.
  Unknown aliases may initially be discovered independently. Discord can still
  return 429s; the REST client records and respects their retry delay.
- The global budget applies to interaction endpoints as well. The invalid
  request safety counter in the REST client remains process-local, and
  unauthenticated requests may be subject to an IP-wide Discord limit.
- Separate bot namespaces do not coordinate shared-scope resource limits with
  each other. Applications using multiple bots against the same shared resource
  must account for those limits separately.
- All workers must use the same `GlobalLimit`, `CacheTTL`, and `ProbeLease` settings. Redis
  metadata is pruned during subsequent activity, so an idle database can retain
  expired fields until another request arrives. The small per-namespace global
  configuration key remains until an administrator removes it.

For direct limiter use, call `AcquireRoute`, record the response with
`UpdateRouteContext`, and invoke the returned release function on every exit
path. The older `Wait` and `WaitRoute` interfaces cannot retain an in-flight
discovery reservation across an HTTP request.
