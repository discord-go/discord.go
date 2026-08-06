// discord.go code
package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/discord-go/discord.go/client"
	"github.com/discord-go/discord.go/events"

	"musicbot/internal/player"
)

// HandleQueue handles the !queue command.
func HandleQueue(ctx context.Context, bot *client.Client, e *events.MessageCreate, manager *player.QueueManager) {
	p := manager.GetPlayer(e.ChannelID)
	q := p.GetQueue()

	if len(q) == 0 {
		reply(ctx, bot, e, "📭 The queue is currently empty.")
		return
	}

	var sb strings.Builder
	sb.WriteString("**Current Queue:**\n")
	for i, track := range q {
		if i >= 10 {
			sb.WriteString(fmt.Sprintf("...and %d more tracks.\n", len(q)-10))
			break
		}
		sb.WriteString(fmt.Sprintf("%d. %s (requested by %s)\n", i+1, track.Title, track.Requested))
	}

	reply(ctx, bot, e, sb.String())
}
