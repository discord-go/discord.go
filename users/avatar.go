package users

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// AvatarURLOptions controls the CDN representation returned by AvatarURL.
type AvatarURLOptions struct {
	Extension string
	Size      int
}

// AvatarURL returns a Discord CDN URL for this user's avatar.
func (u User) AvatarURL(options AvatarURLOptions) string {
	if u.Avatar == nil || *u.Avatar == "" {
		index := (uint64(u.ID) >> 22) % 6
		return fmt.Sprintf("https://cdn.discordapp.com/embed/avatars/%d.png", index)
	}
	extension := strings.TrimPrefix(strings.ToLower(options.Extension), ".")
	if extension == "" {
		extension = "png"
	}
	if extension != "png" && extension != "jpg" && extension != "jpeg" && extension != "webp" && extension != "gif" {
		extension = "png"
	}
	result := fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.%s", u.ID, url.PathEscape(*u.Avatar), extension)
	if options.Size > 0 {
		result += "?size=" + strconv.Itoa(options.Size)
	}
	return result
}

// DisplayAvatarURL is a Discord.js-style alias for AvatarURL.
func (u User) DisplayAvatarURL(options AvatarURLOptions) string {
	return u.AvatarURL(options)
}
