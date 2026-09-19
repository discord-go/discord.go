package messages

import (
	"encoding/json"
	"testing"

	"github.com/discord-go/discord.go/snowflake"
)

// TestAllowedMentionsMarshalsUsersAsStrings pins the live 50035 incident:
// allowed_mentions.users serialized as a JSON number array was rejected by
// Discord with Invalid Form Body, so reminder pings never delivered.
func TestAllowedMentionsMarshalsUsersAsStrings(t *testing.T) {
	owner, err := snowflake.Parse("1528182903698100354")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	payload := MessageSend{
		Content: "probe",
		AllowedMentions: &AllowedMentions{
			Parse: []AllowedMentionType{AllowedMentionTypeUser},
			Users: []snowflake.ID{owner},
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := `{"content":"probe","allowed_mentions":{"parse":["users"],"users":["1528182903698100354"]}}`
	if string(data) != want {
		t.Errorf("marshal = %s, want %s", data, want)
	}
}
