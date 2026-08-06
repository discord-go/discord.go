package events

import "github.com/discord-go/discord.go/channels"

// ChannelCreate represents the CHANNEL_CREATE event.
type ChannelCreate struct {
	channels.Channel
}

// ChannelUpdate represents the CHANNEL_UPDATE event.
type ChannelUpdate struct {
	channels.Channel
}
