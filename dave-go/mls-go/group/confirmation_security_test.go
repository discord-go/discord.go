package group

import (
	"errors"
	"strings"
	"testing"

	"github.com/discord-go/discord.go/dave-go/mls-go/ciphersuite"
	"github.com/discord-go/discord.go/dave-go/mls-go/treesync"
)

func TestInvalidConfirmationDoesNotExposeSecrets(t *testing.T) {
	alice, bob, alicePriv, _ := setupTwoMemberGroup(t)
	key := ciphersuite.NewSignaturePrivateKey(alicePriv.SignatureKey)
	staged, err := alice.Commit(key, key.PublicKey(), nil)
	if err != nil {
		t.Fatal(err)
	}
	content := *staged.authenticatedContent
	content.Auth.ConfirmationTag = append([]byte(nil), content.Auth.ConfirmationTag...)
	if len(content.Auth.ConfirmationTag) == 0 {
		t.Fatal("missing confirmation tag")
	}
	content.Auth.ConfirmationTag[0] ^= 1
	err = bob.ProcessReceivedCommit(&content, treesync.LeafIndex(content.Content.Sender.LeafIndex), bob.MyLeafEncryptionKey())
	if !errors.Is(err, ErrConfirmationTagMismatch) {
		t.Fatalf("expected confirmation mismatch, got %v", err)
	}
	for _, secret := range []string{"initSecret", "commitSecret", "init_secret", "commit_secret"} {
		if strings.Contains(err.Error(), secret) {
			t.Fatalf("error exposes %s", secret)
		}
	}
}
