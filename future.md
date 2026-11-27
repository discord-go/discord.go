# Future work

These are priorities for future development, not completed features.

## REST coordination and distributed rate limits

- Continue testing concurrent requests, out-of-order responses, and changing
  server limits. Verify whether interaction endpoints can safely bypass the
  bot global token bucket.
- Test the opt-in Redis limiter with failover, lost writes, outages, different
  clients, and multi-host deployments. Local tests cover a three-node cluster
  and a temporary outage; `ratelimit.Store` alone does not provide atomic
  reservations or cross-process global coordination.
- Bound route and bucket state growth without discarding active cooldowns or
  in-flight reservations.

## Voice and MLS

- Investigate recording packet gaps observed during live voice playback and
  compare networks and voice regions. Repeat sustained listener and recorder
  rejoin tests; see `docs/examples/templates/voice-test-results.md`.
- Continue MLS protocol and security review, including interoperability and
  commit validation. The fixes and remaining review areas are described in
  `dave-go/PORTING.md`.

## Discord API coverage

Review API changes and the coverage inventory, and add useful endpoints,
models, and tests. `Application.FlagsNew` remains an `int` for source
compatibility, so 32-bit builds cannot decode values above the signed `int`
range. Design a compatible full-width representation or address it in a
breaking release.

## Benchmarks

Use repeatable equivalent workloads, including complete model decoding and
representative bot traffic. Record hardware, Go version, allocations, and
multiple runs; separate CPU and request overhead from live network latency and
Discord rate limits. Exploratory comparisons are kept locally under the
ignored `benchmarks/` directory and are not part of the release.
