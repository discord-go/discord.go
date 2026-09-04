package bot

import (
	"encoding/json"
	"testing"

	"github.com/discord-go/discord.go/events"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/snowflake"
)

func TestTypedContexts_Decode(t *testing.T) {
	raw := json.RawMessage(`{"id":"123","channel_id":"456","guild_id":"789","content":"hello"}`)
	base := BaseContext{raw: raw}

	// MessageUpdateContext embeds BaseContext, so Decode is promoted.
	muc := &MessageUpdateContext{BaseContext: base}
	var mu messages.Message
	if err := muc.Decode(&mu); err != nil {
		t.Fatalf("MessageUpdateContext.Decode: %v", err)
	}
	if mu.ID != snowflake.ID(123) || mu.ChannelID != snowflake.ID(456) || mu.Content != "hello" {
		t.Errorf("decoded message = %+v", mu)
	}

	// MessageDeleteContext.
	mdc := &MessageDeleteContext{BaseContext: base}
	var md events.MessageDelete
	if err := mdc.Decode(&md); err != nil {
		t.Fatalf("MessageDeleteContext.Decode: %v", err)
	}
	if md.ID != snowflake.ID(123) || md.ChannelID != snowflake.ID(456) || md.GuildID != snowflake.ID(789) {
		t.Errorf("decoded delete = %+v", md)
	}

	// ChannelUpdateContext.
	cuc := &ChannelUpdateContext{BaseContext: base}
	var cu struct {
		ID snowflake.ID `json:"id,string"`
	}
	if err := cuc.Decode(&cu); err != nil {
		t.Fatalf("ChannelUpdateContext.Decode: %v", err)
	}
	if cu.ID != snowflake.ID(123) {
		t.Errorf("decoded channel id = %s", cu.ID)
	}
}

func TestBaseContext_DecodeNilAndRaw(t *testing.T) {
	var nilCtx *BaseContext
	if err := nilCtx.Decode(&struct{}{}); err == nil {
		t.Error("Expected error decoding on nil context")
	}
	if got := nilCtx.Raw(); got != nil {
		t.Errorf("nil Raw = %v, want nil", got)
	}

	raw := json.RawMessage(`{"a":1}`)
	base := BaseContext{raw: raw}
	var v struct {
		A int `json:"a"`
	}
	if err := base.Decode(&v); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if v.A != 1 {
		t.Errorf("A = %d, want 1", v.A)
	}
	// Raw returns a copy, not the backing slice.
	got := base.Raw()
	got[0] = 'X'
	if base.raw[0] == 'X' {
		t.Error("Raw returned the backing slice, want a copy")
	}
}

func TestTypedContext_DecodeUnmodeledField(t *testing.T) {
	raw := json.RawMessage(`{"id":"1","channel_id":"2","guild_id":"3","content":"x","unmodeled":{"deep":true}}`)
	base := BaseContext{raw: raw}
	muc := &MessageUpdateContext{BaseContext: base}
	var extra struct {
		Unmodeled struct {
			Deep bool `json:"deep"`
		} `json:"unmodeled"`
	}
	if err := muc.Decode(&extra); err != nil {
		t.Fatalf("Decode unmodeled field: %v", err)
	}
	if !extra.Unmodeled.Deep {
		t.Error("Expected unmodeled.deep = true")
	}
}
