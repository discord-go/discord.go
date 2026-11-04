# Go 1.22 MLS implementation

This local module implements the MLS protocol used by the adjacent DAVE
session module. It follows [RFC 9420](https://www.rfc-editor.org/rfc/rfc9420)
and adapts the upstream
[`mls-go`](https://github.com/thomas-vilte/mls-go) implementation for Go
1.22 compatibility.

The compatibility port replaces the Go 1.26 `crypto/hpke` and `crypto/hkdf`
APIs, newer ECDSA key parsing APIs, and benchmark-only `testing.B.Loop` calls.
The internal HPKE implementation is adapted from the Go Authors' RFC 9180
implementation; its BSD license notice is preserved in
[`internal/hpke/LICENSE`](internal/hpke/LICENSE).

Run the MLS tests from this directory:

```sh
go test ./...
```

This module is currently maintained as part of the parent repository and is
not published independently. The upstream project remains the protocol
reference; see its [source](https://github.com/thomas-vilte/mls-go) and the
included license for attribution and terms.
