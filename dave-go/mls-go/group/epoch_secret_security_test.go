package group

import (
	"bytes"
	"testing"

	"github.com/discord-go/discord.go/dave-go/mls-go/ciphersuite"
	"github.com/discord-go/discord.go/dave-go/mls-go/credentials"
	"github.com/discord-go/discord.go/dave-go/mls-go/keypackages"
	"github.com/discord-go/discord.go/dave-go/mls-go/schedule"
)

func TestInitialEpochSecretsAreNotDerivedFromPublicInputs(t *testing.T) {
	for _, cs := range []ciphersuite.CipherSuite{
		ciphersuite.MLS128DHKEMX25519,
		ciphersuite.MLS128DHKEMP256,
		ciphersuite.MLS128DHKEMX25519ChaCha20,
		ciphersuite.MLS256DHKEMP521AES256GCM,
	} {
		t.Run(cs.String(), func(t *testing.T) {
			cred, _, err := credentials.GenerateCredentialWithKeyForCS([]byte("creator"), cs)
			if err != nil {
				t.Fatal(err)
			}
			kp, priv, err := keypackages.Generate(cred, cs)
			if err != nil {
				t.Fatal(err)
			}
			id, err := NewGroupIDRandom()
			if err != nil {
				t.Fatal(err)
			}
			first, err := NewGroup(id, cs, kp, priv)
			if err != nil {
				t.Fatal(err)
			}
			second, err := NewGroup(id, cs, kp, priv)
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Equal(first.epochSecrets.EncryptionSecret.AsSlice(), second.epochSecrets.EncryptionSecret.AsSlice()) {
				t.Fatal("groups with identical public inputs have identical secrets")
			}
			public := schedule.NewKeySchedule(cs, ciphersuite.ZeroSecretCS(cs))
			public.SetCommitSecret(ciphersuite.ZeroSecretCS(cs))
			context := first.groupContext.Marshal()
			if _, err := public.ComputeJoinerSecret(context); err != nil {
				t.Fatal(err)
			}
			if _, err := public.ComputePskSecret(nil); err != nil {
				t.Fatal(err)
			}
			if _, err := public.ComputeEpochSecret(context); err != nil {
				t.Fatal(err)
			}
			rebuilt, err := public.DeriveEpochSecrets()
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Equal(first.epochSecrets.InitSecret.AsSlice(), rebuilt.InitSecret.AsSlice()) ||
				bytes.Equal(first.epochSecrets.EncryptionSecret.AsSlice(), rebuilt.EncryptionSecret.AsSlice()) {
				t.Fatal("initial secrets can be rebuilt from public inputs")
			}
		})
	}
}

func TestWelcomeDerivesMatchingSecretsWithoutCopying(t *testing.T) {
	alice, bob := newTestUser(t, "alice"), newTestUser(t, "bob")
	id, err := NewGroupIDRandom()
	if err != nil {
		t.Fatal(err)
	}
	creator, err := NewGroup(id, ciphersuite.MLS128DHKEMP256, alice.kp, alice.priv)
	if err != nil {
		t.Fatal(err)
	}
	// Rebuild the former zero-seeded epoch using only the public context.
	public := schedule.NewKeySchedule(creator.cipherSuite, ciphersuite.ZeroSecretCS(creator.cipherSuite))
	public.SetCommitSecret(ciphersuite.ZeroSecretCS(creator.cipherSuite))
	initialContext := creator.groupContext.Marshal()
	if _, err := public.ComputeJoinerSecret(initialContext); err != nil {
		t.Fatal(err)
	}
	if _, err := public.ComputePskSecret(nil); err != nil {
		t.Fatal(err)
	}
	if _, err := public.ComputeEpochSecret(initialContext); err != nil {
		t.Fatal(err)
	}
	publicEpoch, err := public.DeriveEpochSecrets()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := creator.AddMember(bob.kp); err != nil {
		t.Fatal(err)
	}
	commit, err := creator.Commit(alice.sigPriv, alice.sigPub, nil)
	if err != nil {
		t.Fatal(err)
	}
	joinerSecret := commit.JoinerSecret()
	if err := creator.MergeCommit(commit); err != nil {
		t.Fatal(err)
	}
	// An add-only commit can omit the update path. Its next epoch must still
	// depend on the creator's private initial secret.
	publicNext := schedule.NewKeySchedule(creator.cipherSuite, publicEpoch.InitSecret)
	publicNext.SetCommitSecret(ciphersuite.ZeroSecretCS(creator.cipherSuite))
	newContext := creator.groupContext.Marshal()
	if _, err := publicNext.ComputeJoinerSecret(newContext); err != nil {
		t.Fatal(err)
	}
	if _, err := publicNext.ComputePskSecret(nil); err != nil {
		t.Fatal(err)
	}
	if _, err := publicNext.ComputeEpochSecret(newContext); err != nil {
		t.Fatal(err)
	}
	publicNextSecrets, err := publicNext.DeriveEpochSecrets()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(publicNextSecrets.EncryptionSecret.AsSlice(), creator.epochSecrets.EncryptionSecret.AsSlice()) {
		t.Fatal("add-only epoch secrets can be rebuilt from public inputs")
	}
	welcome, err := creator.CreateWelcomeWithOpts([]*keypackages.KeyPackage{bob.kp}, alice.sigPriv, WithJoinerSecret(joinerSecret), WithStagedCommit(commit))
	if err != nil {
		t.Fatal(err)
	}
	joiner, err := JoinFromWelcome(welcome, bob.kp, bob.priv, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(creator.epochSecrets.EncryptionSecret.AsSlice(), joiner.epochSecrets.EncryptionSecret.AsSlice()) ||
		!bytes.Equal(creator.epochSecrets.InitSecret.AsSlice(), joiner.epochSecrets.InitSecret.AsSlice()) {
		t.Fatal("joiner did not derive the creator's epoch secrets")
	}
}

func TestReinitSecretsRequireMoreThanResumptionPSK(t *testing.T) {
	kp, priv := makeSingleKP(t)
	proposal := makeReInitProposal([]byte("successor"))
	psk := ciphersuite.NewSecret(bytes.Repeat([]byte{7}, proposal.CipherSuite.HashLength()))
	g, err := NewGroupFromReInit(proposal, psk, []byte("previous"), 1, kp, priv)
	if err != nil {
		t.Fatal(err)
	}
	rawPSK := g.keySchedule.GetRawPskSecret()
	if rawPSK == nil {
		t.Fatal("missing reinit PSK material")
	}
	public := schedule.NewKeySchedule(proposal.CipherSuite, ciphersuite.ZeroSecretCS(proposal.CipherSuite))
	public.SetCommitSecret(ciphersuite.ZeroSecretCS(proposal.CipherSuite))
	gc := g.groupContext.Marshal()
	if _, err := public.ComputeJoinerSecret(gc); err != nil {
		t.Fatal(err)
	}
	if err := public.SetPskSecretDirect(rawPSK); err != nil {
		t.Fatal(err)
	}
	if _, err := public.ComputeEpochSecret(gc); err != nil {
		t.Fatal(err)
	}
	rebuilt, err := public.DeriveEpochSecrets()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(g.epochSecrets.EncryptionSecret.AsSlice(), rebuilt.EncryptionSecret.AsSlice()) {
		t.Fatal("reinit epoch can be rebuilt using zero seed and the PSK")
	}
}
