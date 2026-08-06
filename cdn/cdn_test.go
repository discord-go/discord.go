package cdn

import "testing"

func TestAvatarURL(t *testing.T) {
	got := Avatar("1", "a_hash", Options{Extension: ExtensionWebP, Size: 512, Animated: true})
	want := "https://cdn.discordapp.com/avatars/1/a_hash.webp?animated=true&size=512"
	if got != want {
		t.Fatalf("URL = %q, want %q", got, want)
	}
}
