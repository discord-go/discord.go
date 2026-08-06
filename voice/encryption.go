package voice

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"errors"
)

// Encrypter handles audio encryption for voice.
type Encrypter interface {
	Encrypt(message, nonce []byte) []byte
	Decrypt(box, nonce []byte) ([]byte, error)
}

// AES-256-GCM encrypter for aead_aes256_gcm_rtpsize mode.
type aesGCMEncrypter struct {
	gcm cipher.AEAD
}

// NewAESGCMEncrypter creates an AES-256-GCM encrypter.
func NewAESGCMEncrypter(key [32]byte) (Encrypter, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &aesGCMEncrypter{gcm: gcm}, nil
}

func (e *aesGCMEncrypter) Encrypt(message, nonce []byte) []byte {
	// GCM nonce must be 12 bytes
	n := make([]byte, 12)
	copy(n, nonce)
	return e.gcm.Seal(nil, n, message, nil)
}

func (e *aesGCMEncrypter) Decrypt(box, nonce []byte) ([]byte, error) {
	n := make([]byte, 12)
	copy(n, nonce)
	out, err := e.gcm.Open(nil, n, box, nil)
	if err != nil {
		return nil, errors.New("voice: AES-GCM decryption failed")
	}
	return out, nil
}

// EncryptAEAD encrypts an opus frame for aead_aes256_gcm_rtpsize mode.
// header is the 12-byte RTP header (used as AAD), opus is the payload.
// Returns the encrypted packet: header + ciphertext + 4-byte nonce suffix.
func EncryptAEAD(key [32]byte, header []byte, opus []byte, nonceCounter uint32) ([]byte, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Build 12-byte nonce: 4-byte big-endian counter, zero-padded to 12 bytes
	nonce := make([]byte, 12)
	binary.BigEndian.PutUint32(nonce[0:4], nonceCounter)

	// Encrypt with AAD = RTP header
	ciphertext := gcm.Seal(nil, nonce, opus, header)

	// Build final packet: header + ciphertext + 4-byte nonce suffix
	nonceSuffix := make([]byte, 4)
	binary.BigEndian.PutUint32(nonceSuffix, nonceCounter)

	packet := make([]byte, len(header)+len(ciphertext)+4)
	copy(packet, header)
	copy(packet[len(header):], ciphertext)
	copy(packet[len(header)+len(ciphertext):], nonceSuffix)

	return packet, nil
}

// DecryptAEAD decrypts an opus frame for aead_aes256_gcm_rtpsize mode.
func DecryptAEAD(key [32]byte, header []byte, ciphertext []byte, nonce []byte) ([]byte, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return gcm.Open(nil, nonce, ciphertext, header)
}
