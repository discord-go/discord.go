// Package main demonstrates how to create, register, and handle Discord slash
// commands using the discord.go framework.
//
// Examples included:
// 1. Creating a Router with bot.NewRouter()
// 2. Registering multiple slash commands with options (string, user, integer, boolean)
// 3. Global middleware vs. per-command middleware (bot.RequirePermissions, bot.GuildOnly)
// 4. Reading interaction options (ctx.GetStringOption, ctx.GetUserID, ctx.GetIntOption, ctx.GetBoolOption)
// 5. Responding to interactions (ctx.Reply, ctx.ReplyEphemeral, ctx.ReplyEmbed)
// 6. Deferring responses and using followups (ctx.Defer, ctx.Followup)
// 7. Auto slash command registration by attaching the router to the bot instance
//
// Usage:
//
//	export DISCORD_TOKEN="your-bot-token-here"
//	go run main.go
//
// discord.go code
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/permissions"
)

// globalLoggingMiddleware demonstrates global middleware applied across all commands.
// It logs each slash command invocation before executing the underlying handler chain.
func globalLoggingMiddleware(next bot.CommandHandler) bot.CommandHandler {
	return func(ctx *bot.InteractionContext) {
		invoker := "Unknown"
		if u := ctx.User(); u != nil {
			invoker = u.Username
		}

		log.Printf("[Global Middleware] Command '/%s' invoked by %s", ctx.CommandName(), invoker)
		next(ctx)
	}
}

func main() {
	// Read bot token from DISCORD_TOKEN environment variable.
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("Error: DISCORD_TOKEN environment variable is not set. Please export DISCORD_TOKEN before running.")
	}

	// 1. Create a new command Router with bot.NewRouter()
	router := bot.NewRouter()

	// 4. Attach Global Middleware to the router.
	// Global middleware executes for all commands registered on this router.
	router.Use(globalLoggingMiddleware)

	// 2. Register /hello command with options (string, integer, boolean)
	// Demonstrates string option, integer option, and boolean option.
	router.Command(
		"hello",
		"Greets the user",
		handleHello,
		interactions.ApplicationCommandOption{
			Type:        interactions.ApplicationCommandOptionTypeString,
			Name:        "name",
			Description: "Name of the person to greet",
			Required:    false,
		},
		interactions.ApplicationCommandOption{
			Type:        interactions.ApplicationCommandOptionTypeInteger,
			Name:        "count",
			Description: "Number of times to repeat the greeting (1-5)",
			Required:    false,
		},
		interactions.ApplicationCommandOption{
			Type:        interactions.ApplicationCommandOptionTypeBoolean,
			Name:        "shout",
			Description: "Whether to output the greeting in UPPERCASE",
			Required:    false,
		},
	)

	// 2. Register /userinfo command with user option
	// Demonstrates user option type (interactions.ApplicationCommandOptionTypeUser).
	router.Command(
		"userinfo",
		"Shows info about a user",
		handleUserInfo,
		interactions.ApplicationCommandOption{
			Type:        interactions.ApplicationCommandOptionTypeUser,
			Name:        "user",
			Description: "The user to inspect",
			Required:    false,
		},
	)

	// 2 & 3 & 4. Register /kick command with user and string options + per-command middleware
	// Demonstrates per-command middleware chaining (.Use):
	// - bot.RequirePermissions(permissions.KickMembers) verifies invoker permissions
	// - bot.GuildOnly() ensures the command is only executed within a server
	router.Command(
		"kick",
		"Kicks a user from the server",
		handleKick,
		interactions.ApplicationCommandOption{
			Type:        interactions.ApplicationCommandOptionTypeUser,
			Name:        "user",
			Description: "The user to kick",
			Required:    true,
		},
		interactions.ApplicationCommandOption{
			Type:        interactions.ApplicationCommandOptionTypeString,
			Name:        "reason",
			Description: "Reason for kicking the user",
			Required:    false,
		},
	).
		Use(bot.RequirePermissions(permissions.KickMembers)).
		Use(bot.GuildOnly())

	// 2 & 3 & 4. Register /serverinfo command + per-command middleware
	// Demonstrates per-command middleware (bot.GuildOnly) without parameters.
	router.Command(
		"serverinfo",
		"Shows server info (guild only)",
		handleServerInfo,
	).Use(bot.GuildOnly())

	// 8. Auto slash command registration by attaching the router to the bot
	// bot.WithRouter(router) registers slash commands with Discord upon gateway READY.
	b := bot.New(
		token,
		bot.WithIntents(intents.Guilds|intents.GuildMessages),
		bot.WithRouter(router),
	)

	// Register a READY handler for startup feedback.
	b.OnReady(func(ctx *bot.ReadyContext) {
		log.Printf("🤖 Bot is connected as %s#%s (ID: %s)",
			ctx.User.Username, ctx.User.Discriminator, ctx.User.ID)
		log.Println("⚡ Slash commands have been automatically registered with Discord!")
	})

	log.Println("🚀 Connecting bot to Discord gateway...")

	// b.Run() starts the bot connection and blocks until SIGINT/SIGTERM is received.
	if err := b.Run(); err != nil {
		log.Fatalf("Fatal error running bot: %v", err)
	}
}

// handleHello demonstrates reading string, integer, and boolean options,
// and responding publicly using ctx.Reply().
func handleHello(ctx *bot.InteractionContext) {
	// 5. Reading string, integer, and boolean options
	name := ctx.GetStringOption("name")
	if name == "" {
		if u := ctx.User(); u != nil {
			name = u.Username
		} else {
			name = "friend"
		}
	}

	shout := ctx.GetBoolOption("shout")
	count := ctx.GetIntOption("count")

	greeting := fmt.Sprintf("Hello, %s! 👋", name)
	if shout {
		greeting = fmt.Sprintf("HELLO, %s! 👋", name)
	}

	if count > 1 {
		if count > 5 {
			count = 5 // Cap repetitions to avoid channel spam
		}
		out := ""
		for i := int64(0); i < count; i++ {
			if i > 0 {
				out += "\n"
			}
			out += greeting
		}
		greeting = out
	}

	// 6. Using ctx.Reply() for public interaction responses
	if err := ctx.Reply(greeting); err != nil {
		log.Printf("Error sending /hello reply: %v", err)
	}
}

// handleUserInfo demonstrates deferring interaction responses with ctx.Defer(),
// reading user options with ctx.GetUserID(), and following up with ctx.Followup().
func handleUserInfo(ctx *bot.InteractionContext) {
	// 7. Deferring response to show a "bot is thinking..." status
	if err := ctx.Defer(); err != nil {
		log.Printf("Error deferring interaction: %v", err)
		return
	}

	// 5. Reading user option as a snowflake ID with ctx.GetUserID()
	targetID := ctx.GetUserID("user")
	if targetID == 0 {
		if u := ctx.User(); u != nil {
			targetID = u.ID
		}
	}

	// 7. Following up after deferral using ctx.Followup()
	msgText := fmt.Sprintf("📋 **User Information**\n• User Mention: <@%s>\n• User ID: `%s`", targetID, targetID)
	if _, err := ctx.Followup(msgText); err != nil {
		log.Printf("Error sending /userinfo followup: %v", err)
	}
}

// handleKick demonstrates reading user and string options with per-command permission middleware
// and responding ephemerally with ctx.ReplyEphemeral().
func handleKick(ctx *bot.InteractionContext) {
	// 5. Reading user option and string option
	targetID := ctx.GetUserID("user")
	reason := ctx.GetStringOption("reason")
	if reason == "" {
		reason = "No reason provided"
	}

	// 6. Using ctx.ReplyEphemeral() for private responses visible only to the invoker
	respText := fmt.Sprintf("👢 **User Kicked Successfully**\n• Target User: <@%s>\n• Reason: %s", targetID, reason)
	if err := ctx.ReplyEphemeral(respText); err != nil {
		log.Printf("Error sending /kick reply: %v", err)
	}
}

// handleServerInfo demonstrates constructing rich embeds with messages.NewEmbedBuilder()
// and replying using ctx.ReplyEmbed().
func handleServerInfo(ctx *bot.InteractionContext) {
	guildID := "Unknown"
	if ctx.InGuild() {
		guildID = ctx.GuildID().String()
	}

	// 6. Using messages.NewEmbedBuilder() for rich formatted embeds
	embed := messages.NewEmbedBuilder().
		SetTitle("🏰 Server Information").
		SetDescription("Overview of the current Discord guild.").
		SetColor(0x3498DB). // Vibrant Blue
		AddField("Guild ID", fmt.Sprintf("`%s`", guildID), true).
		AddField("Enforced Scope", "Guild Only (`bot.GuildOnly` middleware)", true).
		SetFooter("discord.go Slash Commands Example", "").
		Build()

	// 6. Using ctx.ReplyEmbed() to respond with an embed
	if err := ctx.ReplyEmbed(embed); err != nil {
		log.Printf("Error sending /serverinfo embed reply: %v", err)
	}
}
