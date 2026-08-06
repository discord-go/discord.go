// discord.go code
package commands

import (
	"context"

	"github.com/discord-go/discord.go/client"
	"github.com/discord-go/discord.go/events"

	"musicbot/internal/player"
)

// HandleStop handles the !stop command.
func HandleStop(ctx context.Context, bot *client.Client, e *events.MessageCreate, manager *player.QueueManager) {
	p := manager.GetPlayer(e.ChannelID)
	p.Stop()
	reply(ctx, bot, e, "⏹️ Stopped playback and cleared the queue.")
}
