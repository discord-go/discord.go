package session

import (
	"bytes"
	"context"
	"testing"

	"github.com/discord-go/discord.go/dave-go/mls-go/framing"
)

// TestExternalAddProposalRoundTripStable checks that re-marshaling an external
// Add proposal (as processProposalBatchLocked does before computing its
// ProposalRef) preserves the exact AuthenticatedContent bytes. If it does not,
// the bot's ProposalRef differs from the gateway's and the gateway silently
// rejects every commit that references the proposal.
func TestExternalAddProposalRoundTripStable(t *testing.T) {
	cb := &commitCapturingCallbacks{}
	s := New("1520751988764049418", cb)
	defer s.Close()
	s.SetChannelID(1572834879408705586)
	s.OnSelectProtocolAck(1)

	externalPkg, extPriv := buildExternalSenderPackageWithKey(t)
	s.OnDaveMLSExternalSenderPackage(externalPkg)

	human := newMLSClientForTest(t, "1210474295008952332")
	humanKP, err := human.FreshKeyPackageBytes(context.Background())
	if err != nil {
		t.Fatalf("human key package: %v", err)
	}
	groupID, err := human.CreateGroupWithExternalSender(context.Background(), []byte("1572834879408705586"), humanKP, externalPkg)
	if err != nil {
		t.Fatalf("CreateGroupWithExternalSender: %v", err)
	}
	_, welcome, err := human.InviteMember(context.Background(), groupID, cb.lastKeyPackage())
	if err != nil {
		t.Fatalf("InviteMember(bot): %v", err)
	}
	s.OnDaveMLSWelcome(0, welcome)
	s.OnDaveExecuteTransition(0)

	recorderCB := &commitCapturingCallbacks{}
	recorder := New("1527737202484973588", recorderCB)
	defer recorder.Close()
	recorder.SetChannelID(1572834879408705586)
	recorder.OnSelectProtocolAck(1)

	original := externalAddProposalForKP(t, s, extPriv, recorderCB.lastKeyPackage())

	msg, err := framing.UnmarshalMLSMessage(original)
	if err != nil {
		t.Fatalf("UnmarshalMLSMessage(original): %v", err)
	}
	remarshaled := msg.Marshal()
	if !bytes.Equal(original, remarshaled) {
		t.Fatalf("re-marshaled proposal differs from original:\n original=%x\n remarshaled=%x", original, remarshaled)
	}

	ac := func(raw []byte) []byte {
		m, err := framing.UnmarshalMLSMessage(raw)
		if err != nil {
			t.Fatalf("UnmarshalMLSMessage: %v", err)
		}
		pm := m.PublicMessage
		acForRef := &framing.AuthenticatedContent{
			WireFormat: framing.WireFormatPublicMessage,
			Content:    pm.Content,
			Auth:       pm.Auth,
		}
		return acForRef.Marshal()
	}
	if !bytes.Equal(ac(original), ac(remarshaled)) {
		t.Fatal("AuthenticatedContent differs after re-marshal; ProposalRef would diverge")
	}
}
