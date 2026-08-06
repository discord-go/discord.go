package emojis

import (
	"encoding/json"
	"testing"
)

func TestEmoji_UnmarshalJSON(t *testing.T) {
	t.Run("valid emoji with id and roles", func(t *testing.T) {
		data := []byte(`{
			"id": "41771983429993937",
			"name": "LUL",
			"roles": ["41771983429993937", "123456789"]
		}`)
		var e Emoji
		if err := json.Unmarshal(data, &e); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e.ID != 41771983429993937 {
			t.Errorf("expected ID 41771983429993937, got %v", e.ID)
		}
		if len(e.Roles) != 2 {
			t.Fatalf("expected 2 roles, got %v", len(e.Roles))
		}
		if e.Roles[0].ID != 41771983429993937 || e.Roles[1].ID != 123456789 {
			t.Errorf("expected roles [41771983429993937, 123456789], got %v", e.Roles)
		}
	})

	t.Run("unicode emoji without id", func(t *testing.T) {
		data := []byte(`{
			"id": null,
			"name": "🔥"
		}`)
		var e Emoji
		if err := json.Unmarshal(data, &e); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e.ID != 0 {
			t.Errorf("expected ID 0, got %v", e.ID)
		}
		if *e.Name != "🔥" {
			t.Errorf("expected name 🔥, got %v", *e.Name)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		data := []byte(`{
			"id": "invalid"
		}`)
		var e Emoji
		if err := json.Unmarshal(data, &e); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("invalid roles", func(t *testing.T) {
		data := []byte(`{
			"id": "41771983429993937",
			"roles": ["invalid"]
		}`)
		var e Emoji
		if err := json.Unmarshal(data, &e); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("invalid roles type", func(t *testing.T) {
		data := []byte(`{
			"roles": "not an array"
		}`)
		var e Emoji
		if err := json.Unmarshal(data, &e); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		data := []byte(`{invalid}`)
		var e Emoji
		if err := json.Unmarshal(data, &e); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
