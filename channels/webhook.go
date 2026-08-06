package channels

import "github.com/discord-go/discord.go/snowflake"

// Webhook represents a Discord webhook.
type Webhook struct {
	ID snowflake.ID `json:"id,string"`
}
