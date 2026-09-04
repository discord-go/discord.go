package gateway

import (
	"context"
	"errors"
	"testing"

	"github.com/discord-go/discord.go/cache"
	"github.com/discord-go/discord.go/channels"
	"github.com/discord-go/discord.go/guilds"
	"github.com/discord-go/discord.go/users"
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

// TestCacheHydration_RoleEvents verifies that GUILD_ROLE_CREATE/UPDATE merge
// into the cached guild's Roles and GUILD_ROLE_DELETE removes them, so
// permission resolution stays fresh without a reconnect.
func TestCacheHydration_RoleEvents(t *testing.T) {
	store := cache.NewMemoryCache()
	c := NewClient(&sequenceConnection{}, NewDispatcher())
	c.Cache = store

	runPayloads(t, c,
		[]byte(`{"t":"GUILD_CREATE","s":1,"op":0,"d":{"id":"123","name":"g","roles":[{"id":"123","name":"@everyone","permissions":"1024"}]}}`),
		[]byte(`{"t":"GUILD_ROLE_CREATE","s":2,"op":0,"d":{"guild_id":"123","role":{"id":"55","name":"Mods","permissions":"2080"}}}`),
		[]byte(`{"t":"GUILD_ROLE_UPDATE","s":3,"op":0,"d":{"guild_id":"123","role":{"id":"55","name":"Mods","permissions":"2084"}}}`),
	)

	g, ok := store.GetGuild("123")
	if !ok {
		t.Fatalf("guild 123 not cached")
	}
	guild := g.(*guilds.Guild)
	if len(guild.Roles) != 2 {
		t.Fatalf("guild roles = %d, want 2", len(guild.Roles))
	}
	if guild.Roles[1].ID.String() != "55" || guild.Roles[1].Permissions.String() != "2084" {
		t.Fatalf("role 55 not merged/updated: %+v", guild.Roles[1])
	}

	runPayloads(t, c,
		[]byte(`{"t":"GUILD_ROLE_DELETE","s":4,"op":0,"d":{"guild_id":"123","role_id":"55"}}`),
	)
	g, _ = store.GetGuild("123")
	guild = g.(*guilds.Guild)
	if len(guild.Roles) != 1 {
		t.Fatalf("guild roles after delete = %d, want 1", len(guild.Roles))
	}
	if _, ok := store.GetRole("55"); ok {
		t.Fatalf("role 55 still in role cache after GUILD_ROLE_DELETE")
	}
}

// TestCacheHydration_GuildUpdate verifies that GUILD_REFRESH... GUILD_UPDATE
// refreshes the cached guild (owner transfers, role changes).
func TestCacheHydration_GuildUpdate(t *testing.T) {
	store := cache.NewMemoryCache()
	c := NewClient(&sequenceConnection{}, NewDispatcher())
	c.Cache = store

	runPayloads(t, c,
		[]byte(`{"t":"GUILD_CREATE","s":1,"op":0,"d":{"id":"123","name":"g","owner_id":"999"}}`),
		[]byte(`{"t":"GUILD_UPDATE","s":2,"op":0,"d":{"id":"123","name":"renamed","owner_id":"42","roles":[{"id":"123","name":"@everyone","permissions":"1024"}]}}`),
	)

	g, ok := store.GetGuild("123")
	if !ok {
		t.Fatalf("guild 123 not cached")
	}
	guild := g.(*guilds.Guild)
	if guild.OwnerID.String() != "42" {
		t.Fatalf("owner_id = %s, want 42 after GUILD_UPDATE", guild.OwnerID)
	}
	if len(guild.Roles) != 1 {
		t.Fatalf("roles = %d, want 1 after GUILD_UPDATE", len(guild.Roles))
	}
}

// TestCacheHydration_MembersChunk verifies that GUILD_MEMBERS_CHUNK hydrates
// the member cache.
func TestCacheHydration_MembersChunk(t *testing.T) {
	store := cache.NewMemoryCache()
	c := NewClient(&sequenceConnection{}, NewDispatcher())
	c.Cache = store

	runPayloads(t, c,
		[]byte(`{"t":"GUILD_MEMBERS_CHUNK","s":1,"op":0,"d":{"guild_id":"123","members":[{"user":{"id":"42","username":"a"},"roles":["55"]},{"user":{"id":"43","username":"b"},"roles":[]}]}}`),
	)

	for _, uid := range []string{"42", "43"} {
		if _, ok := store.GetMember("123", uid); !ok {
			t.Fatalf("member %s not hydrated from GUILD_MEMBERS_CHUNK", uid)
		}
	}
}

// TestCacheHydration_EmojisAndUser verifies GUILD_EMOJIS_UPDATE refreshes the
// cached guild's emoji array and USER_UPDATE refreshes the user cache.
func TestCacheHydration_EmojisAndUser(t *testing.T) {
	store := cache.NewMemoryCache()
	c := NewClient(&sequenceConnection{}, NewDispatcher())
	c.Cache = store

	runPayloads(t, c,
		[]byte(`{"t":"GUILD_CREATE","s":1,"op":0,"d":{"id":"123","name":"g","emojis":[{"id":"900","name":"old"}]}}`),
		[]byte(`{"t":"GUILD_EMOJIS_UPDATE","s":2,"op":0,"d":{"guild_id":"123","emojis":[{"id":"901","name":"new"}]}}`),
		[]byte(`{"t":"USER_UPDATE","s":3,"op":0,"d":{"id":"42","username":"renamed"}}`),
	)

	g, ok := store.GetGuild("123")
	if !ok {
		t.Fatalf("guild 123 not cached")
	}
	guild := g.(*guilds.Guild)
	if len(guild.Emojis) != 1 || guild.Emojis[0].ID.String() != "901" {
		t.Fatalf("guild emojis = %+v, want one emoji id 901", guild.Emojis)
	}
	u, ok := store.GetUser("42")
	if !ok {
		t.Fatalf("user 42 not cached after USER_UPDATE")
	}
	if user, ok := u.(*users.User); !ok || user.Username != "renamed" {
		t.Fatalf("cached user = %#v, want username renamed", u)
	}
}

// TestCacheHydration_Threads verifies THREAD_CREATE/UPDATE cache threads as
// channels and THREAD_DELETE removes them.
func TestCacheHydration_Threads(t *testing.T) {
	store := cache.NewMemoryCache()
	c := NewClient(&sequenceConnection{}, NewDispatcher())
	c.Cache = store

	runPayloads(t, c,
		[]byte(`{"t":"THREAD_CREATE","s":1,"op":0,"d":{"id":"888","type":11,"name":"thread","guild_id":"123"}}`),
		[]byte(`{"t":"THREAD_UPDATE","s":2,"op":0,"d":{"id":"888","type":11,"name":"renamed","guild_id":"123"}}`),
	)

	ch, ok := store.GetChannel("888")
	if !ok {
		t.Fatalf("thread 888 not cached")
	}
	typed, ok := ch.(*channels.Channel)
	if !ok {
		t.Fatalf("cached thread type = %T, want *channels.Channel", ch)
	}
	if typed.Name == nil || *typed.Name != "renamed" {
		t.Fatalf("thread name = %v, want renamed", typed.Name)
	}

	runPayloads(t, c,
		[]byte(`{"t":"THREAD_DELETE","s":3,"op":0,"d":{"id":"888","guild_id":"123"}}`),
	)
	if _, ok := store.GetChannel("888"); ok {
		t.Fatalf("thread 888 still cached after THREAD_DELETE")
	}
}
