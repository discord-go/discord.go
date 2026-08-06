package rest

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/discord-go/discord.go/guilds"
	"github.com/discord-go/discord.go/snowflake"
)

func TestCriticalPayloadShapes(t *testing.T) {
	automod, err := json.Marshal(CreateAutoModerationRuleParams{
		Name:            "links",
		TriggerMetadata: &AutoModerationTriggerMetadata{KeywordFilter: []string{"token"}},
		Actions:         []AutoModerationAction{{Type: 1, Metadata: &AutoModerationActionMetadata{CustomMessage: "blocked"}}},
		ExemptRoles:     snowflake.IDs{1, 2},
	})
	if err != nil || !strings.Contains(string(automod), `"keyword_filter":["token"]`) || !strings.Contains(string(automod), `"exempt_roles":["1","2"]`) {
		t.Fatalf("automod payload = %s, err=%v", automod, err)
	}
	event, err := json.Marshal(CreateScheduledEventParams{
		Name:           "event",
		EntityMetadata: &guilds.ScheduledEventEntityMetadata{Location: "online"},
		RecurrenceRule: &guilds.ScheduledEventRecurrenceRule{Frequency: 1, Interval: 1},
	})
	if err != nil || !strings.Contains(string(event), `"location":"online"`) || !strings.Contains(string(event), `"recurrence_rule"`) {
		t.Fatalf("scheduled event payload = %s, err=%v", event, err)
	}
	sound, err := json.Marshal(CreateSoundboardSoundParams{Name: "beep", Sound: []byte("audio")})
	if err != nil || !strings.Contains(string(sound), `"sound":"data:audio/ogg;base64,`) {
		t.Fatalf("soundboard payload = %s, err=%v", sound, err)
	}
}
