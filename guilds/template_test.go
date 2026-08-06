package guilds

import (
	"encoding/json"
	"testing"
)

func TestTemplateUnmarshalJSON(t *testing.T) {
	data := []byte(`{
		"code": "2TffvPucqHkN",
		"name": "Template Name",
		"description": "Template description",
		"usage_count": 0,
		"creator_id": "123456789012345678",
		"creator": {
			"id": "123456789012345678",
			"username": "user",
			"discriminator": "0001",
			"avatar": "hash"
		},
		"created_at": "2021-01-01T00:00:00+00:00",
		"updated_at": "2021-01-01T00:00:00+00:00",
		"source_guild_id": "987654321098765432",
		"serialized_source_guild": {
			"name": "Guild Name",
			"description": null,
			"verification_level": 0,
			"default_message_notifications": 0,
			"explicit_content_filter": 0,
			"preferred_locale": "en-US",
			"afk_timeout": 300,
			"roles": [],
			"channels": [],
			"afk_channel_id": null,
			"system_channel_id": null,
			"system_channel_flags": 0
		},
		"is_dirty": null
	}`)

	var tmpl Template
	if err := json.Unmarshal(data, &tmpl); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tmpl.Code != "2TffvPucqHkN" {
		t.Errorf("expected 2TffvPucqHkN, got %v", tmpl.Code)
	}
	if tmpl.CreatorID.String() != "123456789012345678" {
		t.Errorf("expected 123456789012345678, got %v", tmpl.CreatorID)
	}
	if tmpl.SourceGuildID.String() != "987654321098765432" {
		t.Errorf("expected 987654321098765432, got %v", tmpl.SourceGuildID)
	}

	// Test invalid JSON syntax
	if err := tmpl.UnmarshalJSON([]byte(`{`)); err == nil {
		t.Error("expected error for invalid JSON syntax")
	}

	// Test invalid IDs inside JSON
	if err := json.Unmarshal([]byte(`{"creator_id": "invalid"}`), &tmpl); err == nil {
		t.Error("expected error for invalid creator_id")
	}
	if err := json.Unmarshal([]byte(`{"source_guild_id": "invalid"}`), &tmpl); err == nil {
		t.Error("expected error for invalid source_guild_id")
	}
}
