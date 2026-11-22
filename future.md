# Future work

These are priorities for future development, not completed features.

## Distributed rate limiting

Build an opt-in limiter for bots running across multiple processes or servers.
It should coordinate atomic reservations for Discord route buckets and a shared
global request budget, honor server-provided limits and retry delays, and handle
cancellation and coordination failures. The existing `ratelimit.Store` alone
does not provide atomic reservations or cross-process global coordination.

## Better benchmarks

Create repeatable benchmarks for representative bot workloads and compare
libraries using equivalent inputs and rate-limit settings. Record the hardware,
Go version, allocations, and multiple runs; separate local CPU and request
overhead from network latency and live Discord limits.

## Discord API coverage

Review Discord API changes and the existing coverage inventory, then add
missing endpoints, models, and behavior where useful. Keep the coverage record,
documentation, and regression tests up to date as support is added.

Additional investigation items are listed in
[the next-session plan](docs/plans/next-session.md).
