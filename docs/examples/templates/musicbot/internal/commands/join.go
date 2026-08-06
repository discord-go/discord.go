// discord.go code
package commands

import (
	"context"
	"fmt"
	"log"

	"github.com/discord-go/discord.go/client"
	"github.com/discord-go/discord.go/events"
	"github.com/discord-go/discord.go/snowflake"
)

// HandleJoin handles the !join command.
func HandleJoin(ctx context.Context, bot *client.Client, e *events.MessageCreate, vcID snowflake.ID, ok bool) {
	if !ok {
		reply(ctx, bot, e, "❌ You must be in a voice channel to use this command!")
		return
	}

	log.Printf("DEBUG: Joining voice channel GuildID=%d, ChannelID=%d", e.GuildID, vcID)
	err := bot.Gateway.JoinVoiceChannel(e.GuildID, vcID, false, false)
	if err != nil {
		log.Printf("DEBUG: JoinVoiceChannel error: %v", err)
		reply(ctx, bot, e, fmt.Sprintf("❌ Error joining voice channel: %v", err))
		return
	}
	reply(ctx, bot, e, "✅ Joined your voice channel!")
}
