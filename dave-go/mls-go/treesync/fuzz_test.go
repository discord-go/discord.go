package treesync

import (
	"testing"

	"github.com/discord-go/discord.go/dave-go/mls-go/ciphersuite"
)

func FuzzUnmarshalRatchetTree(f *testing.F) {
	f.Add([]byte{0x00, 0x01, 0x00, 0x01, 0x02, 0x03, 0x04})
	f.Add([]byte{})

	f.Fuzz(func(_ *testing.T, data []byte) {
		_, _ = UnmarshalTreeFromExtension(data, ciphersuite.MLS128DHKEMP256)
	})
}
