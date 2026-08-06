package emojis

import (
	"encoding/json"
	"testing"
)

func TestSticker_UnmarshalJSON(t *testing.T) {
	data := []byte(`{
		"id": "749054660769218631",
		"name": "Wave",
		"tags": "wumpus, hello",
		"type": 1,
		"format_type": 3,
		"description": "Wumpus waves hello",
		"pack_id": "847199849233514549",
		"sort_value": 12
	}`)
	var s Sticker
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.ID != 749054660769218631 {
		t.Errorf("expected ID 749054660769218631, got %v", s.ID)
	}
	if s.Type != StickerTypeStandard {
		t.Errorf("expected type 1, got %v", s.Type)
	}
	if s.FormatType != StickerFormatTypeLottie {
		t.Errorf("expected format type 3, got %v", s.FormatType)
	}
}

func TestStickerItem_UnmarshalJSON(t *testing.T) {
	data := []byte(`{
		"id": "749054660769218631",
		"name": "Wave",
		"format_type": 3
	}`)
	var s StickerItem
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.ID != 749054660769218631 {
		t.Errorf("expected ID 749054660769218631, got %v", s.ID)
	}
	if s.FormatType != StickerFormatTypeLottie {
		t.Errorf("expected format type 3, got %v", s.FormatType)
	}
}
