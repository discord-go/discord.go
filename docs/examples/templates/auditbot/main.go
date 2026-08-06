// discord.go code
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/discord-go/discord.go/auditlog"
	"github.com/discord-go/discord.go/client"
	"github.com/discord-go/discord.go/events"
	"github.com/discord-go/discord.go/gateway"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/snowflake"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN environment variable is required")
	}

	logChannelIDStr := os.Getenv("LOG_CHANNEL_ID")
	if logChannelIDStr == "" {
		log.Println("WARNING: LOG_CHANNEL_ID environment variable is not set. The bot will only print to the console.")
	}
	var logChannelID snowflake.ID
	if logChannelIDStr != "" {
		parsed, err := snowflake.Parse(logChannelIDStr)
		if err != nil {
			log.Fatalf("Invalid LOG_CHANNEL_ID: %v", err)
		}
		logChannelID = parsed
	}

	botIntents := intents.Guilds | intents.GuildBans | intents.GuildVoiceStates | intents.GuildMessages | intents.MessageContent

	bot := client.New(token, client.WithIntents(botIntents))

	bot.Gateway.Dispatcher.AddHandler(func(data []byte) {
		var payload gateway.GatewayPayload
		if err := json.Unmarshal(data, &payload); err != nil {
			return
		}

		if payload.Type != nil {
			if *payload.Type == "GUILD_AUDIT_LOG_ENTRY_CREATE" {
				var entry events.GuildAuditLogEntryCreate
				if err := json.Unmarshal(payload.Data, &entry); err != nil {
					log.Printf("Error unmarshalling audit log entry: %v", err)
					return
				}
				handleAuditLogEntry(bot, logChannelID, entry)
			} else if *payload.Type == "VOICE_STATE_UPDATE" {
				var vs struct {
					UserID    snowflake.ID  `json:"user_id,string"`
					ChannelID *snowflake.ID `json:"channel_id,string"`
				}
				if err := json.Unmarshal(payload.Data, &vs); err != nil {
					log.Printf("Error unmarshalling voice state: %v", err)
					return
				}
				handleVoiceState(bot, logChannelID, vs.UserID, vs.ChannelID)
			} else if *payload.Type == "THREAD_CREATE" {
				var thread struct {
					ID       snowflake.ID `json:"id,string"`
					Name     string       `json:"name"`
					OwnerID  snowflake.ID `json:"owner_id,string"`
					ParentID snowflake.ID `json:"parent_id,string"`
				}
				if err := json.Unmarshal(payload.Data, &thread); err != nil {
					log.Printf("ERROR unmarshalling THREAD_CREATE: %v (payload: %s)", err, string(payload.Data))
				} else {
					handleThreadEvent(bot, logChannelID, "Created", thread.ID, thread.Name, thread.OwnerID, thread.ParentID)
				}
			} else if *payload.Type == "THREAD_DELETE" {
				var thread struct {
					ID       snowflake.ID `json:"id,string"`
					ParentID snowflake.ID `json:"parent_id,string"`
				}
				if err := json.Unmarshal(payload.Data, &thread); err != nil {
					log.Printf("ERROR unmarshalling THREAD_DELETE: %v", err)
				} else {
					handleThreadEvent(bot, logChannelID, "Deleted", thread.ID, "Unknown", 0, thread.ParentID)
				}
			}
		}
	})

	log.Println("Starting AuditBot...")
	if err := bot.Gateway.Start(context.Background()); err != nil {
		log.Fatalf("Error starting bot: %v", err)
	}

	log.Println("Bot is running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func handleAuditLogEntry(bot *client.Client, logChannelID snowflake.ID, entry events.GuildAuditLogEntryCreate) {
	log.Printf("Received Audit Log Entry: ActionType=%d", entry.ActionType)

	var title string
	var color int
	var description string

	// Safely dereference pointers
	userIDStr := "Unknown"
	if entry.UserID != nil {
		userIDStr = fmt.Sprintf("<@%v>", *entry.UserID)
	}
	targetIDStr := "Unknown"
	if entry.TargetID != nil {
		targetIDStr = *entry.TargetID
	}

	switch entry.ActionType {
	case auditlog.GUILD_UPDATE:
		title = "Server Updated"
		color = 0x3498DB // Blue
		description = fmt.Sprintf("Server settings were updated by %s.", userIDStr)

	case auditlog.CHANNEL_CREATE:
		title = "Channel Created"
		color = 0x00FF00 // Green
		description = fmt.Sprintf("A new channel (ID: %s) was created by %s.", targetIDStr, userIDStr)
	case auditlog.CHANNEL_UPDATE:
		title = "Channel Updated"
		color = 0x3498DB
		description = fmt.Sprintf("Channel (ID: %s) was updated by %s.", targetIDStr, userIDStr)
	case auditlog.CHANNEL_DELETE:
		title = "Channel Deleted"
		color = 0xFF0000 // Red
		description = fmt.Sprintf("A channel (ID: %s) was deleted by %s.", targetIDStr, userIDStr)
	case auditlog.CHANNEL_OVERWRITE_CREATE:
		title = "Channel Overwrite Created"
		color = 0x00FF00
		description = fmt.Sprintf("A permissions overwrite for channel %s was created by %s.", targetIDStr, userIDStr)
	case auditlog.CHANNEL_OVERWRITE_UPDATE:
		title = "Channel Overwrite Updated"
		color = 0x3498DB
		description = fmt.Sprintf("A permissions overwrite for channel %s was updated by %s.", targetIDStr, userIDStr)
	case auditlog.CHANNEL_OVERWRITE_DELETE:
		title = "Channel Overwrite Deleted"
		color = 0xFF0000
		description = fmt.Sprintf("A permissions overwrite for channel %s was deleted by %s.", targetIDStr, userIDStr)

	case auditlog.MEMBER_KICK:
		title = "Member Kicked"
		color = 0xFFA500 // Orange
		description = fmt.Sprintf("User <@%s> was kicked by %s.", targetIDStr, userIDStr)
	case auditlog.MEMBER_PRUNE:
		title = "Members Pruned"
		color = 0xFF0000
		description = fmt.Sprintf("Members were pruned by %s.", userIDStr)
	case auditlog.MEMBER_BAN_ADD:
		title = "Member Banned"
		color = 0xFF0000 // Red
		description = fmt.Sprintf("User <@%s> was banned by %s.", targetIDStr, userIDStr)
	case auditlog.MEMBER_BAN_REMOVE:
		title = "Member Unbanned"
		color = 0x00FF00 // Green
		description = fmt.Sprintf("User <@%s> was unbanned by %s.", targetIDStr, userIDStr)
	case auditlog.MEMBER_UPDATE:
		title = "Member Updated"
		color = 0x3498DB
		description = fmt.Sprintf("User <@%s> was updated by %s.", targetIDStr, userIDStr)
	case auditlog.MEMBER_ROLE_UPDATE:
		title = "Member Roles Updated"
		color = 0x3498DB
		description = fmt.Sprintf("Roles for <@%s> were updated by %s.", targetIDStr, userIDStr)
	case auditlog.MEMBER_MOVE:
		title = "Member Moved"
		color = 0xFFA500
		description = fmt.Sprintf("User <@%s> was moved to another voice channel by %s.", targetIDStr, userIDStr)
	case auditlog.MEMBER_DISCONNECT:
		title = "Member Disconnected"
		color = 0xFFA500
		description = fmt.Sprintf("User <@%s> was disconnected from voice by %s.", targetIDStr, userIDStr)

	case auditlog.ROLE_CREATE:
		title = "Role Created"
		color = 0x00FF00
		description = fmt.Sprintf("A role (ID: %s) was created by %s.", targetIDStr, userIDStr)
	case auditlog.ROLE_UPDATE:
		title = "Role Updated"
		color = 0x3498DB
		description = fmt.Sprintf("Role (ID: %s) was updated by %s.", targetIDStr, userIDStr)
	case auditlog.ROLE_DELETE:
		title = "Role Deleted"
		color = 0xFF0000
		description = fmt.Sprintf("Role (ID: %s) was deleted by %s.", targetIDStr, userIDStr)

	case auditlog.INVITE_CREATE:
		title = "Invite Created"
		color = 0x00FF00
		description = fmt.Sprintf("An invite was created by %s.", userIDStr)
	case auditlog.INVITE_UPDATE:
		title = "Invite Updated"
		color = 0x3498DB
		description = fmt.Sprintf("An invite was updated by %s.", userIDStr)
	case auditlog.INVITE_DELETE:
		title = "Invite Deleted"
		color = 0xFF0000
		description = fmt.Sprintf("An invite was deleted by %s.", userIDStr)

	case auditlog.WEBHOOK_CREATE:
		title = "Webhook Created"
		color = 0x00FF00
		description = fmt.Sprintf("Webhook (ID: %s) was created by %s.", targetIDStr, userIDStr)
	case auditlog.WEBHOOK_UPDATE:
		title = "Webhook Updated"
		color = 0x3498DB
		description = fmt.Sprintf("Webhook (ID: %s) was updated by %s.", targetIDStr, userIDStr)
	case auditlog.WEBHOOK_DELETE:
		title = "Webhook Deleted"
		color = 0xFF0000
		description = fmt.Sprintf("Webhook (ID: %s) was deleted by %s.", targetIDStr, userIDStr)

	case auditlog.EMOJI_CREATE:
		title = "Emoji Created"
		color = 0x00FF00
		description = fmt.Sprintf("An emoji (ID: %s) was created by %s.", targetIDStr, userIDStr)
	case auditlog.EMOJI_UPDATE:
		title = "Emoji Updated"
		color = 0x3498DB
		description = fmt.Sprintf("An emoji (ID: %s) was updated by %s.", targetIDStr, userIDStr)
	case auditlog.EMOJI_DELETE:
		title = "Emoji Deleted"
		color = 0xFF0000
		description = fmt.Sprintf("An emoji (ID: %s) was deleted by %s.", targetIDStr, userIDStr)

	case auditlog.MESSAGE_DELETE:
		title = "Message Deleted"
		color = 0xFF0000
		description = fmt.Sprintf("A message was deleted by %s in channel %s.", userIDStr, targetIDStr)
	case auditlog.MESSAGE_BULK_DELETE:
		title = "Messages Bulk Deleted"
		color = 0xFF0000
		description = fmt.Sprintf("Messages were bulk deleted by %s in channel %s.", userIDStr, targetIDStr)
	case auditlog.MESSAGE_PIN:
		title = "Message Pinned"
		color = 0x3498DB
		description = fmt.Sprintf("A message was pinned by %s in channel %s.", userIDStr, targetIDStr)
	case auditlog.MESSAGE_UNPIN:
		title = "Message Unpinned"
		color = 0x3498DB
		description = fmt.Sprintf("A message was unpinned by %s in channel %s.", userIDStr, targetIDStr)

	case auditlog.THREAD_CREATE, auditlog.THREAD_UPDATE, auditlog.THREAD_DELETE:
		// We ignore these here because we handle THREAD_CREATE and THREAD_DELETE natively
		// via the gateway events for richer formatting (names, parent channels, etc).
		return

	case auditlog.STICKER_CREATE:
		title = "Sticker Created"
		color = 0x00FF00
		description = fmt.Sprintf("A sticker (ID: %s) was created by %s.", targetIDStr, userIDStr)
	case auditlog.STICKER_UPDATE:
		title = "Sticker Updated"
		color = 0x3498DB
		description = fmt.Sprintf("Sticker (ID: %s) was updated by %s.", targetIDStr, userIDStr)
	case auditlog.STICKER_DELETE:
		title = "Sticker Deleted"
		color = 0xFF0000
		description = fmt.Sprintf("Sticker (ID: %s) was deleted by %s.", targetIDStr, userIDStr)

	default:
		// For unhandled types, just log it.
		title = fmt.Sprintf("Audit Log Event: %d", entry.ActionType)
		color = 0x808080 // Gray
		description = fmt.Sprintf("Action type `%d` performed by %s on target `%s`.", entry.ActionType, userIDStr, targetIDStr)
	}

	if entry.Reason != "" {
		description += fmt.Sprintf("\n**Reason:** %s", entry.Reason)
	}

	embed := messages.Embed{
		Title:       title,
		Description: description,
		Color:       color,
	}

	// If no log channel configured, just print to console
	if logChannelID == 0 {
		log.Printf("[AUDIT] %s: %s", title, description)
		return
	}

	// Send to Discord
	sendReq := messages.MessageSend{
		Embeds: []messages.Embed{embed},
	}

	_, err := bot.Rest.CreateMessageComplex(context.Background(), logChannelID, sendReq)
	if err != nil {
		log.Printf("Failed to send audit log embed: %v", err)
	}
}

func handleVoiceState(bot *client.Client, logChannelID snowflake.ID, userID snowflake.ID, channelID *snowflake.ID) {
	title := "Voice Activity"
	color := 0x3498DB
	description := ""

	if channelID == nil {
		title = "Left Voice Channel"
		color = 0xFF0000
		description = fmt.Sprintf("<@%v> disconnected from voice.", userID)
	} else {
		title = "Joined/Moved Voice Channel"
		color = 0x00FF00
		description = fmt.Sprintf("<@%v> joined/moved to channel <#%v>.", userID, *channelID)
	}

	embed := messages.Embed{
		Title:       title,
		Description: description,
		Color:       color,
	}

	if logChannelID == 0 {
		log.Printf("[VOICE] %s: %s", title, description)
		return
	}

	sendReq := messages.MessageSend{
		Embeds: []messages.Embed{embed},
	}
	_, err := bot.Rest.CreateMessageComplex(context.Background(), logChannelID, sendReq)
	if err != nil {
		log.Printf("Failed to send voice log embed: %v", err)
	}
}

func handleThreadEvent(bot *client.Client, logChannelID snowflake.ID, action string, threadID snowflake.ID, name string, ownerID snowflake.ID, parentID snowflake.ID) {
	title := fmt.Sprintf("Thread %s", action)
	color := 0x3498DB // Blue default

	if action == "Created" {
		color = 0x00FF00
	} else if action == "Deleted" {
		color = 0xFF0000
	}

	ownerStr := "Unknown"
	if ownerID != 0 {
		ownerStr = fmt.Sprintf("<@%v>", ownerID)
	}

	description := fmt.Sprintf("Thread <#%v> (`%s`) was **%s** under channel <#%v>.\n**Thread Owner:** %s", threadID, name, action, parentID, ownerStr)

	embed := messages.Embed{
		Title:       title,
		Description: description,
		Color:       color,
	}

	if logChannelID == 0 {
		log.Printf("[THREAD] %s: %s", title, description)
		return
	}

	sendReq := messages.MessageSend{
		Embeds: []messages.Embed{embed},
	}
	_, err := bot.Rest.CreateMessageComplex(context.Background(), logChannelID, sendReq)
	if err != nil {
		log.Printf("Failed to send thread log embed: %v", err)
	}
}
