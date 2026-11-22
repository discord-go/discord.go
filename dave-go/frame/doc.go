// Package frame implements the DAVE encrypted media frame format.
//
// A frame contains encrypted and unencrypted regions followed by a footer
// with a truncated authentication tag, nonce, range metadata, and marker.
// Unencrypted regions are authenticated as additional data.
//
// See the DAVE protocol's payload format:
// https://github.com/discord-go/discord.go/blob/main/dave-go/protocol.md.
package frame
