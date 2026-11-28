package interactions

import (
	"crypto/ed25519"
	"encoding/hex"
	"strconv"
	"time"
)

// MaxTimestampSkew is the maximum age a signed request timestamp may have
// relative to the verifier's clock before it is rejected as a potential
// replay. Discord signs requests with a Unix timestamp and recommends
// rejecting timestamps that are too old.
const MaxTimestampSkew = 5 * time.Minute

// VerifySignature verifies an HTTP request signature using ed25519.
//
// It does NOT validate the freshness of the timestamp. Callers that use this
// function directly are responsible for rejecting stale timestamps to prevent
// replay attacks. Prefer VerifyRequest for incoming HTTP requests, which
// enforces both the signature and a timestamp freshness check.
func VerifySignature(publicKeyHex string, timestamp string, signatureHex string, body []byte) bool {
	var publicKey [ed25519.PublicKeySize]byte
	var signature [ed25519.SignatureSize]byte
	if len(publicKeyHex) != 2*len(publicKey) || len(signatureHex) != 2*len(signature) {
		return false
	}
	if _, err := hex.Decode(publicKey[:], []byte(publicKeyHex)); err != nil {
		return false
	}
	if _, err := hex.Decode(signature[:], []byte(signatureHex)); err != nil {
		return false
	}

	message := make([]byte, len(timestamp)+len(body))
	copy(message, timestamp)
	copy(message[len(timestamp):], body)
	return ed25519.Verify(publicKey[:], message, signature[:])
}

// VerifyRequest verifies an incoming Discord interaction request, checking
// both the Ed25519 signature and the freshness of the timestamp to prevent
// replay attacks. now is the verifier's current time; pass time.Now() in
// production and a fixed value in tests.
//
// The timestamp header from Discord is a decimal Unix timestamp in seconds.
// Requests whose timestamp differs from now by more than MaxTimestampSkew in
// either direction are rejected.
func VerifyRequest(publicKeyHex string, timestamp string, signatureHex string, body []byte, now time.Time) bool {
	return freshTimestamp(timestamp, now) && VerifySignature(publicKeyHex, timestamp, signatureHex, body)
}

func freshTimestamp(timestamp string, now time.Time) bool {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}

	skew := now.Sub(time.Unix(ts, 0))
	return skew >= -MaxTimestampSkew && skew <= MaxTimestampSkew
}
