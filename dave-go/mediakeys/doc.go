// Package mediakeys derives DAVE media keys from the MLS epoch exporter secret.
// It maintains sender ratchets for deriving generation-specific keys.
//
// See the DAVE protocol's sender key derivation and key rotation sections:
// https://github.com/discord-go/discord.go/blob/main/dave-go/protocol.md.
package mediakeys
