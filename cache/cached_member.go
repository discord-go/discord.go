package cache

import (
	"github.com/discord-go/discord.go/permissions"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

// CachedMember pairs a guild member with pre-computed permission data.
// The Permissions field holds the member's guild-level permissions
// (owner bypass, @everyone base, role OR, and administrator shortcut)
// so callers can check permissions without rebuilding role maps on every
// call. Channel-specific overwrites are not included; resolve those with
// the bot's permission helpers, which layer channel overwrites on top of
// this value.
type CachedMember struct {
	// Member is the underlying guild member.
	Member *users.Member
	// GuildID is the guild the member belongs to.
	GuildID snowflake.ID
	// Permissions holds the pre-computed guild-level permissions.
	Permissions permissions.Permission
}
