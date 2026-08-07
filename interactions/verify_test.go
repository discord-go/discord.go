package interactions

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"testing"
	"time"
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

func TestVerifyRequest(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	publicKeyHex := hex.EncodeToString(publicKey)
	body := []byte(`{"type": 1}`)

	// Use a timestamp close to "now" so the freshness check passes.
	now := time.Unix(1700000000, 0)
	timestamp := "1700000000"

	message := append([]byte(timestamp), body...)
	signature := ed25519.Sign(privateKey, message)
	signatureHex := hex.EncodeToString(signature)

	t.Run("Valid Fresh Request", func(t *testing.T) {
		if !VerifyRequest(publicKeyHex, timestamp, signatureHex, body, now) {
			t.Errorf("Expected fresh request to be valid")
		}
	})

	t.Run("Replay Attack - Stale Timestamp", func(t *testing.T) {
		// A validly signed request with an old timestamp must be rejected.
		staleNow := now.Add(MaxTimestampSkew + time.Second)
		if VerifyRequest(publicKeyHex, timestamp, signatureHex, body, staleNow) {
			t.Errorf("Expected stale timestamp to be rejected as a replay")
		}
	})

	t.Run("Future Timestamp Rejected", func(t *testing.T) {
		futureNow := now.Add(-(MaxTimestampSkew + time.Second))
		if VerifyRequest(publicKeyHex, timestamp, signatureHex, body, futureNow) {
			t.Errorf("Expected future timestamp to be rejected")
		}
	})

	t.Run("Invalid Timestamp Format", func(t *testing.T) {
		if VerifyRequest(publicKeyHex, "not-a-number", signatureHex, body, now) {
			t.Errorf("Expected invalid timestamp format to be rejected")
		}
	})

	t.Run("Tampered Body", func(t *testing.T) {
		if VerifyRequest(publicKeyHex, timestamp, signatureHex, []byte(`{"type": 2}`), now) {
			t.Errorf("Expected tampered body to be rejected")
		}
	})

	t.Run("Boundary - Exactly At Skew Limit", func(t *testing.T) {
		// A timestamp exactly MaxTimestampSkew old should still be accepted.
		boundaryNow := now.Add(MaxTimestampSkew)
		if !VerifyRequest(publicKeyHex, timestamp, signatureHex, body, boundaryNow) {
			t.Errorf("Expected request at exactly the skew boundary to be valid")
		}
	})
}
