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

These modules are nested in the parent repository and currently use local
`replace` directives. Before publishing a parent-library release, publish the
nested modules with repository-prefixed tags such as `dave-go/v0.1.0`,
`dave-go/mls-go/v0.1.0`, and `dave-go/godave/v0.1.0`.

## References

- Discord DAVE protocol: https://github.com/discord/dave-protocol
- Upstream Go DAVE implementation: https://github.com/thomas-vilte/dave-go
- MLS standard: https://www.rfc-editor.org/rfc/rfc9420.html
- HPKE standard: https://www.rfc-editor.org/rfc/rfc9180.html
