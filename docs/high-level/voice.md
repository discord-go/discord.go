# Voice State Control

## Overview

The high-level voice API controls gateway voice state: join a guild voice
channel, move or mute/deafen the bot, and leave. It does not play audio or
construct a voice UDP session. After the join request, applications that need
audio must consume voice gateway events and use the lower-level `voice` package
or an external player.

## Architecture

`JoinVoiceChannel` sends the gateway voice-state update through the single
gateway client or the shard responsible for the guild. Discord then emits
`VOICE_STATE_UPDATE` and `VOICE_SERVER_UPDATE`. The bot exposes those events as
generic `EventContext` subscriptions because the high-level package deliberately
does not own the full voice transport.

The bot also maintains a built-in voice tracker from `VOICE_STATE_UPDATE`
dispatches. It answers "who is connected to channel X" without REST calls, and
a typed handler is available for applications that react to joins, moves, and
leaves. See [Observing voice states](#observing-voice-states).

`LeaveVoiceChannel` is shorthand for joining with channel ID zero. A sharded bot
routes the request using the guild ID and shard count.

## Quick Start

This complete program joins a configured voice channel and logs the two events
needed to continue with a lower-level voice session.

```go
package main

import (
	"log"
	"os"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/snowflake"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	guildText := os.Getenv("VOICE_GUILD_ID")
	channelText := os.Getenv("VOICE_CHANNEL_ID")
	if token == "" || guildText == "" || channelText == "" {
		log.Fatal("DISCORD_TOKEN, VOICE_GUILD_ID, and VOICE_CHANNEL_ID are required")
	}
	guildID, err := snowflake.Parse(guildText)
	if err != nil {
		log.Fatal(err)
	}
	channelID, err := snowflake.Parse(channelText)
	if err != nil {
		log.Fatal(err)
	}

	b := bot.New(token, bot.WithIntents(intents.Guilds|intents.GuildVoiceStates))
	b.OnVoiceStateUpdate(func(ctx *bot.EventContext) {
		log.Printf("voice state event: %s", ctx.Name)
	})
	b.OnVoiceServerUpdate(func(ctx *bot.EventContext) {
		log.Printf("voice server event: %s", ctx.Name)
	})
	b.OnReady(func(ctx *bot.ReadyContext) {
		if err := ctx.Bot.JoinVoiceChannel(guildID, channelID, false, true); err != nil {
			log.Printf("join voice: %v", err)
			return
		}
		log.Println("voice state requested")
	})
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

The example only requests the voice state. Follow the [voice low-level guide](../low-level/voice/README.md)
when you need an audio connection.

## Creating/Configuration

Request the `GuildVoiceStates` intent and ensure the bot has `Connect` and,
when applicable, `Speak` permissions in the target channel. Call
`JoinVoiceChannel(guildID, channelID, selfMute, selfDeaf)` after the gateway is
running, normally from `OnReady` or a command handler. Use `LeaveVoiceChannel`
with the guild ID to disconnect.

For sharding, `WithShards` causes the bot to choose the correct shard. The
high-level method does not expose the raw session description, UDP endpoint, or
Opus encoder.

## Using

### Basic: join and leave

Join with `selfMute` and `selfDeaf` set to the desired initial state. Leave with
`LeaveVoiceChannel` when the feature is disabled.

### Intermediate: observe events

Subscribe to `OnVoiceStateUpdate` and `OnVoiceServerUpdate`; use `EventContext.Raw`
or `Decode` to inspect their JSON payloads.

### Advanced: create audio transport

Use the event data to initialize the low-level `voice.Client` or delegate to an
external audio service such as a Lavalink adapter. Keep transport lifecycle
separate from the bot's join request.

## Common Patterns

- Join only after `WaitReady` or from `OnReady`.
- Track the guild-to-channel state in application storage so reconnect logic can
  decide whether to rejoin.
- Subscribe before joining so the server update cannot be missed.
- Handle `VOICE_SERVER_UPDATE` as data needed for the voice session, not as an
  acknowledgement that audio is already connected.
- Leave on feature shutdown and close the lower-level voice transport too.

## Best Practices

### Keep state control and audio transport separate

Why: the high-level API only sends gateway voice state.

Pros: command code stays small and transport choices remain flexible.

Cons: an audio feature requires additional event decoding and dependencies.

### Use explicit mute and deaf flags

Why: Discord applies them as part of the voice state.

Pros: predictable user-visible state when joining.

Cons: these flags do not replace channel permissions or an audio transport.

### Reconnect deliberately

Why: gateway reconnects and voice session reconnects have separate state.

Pros: fewer stale sessions and clearer operational recovery.

Cons: the application must retain enough state to know which guilds need a
rejoin.

## Common Mistakes

Incorrect: expecting `JoinVoiceChannel` to play a file.

```go
_ = b.JoinVoiceChannel(guildID, channelID, false, false)
// No audio transport is created by this call.
```

Correct: use the join call to obtain voice events, then create a lower-level
voice session.

```go
b.OnVoiceServerUpdate(handleVoiceServer)
b.OnVoiceStateUpdate(handleVoiceState)
_ = b.JoinVoiceChannel(guildID, channelID, false, false)
```

Incorrect: joining before the gateway is running.

```go
_ = b.JoinVoiceChannel(guildID, channelID, false, false)
_ = b.Run()
```

Correct: join from `OnReady` or after `WaitReady`.

```go
b.OnReady(func(ctx *bot.ReadyContext) {
	_ = ctx.Bot.JoinVoiceChannel(guildID, channelID, false, false)
})
```

## Observing Voice States

The bot tracks every `VOICE_STATE_UPDATE` it receives. The tracker is updated
before handlers run, so queries inside a handler observe the state that
triggered them.

```go
b.OnVoiceStateUpdateTyped(func(ctx *bot.VoiceStateContext) {
	state := ctx.State()
	if state.ChannelID == nil {
		return // the user disconnected
	}
	count := ctx.Bot.CountInChannel(*state.ChannelID)
	log.Printf("%s joined %s (%d connected)", state.UserID, state.ChannelID, count)
})

// Anywhere else: query without REST calls.
channelID := b.VoiceChannelOf(guildID, userID)
members := b.VoiceStatesInChannel(channelID)
```

`OnVoiceStateUpdateTyped` receives a decoded `voice.VoiceState`; a nil or zero
`ChannelID` means the user left voice. `VoiceStatesInChannel`,
`VoiceStateOf`, `VoiceChannelOf`, and `CountInChannel` read the tracker and
never touch the network. The tracker starts empty on process start and is
populated by live events; a bot that needs the initial occupancy of many
channels immediately after startup should also read the `voice_states` array
captured on `guilds.Guild` from `GUILD_CREATE`.

## API Walkthrough

- `JoinVoiceChannel(guildID, channelID snowflake.ID, selfMute, selfDeaf bool) error`
  forwards a voice-state request through the correct gateway client or shard.
- `LeaveVoiceChannel(guildID snowflake.ID) error` sends the same request with a
  zero channel ID and false mute/deaf flags.
- `OnVoiceStateUpdate(func(*EventContext)) func()` subscribes to the generic
  `VOICE_STATE_UPDATE` dispatch and returns an unsubscribe function.
- `OnVoiceStateUpdateTyped(func(*VoiceStateContext)) func()` subscribes to the
  same dispatch with the state already decoded.
- `VoiceStatesInChannel(channelID snowflake.ID) []voice.VoiceState`,
  `VoiceStateOf(guildID, userID snowflake.ID) (voice.VoiceState, bool)`,
  `VoiceChannelOf(guildID, userID snowflake.ID) snowflake.ID`, and
  `CountInChannel(channelID snowflake.ID) int` query the built-in tracker.
- `OnVoiceServerUpdate(func(*EventContext)) func()` does the same for
  `VOICE_SERVER_UPDATE`.
- `EventContext.Name` contains the dispatch name, `Data` contains raw JSON,
  `Decode(any) error` unmarshals it, and `Raw() json.RawMessage` returns a copy.
- `WithShards(int) Option` enables shard-aware routing; `WithIntents` configures
  the `GuildVoiceStates` subscription.

## Examples

- [Voice example](../examples/voice/README.md)
- [Voice low-level guide](../low-level/voice/README.md)
- [Gateway events](../low-level/gateway/events.md)

## Related APIs

- [`lifecycle.md`](lifecycle.md) for readiness and reconnect callbacks.
- [`permissions.md`](permissions.md) for `Connect` and `Speak` policy.
- [`../low-level/voice/README.md`](../low-level/voice/README.md) for audio transport.
