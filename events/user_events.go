package events

import (
	"github.com/discord-go/discord.go/guilds"
	"github.com/discord-go/discord.go/users"
)

// Ready represents the READY event.
type Ready struct {
	V         int            `json:"v"`
	User      users.User     `json:"user"`
	Guilds    []guilds.Guild `json:"guilds"`
	SessionID string         `json:"session_id"`
}
