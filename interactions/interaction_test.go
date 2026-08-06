package interactions

import (
	"encoding/json"
	"github.com/discord-go/discord.go/snowflake"
	"testing"
)

func TestInteractionUnmarshal(t *testing.T) {
	data := []byte(`{"id":"123","application_id":"456","type":2,"token":"abc","version":1}`)
	var i Interaction
	err := json.Unmarshal(data, &i)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if i.ID != snowflake.ID(123) {
		t.Errorf("expected id 123, got %d", i.ID)
	}
	if i.Type != InteractionTypeApplicationCommand {
		t.Errorf("expected type 2, got %d", i.Type)
	}
}
