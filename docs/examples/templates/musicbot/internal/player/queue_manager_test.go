package player

import (
	"testing"

	"github.com/discord-go/discord.go/snowflake"
)

func TestQueueUsesGuildAndStopClearsIt(t *testing.T) {
	m := NewQueueManager()
	guild := m.GetPlayer(snowflake.ID(10))
	guild.Enqueue(&Track{Title: "first"})
	guild.Enqueue(&Track{Title: "second"})
	if got := m.GetPlayer(10).GetQueue(); len(got) != 2 || got[0].Title != "first" {
		t.Fatalf("queue: %+v", got)
	}
	if got := m.GetPlayer(11).GetQueue(); len(got) != 0 {
		t.Fatalf("other guild queue: %+v", got)
	}
	guild.Stop()
	if got := guild.GetQueue(); len(got) != 0 {
		t.Fatalf("queue after stop: %+v", got)
	}
}
