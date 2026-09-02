package gateway

import (
	"context"
	"errors"
	"testing"

	"github.com/discord-go/discord.go/cache"
	"github.com/discord-go/discord.go/channels"
	"github.com/discord-go/discord.go/guilds"
)

// sequenceConnection returns each payload in order, then a terminal error.
type sequenceConnection struct {
	payloads [][]byte
	index    int
}

func (s *sequenceConnection) Read() ([]byte, error) {
	if s.index < len(s.payloads) {
		data := s.payloads[s.index]
		s.index++
		return data, nil
	}
	return nil, errors.New("sequence exhausted")
}

func (s *sequenceConnection) Write([]byte) error { return nil }
func (s *sequenceConnection) Close() error       { return nil }

// runPayloads feeds raw gateway payloads through the real readLoop so cache
// hydration runs exactly as it does in production.
func runPayloads(t *testing.T, c *Client, payloads ...[]byte) {
	t.Helper()
	c.Conn = &sequenceConnection{payloads: payloads}
	// The loop returns when the sequence is exhausted.
	_ = c.readLoop(context.Background())
}

// TestCacheHydration_ChannelEvents verifies that CHANNEL_CREATE and
// CHANNEL_UPDATE hydrate the channel cache and CHANNEL_DELETE removes from
// it, so CachedChannel works without REST round-trips.
func TestCacheHydration_ChannelEvents(t *testing.T) {
	store := cache.NewMemoryCache()
	c := NewClient(&sequenceConnection{}, NewDispatcher())
	c.Cache = store

	runPayloads(t, c,
		[]byte(`{"t":"CHANNEL_CREATE","s":1,"op":0,"d":{"id":"777","type":2,"name":"jtc","guild_id":"123"}}`),
		[]byte(`{"t":"CHANNEL_UPDATE","s":2,"op":0,"d":{"id":"777","type":2,"name":"renamed","guild_id":"123"}}`),
	)

	ch, ok := store.GetChannel("777")
	if !ok {
		t.Fatalf("channel 777 not found in cache after CHANNEL_CREATE/UPDATE")
	}
	typed, ok := ch.(*channels.Channel)
	if !ok {
		t.Fatalf("cached value type = %T, want *channels.Channel", ch)
	}
	if typed.Name == nil || *typed.Name != "renamed" {
		t.Fatalf("cached channel name = %v, want renamed", typed.Name)
	}

	runPayloads(t, c,
		[]byte(`{"t":"CHANNEL_DELETE","s":3,"op":0,"d":{"id":"777","type":2,"guild_id":"123"}}`),
	)
	if _, ok := store.GetChannel("777"); ok {
		t.Fatalf("channel 777 still cached after CHANNEL_DELETE")
	}
}

// TestCacheHydration_GuildCreateChannels verifies that GUILD_CREATE hydrates
// the channel cache from its channels array and captures voice states.
func TestCacheHydration_GuildCreateChannels(t *testing.T) {
	store := cache.NewMemoryCache()
	c := NewClient(&sequenceConnection{}, NewDispatcher())
	c.Cache = store

	runPayloads(t, c,
		[]byte(`{"t":"GUILD_CREATE","s":1,"op":0,"d":{"id":"123","name":"g","channels":[{"id":"777","type":2,"name":"jtc"}],"voice_states":[{"user_id":"42","session_id":"abc","channel_id":"777"}]}}`),
	)

	g, ok := store.GetGuild("123")
	if !ok {
		t.Fatalf("guild 123 not cached")
	}
	typed, ok := g.(*guilds.Guild)
	if !ok {
		t.Fatalf("cached guild type = %T, want *guilds.Guild", g)
	}
	if len(typed.VoiceStates) != 1 || typed.VoiceStates[0].UserID.String() != "42" {
		t.Fatalf("guild VoiceStates = %+v, want one state for user 42", typed.VoiceStates)
	}

	ch, ok := store.GetChannel("777")
	if !ok {
		t.Fatalf("channel 777 not hydrated from GUILD_CREATE")
	}
	if cc, ok := ch.(*channels.Channel); !ok || cc.ID.String() != "777" {
		t.Fatalf("cached channel = %#v, want id 777", ch)
	}
}
