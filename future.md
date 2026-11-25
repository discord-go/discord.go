# Future work

These are priorities for future development, not completed features.

## Distributed rate limiting

The opt-in Redis-backed limiter coordinates bots running across multiple
processes or servers. Test Redis Cluster, failover, outages, and real
multi-host deployments; the current integration checks use an isolated local
Redis server. The existing `ratelimit.Store` alone does not provide atomic
reservations or cross-process global coordination.

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

1. Continue testing high-concurrency REST coordination, out-of-order responses,
   and changing server limits. Review whether interaction endpoints can safely
   bypass the bot global token bucket.
2. Exercise the opt-in distributed rate limiter across deployments and failure
   modes. Review memory growth for long-running clients with many routes.
3. Investigate packet gaps observed during the 2026-11-25 live voice test and
   repeat the comparison on another network or voice region. Listener and
   recorder rejoins succeeded, but recording continuity did not; see
   `docs/examples/templates/voice-test-results.md`.
4. Continue the MLS security review listed in `dave-go/PORTING.md`. Initial
   epoch entropy and repeated proposal handling are fixed in `mls-go v0.2.2`;
   further protocol and interoperability review remains.
5. Revisit model decoding performance with repeatable equivalent workloads.
   The exploratory comparison benchmarks are kept locally under the ignored
   `benchmarks/` directory; they are not included in the release.
