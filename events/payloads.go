package events

import "github.com/discord-go/discord.go/json"

// Event represents a Discord gateway event payload.
type Event struct {
	Op   int             `json:"op"`
	Data json.RawMessage `json:"d"`
	Seq  *int            `json:"s"`
	Type string          `json:"t"`
}
