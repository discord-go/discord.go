// Package main demonstrates a moderation bot built using the discord.go framework.
//
// This example covers:
//   - Slash commands with middleware for permission checks
//   - Deferred responses for slower operations
//   - Direct REST API calls (ban, kick, timeout, unban)
//   - Rich embed responses for all actions
//   - Event handlers for Ready and GuildCreate
//
// Usage:
//
//	export DISCORD_TOKEN="your-bot-token"
//	go run main.go
//
// discord.go code
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/permissions"
	"github.com/discord-go/discord.go/rest"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN environment variable is required")
	}

	// ── Router setup ────────────────────────────────────────────────────
	router := bot.NewRouter()

	// Global middleware: all commands require a server context
	router.Use(bot.GuildOnly())

	// /ban [user] [reason?]
	router.Command("ban", "Ban a member from the server", handleBan,
		interactions.ApplicationCommandOption{
			Type:        interactions.ApplicationCommandOptionTypeUser,
			Name:        "user",
			Description: "The user to ban",
			Required:    true,
		},
		interactions.ApplicationCommandOption{
			Type:        interactions.ApplicationCommandOptionTypeString,
			Name:        "reason",
			Description: "Reason for the ban",
		},
	).Use(bot.RequirePermissions(permissions.BanMembers))

	// /kick [user] [reason?]
	router.Command("kick", "Kick a member from the server", handleKick,
		interactions.ApplicationCommandOption{
			Type:        interactions.ApplicationCommandOptionTypeUser,
			Name:        "user",
			Description: "The user to kick",
			Required:    true,
		},
		interactions.ApplicationCommandOption{
			Type:        interactions.ApplicationCommandOptionTypeString,
			Name:        "reason",
			Description: "Reason for the kick",
		},
	).Use(bot.RequirePermissions(permissions.KickMembers))

	// /timeout [user] [duration] [reason?]
	router.Command("timeout", "Timeout a member (server mute)", handleTimeout,
		interactions.ApplicationCommandOption{
			Type:        interactions.ApplicationCommandOptionTypeUser,
			Name:        "user",
			Description: "The user to timeout",
			Required:    true,
		},
		interactions.ApplicationCommandOption{
			Type:        interactions.ApplicationCommandOptionTypeInteger,
			Name:        "duration",
			Description: "Duration in minutes (1-40320)",
			Required:    true,
		},
		interactions.ApplicationCommandOption{
			Type:        interactions.ApplicationCommandOptionTypeString,
			Name:        "reason",
			Description: "Reason for the timeout",
		},
	).Use(bot.RequirePermissions(permissions.ModerateMembers))

	// /unban [user]
	router.Command("unban", "Remove a ban from a user", handleUnban,
		interactions.ApplicationCommandOption{
			Type:        interactions.ApplicationCommandOptionTypeUser,
			Name:        "user",
			Description: "The user to unban",
			Required:    true,
		},
	).Use(bot.RequirePermissions(permissions.BanMembers))

	// ── Bot setup ───────────────────────────────────────────────────────
	b := bot.New(token,
		bot.WithIntents(intents.Guilds|intents.GuildMessages|intents.GuildMembers),
		bot.WithRouter(router),
	)

	// ── Event handlers ──────────────────────────────────────────────────

	b.OnReady(func(ctx *bot.ReadyContext) {
		log.Printf("✅ Moderation bot online as %s#%s", ctx.User.Username, ctx.User.Discriminator)
	})

	guildCount := 0
	b.OnGuildCreate(func(ctx *bot.GuildContext) {
		guildCount++
		log.Printf("📡 Connected to guild: %s (total: %d)", ctx.Name, guildCount)
	})

	// ── Launch ──────────────────────────────────────────────────────────
	log.Println("🚀 Starting moderation bot...")
	if err := b.Run(); err != nil {
		log.Fatalf("Fatal: %v", err)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Command Handlers
// ────────────────────────────────────────────────────────────────────────────

func handleBan(ctx *bot.InteractionContext) {
	userID := ctx.GetUserID("user")
	reason := ctx.GetStringOption("reason")
	if reason == "" {
		reason = "No reason provided"
	}

	// Defer since API calls can be slow
	if err := ctx.Defer(); err != nil {
		log.Printf("Error deferring ban: %v", err)
		return
	}

	// Execute the ban via REST API
	apiCtx := rest.WithReason(context.Background(), reason)
	err := ctx.Bot.Rest.CreateGuildBan(apiCtx, ctx.GuildID(), userID, rest.CreateBanParams{
		DeleteMessageSeconds: 86400, // Delete messages from last 24h
	})
	if err != nil {
		ctx.Followup(fmt.Sprintf("❌ Failed to ban <@%s>: `%v`", userID, err))
		return
	}

	// Send confirmation embed
	embed := messages.NewEmbedBuilder().
		SetTitle("🔨 Member Banned").
		SetColor(0xE74C3C).
		AddField("User", fmt.Sprintf("<@%s>", userID), true).
		AddField("Moderator", fmt.Sprintf("<@%s>", ctx.Member.User.ID), true).
		AddField("Reason", reason, false).
		SetFooter("Moderation Action", "").
		Build()

	ctx.FollowupEmbed(embed)
}

func handleKick(ctx *bot.InteractionContext) {
	userID := ctx.GetUserID("user")
	reason := ctx.GetStringOption("reason")
	if reason == "" {
		reason = "No reason provided"
	}

	if err := ctx.Defer(); err != nil {
		return
	}

	apiCtx := rest.WithReason(context.Background(), reason)
	err := ctx.Bot.Rest.RemoveGuildMember(apiCtx, ctx.GuildID(), userID)
	if err != nil {
		ctx.Followup(fmt.Sprintf("❌ Failed to kick <@%s>: `%v`", userID, err))
		return
	}

	embed := messages.NewEmbedBuilder().
		SetTitle("👢 Member Kicked").
		SetColor(0xF39C12).
		AddField("User", fmt.Sprintf("<@%s>", userID), true).
		AddField("Moderator", fmt.Sprintf("<@%s>", ctx.Member.User.ID), true).
		AddField("Reason", reason, false).
		SetFooter("Moderation Action", "").
		Build()

	ctx.FollowupEmbed(embed)
}

func handleTimeout(ctx *bot.InteractionContext) {
	userID := ctx.GetUserID("user")
	minutes := ctx.GetIntOption("duration")
	reason := ctx.GetStringOption("reason")
	if reason == "" {
		reason = "No reason provided"
	}

	// Clamp duration: 1 minute to 28 days (Discord limit)
	if minutes < 1 {
		minutes = 1
	}
	if minutes > 40320 {
		minutes = 40320
	}

	if err := ctx.Defer(); err != nil {
		return
	}

	// Calculate timeout end time
	until := time.Now().Add(time.Duration(minutes) * time.Minute)

	apiCtx := rest.WithReason(context.Background(), reason)
	_, err := ctx.Bot.Rest.ModifyGuildMember(apiCtx, ctx.GuildID(), userID, rest.ModifyMemberParams{
		CommunicationDisabledUntil: &until,
	})
	if err != nil {
		ctx.Followup(fmt.Sprintf("❌ Failed to timeout <@%s>: `%v`", userID, err))
		return
	}

	embed := messages.NewEmbedBuilder().
		SetTitle("🔇 Member Timed Out").
		SetColor(0x9B59B6).
		AddField("User", fmt.Sprintf("<@%s>", userID), true).
		AddField("Duration", fmt.Sprintf("%d minutes", minutes), true).
		AddField("Expires", fmt.Sprintf("<t:%d:R>", until.Unix()), true).
		AddField("Moderator", fmt.Sprintf("<@%s>", ctx.Member.User.ID), true).
		AddField("Reason", reason, false).
		SetFooter("Moderation Action", "").
		Build()

	ctx.FollowupEmbed(embed)
}

func handleUnban(ctx *bot.InteractionContext) {
	userID := ctx.GetUserID("user")

	if err := ctx.Defer(); err != nil {
		return
	}

	err := ctx.Bot.Rest.RemoveGuildBan(context.Background(), ctx.GuildID(), userID)
	if err != nil {
		ctx.Followup(fmt.Sprintf("❌ Failed to unban <@%s>: `%v`", userID, err))
		return
	}

	embed := messages.NewEmbedBuilder().
		SetTitle("🔓 Member Unbanned").
		SetColor(0x2ECC71).
		AddField("User", fmt.Sprintf("<@%s>", userID), true).
		AddField("Moderator", fmt.Sprintf("<@%s>", ctx.Member.User.ID), true).
		SetFooter("Moderation Action", "").
		Build()

	ctx.FollowupEmbed(embed)
}
