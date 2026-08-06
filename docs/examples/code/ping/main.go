// Package main demonstrates a simple ping/pong Discord bot built using the discord.go framework.
//
// This example illustrates:
//  1. Configuring bot intents with the intents package
//  2. Creating a bot instance with bot.New() and functional options
//  3. Registering direct event handlers using b.OnMessageCreate() and ctx.Reply()
//  4. Setting up a command Router for slash commands (/ping) and prefix commands (!ping)
//  5. Responding to slash commands with rich embeds using messages.EmbedBuilder
//  6. Running the bot lifecycle with b.Run()
//
// discord.go code
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/messages"
)

func main() {
	// Retrieve the Discord bot token from the environment, defaulting to a placeholder.
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		token = "YOUR_BOT_TOKEN"
	}

	// 1. Define the gateway intents required by the bot.
	// - Guilds: to receive guild events
	// - GuildMessages: to receive message creation events in guilds
	// - MessageContent: privileged intent required to read message text for prefix commands
	botIntents := intents.Guilds | intents.GuildMessages | intents.MessageContent

	// 2. Create a router to manage slash and prefix commands.
	router := bot.NewRouter()

	// 4. Register a slash command (/ping) on the router.
	// Slash commands are automatically registered with Discord upon connection (READY).
	// Here we show responding to a slash command using a rich embed.
	router.Command("ping", "Responds with a Pong embed and status info", func(ctx *bot.InteractionContext) {
		// Construct an embed using messages.NewEmbedBuilder()
		embed := messages.NewEmbedBuilder().
			SetTitle("🏓 Pong!").
			SetDescription("The `/ping` slash command was executed successfully.").
			SetColor(0x5865F2). // Discord blurple color
			AddField("Latency", "Gateway connection is active", true).
			AddField("Framework", "discord.go", true).
			SetTimestamp(time.Now()).
			SetFooter("discord.go ping example", "").
			Build()

		// Send the embed response back to the interaction
		if err := ctx.ReplyEmbed(embed); err != nil {
			log.Printf("Error sending slash command response: %v", err)
		}
	})

	// 5. Register a prefix command (!ping) on the router.
	// Prefix commands receive the MessageContext and any command arguments after the command name.
	router.Prefix("ping", func(ctx *bot.MessageContext, args []string) {
		response := "🏓 Pong! (Prefix Command)"
		if len(args) > 0 {
			response += fmt.Sprintf("\nArguments received: %v", args)
		}

		if _, err := ctx.Reply(response); err != nil {
			log.Printf("Error sending prefix command response: %v", err)
		}
	})

	// 1 & 2. Create the bot instance using bot.New() with configuration options.
	b := bot.New(token,
		bot.WithIntents(botIntents),
		bot.WithPrefix("!"),
		bot.WithRouter(router),
	)

	// Register a callback for when the bot logs in.
	b.OnReady(func(ctx *bot.ReadyContext) {
		log.Printf("Ready! Logged in as %s (ID: %s)", ctx.User.Username, ctx.User.ID)
	})

	// 3. Register a direct event handler for MESSAGE_CREATE events.
	// Bot-authored messages are automatically ignored by OnMessageCreate.
	b.OnMessageCreate(func(ctx *bot.MessageContext) {
		// Example of simple keyword matching outside of the router
		if ctx.Content == "ping" {
			if _, err := ctx.Reply("Pong! (Direct Message Handler)"); err != nil {
				log.Printf("Error replying to message: %v", err)
			}
		}
	})

	// 6. Connect to Discord and block until SIGINT or SIGTERM signal is received.
	log.Println("Starting Discord bot... Press Ctrl+C to exit.")
	if err := b.Run(); err != nil {
		log.Fatalf("Bot runtime error: %v", err)
	}
}
