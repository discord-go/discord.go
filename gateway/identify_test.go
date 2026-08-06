package gateway

import (
	"encoding/json"
	"github.com/discord-go/discord.go/intents"
	"testing"
)

func TestIdentify_JSON(t *testing.T) {
	i := Identify{
		Token: "test_token",
		Properties: IdentifyProperties{
			OS:      "linux",
			Browser: "go",
			Device:  "go",
		},
		Compress:       true,
		LargeThreshold: 250,
		Shard:          []int{0, 1},
		Intents:        intents.Guilds,
	}

	b, err := json.Marshal(i)
	if err != nil {
		t.Fatal(err)
	}

	var i2 Identify
	err = json.Unmarshal(b, &i2)
	if err != nil {
		t.Fatal(err)
	}

	if i2.Token != i.Token {
		t.Errorf("expected %s, got %s", i.Token, i2.Token)
	}
}
