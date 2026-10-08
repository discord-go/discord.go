# Channels

## Overview

The `channels` package models Discord channel-shaped resources. `Channel`
covers guild text, DM, group DM, category, announcement, thread, voice, stage,
forum, and media fields. The package also models permission overwrites,
threads, forum tags, default reactions, invites, and the minimal webhook
reference. Operations belong to [`../rest/`](../rest/README.md).

## Architecture

`ChannelType` constants identify the wire type: guild text 0, DM 1, guild
voice 2, group DM 3, category 4, announcement 5, announcement thread 10,
public thread 11, private thread 12, stage voice 13, directory 14, forum 15,
and media 16. `Channel` uses pointers for fields that are absent on other
channel variants. Thread information is split between `ThreadMetadata` and
`ThreadMember`; forum configuration uses `ForumTag`, `DefaultReaction`, and
the available/applied tag fields.

`Flags` is a `*ChannelFlags` bitfield with named constants instead of raw
shifts: `ChannelFlagPinned` (1 << 1), `ChannelFlagRequireTag` (1 << 4),
`ChannelFlagHideMediaDownloadOptions` (1 << 15),
`ChannelFlagObfuscated` (1 << 17), and `ChannelFlagIsSpoilerChannel`
(1 << 21). `Channel.IsObfuscated` reports the obfuscation flag without
inspecting any other field. `AppPermissions` carries the bot's computed
permissions when the channel comes from interaction `resolved` data (added
July 2026); it is distinct from the top-level `Interaction.AppPermissions`.

### Obfuscated Channels

Starting November 16, 2026, `GET /guilds/{guild.id}/channels` omits channels
the bot cannot `VIEW_CHANNEL`, and the Gateway redacts those channels: `name`
becomes `"___hidden___"`, other sensitive fields are nulled or reduced, the
`ChannelFlagObfuscated` flag is set, and `permission_overwrites` holds a
single overwrite denying `VIEW_CHANNEL` for the guild's `@everyone` role.
`id`, `type`, `position`, and `parent_id` are never obfuscated. When the bot
gains access, Discord dispatches a `CHANNEL_UPDATE` with the full data. Until
then, treat every field except those four as unreliable and surface the
condition with `IsObfuscated` rather than by matching on `name`. Gateway
delivery of obfuscated payloads can be tested early by sending the
`gateway.CapabilityChannelObfuscation` (1 << 15) identify capability; see
[`../gateway/events.md`](../gateway/events.md). Interaction payloads are
built through a separate path and are never obfuscated.

`Invite` may contain an `InviteGuild`, channel, inviter, target user or
application, approximate counts, and expiry. Guild invite listings and the
`INVITE_CREATE` / `INVITE_DELETE` gateway payloads also carry metadata:
`Uses`, `MaxUses`, `MaxAge`, `Temporary`, `CreatedAt`, and the flat
`GuildID`, `ChannelID`, `InviterID`, and `TargetUserID` snowflakes. REST
invite objects from `GET /invites/{code}` omit most metadata, so these
fields stay zero there. `Overwrite` uses
`permissions.Permission` for `Allow` and `Deny`, so you can use permission
constants directly without manual string conversion. The fields use
`json:",string"` so Discord's string-encoded numbers marshal and unmarshal
correctly.

## Quick Start

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/discord-go/discord.go/channels"
)

func main() {
	var channel channels.Channel
	if err := json.Unmarshal([]byte(`{"id":"10","type":0,"guild_id":"20","name":"general","permission_overwrites":[]}`), &channel); err != nil {
		panic(err)
	}
	fmt.Println(channel.ID, channel.Type == channels.ChannelTypeGuildText, *channel.Name, *channel.GuildID)
}
```

## Creating And Using Channels

`Channel` fields are directly constructible, but create and modify requests use
REST parameter types so nullable updates can be represented. `ThreadMetadata`
contains archive and lock state, auto-archive duration, timestamps, and
invitable state. `ThreadMember` may include either IDs only or an embedded
`users.Member`. `Application` is the small application object used by invite
targets, while `Webhook` contains only an ID for channel references.

## Common Patterns

Switch on `Channel.Type` before reading variant-specific pointers. Use
`ParentID` for category or forum relationships and `AvailableTags` for forum
configuration. An invite's counts are approximate and its `ExpiresAt` may be
nil. Use a channel ID plus message ID when constructing a REST message route.

## Best Practices

Check nil pointers for `Name`, `GuildID`, `ParentID`, and timestamps. Treat
permission overwrite strings as decimal unsigned values, preserving their
precision. Do not assume every thread contains a member object; list members
explicitly when necessary.

## Common Mistakes

Channel type 10 is an announcement thread, not a normal announcement channel.
`Position` and `Flags` are optional. Do not detect an obfuscated channel by
its `___hidden___` name or by missing topic/member counts; check
`IsObfuscated` instead, and expect `GetGuildChannels` to omit channels the
bot cannot view. `Invite.Channel` is a pointer and may be
missing in partial responses. This package does not calculate effective
permissions or fetch messages.

## API Walkthrough

The exported types are `Channel`, `ChannelType` and all constants,
`ChannelFlags` and the `ChannelFlag*` constants, `Channel.IsObfuscated`,
`Overwrite` (with `permissions.Permission` for `Allow` and `Deny`),
`ThreadMetadata`, `ThreadMember`, `ForumTag`, `DefaultReaction`, `Invite`,
`InviteGuild`, `Application`, and `Webhook`. There are no constructors
or methods beyond `ChannelFlags.Has` and `Channel.IsObfuscated`.

## Examples

The Quick Start program is complete and runnable. Use
[`../rest/endpoints.md`](../rest/endpoints.md) for channel, thread, invite,
message, and webhook operations.

## Related APIs

- [`../rest/`](../rest/README.md)
- [`../messages/`](../messages/README.md)
- [`../permissions/`](../permissions/README.md)
