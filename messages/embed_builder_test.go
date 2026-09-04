package messages

import (
	"strings"
	"testing"
	"time"
)

func TestEmbedBuilder_Chain(t *testing.T) {
	before := time.Now()
	embed := NewEmbedBuilder().
		SetTitle("Title").
		SetDescription("Desc").
		SetURL("https://example.com").
		SetColor(0x5865F2).
		SetTimestamp(before).
		SetFooterText("Footer").
		SetImage("https://img").
		SetThumbnail("https://thumb").
		SetAuthorName("Author").
		AddInlineField("Name", "Value").
		Build()

	if embed.Title != "Title" {
		t.Errorf("Title = %q, want Title", embed.Title)
	}
	if embed.Description != "Desc" {
		t.Errorf("Description = %q, want Desc", embed.Description)
	}
	if embed.URL != "https://example.com" {
		t.Errorf("URL = %q, want https://example.com", embed.URL)
	}
	if embed.Color != 0x5865F2 {
		t.Errorf("Color = %#x, want 0x5865F2", embed.Color)
	}
	if embed.Timestamp != before.Format(time.RFC3339) {
		t.Errorf("Timestamp = %q, want %q", embed.Timestamp, before.Format(time.RFC3339))
	}
	if embed.Footer == nil || embed.Footer.Text != "Footer" || embed.Footer.IconURL != "" {
		t.Errorf("Footer = %+v, want text-only footer", embed.Footer)
	}
	if embed.Image == nil || embed.Image.URL != "https://img" {
		t.Errorf("Image = %+v, want https://img", embed.Image)
	}
	if embed.Thumbnail == nil || embed.Thumbnail.URL != "https://thumb" {
		t.Errorf("Thumbnail = %+v, want https://thumb", embed.Thumbnail)
	}
	if embed.Author == nil || embed.Author.Name != "Author" || embed.Author.URL != "" || embed.Author.IconURL != "" {
		t.Errorf("Author = %+v, want name-only author", embed.Author)
	}
	if len(embed.Fields) != 1 || embed.Fields[0].Name != "Name" || embed.Fields[0].Value != "Value" || !embed.Fields[0].Inline {
		t.Errorf("Fields = %+v, want one inline field", embed.Fields)
	}
}

func TestEmbedBuilder_SetTimestampNow(t *testing.T) {
	before := time.Now().Truncate(time.Second)
	embed := NewEmbedBuilder().SetTimestampNow().Build()
	after := time.Now().Truncate(time.Second)
	ts, err := time.Parse(time.RFC3339, embed.Timestamp)
	if err != nil {
		t.Fatalf("Timestamp not RFC3339: %v", err)
	}
	if ts.Before(before) || ts.After(after) {
		t.Errorf("Timestamp %v outside window [%v, %v]", ts, before, after)
	}
}

func TestEmbedBuilder_SetColorHex(t *testing.T) {
	cases := []struct {
		in   string
		want int
		ok   bool
	}{
		{"#5865F2", 0x5865F2, true},
		{"0x5865F2", 0x5865F2, true},
		{"5865F2", 0x5865F2, true},
		{"#abcdef", 0xabcdef, true},
		{"#ABCDEF", 0xabcdef, true},
		{"", 0, false},
		{"#12345", 0, false},
		{"#1234567", 0, false},
		{"#12zz56", 0, false},
		{"#12 56", 0, false},
		{"red", 0, false},
	}
	for _, tc := range cases {
		got, ok := parseHexColor(tc.in)
		if ok != tc.ok {
			t.Errorf("parseHexColor(%q) ok = %v, want %v", tc.in, ok, tc.ok)
			continue
		}
		if ok && got != tc.want {
			t.Errorf("parseHexColor(%q) = %#x, want %#x", tc.in, got, tc.want)
		}
	}
}

func TestEmbedBuilder_SetColorHexInvalidLeavesUnchanged(t *testing.T) {
	b := NewEmbedBuilder().SetColor(0x112233).SetColorHex("nope")
	if b.embed.Color != 0x112233 {
		t.Errorf("Color changed to %#x on invalid hex, want 0x112233", b.embed.Color)
	}
}

func TestEmbedBuilder_ClearFields(t *testing.T) {
	embed := NewEmbedBuilder().
		AddField("a", "1", true).
		AddField("b", "2", false).
		ClearFields().
		AddField("c", "3", true).
		Build()
	if len(embed.Fields) != 1 || embed.Fields[0].Name != "c" {
		t.Errorf("Fields = %+v, want only field c", embed.Fields)
	}
}

func TestEmbedBuilder_Validate(t *testing.T) {
	embed := NewEmbedBuilder().
		SetTitle(strings.Repeat("x", 300)).
		Build()
	if err := embed.Validate(); err == nil {
		t.Error("Expected validation error for oversized title")
	}
}
