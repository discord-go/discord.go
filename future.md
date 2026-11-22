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

## Next session: follow-up work

These are candidates for investigation, not completed fixes.

1. Improve high-concurrency REST coordination when a bucket is first seen or
   resets, and test out-of-order responses and changing server limits. Review
   whether interaction endpoints can safely bypass the bot global token bucket.
2. Investigate an opt-in distributed rate limiter with atomic reservations and
   global coordination. The existing shared `Store` interface is insufficient
   on its own. Review memory growth for long-running clients with many routes.
3. Repeat live DAVE voice interoperability tests with two bots and a listener,
   including membership leave/rejoin and recording during playback. Earlier
   multi-bot rejoin results were inconsistent; record what each run verifies.
4. Continue the MLS security review listed in `dave-go/PORTING.md`, especially
   epoch-zero entropy and proposal deduplication.
5. Revisit model decoding performance with repeatable equivalent workloads.
   The exploratory comparison benchmarks are kept locally under the ignored
   `benchmarks/` directory; they are not included in the release.
