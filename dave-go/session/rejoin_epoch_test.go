package session

import (
	"context"
	"crypto/ecdsa"
	"testing"

	"github.com/discord-go/discord.go/dave-go/godave"
	"github.com/discord-go/discord.go/dave-go/mls-go"
	"github.com/discord-go/discord.go/dave-go/mls-go/ciphersuite"
	"github.com/discord-go/discord.go/dave-go/mls-go/framing"
	"github.com/discord-go/discord.go/dave-go/mls-go/group"
	"github.com/discord-go/discord.go/dave-go/mls-go/keypackages"
	memorystore "github.com/discord-go/discord.go/dave-go/mls-go/storage/memory"
)

type commitCapturingCallbacks struct {
	keyPackages [][]byte
	commits     [][]byte
}

func (c *commitCapturingCallbacks) SendMLSKeyPackage(kp []byte) error {
	c.keyPackages = append(c.keyPackages, append([]byte(nil), kp...))
	return nil
}

func (c *commitCapturingCallbacks) SendMLSCommitWelcome(payload []byte) error {
	c.commits = append(c.commits, append([]byte(nil), payload...))
	return nil
}

func (c *commitCapturingCallbacks) SendReadyForTransition(uint16) error   { return nil }
func (c *commitCapturingCallbacks) SendInvalidCommitWelcome(uint16) error { return nil }

func (c *commitCapturingCallbacks) lastKeyPackage() []byte {
	if len(c.keyPackages) == 0 {
		return nil
	}
	return c.keyPackages[len(c.keyPackages)-1]
}

func (c *commitCapturingCallbacks) lastCommit() []byte {
	if len(c.commits) == 0 {
		return nil
	}
	return c.commits[len(c.commits)-1]
}

func newMLSClientForTest(t *testing.T, identity string) *mls.Client {
	t.Helper()
	idBytes, err := userIDToIdentityBytes(godave.UserID(identity))
	if err != nil {
		t.Fatalf("userIDToIdentityBytes(%q): %v", identity, err)
	}
	store := memorystore.NewStore()
	client, err := mls.NewClient(idBytes, ciphersuite.MLS128DHKEMP256,
		mls.WithStorage(store, store),
		mls.WithCacheStrategy(mls.CacheNone),
	)
	if err != nil {
		t.Fatalf("mls.NewClient(%q): %v", identity, err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func splitCommitWelcome(t *testing.T, payload []byte) (commit, welcome []byte) {
	t.Helper()
	// The DAVE op 28 payload is commit_message || welcome_message. Parse the
	// shortest prefix that is a complete MLS commit; the remainder is the
	// Welcome body.
	for n := 4; n <= len(payload); n++ {
		msg, err := framing.UnmarshalMLSMessage(payload[:n])
		if err != nil {
			continue
		}
		if msg.PublicMessage == nil || msg.PublicMessage.Content.ContentType() != framing.ContentTypeCommit {
			continue
		}
		return payload[:n], payload[n:]
	}
	t.Fatalf("could not split commit/welcome payload (%d bytes)", len(payload))
	return nil, nil
}

func commitHasPath(t *testing.T, commitBytes []byte) bool {
	t.Helper()
	msg, err := framing.UnmarshalMLSMessage(commitBytes)
	if err != nil {
		t.Fatalf("UnmarshalMLSMessage(commit): %v", err)
	}
	pm := msg.PublicMessage
	if pm == nil {
		t.Fatal("commit is not a public message")
	}
	body, ok := pm.Content.CommitData()
	if !ok {
		t.Fatal("message is not a commit")
	}
	c, err := group.UnmarshalCommit(body)
	if err != nil {
		t.Fatalf("UnmarshalCommit: %v", err)
	}
	return c.Path != nil
}

func externalProposalBytes(t *testing.T, s *Session, extPriv *ecdsa.PrivateKey, proposal *group.Proposal) []byte {
	t.Helper()
	ctx := context.Background()

	epoch, err := s.mlsClient.client.Epoch(ctx, s.groupID)
	if err != nil {
		t.Fatalf("Epoch: %v", err)
	}
	groupInfoBytes, err := s.mlsClient.client.GroupInfo(ctx, s.groupID)
	if err != nil {
		t.Fatalf("GroupInfo: %v", err)
	}
	info, err := group.UnmarshalGroupInfo(groupInfoBytes)
	if err != nil {
		t.Fatalf("UnmarshalGroupInfo: %v", err)
	}

	content := framing.FramedContent{
		GroupID:           s.groupID,
		Epoch:             epoch,
		Sender:            framing.Sender{Type: framing.SenderTypeExternal, SenderIndex: 0},
		AuthenticatedData: []byte{},
		Body:              framing.ProposalBody{Data: group.ProposalMarshal(proposal)},
	}
	sigKey := ciphersuite.NewSignaturePrivateKey(extPriv)
	pm, err := framing.NewPublicMessage(content, sigKey, info.GroupContext.Marshal(), nil, ciphersuite.MLS128DHKEMP256)
	if err != nil {
		t.Fatalf("NewPublicMessage: %v", err)
	}
	return framing.NewMLSMessagePublic(pm).Marshal()
}

func externalAddProposalForKP(t *testing.T, s *Session, extPriv *ecdsa.PrivateKey, kpBytes []byte) []byte {
	t.Helper()
	kp, err := keypackages.UnmarshalKeyPackage(kpBytes)
	if err != nil {
		t.Fatalf("UnmarshalKeyPackage: %v", err)
	}
	return externalProposalBytes(t, s, extPriv, group.NewAddProposal(kp))
}

func externalRemoveProposal(t *testing.T, s *Session, extPriv *ecdsa.PrivateKey, identity string) []byte {
	t.Helper()
	members, err := s.mlsClient.client.ListMembers(context.Background(), s.groupID)
	if err != nil {
		t.Fatalf("ListMembers: %v", err)
	}
	want, err := userIDToIdentityBytes(godave.UserID(identity))
	if err != nil {
		t.Fatalf("userIDToIdentityBytes(%q): %v", identity, err)
	}
	for _, m := range members {
		if string(m.Identity) == string(want) {
			return externalProposalBytes(t, s, extPriv, group.NewRemoveProposal(group.NewLeafNodeIndex(m.LeafIndex)))
		}
	}
	t.Fatalf("identity %q not found in group", identity)
	return nil
}

// TestRejoinAfterRemovalKeepsEpochUsable reproduces the live failure where a
// listener leaves and rejoins: the bot must be able to commit the re-add and
// the rejoined member must be able to join the resulting welcome and decrypt
// frames from the bot.
func TestRejoinAfterRemovalKeepsEpochUsable(t *testing.T) {
	ctx := context.Background()
	cb := &commitCapturingCallbacks{}
	s := New("1520751988764049418", cb)
	defer s.Close()
	s.SetChannelID(1572834879408705586)
	s.AssignSsrcToCodec(2347, godave.CodecOpus)
	s.OnSelectProtocolAck(1)

	externalPkg, extPriv := buildExternalSenderPackageWithKey(t)
	s.OnDaveMLSExternalSenderPackage(externalPkg)

	human := newMLSClientForTest(t, "1210474295008952332")
	humanKP, err := human.FreshKeyPackageBytes(ctx)
	if err != nil {
		t.Fatalf("human key package: %v", err)
	}
	groupID, err := human.CreateGroupWithExternalSender(ctx, []byte("1572834879408705586"), humanKP, externalPkg)
	if err != nil {
		t.Fatalf("CreateGroupWithExternalSender: %v", err)
	}
	_, welcome, err := human.InviteMember(ctx, groupID, cb.lastKeyPackage())
	if err != nil {
		t.Fatalf("InviteMember(bot): %v", err)
	}
	s.OnDaveMLSWelcome(0, welcome)
	if s.Stats().WelcomesJoined != 1 {
		t.Fatalf("bot did not join initial welcome: %+v", s.Stats())
	}
	s.OnDaveExecuteTransition(0)

	// Recorder joins the established group.
	recorderCB := &commitCapturingCallbacks{}
	recorderSession := New("1527737202484973588", recorderCB)
	defer recorderSession.Close()
	recorderSession.SetChannelID(1572834879408705586)
	recorderSession.OnSelectProtocolAck(1)
	recorderSession.OnDaveMLSExternalSenderPackage(externalPkg)
	s.AddUser("1527737202484973588")
	s.OnDaveMLSProposals(buildProposalBatch(externalAddProposalForKP(t, s, extPriv, recorderCB.lastKeyPackage())))
	recorderWelcome := cb.lastCommit()
	if len(recorderWelcome) == 0 {
		t.Fatal("no commit/welcome for recorder add")
	}
	recorderCommit, recorderWelcome := splitCommitWelcome(t, recorderWelcome)
	if commitHasPath(t, recorderCommit) {
		t.Fatal("add-only commit must omit the UpdatePath (libdave force_path=false)")
	}
	if len(cb.lastCommit()) > 2048 {
		t.Fatalf("add commit+welcome payload = %d bytes, want <= 2048", len(cb.lastCommit()))
	}
	s.OnDaveExecuteTransition(1)
	recorderSession.OnDaveMLSWelcome(0, recorderWelcome)
	recorderSession.OnDaveExecuteTransition(0)
	if !recorderSession.Ready() {
		t.Fatalf("recorder session not ready after welcome: %+v", recorderSession.State())
	}

	// Human leaves: external remove proposal.
	s.RemoveUser("1210474295008952332")
	s.OnDaveMLSProposals(buildProposalBatch(externalRemoveProposal(t, s, extPriv, "1210474295008952332")))
	removeCommit := cb.lastCommit()
	if len(removeCommit) == 0 {
		t.Fatal("no commit for remove")
	}
	if !commitHasPath(t, removeCommit) {
		t.Fatal("remove commit must include an UpdatePath")
	}
	s.OnDaveExecuteTransition(2)

	// Human rejoins with a fresh key package.
	rejoinedCB := &commitCapturingCallbacks{}
	rejoined := New("1210474295008952332", rejoinedCB)
	defer rejoined.Close()
	rejoined.SetChannelID(1572834879408705586)
	rejoined.OnSelectProtocolAck(1)
	rejoined.OnDaveMLSExternalSenderPackage(externalPkg)
	s.AddUser("1210474295008952332")
	s.OnDaveMLSProposals(buildProposalBatch(externalAddProposalForKP(t, s, extPriv, rejoinedCB.lastKeyPackage())))
	rejoinWelcome := cb.lastCommit()
	if len(rejoinWelcome) == 0 {
		t.Fatal("no commit/welcome for rejoin add")
	}
	rejoinCommit, rejoinWelcome := splitCommitWelcome(t, rejoinWelcome)
	if commitHasPath(t, rejoinCommit) {
		t.Fatal("rejoin add-only commit must omit the UpdatePath")
	}
	if len(cb.lastCommit()) > 2048 {
		t.Fatalf("rejoin commit+welcome payload = %d bytes, want <= 2048", len(cb.lastCommit()))
	}
	s.OnDaveExecuteTransition(3)

	rejoined.OnDaveMLSWelcome(0, rejoinWelcome)
	rejoined.OnDaveExecuteTransition(0)
	if !rejoined.Ready() {
		t.Fatalf("rejoined session not ready after welcome: %+v", rejoined.State())
	}

	// The bot must be able to encrypt a frame the rejoined member can decrypt.
	encrypted := make([]byte, s.MaxEncryptedFrameSize(64))
	n, err := s.Encrypt(2347, []byte("audio-after-rejoin"), encrypted)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	decrypted := make([]byte, 128)
	dn, err := rejoined.Decrypt("1520751988764049418", encrypted[:n], decrypted)
	if err != nil {
		t.Fatalf("rejoined member Decrypt: %v", err)
	}
	if got := string(decrypted[:dn]); got != "audio-after-rejoin" {
		t.Fatalf("decrypted = %q", got)
	}

	// And the reverse direction: the rejoined member's frames must decrypt for
	// the bot, proving both sides derived the same epoch sender ratchets.
	const rejoinedSSRC = 9999
	rejoined.AssignSsrcToCodec(rejoinedSSRC, godave.CodecOpus)
	backEncrypted := make([]byte, rejoined.MaxEncryptedFrameSize(64))
	bn, err := rejoined.Encrypt(rejoinedSSRC, []byte("listener-audio"), backEncrypted)
	if err != nil {
		t.Fatalf("rejoined Encrypt: %v", err)
	}
	backDecrypted := make([]byte, 128)
	bdn, err := s.Decrypt("1210474295008952332", backEncrypted[:bn], backDecrypted)
	if err != nil {
		t.Fatalf("bot Decrypt: %v", err)
	}
	if got := string(backDecrypted[:bdn]); got != "listener-audio" {
		t.Fatalf("bot decrypted = %q", got)
	}
}
