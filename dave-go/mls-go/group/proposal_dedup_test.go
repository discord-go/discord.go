package group

import (
	"testing"

	"github.com/discord-go/discord.go/dave-go/mls-go/ciphersuite"
	"github.com/discord-go/discord.go/dave-go/mls-go/framing"
)

func TestLocalProposalEchoSurvivesStateRestore(t *testing.T) {
	alice, bob := newTestUser(t, "alice"), newTestUser(t, "bob")
	id, err := NewGroupIDRandom()
	if err != nil {
		t.Fatal(err)
	}
	g, err := NewGroup(id, ciphersuite.MLS128DHKEMP256, alice.kp, alice.priv)
	if err != nil {
		t.Fatal(err)
	}
	proposal, err := g.AddMember(bob.kp)
	if err != nil {
		t.Fatal(err)
	}
	wire, err := g.SignProposalAsPublicMessage(proposal, alice.sigPriv)
	if err != nil {
		t.Fatal(err)
	}
	state, err := g.MarshalState()
	if err != nil {
		t.Fatal(err)
	}
	g, err = UnmarshalGroupState(state)
	if err != nil {
		t.Fatal(err)
	}
	message, err := framing.UnmarshalMLSMessage(wire)
	if err != nil {
		t.Fatal(err)
	}
	public, ok := message.AsPublic()
	if !ok {
		t.Fatal("expected public message")
	}
	for i := 0; i < 2; i++ {
		if err := g.ProcessPublicMessage(public); err != nil {
			t.Fatalf("echo %d: %v", i, err)
		}
	}
	if len(g.proposals.Proposals) != 1 {
		t.Fatalf("got %d pending proposals", len(g.proposals.Proposals))
	}
	commit, err := g.Commit(alice.sigPriv, alice.sigPub, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(commit.Commit().Proposals) != 1 || commit.Commit().Proposals[0].Proposal == nil {
		t.Fatal("expected the local proposal to be committed inline once")
	}
}

func TestRepeatedRemoteReferenceStoredOnce(t *testing.T) {
	alice := newTestUser(t, "alice")
	id, err := NewGroupIDRandom()
	if err != nil {
		t.Fatal(err)
	}
	g, err := NewGroup(id, ciphersuite.MLS128DHKEMP256, alice.kp, alice.priv)
	if err != nil {
		t.Fatal(err)
	}
	p := NewRemoveProposal(0)
	ref := []byte("same-reference")
	g.StoreProposalWithRef(p, 1, ref)
	g.StoreProposalWithRef(p, 1, ref)
	if len(g.proposals.Proposals) != 1 {
		t.Fatalf("got %d pending proposals", len(g.proposals.Proposals))
	}
	if !g.RevokeProposal(ref) || len(g.proposals.Proposals) != 0 {
		t.Fatal("duplicate reference was not fully revoked")
	}
}

func TestSignedProposalNotPreviouslyPendingIsStoredOnReceipt(t *testing.T) {
	alice := newTestUser(t, "alice")
	id, err := NewGroupIDRandom()
	if err != nil {
		t.Fatal(err)
	}
	g, err := NewGroup(id, ciphersuite.MLS128DHKEMP256, alice.kp, alice.priv)
	if err != nil {
		t.Fatal(err)
	}
	wire, err := g.SignProposalAsPublicMessage(NewRemoveProposal(0), alice.sigPriv)
	if err != nil {
		t.Fatal(err)
	}
	message, err := framing.UnmarshalMLSMessage(wire)
	if err != nil {
		t.Fatal(err)
	}
	public, ok := message.AsPublic()
	if !ok {
		t.Fatal("expected public message")
	}
	if err := g.ProcessPublicMessage(public); err != nil {
		t.Fatal(err)
	}
	if len(g.proposals.Proposals) != 1 {
		t.Fatalf("got %d pending proposals", len(g.proposals.Proposals))
	}
}
