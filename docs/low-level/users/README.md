# Users

## Overview

The `users` package models Discord users and the user-shaped data attached to
guilds, presences, and Gateway events. It exports `User`, `Member`,
`PresenceUpdate`, `Activity`, `ClientStatus`, `AvatarDecorationData`,
`Collectibles`, `Nameplate`, `PrimaryGuild`, and the user `Flag` and
`PremiumType` enums. Avatar URL helpers are included; generic asset builders
are in [`../cdn/`](../cdn/README.md).

## Architecture

`User.ID` is a snowflake. `Avatar`, `Banner`, `GlobalName`, email, decoration,
collectibles, and primary guild data are pointers where Discord can omit or
return null. `Member` represents a user's membership in one guild: it adds
roles, nickname, member avatar and banner, join and premium timestamps,
moderation timeout, mute/deaf state, and effective permission text. Its
unmarshaller parses the string role IDs and nullable times.

`PresenceUpdate` groups a user, guild ID, status, activities, and client
status. `Activity` includes name, type, URL, creation time, application ID,
details, and state. `Flag` is a bitfield with staff, partner, HypeSquad,
verified, moderator, and developer flags. `PremiumType` distinguishes none,
classic, Nitro, and basic.

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/discord-go/discord.go/users"
	"github.com/discord-go/discord.go/snowflake"
)

func main() {
	hash := "avatar-hash"
	u := users.User{ID: snowflake.ID(42), Username: "ada", Avatar: &hash, Flags: users.FlagActiveDeveloper}
	url := u.AvatarURL(users.AvatarURLOptions{Extension: "webp", Size: 128})
	fmt.Println(u.Username, u.Flags&users.FlagActiveDeveloper != 0, url == u.DisplayAvatarURL(users.AvatarURLOptions{Extension: "webp", Size: 128}))
}
```

## Creating And Using Users

Construct models with field literals or decode REST and event JSON. `AvatarURL`
accepts `AvatarURLOptions{Extension, Size}`. An empty or unsupported extension
falls back to PNG; a leading dot is removed. A positive size becomes a query
parameter. If the avatar is nil or empty, the method calculates Discord's
default avatar index from the user ID. `DisplayAvatarURL` is an alias with the
same behavior.

## Common Patterns

Use `Member.Roles` with [`../permissions/`](../permissions/README.md) and the
guild's role objects to calculate access. Use `PresenceUpdate` for status
updates rather than treating a missing activity as offline. Use
`AvatarDecorationData` and `Nameplate` as optional metadata, not as avatar
hashes.

## Best Practices

Check pointers before dereferencing optional profile fields. Treat a presence
as a partial, rapidly changing snapshot. Keep user IDs as `snowflake.ID`; do
not parse them through floating-point JSON values.

## Common Mistakes

`Member.Permissions` is not necessarily a complete channel permission result.
`User.Discriminator` can be empty for modern usernames. `DisplayAvatarURL` does
not create a different URL or download an image, and a nil avatar intentionally
returns a default-avatar URL.

## API Walkthrough

The exported API includes `User`, `User.AvatarURL`,
`User.DisplayAvatarURL`, `AvatarURLOptions`, `Member` and `Member.UnmarshalJSON`,
`PresenceUpdate`, `Activity`, `ClientStatus`, `AvatarDecorationData`,
`Collectibles`, `Nameplate`, `PrimaryGuild`, `Flag` and all flag constants, and
`PremiumType` and its constants.

## Examples

The Quick Start program is complete and runnable. Model JSON can be exchanged
through REST methods such as `rest.GetUser` and `rest.GetGuildMember`.

## Related APIs

- [`../cdn/`](../cdn/README.md)
- [`../cache/`](../cache/README.md)
- [`../guilds/`](../guilds/README.md)
