package intents

import "testing"

func TestIntentConstants(t *testing.T) {
	tests := []struct {
		name     string
		intent   Intent
		expected Intent
	}{
		{"Guilds", Guilds, 1 << 0},
		{"GuildMembers", GuildMembers, 1 << 1},
		{"GuildBans", GuildBans, 1 << 2},
		{"GuildEmojisAndStickers", GuildEmojisAndStickers, 1 << 3},
		{"GuildIntegrations", GuildIntegrations, 1 << 4},
		{"GuildWebhooks", GuildWebhooks, 1 << 5},
		{"GuildInvites", GuildInvites, 1 << 6},
		{"GuildVoiceStates", GuildVoiceStates, 1 << 7},
		{"GuildPresences", GuildPresences, 1 << 8},
		{"GuildMessages", GuildMessages, 1 << 9},
		{"GuildMessageReactions", GuildMessageReactions, 1 << 10},
		{"GuildMessageTyping", GuildMessageTyping, 1 << 11},
		{"DirectMessages", DirectMessages, 1 << 12},
		{"DirectMessageReactions", DirectMessageReactions, 1 << 13},
		{"DirectMessageTyping", DirectMessageTyping, 1 << 14},
		{"MessageContent", MessageContent, 1 << 15},
		{"GuildScheduledEvents", GuildScheduledEvents, 1 << 16},
		{"AutoModerationConfiguration", AutoModerationConfiguration, 1 << 20},
		{"AutoModerationExecution", AutoModerationExecution, 1 << 21},
		{"GuildMessagePolls", GuildMessagePolls, 1 << 24},
		{"DirectMessagePolls", DirectMessagePolls, 1 << 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.intent != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, tt.intent)
			}
		})
	}
}
