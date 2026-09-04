package bot

import (
	"testing"

	"github.com/discord-go/discord.go/cache"
	"github.com/discord-go/discord.go/channels"
	"github.com/discord-go/discord.go/events"
	"github.com/discord-go/discord.go/guilds"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/permissions"
	"github.com/discord-go/discord.go/roles"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

func TestMessageContextMemberPermissions(t *testing.T) {
	guildID := snowflake.ID(100)
	channelID := snowflake.ID(200)
	userID := snowflake.ID(300)

	everyone := permissions.Permission(1 << 10)     // ViewChannel
	modRole := permissions.Permission(1<<5 | 1<<11) // ManageGuild | SendMessages

	b := New("token", WithCache(cache.NewMemoryCache()))
	store := b.cacheStore.(*cache.MemoryCache)
	store.SetGuild(guildID.String(), &guilds.Guild{
		ID:      guildID,
		OwnerID: snowflake.ID(999),
		Roles: []roles.Role{
			{ID: guildID, Permissions: everyone},
			{ID: snowflake.ID(50), Permissions: modRole},
		},
	})
	store.SetChannel(channelID.String(), &channels.Channel{
		ID:      channelID,
		GuildID: guildID,
	})

	// No member data on the event.
	ctx := &MessageContext{
		BaseContext:   BaseContext{Bot: b},
		MessageCreate: &events.MessageCreate{Message: messages.Message{ChannelID: channelID}},
	}
	if got := ctx.MemberPermissions(); got != 0 {
		t.Errorf("MemberPermissions without member = %d, want 0", got)
	}

	// Gateway MESSAGE_CREATE payloads do not carry member.permissions, so the
	// helper must compute from cached guild roles and channel overwrites.
	ctx.MessageCreate.Member = &users.Member{
		User:  &users.User{ID: userID},
		Roles: []snowflake.ID{snowflake.ID(50)},
	}
	ctx.MessageCreate.GuildID = guildID
	if got := ctx.MemberPermissions(); !got.HasAll(modRole) {
		t.Errorf("MemberPermissions = %d, want all of %d", got, modRole)
	}

	// A deny overwrite for the member removes SendMessages in the channel.
	store.SetChannel(channelID.String(), &channels.Channel{
		ID:      channelID,
		GuildID: guildID,
		PermissionOverwrites: []channels.Overwrite{
			{ID: userID, Type: 1, Allow: 0, Deny: permissions.Permission(1 << 11)},
		},
	})
	if got := ctx.MemberPermissions(); got.Has(permissions.SendMessages) {
		t.Errorf("MemberPermissions with member deny = %d, SendMessages should be denied", got)
	}

	// Guild not cached -> zero.
	ctx.MessageCreate.GuildID = snowflake.ID(555)
	if got := ctx.MemberPermissions(); got != 0 {
		t.Errorf("MemberPermissions with unknown guild = %d, want 0", got)
	}
}

func TestChannelPermissionsResolution(t *testing.T) {
	guildID := snowflake.ID(100)
	channelID := snowflake.ID(200)
	userID := snowflake.ID(300)
	botID := snowflake.ID(400)

	everyone := permissions.Permission(1 << 11)     // SendMessages
	modRole := permissions.Permission(1<<5 | 1<<11) // ManageGuild | SendMessages

	b := New("token", WithCache(cache.NewMemoryCache()))
	store := b.cacheStore.(*cache.MemoryCache)

	store.SetGuild(guildID.String(), &guilds.Guild{
		ID:      guildID,
		OwnerID: snowflake.ID(999),
		Roles: []roles.Role{
			{ID: guildID, Permissions: everyone},
			{ID: snowflake.ID(50), Permissions: modRole},
		},
	})
	store.SetChannel(channelID.String(), &channels.Channel{
		ID:      channelID,
		GuildID: guildID,
	})
	store.SetMember(guildID.String(), userID.String(), &users.Member{
		User:  &users.User{ID: userID},
		Roles: []snowflake.ID{snowflake.ID(50)},
	})
	store.SetMember(guildID.String(), botID.String(), &users.Member{
		User:  &users.User{ID: botID},
		Roles: []snowflake.ID{snowflake.ID(50)},
	})

	// Member with the mod role should have ManageGuild | SendMessages.
	got := b.MemberChannelPermissions(guildID, channelID, userID)
	if !got.HasAll(modRole) {
		t.Errorf("member permissions = %d, want all of %d", got, modRole)
	}

	// Bot permissions resolve the same way.
	botPerms := b.ChannelPermissions(channelID)
	if botPerms != 0 {
		t.Errorf("bot permissions without AppID = %d, want 0", botPerms)
	}
	botPermsCached := b.MemberChannelPermissionsFromCache(channelID, botID)
	if !botPermsCached.HasAll(modRole) {
		t.Errorf("cached bot permissions = %d, want all of %d", botPermsCached, modRole)
	}

	// Unknown channel returns zero.
	if got := b.ChannelPermissions(snowflake.ID(999)); got != 0 {
		t.Errorf("unknown channel permissions = %d, want 0", got)
	}
}

func TestMessageContextBotPermissions(t *testing.T) {
	guildID := snowflake.ID(100)
	channelID := snowflake.ID(200)
	botID := snowflake.ID(400)

	b := New("token", WithCache(cache.NewMemoryCache()))
	store := b.cacheStore.(*cache.MemoryCache)
	store.SetGuild(guildID.String(), &guilds.Guild{
		ID:      guildID,
		OwnerID: snowflake.ID(999),
		Roles: []roles.Role{
			{ID: guildID, Permissions: permissions.Permission(1 << 11)},
		},
	})
	store.SetChannel(channelID.String(), &channels.Channel{
		ID:      channelID,
		GuildID: guildID,
	})
	store.SetMember(guildID.String(), botID.String(), &users.Member{
		User:  &users.User{ID: botID},
		Roles: []snowflake.ID{},
	})

	// No cache -> zero.
	ctx := &MessageContext{BaseContext: BaseContext{Bot: b}, MessageCreate: &events.MessageCreate{Message: messages.Message{ChannelID: snowflake.ID(999)}}}
	if got := ctx.BotPermissions(); got != 0 {
		t.Errorf("BotPermissions without cache = %d, want 0", got)
	}

	// With cache -> SendMessages.
	b.appID = botID
	ctx2 := &MessageContext{BaseContext: BaseContext{Bot: b}, MessageCreate: &events.MessageCreate{Message: messages.Message{ChannelID: channelID}}}
	if got := ctx2.BotPermissions(); !got.Has(permissions.SendMessages) {
		t.Errorf("BotPermissions = %d, want SendMessages", got)
	}
}

func TestCachedMemberWithPermissions(t *testing.T) {
	guildID := snowflake.ID(100)
	channelID := snowflake.ID(200)
	userID := snowflake.ID(300)
	everyone := permissions.Permission(1 << 10) // ViewChannel
	modRole := permissions.Build(permissions.ManageGuild, permissions.SendMessages)

	b := New("token", WithCache(cache.NewMemoryCache()))
	store := b.cacheStore.(*cache.MemoryCache)
	store.SetGuild(guildID.String(), &guilds.Guild{
		ID:      guildID,
		OwnerID: snowflake.ID(999),
		Roles: []roles.Role{
			{ID: guildID, Permissions: everyone},
			{ID: snowflake.ID(50), Permissions: modRole},
		},
	})
	store.SetChannel(channelID.String(), &channels.Channel{
		ID:      channelID,
		GuildID: guildID,
	})
	store.SetMember(guildID.String(), userID.String(), &users.Member{
		User:  &users.User{ID: userID},
		Roles: []snowflake.ID{snowflake.ID(50)},
	})

	cm, ok := b.CachedMemberWithPermissions(guildID, userID)
	if !ok {
		t.Fatal("expected cached member with permissions")
	}
	if cm.GuildID != guildID {
		t.Errorf("GuildID = %s, want %s", cm.GuildID, guildID)
	}
	if cm.Member == nil || cm.Member.User.ID != userID {
		t.Errorf("Member = %+v, want user %s", cm.Member, userID)
	}
	if !cm.Permissions.HasAll(modRole) {
		t.Errorf("Permissions = %d, want all of %d", cm.Permissions, modRole)
	}
}

func TestCachedMemberWithPermissionsMissing(t *testing.T) {
	b := New("token", WithCache(cache.NewMemoryCache()))
	store := b.cacheStore.(*cache.MemoryCache)
	store.SetGuild(snowflake.ID(100).String(), &guilds.Guild{ID: snowflake.ID(100), OwnerID: snowflake.ID(999)})

	if _, ok := b.CachedMemberWithPermissions(snowflake.ID(100), snowflake.ID(300)); ok {
		t.Error("expected no cached member when member absent")
	}
	store.SetMember(snowflake.ID(100).String(), snowflake.ID(300).String(), &users.Member{User: &users.User{ID: snowflake.ID(300)}})
	store.DeleteGuild(snowflake.ID(100).String())
	if _, ok := b.CachedMemberWithPermissions(snowflake.ID(100), snowflake.ID(300)); ok {
		t.Error("expected no cached member when guild absent")
	}
}

func TestCachedMemberWithPermissionsOwner(t *testing.T) {
	guildID := snowflake.ID(100)
	ownerID := snowflake.ID(999)
	b := New("token", WithCache(cache.NewMemoryCache()))
	store := b.cacheStore.(*cache.MemoryCache)
	store.SetGuild(guildID.String(), &guilds.Guild{
		ID:      guildID,
		OwnerID: ownerID,
		Roles:   []roles.Role{{ID: guildID, Permissions: 0}},
	})
	store.SetMember(guildID.String(), ownerID.String(), &users.Member{
		User:  &users.User{ID: ownerID},
		Roles: []snowflake.ID{},
	})
	cm, ok := b.CachedMemberWithPermissions(guildID, ownerID)
	if !ok {
		t.Fatal("expected cached member")
	}
	if cm.Permissions != ^permissions.Permission(0) {
		t.Errorf("owner permissions = %d, want all bits", cm.Permissions)
	}
}
