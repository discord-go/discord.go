// Package cdn builds Discord CDN and media-proxy asset URLs.
package cdn

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const (
	BaseURL       = "https://cdn.discordapp.com"
	MediaProxyURL = "https://media.discordapp.net"
)

type ImageExtension string

const (
	ExtensionPNG  ImageExtension = "png"
	ExtensionJPEG ImageExtension = "jpg"
	ExtensionWebP ImageExtension = "webp"
	ExtensionGIF  ImageExtension = "gif"
	ExtensionJSON ImageExtension = "json"
)

type Options struct {
	Extension  ImageExtension
	Size       int
	Animated   bool
	MediaProxy bool
}

func (o Options) suffix(hash string) string {
	extension := o.Extension
	if extension == "" {
		if strings.HasPrefix(hash, "a_") && o.Animated {
			extension = ExtensionGIF
		} else {
			extension = ExtensionPNG
		}
	}
	query := url.Values{}
	if o.Size > 0 {
		query.Set("size", strconv.Itoa(o.Size))
	}
	if o.Animated {
		query.Set("animated", "true")
	}
	if encoded := query.Encode(); encoded != "" {
		return "." + string(extension) + "?" + encoded
	}
	return "." + string(extension)
}

func base(options Options) string {
	if options.MediaProxy {
		return MediaProxyURL
	}
	return BaseURL
}

func Avatar(userID, hash string, options Options) string {
	return fmt.Sprintf("%s/avatars/%s/%s%s", base(options), userID, hash, options.suffix(hash))
}

func GuildIcon(guildID, hash string, options Options) string {
	return fmt.Sprintf("%s/icons/%s/%s%s", base(options), guildID, hash, options.suffix(hash))
}

func GuildSplash(guildID, hash string, options Options) string {
	return fmt.Sprintf("%s/splashes/%s/%s%s", base(options), guildID, hash, options.suffix(hash))
}

func DiscoverySplash(guildID, hash string, options Options) string {
	return fmt.Sprintf("%s/discovery-splashes/%s/%s%s", base(options), guildID, hash, options.suffix(hash))
}

func GuildBanner(guildID, hash string, options Options) string {
	return fmt.Sprintf("%s/banners/%s/%s%s", base(options), guildID, hash, options.suffix(hash))
}

func UserBanner(userID, hash string, options Options) string {
	return fmt.Sprintf("%s/banners/%s/%s%s", base(options), userID, hash, options.suffix(hash))
}

func MemberAvatar(guildID, userID, hash string, options Options) string {
	return fmt.Sprintf("%s/guilds/%s/users/%s/avatars/%s%s", base(options), guildID, userID, hash, options.suffix(hash))
}

func MemberBanner(guildID, userID, hash string, options Options) string {
	return fmt.Sprintf("%s/guilds/%s/users/%s/banners/%s%s", base(options), guildID, userID, hash, options.suffix(hash))
}

func RoleIcon(roleID, hash string, options Options) string {
	return fmt.Sprintf("%s/role-icons/%s/%s%s", base(options), roleID, hash, options.suffix(hash))
}

func Sticker(stickerID string, extension ImageExtension) string {
	if extension == "" {
		extension = ExtensionPNG
	}
	return fmt.Sprintf("%s/stickers/%s.%s", BaseURL, stickerID, extension)
}

func DefaultAvatar(index int) string {
	return fmt.Sprintf("%s/embed/avatars/%d.png", BaseURL, index)
}
