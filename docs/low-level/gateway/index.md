# Gateway

## Overview

The `gateway` package implements the Discord Gateway v10 session protocol. It
provides the transport interface, payload envelope, typed identify/resume
data, a reconnecting `Client`, reflection-based `Dispatcher`, concurrent
`Session`, heartbeat loop, identify tracker, and `ShardManager`.

## Architecture

`Client` owns one `Connection`, dispatcher, optional session, optional
heartbeater, token, intents, shard identity, cache, connection factory, URL,
and compression setting. `NewClient` wraps writes for synchronization but does
not initialize a dispatcher or open a connection. `Start(ctx)` reads payloads,
handles HELLO, identify/resume, dispatches events, updates cache/session state,
and reconnects when a failure is resumable. It returns fatal close and context
errors rather than retrying forever.

`Dispatcher` registers functions with exactly one argument. Dispatch matches
the argument's exact reflection type; handlers with the wrong shape panic at
registration. A panic during a handler call is recovered by the dispatcher.
`Session` stores session ID, resume URL, and sequence safely for concurrent
access. Heartbeats and shard lifecycle are covered in [`heartbeat.md`](heartbeat.md)
and [`shards.md`](shards.md).

## Quick Start

```go
package main

import (
	"context"
	"fmt"

	"github.com/discord-go/discord.go/events"
	"github.com/discord-go/discord.go/gateway"
	"github.com/discord-go/discord.go/intents"
)

type connection struct{ writes int }

func (c *connection) Read() ([]byte, error) { return nil, context.Canceled }
func (c *connection) Write(data []byte) error { c.writes++; return nil }
func (c *connection) Close() error { return nil }

func main() {
	dispatcher := gateway.NewDispatcher()
	dispatcher.AddHandler(func(ready events.Ready) { fmt.Println(ready.SessionID) })
	client := gateway.NewClient(&connection{}, dispatcher)
	client.Token = "example-token"
	client.Intents = intents.Guilds
	client.Session = gateway.NewSession()

	err := client.Send(context.Background(), gateway.GatewayPayload{Op: gateway.OpcodeHeartbeat})
	fmt.Println(err == nil, client.Session.CanResume())
}
```

The example sends a local payload only. A real client needs a WebSocket-backed
`Connection` and a `ConnFactory` before calling `Start`.

## Creating A Client

Set `Token`, `Intents`, `Shard`, `GatewayURL`, `Compressed`, `Session`, and
`ConnFactory` after `NewClient`. `Send` serializes a `GatewayPayload`, rejects
payloads over 4096 bytes, enforces the package's 120-per-minute send window,
and honors context cancellation. `RequestGuildMembers` and
`JoinVoiceChannel` create the relevant control payloads; a zero channel ID in
`JoinVoiceChannel` means leave the channel.

## Using Sessions And Errors

After READY, store the session ID and resume URL in `Session`. `CanResume` is
true only when both are non-empty, and `ToResume(token)` includes the latest
sequence. OP 7 reconnect requests use resume behavior. `ErrInvalidSession`
clears state and identifies again. `ErrFatalClose` represents authentication,
invalid intents, and other non-resumable close codes. `ErrIdentifyRateLimit`
comes from `IdentifyTracker`, which protects the 1000-identifies-per-24-hour
limit.

## Common Patterns

Install handlers before starting, share one dispatcher across shards, and
attach one cache to the shard manager. Decode unknown dispatches from their raw
payload and keep the sequence with the session. Close the active connection
when the context is canceled so a blocked `Read` can exit; context cancellation
alone cannot interrupt every custom connection implementation.

## Best Practices

Use a caller-owned context with a shutdown deadline. Honor Discord's identify
rate limit and shard start delay. Use compression only when the connection
factory supports the selected stream format. Treat Gateway events as ordered
per shard but not globally ordered across shards.

## Common Mistakes

`NewClient` does not dial, start, or create a dispatcher. A nil `Conn` makes
`Send` return an error. Dispatcher matching is exact, so `func(*events.Ready)`
does not handle an `events.Ready` value. Do not reconnect after a fatal close
without fixing authentication or intents.

## API Walkthrough

The public API includes `Connection`, `Client`, `NewClient`, `Start`, `Send`,
`RequestGuildMembers`, `JoinVoiceChannel`, `Dispatcher` and its methods,
`GatewayPayload`, `Opcode` and constants, `Identify`, `IdentifyProperties`,
`Resume`, `Session` and its methods, `IdentifyTracker`, close-code constants,
`ShardManager`, and the Gateway error values. `HelloData`,
`RequestGuildMembersData`, `VoiceStateUpdateData`, `SessionStartLimit`, and
`ShardID` are the supporting payload types.

## Examples

The Quick Start program is complete and runnable without a Discord connection.
For the event envelope and compression details see [`events.md`](events.md).

## Related APIs

- [`events.md`](events.md)
- [`heartbeat.md`](heartbeat.md)
- [`shards.md`](shards.md)
- [`../intents/`](../intents/README.md)
