package endpoints

import "fmt"

const (
	// BaseURL is the Discord REST API v10 base URL.
	BaseURL = "https://discord.com/api/v10"

	// CDNURL is the Discord CDN base URL for images, icons, and avatars.
	CDNURL = "https://cdn.discordapp.com"

	// GatewayURL is the Discord Gateway WebSocket URL.
	GatewayURL = "wss://gateway.discord.gg/?v=10&encoding=json"
)

// ---------- Channel endpoints ----------

// Channel returns the endpoint for a single channel.
func Channel(channelID string) string {
	return fmt.Sprintf("%s/channels/%s", BaseURL, channelID)
}

// ChannelMessages returns the endpoint for messages in a channel.
func ChannelMessages(channelID string) string {
	return fmt.Sprintf("%s/channels/%s/messages", BaseURL, channelID)
}

// ChannelMessage returns the endpoint for a specific message in a channel.
func ChannelMessage(channelID, messageID string) string {
	return fmt.Sprintf("%s/channels/%s/messages/%s", BaseURL, channelID, messageID)
}

// ChannelPermissions returns the endpoint for a channel permission overwrite.
func ChannelPermissions(channelID, overwriteID string) string {
	return fmt.Sprintf("%s/channels/%s/permissions/%s", BaseURL, channelID, overwriteID)
}

// ChannelInvites returns the endpoint for channel invites.
func ChannelInvites(channelID string) string {
	return fmt.Sprintf("%s/channels/%s/invites", BaseURL, channelID)
}

// ChannelPins returns the endpoint for pinned messages in a channel.
func ChannelPins(channelID string) string {
	return fmt.Sprintf("%s/channels/%s/pins", BaseURL, channelID)
}

// ChannelPin returns the endpoint for a specific pinned message.
func ChannelPin(channelID, messageID string) string {
	return fmt.Sprintf("%s/channels/%s/pins/%s", BaseURL, channelID, messageID)
}

// ChannelTyping returns the endpoint for triggering the typing indicator.
func ChannelTyping(channelID string) string {
	return fmt.Sprintf("%s/channels/%s/typing", BaseURL, channelID)
}

// ChannelWebhooks returns the endpoint for webhooks in a channel.
func ChannelWebhooks(channelID string) string {
	return fmt.Sprintf("%s/channels/%s/webhooks", BaseURL, channelID)
}

// ---------- Reaction endpoints ----------

// MessageReactions returns the endpoint for all reactions on a message.
func MessageReactions(channelID, messageID, emoji string) string {
	return fmt.Sprintf("%s/channels/%s/messages/%s/reactions/%s", BaseURL, channelID, messageID, emoji)
}

// MessageReactionUser returns the endpoint for a specific user's reaction.
func MessageReactionUser(channelID, messageID, emoji, userID string) string {
	return fmt.Sprintf("%s/channels/%s/messages/%s/reactions/%s/%s", BaseURL, channelID, messageID, emoji, userID)
}

// ---------- Guild endpoints ----------

// Guild returns the endpoint for a single guild.
func Guild(guildID string) string {
	return fmt.Sprintf("%s/guilds/%s", BaseURL, guildID)
}

// GuildChannels returns the endpoint for channels in a guild.
func GuildChannels(guildID string) string {
	return fmt.Sprintf("%s/guilds/%s/channels", BaseURL, guildID)
}

// GuildMembers returns the endpoint for members of a guild.
func GuildMembers(guildID string) string {
	return fmt.Sprintf("%s/guilds/%s/members", BaseURL, guildID)
}

// GuildMember returns the endpoint for a specific guild member.
func GuildMember(guildID, userID string) string {
	return fmt.Sprintf("%s/guilds/%s/members/%s", BaseURL, guildID, userID)
}

// GuildBans returns the endpoint for guild bans.
func GuildBans(guildID string) string {
	return fmt.Sprintf("%s/guilds/%s/bans", BaseURL, guildID)
}

// GuildBan returns the endpoint for a specific guild ban.
func GuildBan(guildID, userID string) string {
	return fmt.Sprintf("%s/guilds/%s/bans/%s", BaseURL, guildID, userID)
}

// GuildRoles returns the endpoint for roles in a guild.
func GuildRoles(guildID string) string {
	return fmt.Sprintf("%s/guilds/%s/roles", BaseURL, guildID)
}

// GuildRole returns the endpoint for a specific role in a guild.
func GuildRole(guildID, roleID string) string {
	return fmt.Sprintf("%s/guilds/%s/roles/%s", BaseURL, guildID, roleID)
}

// GuildEmojis returns the endpoint for emojis in a guild.
func GuildEmojis(guildID string) string {
	return fmt.Sprintf("%s/guilds/%s/emojis", BaseURL, guildID)
}

// GuildEmoji returns the endpoint for a specific emoji in a guild.
func GuildEmoji(guildID, emojiID string) string {
	return fmt.Sprintf("%s/guilds/%s/emojis/%s", BaseURL, guildID, emojiID)
}

// GuildAuditLog returns the endpoint for the guild audit log.
func GuildAuditLog(guildID string) string {
	return fmt.Sprintf("%s/guilds/%s/audit-logs", BaseURL, guildID)
}

// GuildInvites returns the endpoint for guild invites.
func GuildInvites(guildID string) string {
	return fmt.Sprintf("%s/guilds/%s/invites", BaseURL, guildID)
}

// GuildWebhooks returns the endpoint for guild webhooks.
func GuildWebhooks(guildID string) string {
	return fmt.Sprintf("%s/guilds/%s/webhooks", BaseURL, guildID)
}

// ---------- User endpoints ----------

// User returns the endpoint for a specific user.
func User(userID string) string {
	return fmt.Sprintf("%s/users/%s", BaseURL, userID)
}

// CurrentUser returns the endpoint for the current authenticated user.
func CurrentUser() string {
	return fmt.Sprintf("%s/users/@me", BaseURL)
}

// CurrentUserGuilds returns the endpoint for the current user's guilds.
func CurrentUserGuilds() string {
	return fmt.Sprintf("%s/users/@me/guilds", BaseURL)
}

// ---------- Webhook endpoints ----------

// Webhook returns the endpoint for a specific webhook.
func Webhook(webhookID string) string {
	return fmt.Sprintf("%s/webhooks/%s", BaseURL, webhookID)
}

// WebhookWithToken returns the endpoint for a webhook with its token.
func WebhookWithToken(webhookID, token string) string {
	return fmt.Sprintf("%s/webhooks/%s/%s", BaseURL, webhookID, token)
}

// ---------- Invite endpoints ----------

// Invite returns the endpoint for a specific invite.
func Invite(inviteCode string) string {
	return fmt.Sprintf("%s/invites/%s", BaseURL, inviteCode)
}

// ---------- Gateway endpoint ----------

// GatewayBot returns the endpoint for fetching the Gateway bot URL.
func GatewayBot() string {
	return fmt.Sprintf("%s/gateway/bot", BaseURL)
}

// ---------- CDN endpoints ----------

// UserAvatar returns the CDN URL for a user's avatar.
func UserAvatar(userID, avatarHash string) string {
	return fmt.Sprintf("%s/avatars/%s/%s.png", CDNURL, userID, avatarHash)
}

// GuildIcon returns the CDN URL for a guild's icon.
func GuildIcon(guildID, iconHash string) string {
	return fmt.Sprintf("%s/icons/%s/%s.png", CDNURL, guildID, iconHash)
}

// DefaultUserAvatar returns the CDN URL for a default user avatar by discriminator index.
func DefaultUserAvatar(index string) string {
	return fmt.Sprintf("%s/embed/avatars/%s.png", CDNURL, index)
}
