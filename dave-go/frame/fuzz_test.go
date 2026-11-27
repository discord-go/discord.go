package frame

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"
)

// FuzzULEB128MatchesUint32 compares nonce decoding with the standard decoder
// and enforces DAVE's 32-bit limit.
func FuzzULEB128MatchesUint32(f *testing.F) {
	for _, seed := range [][]byte{{0}, {0x80, 1}, {0xff, 0xff, 0xff, 0xff, 0x0f}, {0x80, 0x80, 0x80, 0x80, 0x10}, {0x80}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		want, n := binary.Uvarint(data)
		valid := n > 0 && n <= 5 && want <= math.MaxUint32
		got, consumed, err := DecodeULEB128(data)
		if valid && (err != nil || got != uint32(want) || consumed != n) {
			t.Fatalf("DecodeULEB128(%x) = %d, %d, %v; want %d, %d", data, got, consumed, err, want, n)
		}
		if !valid && err == nil {
			t.Fatalf("DecodeULEB128(%x) accepted invalid or overflowing input", data)
		}
	})
}

// FuzzParseAndDecryptFrame exercises malformed untrusted frames and ranges.
func FuzzParseAndDecryptFrame(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0x80, 13, 0xfa, 0xfa})
	valid, err := Encrypt(EncryptParams{Plaintext: []byte("opus"), Key: make([]byte, 16), TruncatedNonce: 42})
	if err != nil {
		f.Fatal(err)
	}
	f.Add(valid)
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 65536 {
			return
		}
		original := append([]byte(nil), data...)
		parsed, err := Parse(data)
		if err == nil {
			if len(parsed.Tag) != 8 || int(parsed.SupplementalSize) > len(data) {
				t.Fatalf("invalid successful parse: %+v", parsed)
			}
			if err := ValidateRanges(parsed.UnencryptedRanges, len(parsed.InterleavedFrame)); err != nil {
				t.Fatalf("successful parse produced invalid ranges: %v", err)
			}
		}
		_, _, _ = Decrypt(DecryptParams{Ciphertext: data, Key: make([]byte, 16)})
		if !bytes.Equal(data, original) {
			t.Fatal("parse or decrypt modified the input")
		}
	})
}

// FuzzEncryptDecryptFrame checks valid frames and rejects modified tags.
func FuzzEncryptDecryptFrame(f *testing.F) {
	f.Add([]byte("opus payload"), uint32(42))
	f.Add([]byte{}, uint32(0))
	f.Fuzz(func(t *testing.T, plain []byte, nonce uint32) {
		if len(plain) > 4096 {
			return
		}
		key := make([]byte, 16)
		sealed, err := Encrypt(EncryptParams{Plaintext: plain, Key: key, TruncatedNonce: nonce})
		if err != nil {
			t.Fatal(err)
		}
		got, gotNonce, err := Decrypt(DecryptParams{Ciphertext: sealed, Key: key})
		if err != nil || gotNonce != nonce || !bytes.Equal(got, plain) {
			t.Fatalf("round trip: nonce=%d error=%v", gotNonce, err)
		}
		sealed[len(plain)] ^= 1 // first byte of the authentication tag
		if _, _, err := Decrypt(DecryptParams{Ciphertext: sealed, Key: key}); err == nil {
			t.Fatal("modified authentication tag was accepted")
		}
	})
}
