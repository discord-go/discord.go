# Voice Guide

## Overview

discord.go provides full voice support including WebSocket connection
management, UDP transport, RTP packet handling, AES-256-GCM transport
encryption, Opus frame sending, and DAVE/MLS end-to-end encryption. This guide
covers the voice connection lifecycle, encryption, and audio handling.

## Voice Connection Lifecycle

Voice connections follow a two-step process:

1. Send a Voice State Update via the main Gateway to join a channel.
2. Receive `VOICE_SERVER_UPDATE` and create a `voice.Client` to connect to the
   voice WebSocket.

```go
// Step 1: Join a voice channel via the main gateway
err := botClient.JoinVoiceChannel(guildID, channelID, false, false)
if err != nil {
    log.Fatal(err)
}

// Step 2: Handle VOICE_SERVER_UPDATE to create the voice client
botClient.OnRawEvent(func(ctx context.Context, event string, data json.RawMessage) {
    if event != "VOICE_SERVER_UPDATE" {
        return
    }
    var vsu struct {
        Token    string `json:"token"`
        GuildID  string `json:"guild_id"`
        Endpoint string `json:"endpoint"`
    }
    json.Unmarshal(data, &vsu)

    voiceClient := voice.NewClient(conn, guildID, channelID, userID)
    voiceClient.Token = vsu.Token
    voiceClient.Endpoint = vsu.Endpoint
    voiceClient.SessionID = sessionID
    // Connect and start sending audio
})
```

## Opus Encoding and Decoding

The library handles transport encryption (AES-256-GCM) automatically.
Applications provide Opus-encoded audio frames via `SendOpus`:

```go
// Send a 20ms Opus frame
err := voiceClient.SendOpus(opusFrame)
if err != nil {
    log.Printf("send opus: %v", err)
}
```

`SendOpus` handles:
- RTP header construction (sequence, timestamp, SSRC)
- DAVE MLS encryption (if a DAVE session is active)
- AES-256-GCM transport encryption with the cached cipher
- Nonce counter management
- Packet assembly (header + ciphertext + nonce suffix)

For receiving audio, set the `OnAudioPacket` callback:

```go
voiceClient.OnAudioPacket = func(userID string, sequence uint16, timestamp uint32, audio []byte) {
    // audio is a decrypted Opus frame
    // Decode with your Opus decoder (e.g., hraban/opus)
}
```

## DAVE Protocol

DAVE (Discord Audio Visual End-to-end) provides MLS-based end-to-end encryption
for voice. The library supports DAVE protocol version 1.

When a DAVE session is active, `SendOpus` encrypts the Opus frame with MLS
before transport encryption. The `DaveSession` field on `voice.Client` manages
the MLS group state.

DAVE decryption happens automatically in the audio read loop before delivering
frames to `OnAudioPacket`.

## Encryption Modes

The library supports `aead_aes256_gcm_rtpsize` mode:

- AES-256-GCM authenticated encryption.
- The RTP header (12 bytes) is used as additional authenticated data (AAD).
- A 4-byte nonce counter is zero-padded to 12 bytes for GCM.
- The nonce is appended as a 4-byte suffix to each packet.

The `cipher.AEAD` is cached on the `voice.Client` when the secret key is
received in `SessionDescription`, avoiding per-packet cipher allocation.

The nonce counter is a `uint32` that increments on each `SendOpus` call. If
the counter reaches its maximum value (`2^32 - 1`), `SendOpus` returns an
error instead of wrapping to zero, which would reuse a nonce and break GCM.
The caller must re-establish the voice session to get a new key. At 20
packets per second, this limit is reached after approximately 6.8 years of
continuous audio.

## UDP Packet Structure

Voice packets follow this layout:

```
[RTP Header (12 bytes)] [Encrypted Payload] [Nonce Suffix (4 bytes)]
```

The RTP header contains:
- Version (2 bits, always 2)
- Padding (1 bit)
- Extension (1 bit)
- CSRC count (4 bits)
- Marker (1 bit)
- Payload type (7 bits, 0x78)
- Sequence number (16 bits)
- Timestamp (32 bits)
- SSRC (32 bits)

## Speaking Flags

Use speaking flags to indicate the bot is transmitting audio:

```go
// The voice client sets speaking state automatically when sending audio
// Speaking flags: SpeakingFlagMicrophone, SpeakingFlagSoundshare
```

## Voice Receiving

The audio read loop runs in a goroutine after the voice connection is
established. It:

1. Reads UDP packets from the voice server.
2. Extracts the RTP header and nonce suffix.
3. Decrypts the transport encryption using the cached AEAD.
4. If DAVE is active, decrypts the end-to-end encryption.
5. Strips RTP extension data.
6. Delivers the decrypted Opus frame to `OnAudioPacket`.

## Common Patterns

- Always join via the main gateway first, then create the voice client.
- Use `context.Context` with deadlines for voice operations.
- Send Opus frames at 20ms intervals (50 packets/second).
- Handle `VOICE_SERVER_UPDATE` and `VOICE_STATE_UPDATE` events.

## Best Practices

- Keep the voice WebSocket and UDP connection alive with heartbeats.
- Monitor `unackedHeartbeats` for connection health.
- Use DAVE end-to-end encryption when available.
- Close the voice client when leaving the channel.

## Common Mistakes

- Not waiting for `VOICE_SERVER_UPDATE` before creating the voice client.
- Sending non-Opus data to `SendOpus`.
- Not setting `OnAudioPacket` before connecting if receiving is needed.
- Forgetting to close the voice client on shutdown.
