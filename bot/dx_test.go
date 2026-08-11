package bot

import (
	"encoding/json"
	"testing"

	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/messages"
)

func mustMarshal(t *testing.T, v interface{}) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	return b
}

func TestModalPrefix(t *testing.T) {
	router := NewRouter()
	called := make(chan string, 1)

	router.ModalPrefix("supreq_stop_modal_", func(ctx *InteractionContext) {
		called <- "stop"
	})
	router.ModalPrefix("supreq_cost_modal_", func(ctx *InteractionContext) {
		called <- "cost"
	})

	ic := newInteractionContext(BaseContext{}, &interactions.Interaction{
		Type: interactions.InteractionTypeModalSubmit,
		Data: mustMarshal(t, interactionData{CustomID: "supreq_cost_modal_abc123"}),
	})

	router.handleInteraction(ic)

	select {
	case result := <-called:
		if result != "cost" {
			t.Errorf("expected cost modal handler, got %s", result)
		}
	default:
		t.Error("no modal handler was called")
	}
}

func TestPrefixLongestMatch(t *testing.T) {
	router := NewRouter()
	called := make(chan string, 1)

	// Register a shorter prefix first, then a longer one.
	// The longer prefix should win for IDs that match both.
	router.ButtonPrefix("supreq_cost_", func(ctx *InteractionContext) {
		called <- "short"
	})
	router.ButtonPrefix("supreq_cost_done_", func(ctx *InteractionContext) {
		called <- "long"
	})

	ic := newInteractionContext(BaseContext{}, &interactions.Interaction{
		Type: interactions.InteractionTypeMessageComponent,
		Data: mustMarshal(t, interactionData{
			CustomID:      "supreq_cost_done_abc",
			ComponentType: components.ComponentTypeButton,
		}),
	})

	router.handleInteraction(ic)

	select {
	case result := <-called:
		if result != "long" {
			t.Errorf("expected longest prefix match (long), got %s", result)
		}
	default:
		t.Error("no button handler was called")
	}
}

func TestReplyEphemeralComplex(t *testing.T) {
	b := New("MTIz.NjQ1.abc123", WithCommandSyncDisabled())

	// Verify it returns error for nil data.
	err := (&InteractionContext{BaseContext: BaseContext{Bot: b}}).ReplyEphemeralComplex(nil)
	if err == nil {
		t.Error("expected error for nil data")
	}
}

func TestMessageFirstEmbed(t *testing.T) {
	// Test with embeds
	msg := &messages.Message{
		Embeds: []messages.Embed{{Title: "Test"}},
	}
	embed, ok := msg.FirstEmbed()
	if !ok {
		t.Error("expected FirstEmbed to return true")
	}
	if embed.Title != "Test" {
		t.Errorf("expected title 'Test', got %s", embed.Title)
	}

	// Test without embeds
	msg2 := &messages.Message{}
	_, ok2 := msg2.FirstEmbed()
	if ok2 {
		t.Error("expected FirstEmbed to return false for empty embeds")
	}

	// Test nil message
	var msg3 *messages.Message
	_, ok3 := msg3.FirstEmbed()
	if ok3 {
		t.Error("expected FirstEmbed to return false for nil message")
	}
}
