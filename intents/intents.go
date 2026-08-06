package intents

// Intent represents a Discord gateway intent.
type Intent uint64

// Discord Gateway Intents.
const (
	Guilds                      Intent = 1 << 0
	GuildMembers                Intent = 1 << 1
	GuildBans                   Intent = 1 << 2
	GuildEmojisAndStickers      Intent = 1 << 3
	GuildIntegrations           Intent = 1 << 4
	GuildWebhooks               Intent = 1 << 5
	GuildInvites                Intent = 1 << 6
	GuildVoiceStates            Intent = 1 << 7
	GuildPresences              Intent = 1 << 8
	GuildMessages               Intent = 1 << 9
	GuildMessageReactions       Intent = 1 << 10
	GuildMessageTyping          Intent = 1 << 11
	DirectMessages              Intent = 1 << 12
	DirectMessageReactions      Intent = 1 << 13
	DirectMessageTyping         Intent = 1 << 14
	MessageContent              Intent = 1 << 15
	GuildScheduledEvents        Intent = 1 << 16
	AutoModerationConfiguration Intent = 1 << 20
	AutoModerationExecution     Intent = 1 << 21
	GuildMessagePolls           Intent = 1 << 24
	DirectMessagePolls          Intent = 1 << 25
)
