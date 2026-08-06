// discord.go code
package commands

import (
	"context"
	"fmt"
	"log"

	"github.com/discord-go/discord.go/client"
	"github.com/discord-go/discord.go/events"
	"musicbot/internal/player"
)

// HandleLeave handles the !leave command.
func HandleLeave(ctx context.Context, bot *client.Client, e *events.MessageCreate, manager *player.QueueManager) {
	// Stop any playing audio
	p := manager.GetPlayer(e.GuildID)
	p.Stop()

	log.Printf("DEBUG: Leaving voice channel GuildID=%d", e.GuildID)
	// Sending channelID as 0 removes the bot from the voice channel.
	err := bot.Gateway.JoinVoiceChannel(e.GuildID, 0, false, false)
	if err != nil {
		log.Printf("DEBUG: LeaveVoiceChannel error: %v", err)
		reply(ctx, bot, e, fmt.Sprintf("❌ Error leaving voice channel: %v", err))
		return
	}
	reply(ctx, bot, e, "👋 Left the voice channel.")
}
