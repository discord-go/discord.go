# Changelog

All notable changes to discord.go are documented here. The format follows
Keep a Changelog, and versions use the repository's `vX.Y.Z-<codename>` tags.

## [v0.18.0-walk] - 2026-11-22

[v0.18.0-walk]: https://github.com/discord-go/discord.go/compare/v0.17.1-left...v0.18.0-walk

### Fixed

- Paginator interactions now handle acknowledgement errors and timeouts more
  reliably, and clean up their collectors and controls (PR #15).
- Gateway heartbeat shutdown no longer loses a stop signal sent before the
  loop is ready, and repeated stops are safe.
- REST rate limits distinguish HTTP methods and Discord's major resources.
  Rate-limit responses update shared route or global cooldowns, including
  retry delays from response bodies. Multipart requests follow the same retry
  budget and invalid-request handling as ordinary requests.
- Waiters on an exhausted rate-limit bucket can cancel independently, and
  out-of-order responses cannot increase a known remaining count prematurely.

### Changed

- Reduced allocations for snowflake ID and ID-list JSON encoding and decoding
  while retaining the existing wire formats and error handling.
- Reviewed package documentation and corrected DAVE, MLS, rate-limit, and
  internal package descriptions. The existing package structure is unchanged.
- Documented that sharing a rate-limit store alone does not coordinate
  reservations or the global limit between processes. The default limiter
  continues to use a 50 requests/second token bucket with a burst of 50.
- Updated the DAVE, MLS, and GoDave nested module documentation and published
  corresponding module versions.
- Kept the exploratory library-comparison benchmarks local and excluded them
  from Git and CI. Their no-op limiter is for local measurements only.
- Expanded CI's Go test matrix to cover each supported version from 1.22
  through 1.26.

### Compatibility and verification

- Minimum Go version remains 1.22. The root and nested module tests were run
  with Go 1.22 through 1.26; REST and limiter changes were also race tested.
- The default limiter can still receive 429s before route limits are known and
  applies its token bucket to interaction endpoints. Separate processes need
  a distributed limiter for shared coordination.

## [v0.17.1-left] - 2026-11-19

[v0.17.1-left]: https://github.com/discord-go/discord.go/compare/v0.17.0-left...v0.17.1-left

This release removes the local `replace` directives from the published module
metadata and moves local multi-module development to a `go.work` workspace.

### Changed

- The root `go.mod` and `dave-go/go.mod` no longer contain `replace`
  directives. Local builds use `go.work`; consumers resolve the published
  `dave-go v0.2.1`, `dave-go/godave v0.2.0`, and `dave-go/mls-go v0.2.0`
  modules.
- `dave-go` is now `v0.2.1`; its only change is the cleaned `go.mod`/`go.sum`.
- The template `go.mod` files keep their local replaces because they are
  standalone example modules, not published modules.

### Fixed

- The VitePress documentation build no longer reports dead links for the
  template directories; the template index links to the GitHub directories.

## [v0.17.0-left] - 2026-11-19

[v0.17.0-left]: https://github.com/discord-go/discord.go/compare/v0.16.1-wid...v0.17.0-left

This release fixes the DAVE leave/rejoin silence, adds repeatable live voice
scenarios, introduces multi-process shard supervision, and turns the API
coverage inventory into a classified backlog.

### Fixed

- **DAVE rejoin silence.** Add-only MLS commits now omit the `UpdatePath`,
  matching libdave/mlspp's `force_path=false`. The op 28 commit+welcome payload
  drops from roughly 2.2 KB to 1.5 KB, so Discord's voice gateway accepts the
  transition and a listener who leaves and rejoins can decrypt playback again.
  Remove, Update, and empty commits still include the path.
- **DAVE transition handling.** `OnDaveExecuteTransition` now receives the
  transition ID instead of the protocol version. Retained receive epochs are
  pruned during decrypt, and a pending local commit is distinguished from a
  peer epoch before rollback.
- **Recorder WAV output.** The recorder writes a valid `fmt ` chunk, so saved
  recordings open in standard audio tools.

### Added

- **Live voice scenarios.** `run-voice-bots.sh` selects `baseline`,
  `membership`, `reconnect`, `long`, or `failure` with
  `VOICE_TEST_SCENARIO` and verifies non-silent WAVs with FFmpeg. See
  `docs/examples/templates/voice-test-plan.md`.
- **Multi-process sharding.** The optional `shardsupervisor` package adds
  health checks, restart limits, `FileCoordinator` shard claims, and
  cross-process Identify spacing. The runnable example and sharding guide
  document the deployment boundaries.
- **Beginner examples.** Added complete welcome-message and private-ticket
  examples, plus the shard-supervisor example. CI compiles the runnable
  examples and tests the nested template modules.
- **Clean-consumer release check.** `scripts/verify-release.sh` and the Release
  consumer check workflow fetch a tagged version into a fresh Go 1.22 module
  without local replacements and compile the bot, voice, and voiceplayer
  imports.
- **Voice playback guidance.** `voiceplayer` has a compile-checked `Source`
  example; the musicbot template provides the FFmpeg/local-file adapter and
  per-guild queue without adding CGO dependencies to the core package.
- **API coverage backlog.** The scanner now handles local `itoa`, string
  parameters, and no-auth/multipart request helpers. The inventory matches
  191/250 specification operations, and `docs/api-coverage.md` classifies the
  remaining gaps and lists regression tests.
- **DAVE maintenance record.** `dave-go/PORTING.md` records the upstream
  baseline, security review log, and the nested module tag procedure.

### Changed

- Nested DAVE modules are now `dave-go/v0.2.0`, `dave-go/godave/v0.2.0`, and
  `dave-go/mls-go/v0.2.0`.
- The musicbot template accepts local media files directly, in addition to
  yt-dlp URLs and searches.
- The library version constant and README badge now report `v0.17.0-left`.

### Breaking

- Custom `godave.Session` implementations must update
  `OnDaveExecuteTransition` to accept a transition ID (`uint16`) instead of a
  protocol version.

### Compatibility

- Minimum Go version: 1.22.
- Discord API v10.
