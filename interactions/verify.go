package interactions

import (
	"crypto/ed25519"
	"encoding/hex"
)

// VerifySignature verifies an HTTP request signature using ed25519.
func VerifySignature(publicKeyHex string, timestamp string, signatureHex string, body []byte) bool {
	publicKey, err := hex.DecodeString(publicKeyHex)
	if err != nil || len(publicKey) != ed25519.PublicKeySize {
		return false
	}

	signature, err := hex.DecodeString(signatureHex)
	if err != nil || len(signature) != ed25519.SignatureSize {
		return false
	}

	message := append([]byte(timestamp), body...)
	return ed25519.Verify(publicKey, message, signature)
}
