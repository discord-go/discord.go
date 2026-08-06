// discord.go code
package commands

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/discord-go/discord.go/client"
	"github.com/discord-go/discord.go/events"
	"github.com/discord-go/discord.go/snowflake"

	"musicbot/internal/player"
	"musicbot/internal/ytdlp"
)

// HandlePlay handles the !play command.
func HandlePlay(ctx context.Context, bot *client.Client, e *events.MessageCreate, manager *player.QueueManager, vcID snowflake.ID, ok bool) {
	if !ok {
		reply(ctx, bot, e, "❌ You must be in a voice channel to use this command!")
		return
	}

	parts := strings.SplitN(e.Content, " ", 2)
	if len(parts) < 2 {
		reply(ctx, bot, e, "Usage: !play <youtube-url or search>")
		return
	}
	videoURL := parts[1]

	reply(ctx, bot, e, fmt.Sprintf("🔍 Searching for %s...", videoURL))

	streamURL, title, err := ytdlp.ExtractAudioInfo(videoURL)
	if err != nil {
		reply(ctx, bot, e, fmt.Sprintf("❌ Error extracting audio: %v", err))
		return
	}

	p := manager.GetPlayer(e.GuildID)
	track := &player.Track{
		Title:     title,
		StreamURL: streamURL,
		Requested: e.Author.Username,
	}
	p.Enqueue(track)
	reply(ctx, bot, e, fmt.Sprintf("✅ Added to queue: **%s**", title))

	if !p.Playing {
		if p.VoiceClient != nil {
			log.Printf("DEBUG: Already connected, starting playback on existing VoiceClient")
			p.PlayNext(p.VoiceClient)
		} else {
			log.Printf("DEBUG: Joining voice channel GuildID=%d, ChannelID=%d", e.GuildID, vcID)
			err := bot.Gateway.JoinVoiceChannel(e.GuildID, vcID, false, false)
			if err != nil {
				log.Printf("DEBUG: JoinVoiceChannel error: %v", err)
				reply(ctx, bot, e, fmt.Sprintf("❌ Error joining voice channel: %v", err))
			} else {
				log.Printf("DEBUG: Successfully called JoinVoiceChannel!")
			}
		}
	}
}

func reply(ctx context.Context, bot *client.Client, e *events.MessageCreate, content string) {
	_, _ = bot.Rest.CreateMessage(ctx, e.ChannelID, content)
}
