package interactions

import (
	"encoding/json"
	"testing"
)

func TestInteractionResponseUnmarshal(t *testing.T) {
	data := []byte(`{"type":4,"data":{"content":"hello"}}`)
	var r InteractionResponse
	err := json.Unmarshal(data, &r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Type != InteractionCallbackTypeChannelMessageWithSource {
		t.Errorf("unexpected type: %v", r.Type)
	}
	if r.Data == nil || r.Data.Content != "hello" {
		t.Errorf("unexpected data: %+v", r.Data)
	}
}
