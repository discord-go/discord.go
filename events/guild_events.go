package events

import (
	"github.com/discord-go/discord.go/auditlog"
	"github.com/discord-go/discord.go/guilds"
	"github.com/discord-go/discord.go/snowflake"
)

// GuildCreate represents the GUILD_CREATE event.
type GuildCreate struct {
	guilds.Guild
}

// GuildUpdate represents the GUILD_UPDATE event.
type GuildUpdate struct {
	guilds.Guild
}

// GuildDelete represents the GUILD_DELETE event.
type GuildDelete struct {
	guilds.Guild
}

// GuildAuditLogEntryCreate represents the GUILD_AUDIT_LOG_ENTRY_CREATE event.
type GuildAuditLogEntryCreate struct {
	auditlog.AuditLogEntry
	GuildID snowflake.ID `json:"guild_id,string"`
}
