package users

import (
	"strings"
	"testing"

	"github.com/discord-go/discord.go/snowflake"
)

func TestAvatarURL(t *testing.T) {
	hash := "avatarhash"
	user := User{ID: snowflake.ID(123), Avatar: &hash}
	got := user.DisplayAvatarURL(AvatarURLOptions{Extension: "png", Size: 512})
	if got != "https://cdn.discordapp.com/avatars/123/avatarhash.png?size=512" {
		t.Fatalf("avatar URL = %q", got)
	}
	defaultURL := (User{ID: snowflake.ID(123)}).AvatarURL(AvatarURLOptions{})
	if !strings.HasPrefix(defaultURL, "https://cdn.discordapp.com/embed/avatars/") {
		t.Fatalf("default avatar URL = %q", defaultURL)
	}
}
