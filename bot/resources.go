package bot

import (
	"context"
	"time"

	"github.com/discord-go/discord.go/cache"
	"github.com/discord-go/discord.go/channels"
	"github.com/discord-go/discord.go/guilds"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

// GatewayLatency returns the latest heartbeat round-trip time.
func (b *Bot) GatewayLatency() time.Duration {
	b.stateMu.RLock()
	client := b.gwClient
	manager := b.shardManager
	b.stateMu.RUnlock()
	if client == nil && manager != nil {
		client = manager.Shard(0)
	}
	if client == nil || client.Heartbeater == nil {
		return 0
	}
	return client.Heartbeater.Ping()
}

// APILatency measures a lightweight REST request to the current-user endpoint.
func (b *Bot) APILatency(ctx context.Context) (time.Duration, error) {
	started := time.Now()
	_, err := b.Rest.GetCurrentUser(nonNilContext(ctx))
	return time.Since(started), err
}

// User returns a copy of the bot user received during READY.
func (b *Bot) User() *users.User {
	b.stateMu.RLock()
	defer b.stateMu.RUnlock()
	if b.user == nil {
		return nil
	}
	user := *b.user
	return &user
}

// ReadyAt returns the time the current run became ready.
func (b *Bot) ReadyAt() time.Time {
	b.stateMu.RLock()
	defer b.stateMu.RUnlock()
	return b.readyAt
}

// Uptime returns the time since READY, or zero when the bot is not ready.
func (b *Bot) Uptime() time.Duration {
	readyAt := b.ReadyAt()
	if readyAt.IsZero() {
		return 0
	}
	return time.Since(readyAt)
}

// CachedGuild returns a guild from the configured cache.
func (b *Bot) CachedGuild(id snowflake.ID) (*guilds.Guild, bool) {
	store, ok := b.cacheStore.(cache.GuildCache)
	if !ok {
		return nil, false
	}
	value, ok := store.GetGuild(id.String())
	return cachedValue[guilds.Guild](value, ok)
}

// CachedChannel returns a channel from the configured cache.
func (b *Bot) CachedChannel(id snowflake.ID) (*channels.Channel, bool) {
	store, ok := b.cacheStore.(cache.ChannelCache)
	if !ok {
		return nil, false
	}
	value, ok := store.GetChannel(id.String())
	return cachedValue[channels.Channel](value, ok)
}

// CachedUser returns a user from the configured cache.
func (b *Bot) CachedUser(id snowflake.ID) (*users.User, bool) {
	store, ok := b.cacheStore.(cache.UserCache)
	if !ok {
		return nil, false
	}
	value, ok := store.GetUser(id.String())
	return cachedValue[users.User](value, ok)
}

// CachedMember returns a guild member from the configured cache.
func (b *Bot) CachedMember(guildID, userID snowflake.ID) (*users.Member, bool) {
	store, ok := b.cacheStore.(cache.MemberCache)
	if !ok {
		return nil, false
	}
	value, ok := store.GetMember(guildID.String(), userID.String())
	return cachedValue[users.Member](value, ok)
}

// CachedMessage returns a message from the configured cache.
func (b *Bot) CachedMessage(id snowflake.ID) (*messages.Message, bool) {
	store, ok := b.cacheStore.(cache.MessageCache)
	if !ok {
		return nil, false
	}
	value, ok := store.GetMessage(id.String())
	return cachedValue[messages.Message](value, ok)
}

// FetchGuild fetches a guild and updates the configured cache.
func (b *Bot) FetchGuild(ctx context.Context, id snowflake.ID) (*guilds.Guild, error) {
	guild, err := b.Rest.GetGuild(nonNilContext(ctx), id)
	if err == nil {
		if store, ok := b.cacheStore.(cache.GuildCache); ok {
			store.SetGuild(id.String(), guild)
		}
	}
	return guild, err
}

// FetchChannel fetches a channel and updates the configured cache.
func (b *Bot) FetchChannel(ctx context.Context, id snowflake.ID) (*channels.Channel, error) {
	channel, err := b.Rest.GetChannel(nonNilContext(ctx), id)
	if err == nil {
		if store, ok := b.cacheStore.(cache.ChannelCache); ok {
			store.SetChannel(id.String(), channel)
		}
	}
	return channel, err
}

// FetchUser fetches a user and updates the configured cache.
func (b *Bot) FetchUser(ctx context.Context, id snowflake.ID) (*users.User, error) {
	user, err := b.Rest.GetUser(nonNilContext(ctx), id)
	if err == nil {
		if store, ok := b.cacheStore.(cache.UserCache); ok {
			store.SetUser(id.String(), user)
		}
	}
	return user, err
}

// FetchMember fetches a guild member and updates the configured cache.
func (b *Bot) FetchMember(ctx context.Context, guildID, userID snowflake.ID) (*users.Member, error) {
	member, err := b.Rest.GetGuildMember(nonNilContext(ctx), guildID, userID)
	if err == nil {
		if store, ok := b.cacheStore.(cache.MemberCache); ok {
			store.SetMember(guildID.String(), userID.String(), member)
		}
	}
	return member, err
}

// FetchMessage fetches a message and updates the configured cache.
func (b *Bot) FetchMessage(ctx context.Context, channelID, messageID snowflake.ID) (*messages.Message, error) {
	message, err := b.Rest.GetChannelMessage(nonNilContext(ctx), channelID, messageID)
	if err == nil {
		if store, ok := b.cacheStore.(cache.MessageCache); ok {
			store.SetMessage(messageID.String(), message)
		}
	}
	return message, err
}

func nonNilContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func cachedValue[T any](value any, ok bool) (*T, bool) {
	if !ok {
		return nil, false
	}
	switch typed := value.(type) {
	case *T:
		return typed, true
	case T:
		return &typed, true
	default:
		return nil, false
	}
}
