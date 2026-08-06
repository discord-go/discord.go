// discord.go code
package commands

import (
	"context"

	"github.com/discord-go/discord.go/client"
	"github.com/discord-go/discord.go/events"
)

// HandleInfo handles the !info command.
func HandleInfo(ctx context.Context, bot *client.Client, e *events.MessageCreate) {
	infoText := "**Music Bot Template**\n" +
		"A highly organized template for building a music bot using `discord.go`.\n" +
		"Commands:\n" +
		"`!play <youtube-url>` - Add a track to the queue and play it.\n" +
		"`!stop` - Stop playing and clear the queue.\n" +
		"`!queue` - View the current queue.\n" +
		"`!info` - Show this information message."

	reply(ctx, bot, e, infoText)
}
