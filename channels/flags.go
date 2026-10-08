package channels

// ChannelFlags is a bitfield of Discord channel flags. The JSON wire format
// is an integer, so the type stays compatible with encoding/json while giving
// callers named constants instead of raw shifts.
type ChannelFlags int

const (
	// ChannelFlagPinned marks a thread pinned to the top of its parent
	// GUILD_FORUM or GUILD_MEDIA channel.
	ChannelFlagPinned ChannelFlags = 1 << 1

	// ChannelFlagRequireTag requires a tag when creating a thread in a
	// GUILD_FORUM or GUILD_MEDIA channel.
	ChannelFlagRequireTag ChannelFlags = 1 << 4

	// ChannelFlagHideMediaDownloadOptions hides the embedded media download
	// options. Available only for media channels.
	ChannelFlagHideMediaDownloadOptions ChannelFlags = 1 << 15

	// ChannelFlagObfuscated marks channel metadata that Discord redacted
	// because the bot lacks VIEW_CHANNEL on it. The channel's name becomes
	// "___hidden___", other sensitive fields are nulled or reduced, and
	// permission_overwrites holds a single overwrite denying VIEW_CHANNEL
	// for the guild's @everyone role. Only ever set on channels received
	// over the Gateway; the HTTP API omits those channels entirely from
	// GET /guilds/{guild.id}/channels starting November 16, 2026. Detect
	// obfuscation with this flag rather than inspecting name or any other
	// field.
	ChannelFlagObfuscated ChannelFlags = 1 << 17

	// ChannelFlagIsSpoilerChannel marks a channel users must opt in to view
	// (spoiler channel). Can be set on textual guild channels and voice
	// channels except GUILD_STAGE, and only while nsfw is false.
	ChannelFlagIsSpoilerChannel ChannelFlags = 1 << 21
)

// Has reports whether every bit in flag is set.
func (f ChannelFlags) Has(flag ChannelFlags) bool {
	return f&flag == flag
}

// IsObfuscated reports whether Discord redacted this channel's metadata
// because the bot cannot view it. Obfuscated channels keep id, type,
// position, and parent_id; every other field is unreliable until the
// Gateway dispatches a CHANNEL_UPDATE with the full data after the bot
// gains VIEW_CHANNEL.
func (c Channel) IsObfuscated() bool {
	return c.Flags != nil && c.Flags.Has(ChannelFlagObfuscated)
}
