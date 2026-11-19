// Package main demonstrates a private support-ticket bot.
//
// The /ticket command creates a text channel that only the invoking member and
// the bot can see, then posts a greeting. Staff with Manage Channels can see
// and manage the channel through their role permissions.
//
//	export DISCORD_TOKEN='your-bot-token'
//	go run .
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/channels"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/permissions"
	"github.com/discord-go/discord.go/rest"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN environment variable is required")
	}

	router := bot.NewRouter()
	router.Use(bot.GuildOnly())
	router.Command("ticket", "Open a private support ticket", handleTicket)

	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	b.OnReady(func(ctx *bot.ReadyContext) {
		log.Printf("ticket bot online as %s", ctx.User.Username)
	})
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}

func handleTicket(ctx *bot.InteractionContext) {
	user := ctx.User()
	if user == nil {
		_ = ctx.ReplyEphemeral("Could not identify the invoking user.")
		return
	}
	guildID := ctx.GuildID()
	name := "ticket-" + sanitizeChannelName(user.Username)

	overwrites := []channels.Overwrite{
		// @everyone cannot see the ticket.
		{ID: guildID, Type: 0, Deny: permissions.ViewChannel},
		// The member can read and reply.
		{
			ID:    user.ID,
			Type:  1,
			Allow: permissions.ViewChannel | permissions.SendMessages | permissions.ReadMessageHistory,
		},
	}
	channel, err := ctx.Bot.Rest.CreateGuildChannel(context.Background(), guildID, rest.CreateGuildChannelParams{
		Name:                 name,
		Type:                 channels.ChannelTypeGuildText,
		Topic:                fmt.Sprintf("Support ticket for %s", user.Username),
		PermissionOverwrites: overwrites,
	})
	if err != nil {
		_ = ctx.ReplyEphemeral("Could not create the ticket channel: " + err.Error())
		return
	}

	greeting := fmt.Sprintf("Welcome <@%s>. Describe your issue and a staff member will reply here.", user.ID)
	if _, err := ctx.Bot.Rest.CreateMessage(context.Background(), channel.ID, greeting); err != nil {
		log.Printf("send ticket greeting: %v", err)
	}
	_ = ctx.ReplyEphemeral(fmt.Sprintf("Ticket opened: <#%s>", channel.ID))
}

func sanitizeChannelName(name string) string {
	name = strings.ToLower(name)
	var builder strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
			builder.WriteRune(r)
		default:
			builder.WriteByte('-')
		}
	}
	result := strings.Trim(builder.String(), "-")
	if result == "" {
		return "member"
	}
	return result
}
