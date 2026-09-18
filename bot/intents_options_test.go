package bot

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/discord-go/discord.go/intents"
)

// TestWithIntentsMerge pins that WithIntents unions with DefaultIntents
// instead of replacing them, so MessageContent is never dropped by accident.
func TestWithIntentsMerge(t *testing.T) {
	b := New("token", WithIntents(intents.GuildVoiceStates))
	if want := DefaultIntents() | intents.GuildVoiceStates; b.intentsVal != want {
		t.Fatalf("intents = %d, want %d", b.intentsVal, want)
	}

	// Multiple WithIntents calls union with each other and the defaults.
	b = New("token", WithIntents(intents.GuildInvites), WithIntents(intents.GuildMembers))
	if want := DefaultIntents() | intents.GuildInvites | intents.GuildMembers; b.intentsVal != want {
		t.Fatalf("intents = %d, want %d", b.intentsVal, want)
	}

	// Passing the default set is a no-op.
	b = New("token", WithIntents(DefaultIntents()))
	if b.intentsVal != DefaultIntents() {
		t.Fatalf("intents = %d, want %d", b.intentsVal, DefaultIntents())
	}

	// A bot with no intent options keeps the defaults.
	b = New("token")
	if b.intentsVal != DefaultIntents() {
		t.Fatalf("intents = %d, want %d", b.intentsVal, DefaultIntents())
	}
}

// TestWithIntentsExclusive pins that WithIntentsExclusive replaces the
// default set exactly.
func TestWithIntentsExclusive(t *testing.T) {
	replacement := intents.Guilds | intents.GuildMembers | intents.MessageContent
	b := New("token", WithIntentsExclusive(replacement))
	if b.intentsVal != replacement {
		t.Fatalf("intents = %d, want %d", b.intentsVal, replacement)
	}

	// Exclusive wins over an earlier additive call.
	b = New("token", WithIntents(intents.GuildVoiceStates), WithIntentsExclusive(replacement))
	if b.intentsVal != replacement {
		t.Fatalf("intents = %d, want %d", b.intentsVal, replacement)
	}

	// An additive call after exclusive extends the replacement.
	extended := replacement | intents.GuildInvites
	b = New("token", WithIntentsExclusive(replacement), WithIntents(intents.GuildInvites))
	if b.intentsVal != extended {
		t.Fatalf("intents = %d, want %d", b.intentsVal, extended)
	}
}

// TestWithIntentsExclusiveWarnsOnDroppedPrivilegedIntents pins the warning
// behavior: every dropped privileged intent that was enabled before is
// logged, and nothing else warns.
func TestWithIntentsExclusiveWarnsOnDroppedPrivilegedIntents(t *testing.T) {
	capture := func(t *testing.T, opts ...Option) (string, *Bot) {
		t.Helper()
		var buf bytes.Buffer
		logger := log.New(&buf, "", 0)
		b := New("token", append([]Option{WithLogger(logger)}, opts...)...)
		return buf.String(), b
	}

	// Dropping MessageContent (part of the defaults) warns and names it.
	output, _ := capture(t, WithIntentsExclusive(intents.Guilds|intents.GuildMessages))

	if !strings.Contains(output, "MessageContent") {
		t.Errorf("expected a MessageContent drop warning, got %q", output)
	}
	if !strings.Contains(output, "WithIntentsExclusive") {
		t.Errorf("warning should name WithIntentsExclusive, got %q", output)
	}

	// Privileged intents that were never enabled do not warn: only
	// MessageContent was in the default set.
	if strings.Contains(output, "GuildMembers") || strings.Contains(output, "GuildPresences") {
		t.Errorf("unexpected GuildMembers/GuildPresences warning, got %q", output)
	}

	// Non-privileged drops (GuildMessages) do not warn.
	if strings.Contains(output, "GuildMessages") {
		t.Errorf("unexpected non-privileged GuildMessages warning, got %q", output)
	}

	// Keeping every privileged intent stays silent.
	output, _ = capture(t, WithIntentsExclusive(DefaultIntents()))
	if output != "" {
		t.Errorf("expected no warning when privileged intents are retained, got %q", output)
	}

	// A merged-in privileged intent that exclusive drops warns too.
	output, _ = capture(t, WithIntents(intents.GuildMembers), WithIntentsExclusive(intents.Guilds))
	if !strings.Contains(output, "GuildMembers") {
		t.Errorf("expected a GuildMembers drop warning, got %q", output)
	}

	// Plain WithIntents never warns, even when it adds privileged intents.
	output, _ = capture(t, WithIntents(intents.GuildMembers|intents.GuildPresences))
	if output != "" {
		t.Errorf("expected no warning from WithIntents, got %q", output)
	}
}
