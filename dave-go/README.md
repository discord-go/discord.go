# dave-go

This folder contains a Go 1.22-compatible DAVE implementation for
`discord.go`. It is based on the public DAVE protocol and uses the upstream
`dave-go`, `mls-go`, and `godave` projects as references. Upstream versions that
require newer toolchains are not used as dependencies.

The module includes the DAVE voice session, an MLS implementation, and the
small voice-session interface it needs. Each is kept as a local module so the
protocol layers and their licenses remain clear.

## Compatibility work

- MLS HPKE uses a Go 1.22-compatible implementation in
  `mls-go/internal/hpke`, adapted from the Go Authors' RFC 9180 implementation
  and covered by the upstream MLS interoperability vectors.
- HKDF uses `golang.org/x/crypto`; elliptic key parsing uses Go 1.22 APIs.
- The DAVE and MLS module tests pass with the Go 1.22 Docker toolchain.

See [PORTING.md](PORTING.md) for the porting notes and limits. This folder is
not a standalone published module yet; the parent library uses local module
replacements while the integration is validated.
