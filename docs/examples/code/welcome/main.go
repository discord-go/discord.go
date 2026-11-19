// Package main demonstrates a welcome-message bot.
//
// It listens for GUILD_MEMBER_ADD, decodes the raw event, and posts a mention
// in a configured channel. Enable the Guild Members privileged intent in the
// Developer Portal before running it.
//
//	export DISCORD_TOKEN='your-bot-token'
//	export WELCOME_CHANNEL_ID='123456789012345678'
//	go run .
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/snowflake"
)

type memberAdd struct {
	GuildID snowflake.ID `json:"guild_id,string"`
	User    struct {
		ID       snowflake.ID `json:"id,string"`
		Username string       `json:"username"`
	} `json:"user"`
}

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN environment variable is required")
	}
	channelID, err := snowflake.Parse(os.Getenv("WELCOME_CHANNEL_ID"))
	if err != nil {
		log.Fatal("WELCOME_CHANNEL_ID must be a channel snowflake")
	}

	b := bot.New(token, bot.WithIntents(intents.Guilds|intents.GuildMembers))
	b.OnReady(func(ctx *bot.ReadyContext) {
		log.Printf("welcome bot online as %s", ctx.User.Username)
	})
	b.OnGuildMemberAdd(func(ctx *bot.EventContext) {
		var event memberAdd
		if err := ctx.Decode(&event); err != nil {
			log.Printf("decode GUILD_MEMBER_ADD: %v", err)
			return
		}
		message := fmt.Sprintf("Welcome <@%s> to the server!", event.User.ID)
		if _, err := ctx.Bot.Rest.CreateMessage(context.Background(), channelID, message); err != nil {
			log.Printf("send welcome message: %v", err)
		}
	})

	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
