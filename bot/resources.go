package bot

import (
	"context"
	"time"

	"github.com/discord-go/discord.go/cache"
	"github.com/discord-go/discord.go/channels"
	"github.com/discord-go/discord.go/guilds"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/permissions"
	"github.com/discord-go/discord.go/rest"
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

// RestClient returns the REST API client. The client is initialized during
// New and is safe to use immediately after construction, before Start or Run
// is called. This accessor provides a consistent method-based API alongside
// the public Rest field.
func (b *Bot) RestClient() *rest.Client {
	return b.Rest
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

// CachedMemberWithPermissions returns a guild member with pre-computed
// guild-level permissions from the configured cache. It returns false when
// the member, its guild, or the required cache capability is missing. The
// returned permissions include the owner bypass, the @everyone base, the
// member's role OR, and the administrator shortcut; channel overwrites are
// not applied.
func (b *Bot) CachedMemberWithPermissions(guildID, userID snowflake.ID) (*cache.CachedMember, bool) {
	member, ok := b.CachedMember(guildID, userID)
	if !ok || member == nil {
		return nil, false
	}
	guild, ok := b.CachedGuild(guildID)
	if !ok || guild == nil {
		return nil, false
	}
	perms := computeGuildPermissions(guild, member)
	return &cache.CachedMember{
		Member:      member,
		GuildID:     guildID,
		Permissions: perms,
	}, true
}

// computeGuildPermissions resolves a member's guild-level permissions
// without channel overwrites.
func computeGuildPermissions(guild *guilds.Guild, member *users.Member) permissions.Permission {
	if member.User != nil && member.User.ID == guild.OwnerID {
		return ^permissions.Permission(0)
	}

	everyone := permissions.Permission(0)
	rolePerms := make(map[snowflake.ID]permissions.Permission, len(guild.Roles))
	for i := range guild.Roles {
		role := guild.Roles[i]
		rolePerms[role.ID] = role.Permissions
		if role.ID == guild.ID {
			everyone = role.Permissions
		}
	}

	base := everyone
	memberRoles := make([]snowflake.ID, 0, len(member.Roles))
	for _, roleID := range member.Roles {
		if perm, ok := rolePerms[roleID]; ok {
			base.Add(perm)
			memberRoles = append(memberRoles, roleID)
		}
	}

	if base.Has(permissions.Administrator) {
		return ^permissions.Permission(0)
	}
	return base
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

// ChannelPermissions resolves the effective permissions for the bot in a
// channel using cached guild, channel, and member data. It returns zero when
// the required cache entries are missing; callers should treat zero as
// "unknown" rather than "no permissions".
func (b *Bot) ChannelPermissions(channelID snowflake.ID) permissions.Permission {
	return b.MemberChannelPermissionsFromCache(channelID, b.AppID())
}

// MemberChannelPermissionsFromCache resolves the effective permissions for a
// member in a channel using cached guild, channel, and member data. It returns
// zero when the required cache entries are missing; callers should treat zero
// as "unknown" rather than "no permissions".
func (b *Bot) MemberChannelPermissionsFromCache(channelID, userID snowflake.ID) permissions.Permission {
	if b == nil {
		return 0
	}
	channel, ok := b.CachedChannel(channelID)
	if !ok || channel == nil || channel.GuildID.IsZero() {
		return 0
	}
	guild, ok := b.CachedGuild(channel.GuildID)
	if !ok || guild == nil {
		return 0
	}
	member, ok := b.CachedMember(channel.GuildID, userID)
	if !ok || member == nil {
		return 0
	}
	return resolveMemberPermissions(guild, channel, member)
}

// MemberChannelPermissions resolves the effective permissions for a member in
// a channel using cached guild, channel, and member data. It returns zero when
// the required cache entries are missing.
func (b *Bot) MemberChannelPermissions(guildID, channelID, userID snowflake.ID) permissions.Permission {
	if b == nil {
		return 0
	}
	channel, ok := b.CachedChannel(channelID)
	if !ok || channel == nil {
		return 0
	}
	guild, ok := b.CachedGuild(guildID)
	if !ok || guild == nil {
		return 0
	}
	member, ok := b.CachedMember(guildID, userID)
	if !ok || member == nil {
		return 0
	}
	return resolveMemberPermissions(guild, channel, member)
}

func resolveMemberPermissions(guild *guilds.Guild, channel *channels.Channel, member *users.Member) permissions.Permission {
	// Guild owner bypasses all permission checks.
	if member.User != nil && member.User.ID == guild.OwnerID {
		return ^permissions.Permission(0)
	}

	// The @everyone role is the guild ID.
	everyone := permissions.Permission(0)
	rolePerms := make(map[snowflake.ID]permissions.Permission, len(guild.Roles))
	for i := range guild.Roles {
		role := guild.Roles[i]
		rolePerms[role.ID] = role.Permissions
		if role.ID == guild.ID {
			everyone = role.Permissions
		}
	}

	// Base permissions: @everyone plus all of the member's roles.
	base := everyone
	memberRoles := make([]snowflake.ID, 0, len(member.Roles))
	for _, roleID := range member.Roles {
		if perm, ok := rolePerms[roleID]; ok {
			base.Add(perm)
		}
		memberRoles = append(memberRoles, roleID)
	}

	overwrites := make([]permissions.Overwrite, 0, len(channel.PermissionOverwrites))
	for i := range channel.PermissionOverwrites {
		ow := channel.PermissionOverwrites[i]
		overwrites = append(overwrites, permissions.Overwrite{
			ID:    ow.ID,
			Type:  ow.Type,
			Allow: ow.Allow,
			Deny:  ow.Deny,
		})
	}

	var memberID snowflake.ID
	if member.User != nil {
		memberID = member.User.ID
	}
	return permissions.Calculate(memberID, guild.ID, guild.OwnerID, everyone, memberRoles, rolePermsValues(rolePerms, memberRoles), overwrites)
}

func rolePermsValues(rolePerms map[snowflake.ID]permissions.Permission, memberRoles []snowflake.ID) []permissions.Permission {
	values := make([]permissions.Permission, 0, len(memberRoles))
	for _, roleID := range memberRoles {
		if perm, ok := rolePerms[roleID]; ok {
			values = append(values, perm)
		}
	}
	return values
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
