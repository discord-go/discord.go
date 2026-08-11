package messages

import (
	"testing"

	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/snowflake"
)

func TestMessageSendBuilder_Content(t *testing.T) {
	send := NewMessageSendBuilder().
		SetContent("hello").
		Build()

	if send.Content != "hello" {
		t.Errorf("expected content %q, got %q", "hello", send.Content)
	}
}

func TestMessageSendBuilder_Embeds(t *testing.T) {
	e1 := Embed{Title: "first"}
	e2 := Embed{Title: "second"}

	// SetEmbeds replaces
	send := NewMessageSendBuilder().
		SetEmbeds(e1, e2).
		Build()
	if len(send.Embeds) != 2 {
		t.Fatalf("expected 2 embeds, got %d", len(send.Embeds))
	}

	// AddEmbed appends
	send = NewMessageSendBuilder().
		AddEmbed(e1).
		AddEmbed(e2).
		Build()
	if len(send.Embeds) != 2 {
		t.Fatalf("expected 2 embeds via AddEmbed, got %d", len(send.Embeds))
	}
	if send.Embeds[0].Title != "first" {
		t.Errorf("expected first embed title %q, got %q", "first", send.Embeds[0].Title)
	}
}

func TestMessageSendBuilder_Components(t *testing.T) {
	btn := components.Button{CustomID: "btn1"}
	row := components.NewActionRowBuilder().AddComponents(btn).Build()

	// SetComponents replaces
	send := NewMessageSendBuilder().
		SetComponents(row).
		Build()
	if len(send.Components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(send.Components))
	}

	// AddComponent appends
	send = NewMessageSendBuilder().
		AddComponent(row).
		AddComponent(btn).
		Build()
	if len(send.Components) != 2 {
		t.Fatalf("expected 2 components via AddComponent, got %d", len(send.Components))
	}
}

func TestMessageSendBuilder_Flags(t *testing.T) {
	// SetFlags replaces
	send := NewMessageSendBuilder().
		SetFlags(FlagEphemeral).
		Build()
	if send.Flags != FlagEphemeral {
		t.Errorf("expected flags %d, got %d", FlagEphemeral, send.Flags)
	}

	// AddFlag ORs
	send = NewMessageSendBuilder().
		AddFlag(FlagEphemeral).
		AddFlag(FlagIsComponentsV2).
		Build()
	expected := FlagEphemeral | FlagIsComponentsV2
	if send.Flags != expected {
		t.Errorf("expected flags %d, got %d", expected, send.Flags)
	}
}

func TestMessageSendBuilder_AllFields(t *testing.T) {
	am := &AllowedMentions{Parse: []AllowedMentionType{AllowedMentionTypeUser}}
	ref := &MessageReference{MessageID: snowflake.ID(123)}
	poll := &Poll{Question: PollMedia{Text: "test?"}}

	send := NewMessageSendBuilder().
		SetContent("content").
		SetTTS(true).
		SetEmbeds(Embed{Title: "t"}).
		SetComponents(components.Button{CustomID: "c"}).
		SetFlags(FlagEphemeral).
		SetAllowedMentions(am).
		SetMessageReference(ref).
		SetNonce("n1").
		SetEnforceNonce(true).
		SetPoll(poll).
		AddAttachment(AttachmentSend{ID: "0", Filename: "f.txt"}).
		SetStickerIDs(snowflake.ID(1), snowflake.ID(2)).
		Build()

	if send.Content != "content" {
		t.Errorf("content mismatch")
	}
	if !send.TTS {
		t.Error("tts should be true")
	}
	if len(send.Embeds) != 1 {
		t.Errorf("expected 1 embed, got %d", len(send.Embeds))
	}
	if len(send.Components) != 1 {
		t.Errorf("expected 1 component, got %d", len(send.Components))
	}
	if send.Flags != FlagEphemeral {
		t.Error("flags mismatch")
	}
	if send.AllowedMentions == nil || len(send.AllowedMentions.Parse) != 1 {
		t.Error("allowed mentions mismatch")
	}
	if send.MessageReference == nil || send.MessageReference.MessageID != snowflake.ID(123) {
		t.Error("message reference mismatch")
	}
	if send.Nonce != "n1" {
		t.Error("nonce mismatch")
	}
	if !send.EnforceNonce {
		t.Error("enforce nonce should be true")
	}
	if send.Poll == nil || send.Poll.Question.Text != "test?" {
		t.Error("poll mismatch")
	}
	if len(send.Attachments) != 1 {
		t.Error("attachments mismatch")
	}
	if len(send.StickerIDs) != 2 {
		t.Error("sticker IDs mismatch")
	}
}

func TestMessageSendBuilder_Empty(t *testing.T) {
	send := NewMessageSendBuilder().Build()
	if send.Content != "" {
		t.Error("empty builder should have empty content")
	}
	if len(send.Embeds) != 0 {
		t.Error("empty builder should have no embeds")
	}
	if len(send.Components) != 0 {
		t.Error("empty builder should have no components")
	}
	if send.Flags != 0 {
		t.Error("empty builder should have no flags")
	}
}
