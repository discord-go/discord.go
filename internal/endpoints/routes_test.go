package endpoints

import (
	"testing"
)

func TestConstants(t *testing.T) {
	if BaseURL != "https://discord.com/api/v10" {
		t.Errorf("BaseURL = %q, want %q", BaseURL, "https://discord.com/api/v10")
	}
	if CDNURL != "https://cdn.discordapp.com" {
		t.Errorf("CDNURL = %q, want %q", CDNURL, "https://cdn.discordapp.com")
	}
	if GatewayURL != "wss://gateway.discord.gg/?v=10&encoding=json" {
		t.Errorf("GatewayURL = %q, want %q", GatewayURL, "wss://gateway.discord.gg/?v=10&encoding=json")
	}
}

// endpointTest describes a single endpoint function test case.
type endpointTest struct {
	name string
	got  string
	want string
}

func TestChannelEndpoints(t *testing.T) {
	tests := []endpointTest{
		{"Channel", Channel("123"), BaseURL + "/channels/123"},
		{"ChannelMessages", ChannelMessages("123"), BaseURL + "/channels/123/messages"},
		{"ChannelMessage", ChannelMessage("123", "456"), BaseURL + "/channels/123/messages/456"},
		{"ChannelPermissions", ChannelPermissions("123", "789"), BaseURL + "/channels/123/permissions/789"},
		{"ChannelInvites", ChannelInvites("123"), BaseURL + "/channels/123/invites"},
		{"ChannelPins", ChannelPins("123"), BaseURL + "/channels/123/pins"},
		{"ChannelPin", ChannelPin("123", "456"), BaseURL + "/channels/123/pins/456"},
		{"ChannelTyping", ChannelTyping("123"), BaseURL + "/channels/123/typing"},
		{"ChannelWebhooks", ChannelWebhooks("123"), BaseURL + "/channels/123/webhooks"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s() = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestReactionEndpoints(t *testing.T) {
	tests := []endpointTest{
		{"MessageReactions", MessageReactions("123", "456", "🔥"), BaseURL + "/channels/123/messages/456/reactions/🔥"},
		{"MessageReactionUser", MessageReactionUser("123", "456", "🔥", "789"), BaseURL + "/channels/123/messages/456/reactions/🔥/789"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s() = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestGuildEndpoints(t *testing.T) {
	tests := []endpointTest{
		{"Guild", Guild("111"), BaseURL + "/guilds/111"},
		{"GuildChannels", GuildChannels("111"), BaseURL + "/guilds/111/channels"},
		{"GuildMembers", GuildMembers("111"), BaseURL + "/guilds/111/members"},
		{"GuildMember", GuildMember("111", "222"), BaseURL + "/guilds/111/members/222"},
		{"GuildBans", GuildBans("111"), BaseURL + "/guilds/111/bans"},
		{"GuildBan", GuildBan("111", "222"), BaseURL + "/guilds/111/bans/222"},
		{"GuildRoles", GuildRoles("111"), BaseURL + "/guilds/111/roles"},
		{"GuildRole", GuildRole("111", "333"), BaseURL + "/guilds/111/roles/333"},
		{"GuildEmojis", GuildEmojis("111"), BaseURL + "/guilds/111/emojis"},
		{"GuildEmoji", GuildEmoji("111", "444"), BaseURL + "/guilds/111/emojis/444"},
		{"GuildAuditLog", GuildAuditLog("111"), BaseURL + "/guilds/111/audit-logs"},
		{"GuildInvites", GuildInvites("111"), BaseURL + "/guilds/111/invites"},
		{"GuildWebhooks", GuildWebhooks("111"), BaseURL + "/guilds/111/webhooks"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s() = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestUserEndpoints(t *testing.T) {
	tests := []endpointTest{
		{"User", User("555"), BaseURL + "/users/555"},
		{"CurrentUser", CurrentUser(), BaseURL + "/users/@me"},
		{"CurrentUserGuilds", CurrentUserGuilds(), BaseURL + "/users/@me/guilds"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s() = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestWebhookEndpoints(t *testing.T) {
	tests := []endpointTest{
		{"Webhook", Webhook("666"), BaseURL + "/webhooks/666"},
		{"WebhookWithToken", WebhookWithToken("666", "tok"), BaseURL + "/webhooks/666/tok"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s() = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestInviteEndpoint(t *testing.T) {
	got := Invite("abcXYZ")
	want := BaseURL + "/invites/abcXYZ"
	if got != want {
		t.Errorf("Invite() = %q, want %q", got, want)
	}
}

func TestGatewayBot(t *testing.T) {
	got := GatewayBot()
	want := BaseURL + "/gateway/bot"
	if got != want {
		t.Errorf("GatewayBot() = %q, want %q", got, want)
	}
}

func TestCDNEndpoints(t *testing.T) {
	tests := []endpointTest{
		{"UserAvatar", UserAvatar("111", "abc123"), CDNURL + "/avatars/111/abc123.png"},
		{"GuildIcon", GuildIcon("222", "def456"), CDNURL + "/icons/222/def456.png"},
		{"DefaultUserAvatar", DefaultUserAvatar("3"), CDNURL + "/embed/avatars/3.png"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s() = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}
