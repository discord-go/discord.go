package session

import (
	"testing"

	"github.com/discord-go/discord.go/dave-go/godave"
	"github.com/discord-go/discord.go/dave-go/mediakeys"
)

func newRatchet(t *testing.T, b byte) *mediakeys.KeyRatchet {
	t.Helper()
	seed := make([]byte, mediakeys.BaseSecretLen)
	for i := range seed {
		seed[i] = b
	}
	r, err := mediakeys.NewKeyRatchet(seed)
	if err != nil {
		t.Fatalf("NewKeyRatchet: %v", err)
	}

	return r
}

// setupForActivation builds a session whose active epoch is `oldMembers` (with
// the bot as sender) and whose pending epoch is `newMembers`, ready for
// activatePendingEpochLocked. The send counter is advanced so a reset is
// observable.
func setupForActivation(t *testing.T, oldMembers, newMembers []godave.UserID) *Session {
	t.Helper()
	s := New("bot", testCallbacks{})
	bot := s.userID

	oldSenders := map[godave.UserID]*senderState{}
	for _, m := range oldMembers {
		oldSenders[m] = &senderState{ratchet: newRatchet(t, 0x01)}
	}
	newSenders := map[godave.UserID]*senderState{}
	for _, m := range newMembers {
		newSenders[m] = &senderState{ratchet: newRatchet(t, 0x02)}
	}

	s.mu.Lock()
	s.activeEpoch = &epochState{id: 0, senders: oldSenders}
	s.sendRatchet = oldSenders[bot].ratchet
	s.pendingEpoch = &epochState{id: 1, groupID: []byte("g"), senders: newSenders}
	s.pendingGroupID = []byte("g")
	s.mu.Unlock()

	// Advance the counter so we can tell whether activation reset it.
	for range 5 {
		s.sendCounter.Next()
	}

	return s
}

// TestActivate_SkipsSendRetentionOnJoin: when the new epoch adds a member (a
// join), the previous send ratchet must NOT be retained — retaining it would
// black out the joiner (who holds no old-epoch keys) for the whole TTL. Instead
// the bot switches atomically and resets the nonce.
func TestActivate_SkipsSendRetentionOnJoin(t *testing.T) {
	s := setupForActivation(t, []godave.UserID{"bot"}, []godave.UserID{"bot", "joiner"})

	s.mu.Lock()
	s.activatePendingEpochLocked()
	retained := s.retainedSendRatchet
	counter := s.sendCounter.Current()
	sendIsNew := s.sendRatchet == s.activeEpoch.senders["bot"].ratchet
	s.mu.Unlock()

	if retained != nil {
		t.Error("send ratchet must NOT be retained when the new epoch adds a member")
	}
	if counter != 0 {
		t.Errorf("send counter must reset to 0 on atomic switch, got %d", counter)
	}
	if !sendIsNew {
		t.Error("send ratchet must switch to the new epoch's ratchet")
	}
}

// A sender must immediately use the new epoch after every transition.
func TestActivate_SwitchesSendRatchetImmediately(t *testing.T) {
	for _, tt := range []struct {
		name                   string
		oldMembers, newMembers []godave.UserID
	}{
		{"same members", []godave.UserID{"bot", "other"}, []godave.UserID{"bot", "other"}},
		{"member removed", []godave.UserID{"bot", "leaver"}, []godave.UserID{"bot"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := setupForActivation(t, tt.oldMembers, tt.newMembers)
			defer s.Close()
			s.mu.Lock()
			s.activatePendingEpochLocked()
			retained := s.retainedSendRatchet
			sendIsNew := s.sendRatchet == s.activeEpoch.senders["bot"].ratchet
			counter := s.sendCounter.Current()
			s.mu.Unlock()
			if retained != nil || !sendIsNew || counter != 0 {
				t.Fatalf("expected new send key and reset nonce, retained=%v new=%v counter=%d", retained != nil, sendIsNew, counter)
			}
		})
	}
}
