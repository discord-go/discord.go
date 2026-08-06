# Voice

## Overview

The `voice` package implements the independent Discord voice connection. It
contains voice WebSocket payloads, UDP discovery, RTP packet construction,
AES-256-GCM transport encryption, speaking state, voice heartbeats, Opus frame
transport, and DAVE session integration. It does not decode audio, run FFmpeg,
resolve sources, or provide a Lavalink client.

## Architecture

`Client` is created with `NewClient(conn, guildID, channelID, userID)`. Before
`Connect`, call `SetSession(sessionID, token, endpoint)` with data collected
from Gateway `VOICE_STATE_UPDATE` and `VOICE_SERVER_UPDATE` events. Connect
performs the voice HELLO/IDENTIFY/READY handshake, UDP IP discovery, protocol
selection, session description, and heartbeat setup. `SendOpus` accepts one
already encoded Opus frame and encrypts it as `aead_aes256_gcm_rtpsize`.

`Connection` is the voice WebSocket abstraction. `UDPConnection` handles the
IP discovery packet and UDP writes. `RTPHeader` has version 2 and payload type
120 by default through `NewRTPHeader`; `BuildAudioPacket` combines a header and
payload. The `EncryptAEAD` packet layout is RTP header, GCM ciphertext, and a
4-byte big-endian nonce counter suffix. `DecryptAEAD` expects the 12-byte RTP
header as AAD and a 12-byte GCM nonce.

## Quick Start

```go
package main

import (
	"encoding/binary"
	"fmt"

	"github.com/discord-go/discord.go/voice"
)

func main() {
	key := [32]byte{1}
	header := voice.NewRTPHeader(1, 960, 42)
	plain := []byte{1, 2, 3}
	packet, err := voice.EncryptAEAD(key, header.Marshal(), plain, 7)
	if err != nil {
		panic(err)
	}
	nonce := make([]byte, 12)
	binary.BigEndian.PutUint32(nonce, 7)
	sealed := packet[12 : len(packet)-4]
	opened, err := voice.DecryptAEAD(key, packet[:12], sealed, nonce)
	if err != nil {
		panic(err)
	}
	parsed, _ := voice.ParseRTPHeader(packet)
	fmt.Println(parsed.Sequence, string(opened), len(voice.BuildAudioPacket(header, plain)))
}
```

## Creating A Voice Client

The connection must support text and binary reads and writes. `Connect` returns
an error if the client is not idle, the context is canceled, the server sends
an unexpected opcode, the encryption mode is unsupported, or the session key is
invalid. `Disconnect` closes the voice connection and moves through the state
machine. `GetState` and `ConnectionState.IsActive` expose lifecycle state;
`String` provides a diagnostic name. `Reconnect` requires a
`ConnFactory(endpoint)` and preserves the DAVE session.

## Using Audio And Speaking State

Call `SetSpeaking` with `Microphone`, `Soundshare`, `Priority`, or a bitwise
combination. Maintain sequence, timestamp, SSRC, and nonce counter as the
client does; callers should send frames through `SendOpus` rather than building
encrypted packets around stale counters. Received audio can be delivered
through `OnAudioPacket` after SSRC mapping. `NewAESGCMEncrypter` exposes a
generic AES-GCM `Encrypter` with 12-byte nonce normalization.

## Common Patterns

Join voice through Gateway first, collect both voice state and server update,
then call `SetSession` and `Connect`. Use `VoiceRegion` for REST region data.
Use DAVE methods (`SendMLSKeyPackage`, `SendMLSCommitWelcome`, transition and
epoch notifications) only when the session negotiation requires them.

## Best Practices

Keep voice contexts bounded and close UDP resources on all failure paths. Feed
valid Opus frames at the expected cadence. Do not log session tokens or secret
keys. Test RTP parsing with truncated packets and tampered ciphertext.

## Common Mistakes

`Connect` cannot work with only a guild/channel ID; it needs endpoint, token,
and session information. `SendOpus` does not encode PCM or validate that bytes
are Opus. RTP headers are not encrypted but are authenticated as GCM AAD.
`StateIdle` is the only inactive state; do not call `Connect` on an active
client.

## API Walkthrough

The public API is `Client`, `NewClient`, `Connect`, `Disconnect`, `Reconnect`,
`SetSession`, `SetSpeaking`, `SendOpus`, all DAVE send methods, `Connection`,
`ConnectionState` and its constants/methods, `VoiceState`, `VoiceServerUpdate`,
`VoiceRegion`, voice `Payload` and handshake data types, all voice opcodes,
`UDPConnection` and its methods, `RTPHeader` and its constructor/parser/
marshal methods, `BuildAudioPacket`, `Encrypter`, `NewAESGCMEncrypter`,
`EncryptAEAD`, and `DecryptAEAD`.

## Examples

The Quick Start program is complete and runnable locally and demonstrates RTP
serialization plus authenticated encryption. Gateway coordination is covered
in [`../gateway/`](../gateway/README.md).

## Related APIs

- [`../gateway/`](../gateway/README.md)
- [`../rest/`](../rest/README.md)
- [`../events/`](../events/README.md)
