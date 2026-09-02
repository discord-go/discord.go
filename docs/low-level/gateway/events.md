# Gateway Events

## Overview

Gateway traffic is an envelope rather than a stream of already typed Go
values. `gateway.GatewayPayload` contains opcode `Op`, raw `Data`, optional
sequence, and optional event type. Dispatch events use opcode 0 and put the
event name in `Type`; control payloads use the constants in `gateway/opcode.go`.

## Architecture

The client reads JSON, handles control opcodes, updates the session sequence,
and then dispatches event data. `events.Event` is a second convenient envelope
for decoding an incoming payload; its `Data` is `json.RawMessage`. Typed
wrappers are documented in [`../events/`](../events/README.md). Keeping data
raw until the event name is known allows applications to support event types
that the repository does not yet model.

The important control flow is: HELLO supplies the heartbeat interval and
causes identify or resume; HEARTBEAT sends an immediate sequence heartbeat;
HEARTBEAT_ACK marks the heartbeat alive; RECONNECT closes and resumes;
INVALID_SESSION either resets and identifies or attempts a resume according to
Discord's flag; and DISPATCH updates session/cache state before handler calls.
`Client.Compressed` enables the repository's Gateway compression stream for
the configured connection.

## Quick Start

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/discord-go/discord.go/events"
)

func main() {
	var event events.Event
	if err := json.Unmarshal([]byte(`{"op":0,"d":{"id":"55","channel_id":"9","content":"hello"},"s":12,"t":"MESSAGE_CREATE"}`), &event); err != nil {
		panic(err)
	}
	var message events.MessageCreate
	if err := json.Unmarshal(event.Data, &message); err != nil {
		panic(err)
	}
	fmt.Println(event.Op, event.Type, *event.Seq, message.Content)
}
```

## Using Control Payloads

Construct a `GatewayPayload` with an opcode and marshaled data for identify,
resume, presence, voice state, or member requests. `Client.Send` enforces the
4096-byte payload limit and a 120-per-minute send window. Heartbeat payloads
are normally owned by `Heartbeater`; do not create a second loop for the same
connection.

## Common Patterns

Decode the envelope, store `Seq`, switch on `Type`, and decode `Data` into the
matching `events` type. For an unknown event, retain the raw bytes and log the
type rather than treating it as malformed. A `MESSAGE_UPDATE` can be partial;
merge it into cached state instead of replacing every field with zero values.

## Best Practices

Dispatch quickly and move slow work to a queue while preserving ordering where
the application needs it. Keep raw payload bytes until handlers have finished.
Enable only the intents needed by the event families. Close the connection on
heartbeat failure so the client's reconnect path can run.

## Common Mistakes

Sequence `s` is absent on some control payloads. `Data` is not guaranteed to
match an event model just because `Type` is familiar. Do not decode compressed
frames as plain JSON; configure `Client.Compressed` and use a compatible
connection path.

## API Walkthrough

The page covers `gateway.GatewayPayload` and the opcode constants, plus
`events.Event` and typed wrappers such as `Ready`, message, channel, guild,
interaction, reaction, and audit-log event values. Heartbeat and session
methods are in [`heartbeat.md`](heartbeat.md) and [`../gateway/`](README.md).

When `Client.Cache` is set, the client hydrates the cache during dispatch:
`GUILD_CREATE` stores the guild (including its `voice_states`, `members`,
`channels`, `threads`, and `presences` arrays) and every channel it carries,
`CHANNEL_CREATE` and `CHANNEL_UPDATE` store the channel, and `CHANNEL_DELETE`,
`ROLE_DELETE`, `MESSAGE_DELETE`, and `GUILD_DELETE` evict.

## Examples

The Quick Start program is complete and runnable from a local JSON payload.

## Related APIs

- [`../events/`](../events/README.md)
- [`README.md`](README.md)
- [`heartbeat.md`](heartbeat.md)
