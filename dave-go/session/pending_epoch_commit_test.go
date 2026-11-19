package session

import (
	"context"
	"testing"
)

// A pending epoch obtained from a Welcome is not a locally-created commit:
// another accepted commit must be applied without restoring a nonexistent
// pre-commit snapshot, even if the first transition has not executed yet.
func TestPeerCommitOverUnexecutedWelcome(t *testing.T) {
	ctx := context.Background()
	cb := &kpCapturingCallbacks{}
	s := New("123456789", cb)
	defer s.Close()
	s.SetChannelID(987654321)
	s.OnSelectProtocolAck(1)
	external := buildExternalSenderPackage(t)
	setExternalSenderPackage(t, s, external)
	peer := peerWithIdentity(t, "987654321")
	kp, err := peer.FreshKeyPackageBytes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	gid, err := peer.CreateGroupWithExternalSender(ctx, []byte("12345678"), kp, external)
	if err != nil {
		t.Fatal(err)
	}
	_, welcome, err := peer.InviteMember(ctx, gid, cb.lastKeyPackage())
	if err != nil {
		t.Fatal(err)
	}
	s.OnDaveMLSWelcome(5, welcome)
	if s.Stats().WelcomesJoined != 1 {
		t.Fatal("welcome did not join")
	}
	commit, _, err := peer.CommitPendingProposals(ctx, gid)
	if err != nil {
		t.Fatal(err)
	}
	s.OnDaveMLSPrepareCommitTransition(6, commit)
	if got := s.Stats().CommitsProcessed; got != 1 {
		t.Fatalf("processed %d peer commits, want 1", got)
	}
	want, err := peer.Epoch(ctx, gid)
	if err != nil {
		t.Fatal(err)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.pendingEpoch == nil || s.pendingEpoch.id != want {
		t.Fatalf("pending epoch = %+v, want %d", s.pendingEpoch, want)
	}
}
