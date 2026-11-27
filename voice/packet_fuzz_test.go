package voice

import (
	"bytes"
	"testing"
)

// FuzzParseRTPHeader checks malformed packet handling and the fixed header's
// round-trip representation. Extension, CSRC and padding bytes remain payload.
func FuzzParseRTPHeader(f *testing.F) {
	f.Add([]byte{})
	f.Add(NewRTPHeader(12, 34, 56).Marshal())
	f.Add([]byte{0x90, 0x78, 0, 1, 0, 0, 0, 2, 0, 0, 0, 3, 0xbe, 0xde, 0, 0})
	f.Fuzz(func(t *testing.T, packet []byte) {
		if len(packet) > 65536 {
			return
		}
		original := append([]byte(nil), packet...)
		header, err := ParseRTPHeader(packet)
		if err == nil {
			if header == nil || header.Version != 2 || len(packet) < 12 {
				t.Fatal("successful parse returned an invalid header")
			}
			if got := header.Marshal(); !bytes.Equal(got, packet[:12]) {
				t.Fatalf("header round trip = %x, want %x", got, packet[:12])
			}
		}
		if !bytes.Equal(packet, original) {
			t.Fatal("parsing modified the input")
		}
	})
}
