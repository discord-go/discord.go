package bot

import (
	"testing"

	"github.com/discord-go/discord.go/cache"
)

// feedDispatch sends a raw gateway dispatch through the bot's dispatch path,
// exercising both gateway hydration and bot-layer cache updates. The outer
// envelope's t is discarded; the inner payload supplies the real event type.
func feedDispatch(t *testing.T, b *Bot, eventType string, payload string) {
	t.Helper()
	b.handleRawDispatch([]byte(`{"t":"` + eventType + `","s":1,"op":0,"d":` + payload + `}`))
}

// TestBotMessageCache verifies MESSAGE_CREATE stores, MESSAGE_UPDATE merges,
// and MESSAGE_DELETE removes from the configured message cache.
func TestBotMessageCache(t *testing.T) {
	b := New("token", WithCache(cache.NewMemoryCache()))
	store := b.cacheStore.(*cache.MemoryCache)

	feedDispatch(t, b, "MESSAGE_CREATE", `{"id":"500","channel_id":"200","guild_id":"100","content":"v1","author":{"id":"5","username":"u"}}`)
	if _, ok := store.GetMessage("500"); !ok {
		t.Fatalf("message 500 not cached after MESSAGE_CREATE")
	}

	feedDispatch(t, b, "MESSAGE_UPDATE", `{"id":"500","channel_id":"200","content":"v2"}`)
	cached, ok := b.CachedMessage(500)
	if !ok {
		t.Fatalf("message 500 missing after MESSAGE_UPDATE")
	}
	if cached.Content != "v2" {
		t.Errorf("message content = %q, want v2", cached.Content)
	}
	if cached.Author == nil || cached.Author.ID.String() != "5" {
		t.Errorf("partial update dropped cached author: %+v", cached.Author)
	}

	feedDispatch(t, b, "MESSAGE_DELETE", `{"id":"500","channel_id":"200"}`)
	if _, ok := store.GetMessage("500"); ok {
		t.Fatalf("message 500 still cached after MESSAGE_DELETE")
	}
}
