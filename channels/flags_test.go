package channels_test

import (
	"encoding/json"
	"testing"

	"github.com/discord-go/discord.go/channels"
	"github.com/discord-go/discord.go/permissions"
)

// TestObfuscatedChannel pins the Gateway channel obfuscation payload shape
// that becomes mandatory on November 16, 2026: name is "___hidden___", flags
// carry CHANNEL_OBFUSCATED, and permission_overwrites collapses to a single
// VIEW_CHANNEL deny for @everyone. The decoder must accept it and the flag
// must be detectable without inspecting any other field.
func TestObfuscatedChannel(t *testing.T) {
	var ch channels.Channel

	err := json.Unmarshal([]byte(`{
		"id": "123",
		"type": 0,
		"guild_id": "456",
		"name": "___hidden___",
		"flags": 131072,
		"permission_overwrites": [
			{
				"id": "456",
				"type": 0,
				"allow": "0",
				"deny": "1024"
			}
		]
	}`), &ch)

	if err != nil {
		t.Fatal(err)
	}

	if ch.Flags == nil || *ch.Flags != 1<<17 {
		t.Fatal("expected obfuscated channel flag")
	}
	if !ch.IsObfuscated() {
		t.Error("IsObfuscated() = false, want true")
	}
	if ch.Name == nil || *ch.Name != "___hidden___" {
		t.Errorf("expected redacted name, got %v", ch.Name)
	}
	if len(ch.PermissionOverwrites) != 1 {
		t.Fatalf("expected a single overwrite, got %d", len(ch.PermissionOverwrites))
	}
	if ch.PermissionOverwrites[0].Deny&permissions.ViewChannel == 0 {
		t.Error("expected the overwrite to deny VIEW_CHANNEL")
	}
}

func TestObfuscatedChannelFlagOnUnobfuscatedChannel(t *testing.T) {
	var ch channels.Channel
	if err := json.Unmarshal([]byte(`{"id":"1","type":0,"name":"general"}`), &ch); err != nil {
		t.Fatal(err)
	}
	if ch.IsObfuscated() {
		t.Error("IsObfuscated() = true for a plain channel, want false")
	}
	var redacted channels.Channel
	if err := json.Unmarshal([]byte(`{"id":"1","type":0,"flags":131072}`), &redacted); err != nil {
		t.Fatal(err)
	}
	if !redacted.IsObfuscated() {
		t.Error("IsObfuscated() = false when only the flag is set, want true")
	}
}

func TestChannelFlagValues(t *testing.T) {
	cases := map[string]struct {
		flag channels.ChannelFlags
		want channels.ChannelFlags
	}{
		"pinned":                      {channels.ChannelFlagPinned, 1 << 1},
		"require_tag":                 {channels.ChannelFlagRequireTag, 1 << 4},
		"hide_media_download_options": {channels.ChannelFlagHideMediaDownloadOptions, 1 << 15},
		"obfuscated":                  {channels.ChannelFlagObfuscated, 1 << 17},
		"is_spoiler_channel":          {channels.ChannelFlagIsSpoilerChannel, 1 << 21},
	}
	for name, tc := range cases {
		if tc.flag != tc.want {
			t.Errorf("%s: got %d, want %d", name, tc.flag, tc.want)
		}
	}

	combined := channels.ChannelFlagPinned | channels.ChannelFlagRequireTag
	if !combined.Has(channels.ChannelFlagPinned) {
		t.Error("Has(ChannelFlagPinned) = false on a combined bitfield")
	}
	if combined.Has(channels.ChannelFlagObfuscated) {
		t.Error("Has(ChannelFlagObfuscated) = true on a combined bitfield")
	}
	if !channels.ChannelFlags(0).Has(0) {
		t.Error("Has(0) = false, want true for the empty bit test")
	}
}

// TestChannelResolvedAppPermissions covers the July 2026 resolved-channel
// app_permissions field that interaction payloads carry on channel objects.
func TestChannelResolvedAppPermissions(t *testing.T) {
	var ch channels.Channel
	err := json.Unmarshal([]byte(`{
		"id": "123",
		"type": 0,
		"permissions": "1024",
		"app_permissions": "16384"
	}`), &ch)
	if err != nil {
		t.Fatal(err)
	}
	if ch.AppPermissions == nil || *ch.AppPermissions != "16384" {
		t.Errorf("expected app_permissions 16384, got %v", ch.AppPermissions)
	}
	if ch.Permissions == nil || *ch.Permissions != "1024" {
		t.Errorf("expected permissions 1024, got %v", ch.Permissions)
	}
}
