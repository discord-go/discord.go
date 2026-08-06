package messages

import "github.com/discord-go/discord.go/emojis"

// ReactionCountDetails represents the details of reaction counts.
type ReactionCountDetails struct {
	Burst  int `json:"burst"`
	Normal int `json:"normal"`
}

// Reaction represents a reaction to a message.
type Reaction struct {
	Count        int                  `json:"count"`
	CountDetails ReactionCountDetails `json:"count_details"`
	Me           bool                 `json:"me"`
	MeBurst      bool                 `json:"me_burst"`
	Emoji        emojis.Emoji         `json:"emoji"`
	BurstColors  []string             `json:"burst_colors,omitempty"`
}
