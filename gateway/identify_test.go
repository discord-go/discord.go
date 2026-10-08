package gateway

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/discord-go/discord.go/intents"
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

func TestIdentify_CapabilitiesJSON(t *testing.T) {
	i := Identify{
		Token:        "test_token",
		Intents:      intents.Guilds,
		Capabilities: CapabilityChannelObfuscation,
	}

	b, err := json.Marshal(i)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"capabilities":32768`) {
		t.Errorf("expected capabilities 32768 in identify payload, got %s", b)
	}

	var zero Identify
	zero.Token = "t"
	zero.Intents = intents.Guilds
	b, err = json.Marshal(zero)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "capabilities") {
		t.Errorf("expected capabilities to be omitted when unset, got %s", b)
	}
}

func TestCapabilityChannelObfuscationValue(t *testing.T) {
	if CapabilityChannelObfuscation != 1<<15 {
		t.Errorf("CapabilityChannelObfuscation = %d, want %d", CapabilityChannelObfuscation, 1<<15)
	}
}
