package bot

import (
	"encoding/json"
	"testing"

	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/rest"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
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

func TestRestClientAccessor(t *testing.T) {
	b := New("MTIz.NjQ1.abc123", WithCommandSyncDisabled())
	if b.RestClient() == nil {
		t.Fatal("expected RestClient() to return non-nil client")
	}
	if b.RestClient() != b.Rest {
		t.Error("expected RestClient() to return the same client as the Rest field")
	}
}

func TestStringPtr(t *testing.T) {
	p := rest.StringPtr("hello")
	if p == nil {
		t.Fatal("expected non-nil pointer")
	}
	if *p != "hello" {
		t.Errorf("expected *p = hello, got %s", *p)
	}

	empty := rest.StringPtr("")
	if empty == nil || *empty != "" {
		t.Error("expected non-nil pointer to empty string")
	}
}

func TestInteractionContextUser(t *testing.T) {
	// Guild interaction: user nested inside Member.
	guildUser := &users.User{ID: snowflake.ID(1), Username: "guilduser"}
	ic := newInteractionContext(BaseContext{}, &interactions.Interaction{
		Type:   interactions.InteractionTypeApplicationCommand,
		Member: &users.Member{User: guildUser},
	})
	if u := ic.User(); u == nil || u.ID != snowflake.ID(1) {
		t.Errorf("expected guild user, got %v", u)
	}

	// DM interaction: user on top-level field.
	dmUser := &users.User{ID: snowflake.ID(2), Username: "dmuser"}
	ic2 := newInteractionContext(BaseContext{}, &interactions.Interaction{
		Type: interactions.InteractionTypeApplicationCommand,
		User: dmUser,
	})
	if u := ic2.User(); u == nil || u.ID != snowflake.ID(2) {
		t.Errorf("expected DM user, got %v", u)
	}

	// No user at all.
	ic3 := newInteractionContext(BaseContext{}, &interactions.Interaction{
		Type: interactions.InteractionTypeApplicationCommand,
	})
	if u := ic3.User(); u != nil {
		t.Errorf("expected nil user, got %v", u)
	}
}

func TestInteractionContextGuildID(t *testing.T) {
	// Guild interaction.
	guildID := snowflake.ID(999)
	ic := newInteractionContext(BaseContext{}, &interactions.Interaction{
		Type:    interactions.InteractionTypeApplicationCommand,
		GuildID: &guildID,
	})
	if got := ic.GuildID(); got != guildID {
		t.Errorf("expected guild ID %s, got %s", guildID, got)
	}

	// DM interaction — no guild ID.
	ic2 := newInteractionContext(BaseContext{}, &interactions.Interaction{
		Type: interactions.InteractionTypeApplicationCommand,
	})
	if got := ic2.GuildID(); got != 0 {
		t.Errorf("expected zero guild ID, got %s", got)
	}
}

func TestInteractionContextChannelID(t *testing.T) {
	channelID := snowflake.ID(555)
	ic := newInteractionContext(BaseContext{}, &interactions.Interaction{
		Type:      interactions.InteractionTypeApplicationCommand,
		ChannelID: &channelID,
	})
	if got := ic.ChannelID(); got != channelID {
		t.Errorf("expected channel ID %s, got %s", channelID, got)
	}

	// No channel ID.
	ic2 := newInteractionContext(BaseContext{}, &interactions.Interaction{
		Type: interactions.InteractionTypeApplicationCommand,
	})
	if got := ic2.ChannelID(); got != 0 {
		t.Errorf("expected zero channel ID, got %s", got)
	}
}

func TestFocusedOptionString(t *testing.T) {
	// Autocomplete interaction with a focused string option.
	ic := newInteractionContext(BaseContext{}, &interactions.Interaction{
		Type: interactions.InteractionTypeApplicationCommandAutocomplete,
		Data: mustMarshal(t, interactionData{
			Options: []interactions.ApplicationCommandInteractionDataOption{
				{Name: "query", Type: interactions.ApplicationCommandOptionTypeString, Value: "hello", Focused: true},
			},
		}),
	})
	if got := ic.FocusedOptionString(); got != "hello" {
		t.Errorf("expected focused string 'hello', got %q", got)
	}

	// No focused option.
	ic2 := newInteractionContext(BaseContext{}, &interactions.Interaction{
		Type: interactions.InteractionTypeApplicationCommandAutocomplete,
		Data: mustMarshal(t, interactionData{
			Options: []interactions.ApplicationCommandInteractionDataOption{
				{Name: "query", Type: interactions.ApplicationCommandOptionTypeString, Value: "hello"},
			},
		}),
	})
	if got := ic2.FocusedOptionString(); got != "" {
		t.Errorf("expected empty string for no focused option, got %q", got)
	}

	// Focused option with non-string value.
	ic3 := newInteractionContext(BaseContext{}, &interactions.Interaction{
		Type: interactions.InteractionTypeApplicationCommandAutocomplete,
		Data: mustMarshal(t, interactionData{
			Options: []interactions.ApplicationCommandInteractionDataOption{
				{Name: "count", Type: interactions.ApplicationCommandOptionTypeInteger, Value: float64(42), Focused: true},
			},
		}),
	})
	if got := ic3.FocusedOptionString(); got != "" {
		t.Errorf("expected empty string for non-string focused value, got %q", got)
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

func TestButtonPrefixSuffix(t *testing.T) {
	router := NewRouter()
	var gotSuffix string
	router.ButtonPrefix("ticket:close:", func(ctx *InteractionContext) {
		gotSuffix = ctx.Suffix()
	})

	ic := newInteractionContext(BaseContext{}, &interactions.Interaction{
		Type: interactions.InteractionTypeMessageComponent,
		Data: mustMarshal(t, interactionData{
			CustomID:      "ticket:close:confirm:123456789",
			ComponentType: components.ComponentTypeButton,
		}),
	})
	router.handleInteraction(ic)

	if gotSuffix != "confirm:123456789" {
		t.Errorf("suffix = %q, want %q", gotSuffix, "confirm:123456789")
	}
}

func TestExactButtonSuffixIsEmpty(t *testing.T) {
	router := NewRouter()
	var gotSuffix string
	router.Button("confirm:42", func(ctx *InteractionContext) {
		gotSuffix = ctx.Suffix()
	})

	ic := newInteractionContext(BaseContext{}, &interactions.Interaction{
		Type: interactions.InteractionTypeMessageComponent,
		Data: mustMarshal(t, interactionData{
			CustomID:      "confirm:42",
			ComponentType: components.ComponentTypeButton,
		}),
	})
	router.handleInteraction(ic)

	if gotSuffix != "" {
		t.Errorf("suffix = %q, want empty for exact-match route", gotSuffix)
	}
}

func TestSelectPrefixSuffix(t *testing.T) {
	router := NewRouter()
	var gotSuffix string
	router.SelectPrefix("role:assign:", func(ctx *InteractionContext) {
		gotSuffix = ctx.Suffix()
	})

	ic := newInteractionContext(BaseContext{}, &interactions.Interaction{
		Type: interactions.InteractionTypeMessageComponent,
		Data: mustMarshal(t, interactionData{
			CustomID:      "role:assign:admin",
			ComponentType: components.ComponentTypeStringSelect,
		}),
	})
	router.handleInteraction(ic)

	if gotSuffix != "admin" {
		t.Errorf("suffix = %q, want %q", gotSuffix, "admin")
	}
}

func TestModalPrefixSuffix(t *testing.T) {
	router := NewRouter()
	var gotSuffix string
	router.ModalPrefix("ticket:modal:", func(ctx *InteractionContext) {
		gotSuffix = ctx.Suffix()
	})

	ic := newInteractionContext(BaseContext{}, &interactions.Interaction{
		Type: interactions.InteractionTypeModalSubmit,
		Data: mustMarshal(t, interactionData{CustomID: "ticket:modal:abc123"}),
	})
	router.handleInteraction(ic)

	if gotSuffix != "abc123" {
		t.Errorf("suffix = %q, want %q", gotSuffix, "abc123")
	}
}
