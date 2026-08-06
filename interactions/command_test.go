package interactions

import (
	"encoding/json"
	"github.com/discord-go/discord.go/snowflake"
	"testing"
)

func TestApplicationCommandUnmarshal(t *testing.T) {
	data := []byte(`{"id":"111","name":"test","description":"desc","options":[{"type":3,"name":"opt","description":"optdesc"}]}`)
	var c ApplicationCommand
	err := json.Unmarshal(data, &c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.ID != snowflake.ID(111) || c.Name != "test" || len(c.Options) != 1 {
		t.Errorf("unexpected command: %+v", c)
	}
	if c.Options[0].Type != ApplicationCommandOptionTypeString {
		t.Errorf("unexpected option type: %v", c.Options[0].Type)
	}
}
