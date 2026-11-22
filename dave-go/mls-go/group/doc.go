// Package group implements the MLS group state machine.
//
// It handles group membership, proposals, commits, Welcome messages, and epoch
// changes. Applications using the high-level MLS client generally do not need
// to manage Group directly.
//
// See https://www.rfc-editor.org/rfc/rfc9420#section-12.
package group
