// Package mls implements Messaging Layer Security (MLS) group operations.
//
// Client creates and joins groups, handles proposals and commits, and sends and
// receives protected messages. The default storage is in memory; applications
// can provide persistent storage with WithStorage. Group state contains
// sensitive key material and must be protected.
//
// See https://www.rfc-editor.org/rfc/rfc9420.
package mls
