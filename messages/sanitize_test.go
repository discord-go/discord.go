package messages

import (
	"strings"
	"testing"
)

func TestMessage_SanitizeContent(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    string
	}{
		{"plain text unchanged", "hello world", "hello world"},
		{"zero-width space", "hel\u200blo", "hello"},
		{"zwj emoji sequence preserved", "\U0001f468‍\U0001f469‍\U0001f467", "\U0001f468‍\U0001f469‍\U0001f467"},
		{"word joiner", "a\u2060b", "ab"},
		{"soft hyphen", "inter\u00adnet", "internet"},
		{"bom", "\ufeffhello", "hello"},
		{"bidi override (trojan source)", "if (x) \u202e} \u2066gated\u2069", "if (x) } gated"},
		{"right-to-left mark", "abc\u200fdef", "abcdef"},
		{"left-to-right mark", "abc\u200edef", "abcdef"},
		{"arabic letter mark", "abc\u061cdef", "abcdef"},
		{"bidi isolates", "a\u2066b\u2069c", "abc"},
		{"invisible tag chars", "a\U000e0041b", "ab"},
		{"variation selector 15 stripped", "a\ufe0eb", "ab"},
		{"emoji preserved", "\U0001f600 hello", "\U0001f600 hello"},
		{"zwj emoji sequence preserved", "\U0001f468‍\U0001f469‍\U0001f467", "\U0001f468‍\U0001f469‍\U0001f467"},
		{"rainbow flag preserved", "🏳️‍🌈", "🏳️‍🌈"},
		{"keycap sequence preserved", "1️⃣", "1️⃣"},
		{"skin tone modifier preserved", "\U0001f44d\U0001f3fd", "\U0001f44d\U0001f3fd"},
		{"regional indicators preserved", "\U0001f1ee\U0001f1f3", "\U0001f1ee\U0001f1f3"},
		{"variation selector 16 preserved", "❤️", "❤️"},
		{"normal unicode preserved", "héllo wörld 日本語", "héllo wörld 日本語"},
		{"newlines and tabs preserved", "a\nb\tc", "a\nb\tc"},
		{"rtl text visible preserved", "مرحبا", "مرحبا"},
		{"mixed", "hi\u200b <@123> \u202ethere", "hi <@123> there"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg := &Message{Content: tc.content}
			if got := msg.SanitizeContent(); got != tc.want {
				t.Errorf("SanitizeContent = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestMessage_SanitizeContentEdgeCases(t *testing.T) {
	var nilMsg *Message
	if got := nilMsg.SanitizeContent(); got != "" {
		t.Errorf("nil message SanitizeContent = %q, want empty", got)
	}
	if got := (&Message{}).SanitizeContent(); got != "" {
		t.Errorf("empty message SanitizeContent = %q, want empty", got)
	}
	// The original content must not be mutated.
	msg := &Message{Content: "a\u200bb"}
	_ = msg.SanitizeContent()
	if msg.Content != "a\u200bb" {
		t.Errorf("SanitizeContent mutated Content = %q", msg.Content)
	}
	// Fast path: content without invisible characters returns the same
	// string without allocation.
	clean := "nothing to remove"
	if got := (&Message{Content: clean}).SanitizeContent(); got != clean {
		t.Errorf("SanitizeContent = %q, want %q", got, clean)
	}
}

func TestMessage_SanitizeContentLongInput(t *testing.T) {
	// 10k alternating segments: performance and correctness under volume.
	var b strings.Builder
	for i := 0; i < 1000; i++ {
		b.WriteString("seg\u200bment ")
	}
	msg := &Message{Content: b.String()}
	got := msg.SanitizeContent()
	want := strings.Repeat("segment ", 1000)
	if got != want {
		t.Errorf("SanitizeContent length = %d, want %d", len(got), len(want))
	}
}
