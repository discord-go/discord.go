package client

import (
	"github.com/discord-go/discord.go/cache"
	"github.com/discord-go/discord.go/intents"
)

// Config holds the configuration states for the client.
type Config struct {
	Token   string
	Cache   cache.Cache
	Intents intents.Intent
}
