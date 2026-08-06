# Voice

## Overview

Voice has two related but separate parts in this repository. `bot.JoinVoiceChannel` changes the bot's main-Gateway voice state. The independent `voice.Client` consumes the resulting voice session data, performs the voice WebSocket and UDP handshake, and accepts encoded Opus frames. This guide provides a complete runnable control-plane example and explains the transport boundary without pretending that a generic application has an audio source or WebSocket adapter.

## Prerequisites

- Go `1.26.4` or newer.
- `DISCORD_TOKEN` set to a bot token.
- `VOICE_GUILD_ID` and `VOICE_CHANNEL_ID` set to numeric Discord snowflake IDs.
- The bot installed in the guild and allowed to View Channel and Connect to the voice channel.
- `Guilds` and `GuildVoiceStates` enabled in the Portal and selected by the bot.
- For audio playback, an application-level Opus encoder/source and a voice WebSocket connection adapter.

## Architecture

The main Gateway sends a voice state update when `JoinVoiceChannel` is called. Discord returns a `VOICE_STATE_UPDATE` containing the bot's session ID and a `VOICE_SERVER_UPDATE` containing the voice token and endpoint. The application combines those values with guild, channel, and user IDs to configure `voice.NewClient`. That client performs UDP discovery, encryption, DAVE session handling, heartbeats, and Opus transport. Leaving the channel and disconnecting the voice client are separate cleanup operations.

## Quick Start

The complete program below joins a channel and logs the two control-plane payloads:

```bash
export DISCORD_TOKEN='replace-with-a-bot-token'
export VOICE_GUILD_ID='123456789012345678'
export VOICE_CHANNEL_ID='234567890123456789'
go run ./path/to/your/voice-example
```

Use a real voice-channel ID in a test guild. The program does not send audio.

## Complete Runnable Example

This complete `package main` program uses current repository APIs, imports, and `main`. It is intentionally a control-plane example: it can join and receive the required session payloads without making up an audio source.

```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/voice"
)

type voiceStatePayload struct {
	GuildID   snowflake.ID  `json:"guild_id,string"`
	ChannelID *snowflake.ID `json:"channel_id,string"`
	UserID    snowflake.ID  `json:"user_id,string"`
	SessionID string         `json:"session_id"`
}

func requiredID(name string) snowflake.ID {
	value := os.Getenv(name)
	id, err := snowflake.Parse(value)
	if err != nil || id == 0 {
		log.Fatalf("%s must be a valid snowflake", name)
	}
	return id
}

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}
	guildID := requiredID("VOICE_GUILD_ID")
	channelID := requiredID("VOICE_CHANNEL_ID")
	runCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	b := bot.New(token, bot.WithIntents(intents.Guilds|intents.GuildVoiceStates))
	b.OnReady(func(ctx *bot.ReadyContext) {
		log.Printf("ready as %s; joining voice channel", ctx.User.Username)
		if err := ctx.Bot.JoinVoiceChannel(guildID, channelID, false, false); err != nil {
			log.Printf("join voice channel: %v", err)
		}
	})
	b.OnVoiceStateUpdate(func(event *bot.EventContext) {
		var state voiceStatePayload
		if err := event.Decode(&state); err != nil {
			log.Printf("decode voice state: %v", err)
			return
		}
		if state.UserID == b.AppID() {
			log.Printf("voice state: guild=%s channel=%v session=%s", state.GuildID.String(), state.ChannelID, state.SessionID)
		}
	})
	b.OnVoiceServerUpdate(func(event *bot.EventContext) {
		var update voice.VoiceServerUpdate
		if err := event.Decode(&update); err != nil {
			log.Printf("decode voice server: %v", err)
			return
		}
		log.Printf("voice server: guild=%s endpoint=%v", update.GuildID.String(), update.Endpoint)
	})
	b.OnDisconnect(func() {
		log.Println("main Gateway disconnected; voice transport must also be stopped by the application")
	})

	if err := b.Start(runCtx); err != nil {
		log.Fatal(err)
	}
	<-runCtx.Done()
	if err := b.LeaveVoiceChannel(guildID); err != nil {
		log.Printf("leave voice channel: %v", err)
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := b.Stop(shutdownCtx); err != nil {
		log.Printf("stop bot: %v", err)
	}
}
```

## Explanation

The bot facade does not itself create a `voice.Client`. It only asks Discord to move the bot into the channel. A full voice implementation must wait for both the bot's voice state session ID and the server update token/endpoint, open a `voice.Connection`, construct `voice.NewClient`, set its `SessionID`, `Token`, and `Endpoint`, then call `Connect` with a deadline. The transport must call `Disconnect` and the bot must call `LeaveVoiceChannel` during cleanup.

`voice.Client.SendOpus` expects already encoded Opus data. PCM decoding, media playback, queueing, backpressure, and source cancellation belong to the application or a media service integration.

## Basic Usage

- Select `intents.Guilds|intents.GuildVoiceStates`.
- Call `JoinVoiceChannel(guildID, channelID, selfMute, selfDeaf)` after the bot is running.
- Subscribe to `OnVoiceStateUpdate` and `OnVoiceServerUpdate`.
- Decode `voice.VoiceServerUpdate` with `EventContext.Decode`.
- Call `LeaveVoiceChannel(guildID)` when the bot should leave.

## Intermediate Usage

- Correlate updates by guild ID and ignore updates for other guilds.
- Store the latest session ID, token, and endpoint under a guild-scoped lock.
- Reconnect when Discord supplies a new endpoint, using `voice.Client.Reconnect` and a connection factory.
- Set `voice.Client.OnAudioPacket` when receiving and processing voice audio.
- Use `context.WithTimeout` for `voice.Client.Connect` and retry only after a clean failure.

## Advanced Usage

- Implement `voice.Connection` over the WebSocket library used by the application and ensure writes are serialized.
- Preserve DAVE session state and support the encryption mode negotiated by the current voice server.
- Run audio production through a bounded queue and stop it when `voice.Client.GetState()` becomes idle.
- Do not assume a voice endpoint remains valid across reconnects; update session data and close stale UDP sockets.
- Track heartbeats, UDP discovery, encryption failures, Opus send errors, and disconnect reasons.

## Common Patterns

- Treat the main-Gateway voice state and the voice WebSocket as two lifecycles.
- Use a per-guild voice-session object to correlate state and server updates.
- Join only after READY and leave before process shutdown completes.
- Use `voice.Client.SetSpeaking` around playback state changes.
- Increment RTP timestamps through `SendOpus`; do not hand-craft packets unless the application owns the complete transport.

## Best Practices

- Never log voice tokens or full session credentials.
- Bound connection and UDP discovery timeouts.
- Close the voice client before closing application-owned audio sources.
- Stop playback producers before calling `Disconnect` to prevent writes to a closed UDP connection.
- Handle channel deletion, guild removal, Gateway reconnect, and endpoint changes as normal lifecycle events.
- Request Connect, Speak, and other voice permissions explicitly; `GuildVoiceStates` only controls event delivery.

## Common Mistakes with wrong/correct examples

### Wrong

```go
client.JoinVoiceChannel(guildID, channelID, false, false)
voiceClient.SendOpus(pcmBytes)
```

### Correct

```go
if err := b.JoinVoiceChannel(guildID, channelID, false, false); err != nil {
	return err
}
// Wait for VOICE_STATE_UPDATE and VOICE_SERVER_UPDATE.
// Encode PCM to Opus in the application, then call voiceClient.SendOpus(opus).
```

### Wrong

```go
_ = voiceClient.Connect(context.Background())
```

### Correct

```go
connectCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
defer cancel()
if err := voiceClient.Connect(connectCtx); err != nil {
	return err
}
```

### Wrong

```go
// Leave the main Gateway voice state and forget the voice client.
_ = b.LeaveVoiceChannel(guildID)
```

### Correct

```go
if voiceClient != nil && voiceClient.GetState().IsActive() {
	_ = voiceClient.Disconnect()
}
_ = b.LeaveVoiceChannel(guildID)
```

The fragments show the required ordering but are not standalone programs; the complete runnable control-plane program is above.

## Expected Result

The bot connects, joins the configured voice channel, logs its voice state session ID, and logs the voice server endpoint. No audio is transmitted by the runnable example. A full media implementation additionally creates and connects `voice.Client`, sends encoded Opus frames, and disconnects it during shutdown.

## Related Pages

- [Examples Overview](README.md)
- [Basic Client](basic-client.md)
- [Gateway](gateway.md)
- [Complete bot voice facade: `bot/voice.go`](../../../bot/voice.go)
- [Voice client source: `voice/client.go`](../../../voice/client.go)
- [Voice server payload: `voice/session.go`](../../../voice/session.go)
