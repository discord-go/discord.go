# Next improvements checklist

This checklist records implementation and verification separately. A checked item
means the work described has been completed; live Discord scenarios require a
test server and credentials supplied outside the repository.

- [x] **Release verification:** `scripts/verify-release.sh` and the Release
  consumer check workflow fetch a tagged version into a fresh Go 1.22 module
  without local replacements and compile the bot, voice, and voiceplayer
  imports. The same fetch resolves the nested DAVE, MLS, and GoDave tags. See
  `docs/releasing.md`.
- [x] **Voice resilience:** `run-voice-bots.sh` selects baseline, membership,
  reconnect, long, and failure scenarios and verifies non-silent WAVs. The
  membership scenario was run live on 2026-11-19; the remaining scenarios
  require the dedicated test guild and credentials.
- [x] **DAVE maintenance:** `dave-go/PORTING.md` records the upstream baseline,
  security review log, the fixed rejoin issue, and the nested module tag
  procedure.
- [x] **API coverage:** the scanner handles local `itoa`, string parameters, and
  no-auth request helpers; the inventory now matches 191/250 operations.
  `docs/api-coverage.md` classifies the remaining gaps and lists regression
  tests.
- [x] **Voice playback:** `voiceplayer` has a compile-checked `Source` example;
  the musicbot template provides the FFmpeg/local-file adapter and per-guild
  queue without adding CGO dependencies to the core package.
- [x] **Multi-process sharding:** `shardsupervisor` adds health checks, restart
  limits, `FileCoordinator` shard claims, and Identify spacing; the example and
  sharding guide document the deployment boundaries.
- [x] **Examples:** CI compiles the runnable examples and tests the template
  modules; the welcome and tickets examples are added and indexed.
