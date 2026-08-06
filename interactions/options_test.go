package interactions

import (
	"encoding/json"
	"testing"
)

func TestApplicationCommandInteractionDataUnmarshal(t *testing.T) {
	data := []byte(`{"id":"222","name":"cmd","type":1,"options":[{"name":"opt1","type":3,"value":"val1"}]}`)
	var d ApplicationCommandInteractionData
	err := json.Unmarshal(data, &d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.ID != "222" || d.Name != "cmd" || len(d.Options) != 1 {
		t.Errorf("unexpected data: %+v", d)
	}
	if d.Options[0].Name != "opt1" || d.Options[0].Type != ApplicationCommandOptionTypeString {
		t.Errorf("unexpected option: %+v", d.Options[0])
	}
}
