package mls

import (
	"context"
	"fmt"
	"testing"

	"github.com/discord-go/discord.go/dave-go/mls-go/ciphersuite"
)

// Exercise a reused leaf slot, with commits made by either remaining member.
// This differs from removing only the last leaf and never welcoming it back.
func TestClientRepeatedMemberRejoin(t *testing.T) {
	for _, committerIndex := range []int{0, 2} {
		t.Run(fmt.Sprint(committerIndex), func(t *testing.T) {
			ctx := context.Background()
			newClient := func(identity string) *Client {
				c, err := NewClient([]byte(identity), ciphersuite.MLS128DHKEMP256, WithCacheStrategy(CacheNone))
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = c.Close() })
				return c
			}
			members := []*Client{newClient("music"), newClient("listener"), newClient("recorder")}
			gid, err := members[0].CreateGroup(ctx)
			if err != nil {
				t.Fatal(err)
			}
			for i := 1; i < len(members); i++ {
				kp, err := members[i].FreshKeyPackageBytes(ctx)
				if err != nil {
					t.Fatal(err)
				}
				commit, welcome, err := members[0].InviteMember(ctx, gid, kp)
				if err != nil {
					t.Fatal(err)
				}
				for j := 1; j < i; j++ {
					if err := members[j].ProcessCommit(ctx, gid, commit); err != nil {
						t.Fatal(err)
					}
				}
				if _, err := members[i].JoinGroup(ctx, welcome); err != nil {
					t.Fatal(err)
				}
			}
			committer, peer := members[committerIndex], members[2-committerIndex]
			for cycle := 0; cycle < 3; cycle++ {
				remove, err := committer.RemoveMember(ctx, gid, []byte("listener"))
				if err != nil {
					t.Fatal(err)
				}
				if err := peer.ProcessCommit(ctx, gid, remove); err != nil {
					t.Fatal(err)
				}
				listener := newClient("listener")
				kp, err := listener.FreshKeyPackageBytes(ctx)
				if err != nil {
					t.Fatal(err)
				}
				add, welcome, err := committer.InviteMember(ctx, gid, kp)
				if err != nil {
					t.Fatal(err)
				}
				if err := peer.ProcessCommit(ctx, gid, add); err != nil {
					t.Fatalf("cycle %d peer: %v", cycle, err)
				}
				if _, err := listener.JoinGroup(ctx, welcome); err != nil {
					t.Fatalf("cycle %d rejoin: %v", cycle, err)
				}
				wire, err := committer.SendMessage(ctx, gid, []byte("audio continues"))
				if err != nil {
					t.Fatal(err)
				}
				for _, recipient := range []*Client{peer, listener} {
					msg, err := recipient.ReceiveMessage(ctx, gid, wire)
					if err != nil {
						t.Fatal(err)
					}
					if string(msg.Plaintext) != "audio continues" {
						t.Fatal("incorrect plaintext after rejoin")
					}
				}
			}
		})
	}
}
