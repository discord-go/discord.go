package group

import (
	"crypto/ecdh"
	"crypto/rand"
	"testing"

	"github.com/discord-go/discord.go/dave-go/mls-go/ciphersuite"
	"github.com/discord-go/discord.go/dave-go/mls-go/keypackages"
)

func TestJoinFromWelcomeRejectsInvalidInputs(t *testing.T) {
	key, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	kp := &keypackages.KeyPackage{CipherSuite: ciphersuite.MLS128DHKEMP256}
	private := &keypackages.KeyPackagePrivateKeys{InitKey: key}
	goodSuite := &Welcome{CipherSuite: ciphersuite.MLS128DHKEMP256}
	for _, tt := range []struct {
		name    string
		welcome *Welcome
		key     *keypackages.KeyPackage
		private *keypackages.KeyPackagePrivateKeys
	}{
		{"nil welcome", nil, kp, private},
		{"unknown suite", &Welcome{CipherSuite: 0x3030}, kp, private},
		{"different suite", &Welcome{CipherSuite: ciphersuite.MLS128DHKEMX25519}, kp, private},
		{"nil key package", goodSuite, nil, private},
		{"nil private keys", goodSuite, kp, nil},
		{"nil init key", goodSuite, kp, &keypackages.KeyPackagePrivateKeys{}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := JoinFromWelcome(tt.welcome, tt.key, tt.private, nil); err == nil {
				t.Fatal("accepted invalid welcome inputs")
			}
		})
	}
}

func TestUnmarshalWelcomeRejectsTruncatedSecretReference(t *testing.T) {
	// Supported cipher suite, a one-byte secrets vector containing a truncated
	// MLS varint, and an empty encrypted group-info vector.
	if _, err := UnmarshalWelcome([]byte{0, 2, 1, 0x40, 0}); err == nil {
		t.Fatal("accepted a truncated key package reference")
	}
}
