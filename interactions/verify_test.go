package interactions

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"testing"
)

func TestVerifySignature(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	publicKeyHex := hex.EncodeToString(publicKey)
	timestamp := "1620000000"
	body := []byte(`{"type": 1}`)

	message := append([]byte(timestamp), body...)
	signature := ed25519.Sign(privateKey, message)
	signatureHex := hex.EncodeToString(signature)

	t.Run("Valid Signature", func(t *testing.T) {
		if !VerifySignature(publicKeyHex, timestamp, signatureHex, body) {
			t.Errorf("Expected signature to be valid")
		}
	})

	t.Run("Invalid Body", func(t *testing.T) {
		if VerifySignature(publicKeyHex, timestamp, signatureHex, []byte(`{"type": 2}`)) {
			t.Errorf("Expected signature to be invalid due to tampered body")
		}
	})

	t.Run("Invalid Timestamp", func(t *testing.T) {
		if VerifySignature(publicKeyHex, "1620000001", signatureHex, body) {
			t.Errorf("Expected signature to be invalid due to tampered timestamp")
		}
	})

	t.Run("Invalid Public Key Length", func(t *testing.T) {
		if VerifySignature(publicKeyHex[:10], timestamp, signatureHex, body) {
			t.Errorf("Expected signature to be invalid due to short public key")
		}
	})

	t.Run("Invalid Hex Public Key", func(t *testing.T) {
		if VerifySignature("invalidhex", timestamp, signatureHex, body) {
			t.Errorf("Expected signature to be invalid due to invalid hex format")
		}
	})

	t.Run("Invalid Signature Length", func(t *testing.T) {
		if VerifySignature(publicKeyHex, timestamp, signatureHex[:10], body) {
			t.Errorf("Expected signature to be invalid due to short signature length")
		}
	})
}
