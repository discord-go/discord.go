package client

import (
	"github.com/discord-go/discord.go/intents"
	"testing"
)

type mockCache struct{}

func (m *mockCache) Get(key string) (any, bool) { return nil, false }
func (m *mockCache) Set(key string, value any)  {}
func (m *mockCache) Delete(key string)          {}
func (m *mockCache) Clear()                     {}

func TestNew(t *testing.T) {
	token := "test-token"
	mockC := &mockCache{}
	intent := intents.GuildMessages

	client := New(token, WithCache(mockC), WithIntents(intent))

	if client.Rest == nil {
		t.Errorf("Expected Rest client to be initialized")
	}

	if client.Rest.Token != token {
		t.Errorf("Expected token %q, got %q", token, client.Rest.Token)
	}

	if client.Cache != mockC {
		t.Errorf("Expected cache to be %v, got %v", mockC, client.Cache)
	}
}

func TestOptions(t *testing.T) {
	cfg := &Config{}
	mockC := &mockCache{}
	WithCache(mockC)(cfg)
	if cfg.Cache != mockC {
		t.Errorf("Expected cache to be %v, got %v", mockC, cfg.Cache)
	}

	intent := intents.GuildMessages
	WithIntents(intent)(cfg)
	if cfg.Intents != intent {
		t.Errorf("Expected intents to be %v, got %v", intent, cfg.Intents)
	}
}
