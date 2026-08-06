# Intents

## Overview

Gateway intents select which event families Discord sends to a bot. The
`intents.Intent` type is a `uint64` bitfield with constants for guilds, members,
bans, emojis and stickers, integrations, webhooks, invites, voice states,
presences, guild and direct messages, reactions, typing, message content,
scheduled events, AutoMod, and polls.

## Architecture

The zero value requests no intent bits. Combine constants with `|`, test with
`Has`, and mutate a pointer with `Add` or `Remove`. The methods do not contact
Discord and do not know whether a privileged intent is enabled in the
Developer Portal. The Gateway identify payload receives the final bitfield.

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/discord-go/discord.go/intents"
)

func main() {
	wanted := intents.Guilds | intents.GuildMessages | intents.DirectMessages
	wanted.Add(intents.MessageContent)
	fmt.Println(wanted.Has(intents.Guilds), wanted.Has(intents.GuildPresences))
	wanted.Remove(intents.MessageContent)
	fmt.Println(wanted.Has(intents.MessageContent))
}
```

## Creating An Intent Set

Use `|` for a literal set and `Add` when a feature conditionally enables a
family. The constants are `Guilds`, `GuildMembers`, `GuildBans`,
`GuildEmojisAndStickers`, `GuildIntegrations`, `GuildWebhooks`, `GuildInvites`,
`GuildVoiceStates`, `GuildPresences`, `GuildMessages`,
`GuildMessageReactions`, `GuildMessageTyping`, `DirectMessages`,
`DirectMessageReactions`, `DirectMessageTyping`, `MessageContent`,
`GuildScheduledEvents`, `AutoModerationConfiguration`,
`AutoModerationExecution`, `GuildMessagePolls`, and `DirectMessagePolls`.

## Using Intents

Pass the bitfield to `client.WithIntents` or
`gateway.NewShardManager(token, shardCount, gatewayIntents)`. The Gateway
manager sends it during identify. A missing intent can look like a cache bug
because the corresponding dispatch never arrives.

## Common Patterns

Use the smallest set that supports the application. Enable `GuildMembers`,
`GuildPresences`, and `MessageContent` only when required and approved. Keep
the configured value near the code that registers dependent handlers so a
review can compare event use and intent use.

## Best Practices

Enable privileged intents in both the Developer Portal and the identify
payload. Log the numeric bitfield at startup, but never log the token. Treat
intent changes as deployment configuration changes because they alter cache
completeness and event volume.

## Common Mistakes

`Has` is a bit test for any supplied bit, not a portal permission check. Adding
an intent in code cannot bypass Discord approval. Do not request all intents
just to make a cache appear complete; REST fallback is safer for data that is
not delivered.

## API Walkthrough

The public API is `Intent`, all 21 intent constants, `Intent.Add`,
`Intent.Remove`, and `Intent.Has`. There are no constructors or errors.

## Examples

The Quick Start program is complete and runnable. Client wiring is shown in
[`../client/`](../client/README.md), and Gateway identify behavior is covered
by [`../gateway/`](../gateway/README.md).

## Related APIs

- [`../client/`](../client/README.md)
- [`../gateway/`](../gateway/README.md)
- [`../cache/`](../cache/README.md)
