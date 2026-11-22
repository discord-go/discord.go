// Package ciphersuite implements cryptographic operations used by MLS cipher
// suites.
//
// It provides key derivation, signatures, authenticated encryption, and HPKE
// operations for the supported suites. Check IsSupported before selecting a
// suite.
//
// See https://www.rfc-editor.org/rfc/rfc9420#section-5.
package ciphersuite
