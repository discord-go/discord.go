# Go 1.22 DAVE implementation

This folder contains a Go 1.22-targeted implementation of Discord's DAVE voice
encryption support. The upstream `dave-go`, `mls-go`, and `godave` projects are
protocol and API references. Their source and dependency requirements were
reviewed and adapted for this port.

## Compatibility findings

- The current upstream DAVE session module declares Go 1.26.
- Its MLS module used `crypto/hpke` and `crypto/hkdf`, which are unavailable in
  Go 1.22. This port replaces them with a compatible implementation, verified
  against the MLS interoperability vectors.
- The current GoDave module declares Go 1.24, so its small public interface is
  also carried here and targeted at Go 1.22.
- The parent library now uses `golang.org/x/time v0.10.0`, whose module
  directive supports Go 1.18.

The MLS module now has a local RFC 9180 implementation adapted from the Go
Authors' standard-library implementation. It replaces APIs unavailable in Go
1.22, including `crypto/hpke`, `crypto/hkdf`, `ecdh.KeyExchanger`, and newer
ECDSA point parsing. The MLS interoperability-vector suite passes under Go
1.22. An earlier candidate HPKE dependency implemented an obsolete draft and
failed those vectors, so it was rejected.

The DAVE and MLS package tests pass under Go 1.22. Live Discord voice
interoperability was also tested with the musicbot and recorderbot: both bots
connected, completed DAVE setup, and joined a voice channel; audio sent by the
musicbot was received and saved by the recorderbot as a WAV file. This verifies
the tested connection and audio path against Discord.

These modules are nested in the parent repository and use a `go.work`
workspace for in-repository development. Published module versions are
resolved from their respective tags. A clean consumer check compiles the
parent module and its nested dependencies with Go 1.22.0.

## References

- Discord DAVE protocol: https://github.com/discord/dave-protocol
- Upstream Go DAVE implementation: https://github.com/thomas-vilte/dave-go
- MLS standard: https://www.rfc-editor.org/rfc/rfc9420.html
- HPKE standard: https://www.rfc-editor.org/rfc/rfc9180.html

## Maintaining and releasing the port

The original port did not preserve a recorded upstream commit for every
module, so its precise ancestry has not been established. For orientation,
representative files match the structure of `thomas-vilte/dave-go` v0.5.1
(commit `7a73ceacb28b3e1f1d01fcf5e777b592eb779cd5`), its declared MLS
v1.6.0 (`9d524c5b1de7e0954b1ea976e9744460b42f5619`), and GoDave v0.2.0
(`58b63fa76f6290b8001a1753dec2a2b7b56f3542`), with local changes. These are
comparison points, not a claim of exact provenance. Compare the entire trees
before treating a particular upstream release as the port's baseline.

For every update:

1. Review Discord DAVE changes and upstream DAVE, MLS, and GoDave releases and
   security advisories. Record upstream commit IDs, local modifications, and
   licenses for any code incorporated. Review crypto changes carefully.
2. Run each nested module's tests and interoperability vectors under the minimum
   supported Go version and the current version. Run the voice integration
   scenarios in `docs/examples/templates/voice-test-plan.md` when credentials
   and a test server are available; do not claim scenarios that were not run.
3. Commit changes, create new nested module tags for each changed module, and
   push them. Update parent and nested `go.mod` requirements to those published
   versions; never move an existing published tag.
4. Push the parent version tag, run the clean consumer check described in
   `docs/releasing.md`, and publish the parent release after it passes.

### Fixed live membership issue (2026-11-19)

A listener could hear playback initially but heard silence after leaving and
rejoining. The bot's add-only MLS commit carried an `UpdatePath`, making the
op 28 `commit + welcome` payload about 2.2 KB. Discord's voice gateway silently
dropped that payload, never announced the commit, and the bot's session entered
recovery while the rejoined listener had no keys for the still-active epoch.
The fix omits the `UpdatePath` for add-only commits, matching libdave/mlspp's
`force_path=false`; Remove, Update, and empty commits still include it. The
live `membership` scenario in `docs/examples/templates/voice-test-plan.md`
verifies non-silent audio before and after the rejoin.

### Upstream baseline

| Module | Upstream | Version | Commit | Local changes |
| --- | --- | --- | --- | --- |
| DAVE session | `thomas-vilte/dave-go` | v0.5.1 | `7a73ceacb28b3e1f1d01fcf5e777b592eb779cd5` | Go 1.22 compatibility, transition-ID handling, retained epochs, pending-commit reconciliation, add-only commit path |
| MLS | `thomas-vilte/mls-go` | v1.6.0 | `9d524c5b1de7e0954b1ea976e9744460b42f5619` | Local RFC 9180 HPKE/HKDF, confirmation-error redaction, parent-hash and unmerged-leaf validation |
| GoDave | `thomas-vilte/godave` | v0.2.0 | `58b63fa76f6290b8001a1753dec2a2b7b56f3542` | Go 1.22 interface port |

These are comparison points, not a claim of exact provenance. Compare the
entire trees before treating a particular upstream release as the port's
baseline.

### Security and protocol review log

| Upstream change | Area | Status |
| --- | --- | --- |
| DAVE `e2341dc0` | Retained receive epoch expiry | Ported: retained epochs are pruned during decrypt. |
| DAVE `7a8e9201`, `9bf0fbe0` | Pending local commit vs peer epoch | Ported: rollback only when a local commit is pending. |
| MLS `0b565769` | Secrets in confirmation errors | Ported: sentinel error without secret data; debug key logging removed. |
| MLS `3a49c697` | Welcome signer validation | Ported: external sender and GroupInfo signer are checked. |
| MLS `4d4c4bd7` | Epoch-zero entropy | Review pending: confirm the local RFC 9180 path matches the upstream fix. |
| MLS `459ca333` | Proposal signatures | Ported: external proposals are verified before caching. |
| MLS `91c8bddf` | Commit paths | Ported: add-only commits omit the path; Remove/Update/empty commits keep it. |
| MLS `c4a0c225` | Rejected commit state | Ported: rejected commits roll back to the pre-commit snapshot. |
| MLS `d26fd13b` | Proposal deduplication | Review pending: confirm duplicate refs are rejected in every proposal path. |

Re-review this table against the upstream comparison histories before each
nested module release. A "ported" entry means the local tree contains an
equivalent change and a test; it is not a claim that the upstream patch was
applied byte-for-byte.

### Nested module tag procedure

The three nested modules are published independently. Tag them in dependency
order so the parent `go.mod` can require published versions:

```sh
# 1. MLS and GoDave first.
git tag -a dave-go/mls-go/vX.Y.Z -m 'dave-go/mls-go vX.Y.Z'
git tag -a dave-go/godave/vX.Y.Z -m 'dave-go/godave vX.Y.Z'
git push origin dave-go/mls-go/vX.Y.Z dave-go/godave/vX.Y.Z

# 2. DAVE session, which requires the two tags above.
git tag -a dave-go/vX.Y.Z -m 'dave-go vX.Y.Z'
git push origin dave-go/vX.Y.Z

# 3. Update the parent and nested go.mod requirements to the published
#    versions, run the clean consumer check, then tag the parent.
git tag -a vX.Y.Z-<codename> -m 'vX.Y.Z-<codename>'
git push origin vX.Y.Z-<codename>
```

Never move or delete a published tag. If a nested module did not change, keep
its existing tag and requirement. The clean consumer check in
`docs/releasing.md` fetches the parent and all nested modules without local
replacements; a missing nested tag fails that check before the GitHub release
is published.

See the upstream comparison histories for context:
[MLS changes](https://github.com/thomas-vilte/mls-go/compare/v1.6.0...main)
and [DAVE changes](https://github.com/thomas-vilte/dave-go/compare/v0.5.1...main).
Publish new nested module versions when their contents change, then update the
parent requirements and run the consumer check before releasing the parent.
