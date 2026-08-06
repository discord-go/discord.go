package cache

// Cache represents a generic cache interface.
type Cache interface {
	Get(key string) (any, bool)
	Set(key string, value any)
	Delete(key string)
	Clear()
}

// GuildCache represents a cache for Guild objects.
type GuildCache interface {
	Cache
	GetGuild(id string) (any, bool)
	SetGuild(id string, guild any)
	DeleteGuild(id string)
}

// ChannelCache represents a cache for Channel objects.
type ChannelCache interface {
	Cache
	GetChannel(id string) (any, bool)
	SetChannel(id string, channel any)
	DeleteChannel(id string)
}

// UserCache represents a cache for User objects.
type UserCache interface {
	Cache
	GetUser(id string) (any, bool)
	SetUser(id string, user any)
	DeleteUser(id string)
}

// RoleCache represents a cache for Role objects.
type RoleCache interface {
	Cache
	GetRole(id string) (any, bool)
	SetRole(id string, role any)
	DeleteRole(id string)
}

// MessageCache represents a cache for Message objects.
type MessageCache interface {
	Cache
	GetMessage(id string) (any, bool)
	SetMessage(id string, message any)
	DeleteMessage(id string)
}

// MemberCache represents a cache for Member objects.
type MemberCache interface {
	Cache
	GetMember(guildID, userID string) (any, bool)
	SetMember(guildID, userID string, member any)
	DeleteMember(guildID, userID string)
}
