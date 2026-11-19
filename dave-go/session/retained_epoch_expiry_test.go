package session

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/discord-go/discord.go/dave-go/frame"
	"github.com/discord-go/discord.go/dave-go/godave"
	"github.com/discord-go/discord.go/dave-go/mediakeys"
)

func TestDecryptRejectsExpiredRetainedEpoch(t *testing.T) {
	for _, tt := range []struct {
		name    string
		expiry  time.Time
		allowed bool
	}{
		{"retained", time.Now().Add(time.Minute), true},
		{"expired", time.Now().Add(-time.Second), false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := New("sender", testCallbacks{})
			defer s.Close()
			ratchet, err := mediakeys.NewKeyRatchet(bytes.Repeat([]byte{0xaa}, 16))
			if err != nil {
				t.Fatal(err)
			}
			key, err := ratchet.GetKey(0)
			if err != nil {
				t.Fatal(err)
			}
			packet, err := frame.Encrypt(frame.EncryptParams{Plaintext: []byte("audio"), Key: key})
			if err != nil {
				t.Fatal(err)
			}
			s.retainedEpoch = []*epochState{{expiresAt: tt.expiry, senders: map[godave.UserID]*senderState{
				"sender": {ratchet: ratchet, expander: mediakeys.NewNonceExpander()},
			}}}
			_, err = s.Decrypt("sender", packet, make([]byte, 256))
			if tt.allowed && err != nil {
				t.Fatalf("retained frame: %v", err)
			}
			if !tt.allowed && !errors.Is(err, ErrDecryptionFailed) {
				t.Fatalf("expired frame: %v", err)
			}
		})
	}
}
