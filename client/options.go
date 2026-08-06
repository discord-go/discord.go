package client

import (
	"github.com/discord-go/discord.go/cache"
	"github.com/discord-go/discord.go/intents"
)

// Option is a functional option for configuring the client.
type Option func(*Config)

// WithCache sets the cache for the client.
func WithCache(c cache.Cache) Option {
	return func(cfg *Config) {
		cfg.Cache = c
	}
}

// WithIntents sets the intents for the client.
func WithIntents(i intents.Intent) Option {
	return func(cfg *Config) {
		cfg.Intents = i
	}
}
