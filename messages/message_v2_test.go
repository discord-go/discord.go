package messages

import (
	"encoding/json"
	"testing"

	"github.com/discord-go/discord.go/components"
)

func TestMessageUnmarshalComponentsV2(t *testing.T) {
	payload := []byte(`{"flags":32768,"components":[{"type":10,"content":"hello"},{"type":17,"accent_color":5793266,"components":[{"type":14,"divider":true,"spacing":1}]}]}`)
	var message Message
	if err := json.Unmarshal(payload, &message); err != nil {
		t.Fatal(err)
	}
	if message.Flags != FlagIsComponentsV2 || len(message.Components) != 2 {
		t.Fatalf("message = %#v", message)
	}
	if _, ok := message.Components[0].(components.TextDisplay); !ok {
		t.Fatalf("text display type = %T", message.Components[0])
	}
	container, ok := message.Components[1].(components.Container)
	if !ok || len(container.Components) != 1 {
		t.Fatalf("container = %#v", message.Components[1])
	}
}
