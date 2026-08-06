# CDN

## Overview

The `cdn` package creates URLs without making HTTP requests. It covers user
avatars and banners, guild icons, banners, splashes and discovery splashes,
member avatars and banners, role icons, stickers, and default avatars. The
functions accept string IDs and hashes because a URL builder should not need a
REST model to be useful in templates or serializers.

## Architecture

`BaseURL` is `https://cdn.discordapp.com`; `MediaProxyURL` is
`https://media.discordapp.net`. Asset functions choose the base from
`Options.MediaProxy`. `Options.Extension` selects the suffix, `Size` adds a
`size` query parameter when positive, and `Animated` adds `animated=true`.
When the extension is empty, hashes beginning with `a_` use GIF only when
`Animated` is true; all other empty extensions default to PNG. `Sticker` and
`DefaultAvatar` always use the CDN base and have their own extension rules.

## Quick Start

Every function is pure and returns a string. A zero `Options` value is useful
for ordinary PNG assets.

```go
package main

import (
	"fmt"

	"github.com/discord-go/discord.go/cdn"
)

func main() {
	avatar := cdn.Avatar("123", "a_hash", cdn.Options{Animated: true, Size: 256})
	icon := cdn.GuildIcon("456", "icon-hash", cdn.Options{MediaProxy: true})
	sticker := cdn.Sticker("789", cdn.ExtensionPNG)
	fmt.Println(avatar)
	fmt.Println(icon)
	fmt.Println(sticker, cdn.DefaultAvatar(2))
}
```

## Creating Asset URLs

The asset functions are `Avatar`, `GuildIcon`, `GuildSplash`,
`DiscoverySplash`, `GuildBanner`, `UserBanner`, `MemberAvatar`,
`MemberBanner`, and `RoleIcon`. Each takes the relevant ID and hash plus an
`Options` value. `Sticker(stickerID, extension)` defaults an empty extension
to `ExtensionPNG`; `DefaultAvatar(index)` writes the index verbatim, so validate
an index before exposing it to untrusted URL input.

## Using Options

The exported extension constants are `ExtensionPNG`, `ExtensionJPEG`,
`ExtensionWebP`, `ExtensionGIF`, and `ExtensionJSON`. The builder does not
validate extension strings, sizes, IDs, or hashes. A zero or negative `Size`
omits the query parameter. `Animated` affects both the query and the implicit
extension, but an explicit extension wins. Query parameters are URL-encoded
by `net/url`, so generated URLs have stable escaped ordering.

## Common Patterns

Use `users.User.AvatarURL` when you already have a user model and need its
default-avatar fallback; use this package when you have a raw hash or need
member and guild assets. Cache URL strings, not downloaded bytes, when asset
freshness is controlled by a hash. Select `MediaProxy` for media-proxy
behavior rather than manually replacing the host.

## Best Practices

Treat hashes as opaque values. Keep IDs and hashes separate, and do not
construct a path with string concatenation when one of the typed functions
exists. Use a supported extension for Discord assets and positive power-of-two
sizes accepted by Discord's CDN.

## Common Mistakes

`MediaProxy` is not used by `Sticker` or `DefaultAvatar`. An animated hash does
not automatically become GIF unless `Animated` is set. The package does not
fetch, authenticate, or check whether an asset exists.

## API Walkthrough

The public surface consists of `ImageExtension`, its five constants, `Options`,
the two base URL constants, and the eleven URL functions described above. No
constructor or error return is needed because all functions are deterministic
string formatters.

## Examples

The Quick Start program is complete and runnable. It demonstrates implicit
GIF selection, size parameters, the media proxy, stickers, and default avatars.

## Related APIs

- [`../users/`](../users/README.md)
- [`../emojis/`](../emojis/README.md)
- [`../models/`](../models/README.md)
