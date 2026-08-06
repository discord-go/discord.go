# Heartbeats

## Overview

`gateway.Heartbeater` runs the Gateway heartbeat loop after Discord's HELLO
payload supplies an interval. It sends opcode 1 with the most recent sequence,
tracks acknowledgements, measures the latest round-trip duration, and reports a
zombie connection when an acknowledgement is missing by the next tick.

## Architecture

`NewHeartbeater(conn, interval)` initializes a loop, marks the first
heartbeat as acknowledged so the initial tick is allowed, and starts with an
unknown sequence. `UpdateSequence(seq)` supplies the latest dispatch sequence.
`AckReceived()` marks the last send healthy and updates `Ping()`. `Run(ctx)`
blocks on the ticker, context, stop channel, or connection write error. It
returns `ErrZombieConnection` when the prior heartbeat was not acknowledged.
`Stop()` is non-blocking and causes a running loop to return nil.

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/discord-go/discord.go/gateway"
)

type connection struct{ last []byte }

func (c *connection) Read() ([]byte, error) { return nil, nil }
func (c *connection) Write(data []byte) error { c.last = append(c.last[:0], data...); return nil }
func (c *connection) Close() error { return nil }

func main() {
	conn := &connection{}
	h := gateway.NewHeartbeater(conn, time.Hour)
	h.UpdateSequence(12)
	h.AckReceived()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := h.Run(ctx)
	fmt.Println(err == context.Canceled, h.Ping() == 0, conn.last == nil)
}
```

The long interval keeps the local example from sending a heartbeat before the
context is canceled. In a real client, use the interval from `HelloData`.

## Creating A Heartbeater

The connection must implement `Read`, `Write`, and `Close`; the heartbeater
only calls `Write`. A non-positive interval is passed to `time.NewTicker` and
will panic, so validate the server interval before construction. There is no
default interval and no constructor that starts the loop automatically.

## Using A Heartbeater

Run exactly one loop per Gateway connection, usually in a goroutine owned by
the client. Feed every dispatch sequence to `UpdateSequence` and route
HEARTBEAT_ACK to `AckReceived`. If `Run` returns a zombie or write error,
close the connection and let the Gateway client's reconnect path create a new
heartbeater.

## Common Patterns

Use `Ping` for diagnostics, not as a latency SLA. Treat a zero ping as "no
heartbeat acknowledgement has been measured." Stop the loop during orderly
shutdown before closing the transport.

## Best Practices

Use Discord's interval exactly, avoid duplicate loops, and make connection
closure unblock any reader waiting in another goroutine. Propagate context
errors to the service lifecycle rather than logging them as protocol failures.

## Common Mistakes

Calling `AckReceived` without a sent heartbeat marks the state healthy but does
not create a ping. A zombie error means the transport should be closed; merely
starting another loop on the same connection creates competing heartbeats.

## API Walkthrough

The API is `Heartbeater`, `NewHeartbeater`, `Run`, `Stop`, `UpdateSequence`,
`AckReceived`, `Ping`, `ErrZombieConnection`, and `HelloData`.

## Examples

The Quick Start program is complete and runnable. Full dispatch control flow is
in [`events.md`](events.md) and session reconnect behavior in [`README.md`](README.md).

## Related APIs

- [`README.md`](README.md)
- [`events.md`](events.md)
- [`shards.md`](shards.md)
