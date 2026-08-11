# Events

## Overview

The `events` package contains typed payload models for common Discord Gateway
dispatches. It is transport-neutral: [`../gateway/`](../gateway/README.md)
receives a payload, and application code can decode its `d` field into one of
these types before updating a cache or invoking business logic.

## Architecture

`Event` is the generic envelope with `Op`, `Data`, `Seq`, and `Type`. For a
dispatch, `Op` is 0, `Data` is the raw JSON object, `Seq` is the sequence, and
`Type` is the event name. The typed wrappers are `Ready`, `GuildCreate`,
`GuildUpdate`, `GuildDelete`, `ChannelCreate`, `ChannelUpdate`,
`MessageCreate`, `MessageUpdate`, `MessageDelete`, `MessageReactionAdd`,
`InteractionCreate`, and `GuildAuditLogEntryCreate`.

The wrappers embed the corresponding model where possible. `Ready` includes
`V`, `User`, `Guilds` (the array of partial guild objects Discord sends at
READY time, each with `Unavailable` set to true), and `SessionID`.
`MessageCreate` adds `GuildID`, `MessageReactionAdd` contains IDs and an
`emojis.Emoji`, and `GuildAuditLogEntryCreate` adds `GuildID`.
`MessageCreate.UnmarshalJSON` and the embedded model unmarshalers handle
nested components and string IDs.

## Quick Start

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/discord-go/discord.go/events"
)

func main() {
	wire := []byte(`{"op":0,"d":{"v":10,"session_id":"session","user":{"id":"42","username":"bot"}},"s":7,"t":"READY"}`)
	var envelope events.Event
	if err := json.Unmarshal(wire, &envelope); err != nil {
		panic(err)
	}
	var ready events.Ready
	if err := json.Unmarshal(envelope.Data, &ready); err != nil {
		panic(err)
	}
	fmt.Println(envelope.Type, *envelope.Seq, ready.SessionID, ready.User.ID)
}
```

## Creating And Using Events

Event values are normally decoded, not constructed manually. Use a type switch
after decoding or register typed handlers with `gateway.Dispatcher`. A
`MessageUpdate` may contain only changed message fields because Discord sends a
partial update. A `GuildDelete` can represent an unavailable guild, so inspect
its `Unavailable` field before treating it as a permanent removal.

## Common Patterns

Decode the envelope once, update the session sequence immediately, and pass
`Data` to exactly one typed decoder. For an unknown event, retain the raw bytes
and log the type rather than treating it as malformed. For interactions, use
the `Interaction` model and then decode its `Data` according to the interaction
type.

## Best Practices

Make handlers tolerant of optional pointers and partial updates. Keep event
processing ordered per shard when cache consistency matters. Avoid assuming a
Gateway event is an authoritative REST response; Discord can deliver events
around a write and the API remains eventually consistent.

## Common Mistakes

`Event.Data` is raw JSON, not an already-populated typed object. `Seq` is a
pointer because it is absent on some payloads. Do not decode every dispatch as
`MessageCreate`, and do not use the high-level `bot.EventContext` from this
low-level package.

## API Walkthrough

The complete exported API is `Event`, `Ready` (with `V`, `User`, `Guilds`,
and `SessionID` fields), `GuildCreate`, `GuildUpdate`, `GuildDelete`,
`ChannelCreate`, `ChannelUpdate`, `MessageCreate` and its `UnmarshalJSON`,
`MessageUpdate`, `MessageDelete`, `MessageReactionAdd`, `InteractionCreate`,
and `GuildAuditLogEntryCreate`.

## Examples

The Quick Start program is runnable against no network. Gateway dispatch and
handler registration are covered in [`../gateway/events.md`](../gateway/events.md).

## Related APIs

- [`../gateway/`](../gateway/README.md)
- [`../messages/`](../messages/README.md)
- [`../interactions/`](../interactions/README.md)
