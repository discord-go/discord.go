package treesync

import (
	"encoding/binary"
	"testing"

	"github.com/discord-go/discord.go/dave-go/mls-go/ciphersuite"
)

func TestUnmarshalTreeRejectsImpossibleLeafCounts(t *testing.T) {
	for _, count := range []uint32{0, 1, 2, 0x7fffffff, 0x80000000, 0xfa0000fa, 0xffffffff} {
		data := make([]byte, 4)
		binary.BigEndian.PutUint32(data, count)
		if _, err := UnmarshalTree(data, ciphersuite.MLS128DHKEMP256); err == nil {
			t.Errorf("accepted %d leaves without node data", count)
		}
		if _, err := UnmarshalTreeFromExtension(data, ciphersuite.MLS128DHKEMP256); err == nil {
			t.Errorf("extension parser accepted %d leaves without node data", count)
		}
	}
}
