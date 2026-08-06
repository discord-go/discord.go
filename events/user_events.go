package events

import "github.com/discord-go/discord.go/users"

// Ready represents the READY event.
type Ready struct {
	V         int        `json:"v"`
	User      users.User `json:"user"`
	SessionID string     `json:"session_id"`
}
