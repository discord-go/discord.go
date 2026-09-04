package bot

import (
	"encoding/json"
	"testing"

	"github.com/discord-go/discord.go/interactions"
)

// TestInteractionContext_GetAliases pins the short Get* alias family and the
// coercion contracts the option accessors rely on.
func TestInteractionContext_GetAliases(t *testing.T) {
	raw := []byte(`{
		"type": 2,
		"token": "t",
		"application_id": "1",
		"data": {
			"id": "10",
			"name": "cmd",
			"options": [
				{"name": "flag", "type": 5, "value": true},
				{"name": "text", "type": 3, "value": "hello"},
				{"name": "count", "type": 4, "value": 7},
				{"name": "ratio", "type": 10, "value": 2.5},
				{"name": "target", "type": 6, "value": "9007199254740993"}
			]
		}
	}`)
	var interaction interactions.Interaction
	if err := json.Unmarshal(raw, &interaction); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	ctx := newInteractionContext(BaseContext{}, &interaction)

	if !ctx.GetBool("flag") {
		t.Error("GetBool(flag) = false, want true")
	}
	if ctx.GetBool("missing") {
		t.Error("GetBool(missing) = true, want false")
	}
	if got := ctx.GetString("text"); got != "hello" {
		t.Errorf("GetString(text) = %q, want %q", got, "hello")
	}
	if got := ctx.GetInt("count"); got != 7 {
		t.Errorf("GetInt(count) = %d, want 7", got)
	}
	// Fractional numeric values truncate toward zero, not silently zero.
	rawFractional := []byte(`{
		"type": 2,
		"token": "t",
		"application_id": "1",
		"data": {"id": "10", "name": "cmd", "options": [{"name": "ratio", "type": 10, "value": 3.7}]}
	}`)
	var fractional interactions.Interaction
	if err := json.Unmarshal(rawFractional, &fractional); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	fctx := newInteractionContext(BaseContext{}, &fractional)
	if got := fctx.GetInt("ratio"); got != 3 {
		t.Errorf("GetInt(fractional 3.7) = %d, want 3 (truncate toward zero)", got)
	}
	if got := fctx.GetOption("ratio").Int(); got != 3 {
		t.Errorf("option.Int() on 3.7 = %d, want 3", got)
	}
	if got := ctx.GetFloat("ratio"); got != 2.5 {
		t.Errorf("GetFloat(ratio) = %v, want 2.5", got)
	}
	if got := ctx.GetSnowflake("target"); got.String() != "9007199254740993" {
		t.Errorf("GetSnowflake(target) = %s, want 9007199254740993 (precision must be preserved)", got)
	}
	if !ctx.GetSnowflake("missing").IsZero() {
		t.Error("GetSnowflake(missing) != 0, want zero")
	}
}
