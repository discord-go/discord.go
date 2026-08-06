package events

import "github.com/discord-go/discord.go/interactions"

// InteractionCreate represents the INTERACTION_CREATE event.
type InteractionCreate struct {
	interactions.Interaction
}
