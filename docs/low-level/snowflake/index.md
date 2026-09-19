# Snowflakes

## Overview

Discord snowflakes are 64-bit IDs that encode a timestamp and other internal
bits. The `snowflake` package provides `ID`, parsing, timestamp extraction, and
the `IDs` slice type used by Discord's string-encoded ID arrays. Both `ID` and
`IDs` marshal to JSON as strings, matching Discord's wire format.

## Architecture

`ID` is an alias-like defined type over `uint64`. `Parse(string)` accepts a
base-10 unsigned 64-bit value and returns an error for malformed or overflowing
input. `ID.String()` returns the decimal representation. `ID.Time()` extracts
the timestamp using `DiscordEpoch`, which is 1420070400000 milliseconds
(2015-01-01 UTC). The lower snowflake bits are ignored by `Time`.

`ID` implements `MarshalJSON` and writes the decimal representation as a JSON
string, which matches Discord's wire format; Discord rejects snowflakes sent
as JSON numbers in request bodies with a `50035` Invalid Form Body error. The
method takes precedence over legacy `,string` struct tags and produces the
same output, so existing tagged fields are unaffected. `IDs` implements
`MarshalJSON` by writing every ID as a JSON string. Its `UnmarshalJSON`
accepts arrays whose elements are strings or JSON numbers, which makes it
tolerant of both Discord responses and local test fixtures.

## Quick Start

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/discord-go/discord.go/snowflake"
)

func main() {
	id, err := snowflake.Parse("175928847299117063")
	if err != nil {
		panic(err)
	}
	data, _ := json.Marshal(snowflake.IDs{id, 42})
	var ids snowflake.IDs
	if err := json.Unmarshal(data, &ids); err != nil {
		panic(err)
	}
	fmt.Println(id.String(), id.Time().UTC().Format("2006-01-02"), string(data), ids[1])
}
```

## Creating And Using IDs

Use a numeric conversion only when the source value is already an exact
unsigned integer. Prefer `Parse` for path parameters, environment variables,
and JSON strings. Use `ID.String()` when constructing a Discord route or log
field. `IDs(nil)` marshals as `null` through the standard slice behavior,
whereas an allocated empty slice marshals as `[]`.

## Common Patterns

Keep snowflakes as IDs throughout model and persistence code. Use `ID.Time()`
for approximate creation-time display, not as an authoritative event time.
Any `[]ID` slice string-encodes on the wire because `ID` itself marshals as a
JSON string. Keep using `IDs` for role IDs, member IDs, or other API arrays,
since its unmarshaller also accepts numeric elements.

## Best Practices

Never route snowflakes through `float64`; values above 2^53 lose precision.
Check `Parse` errors and reject empty IDs. Treat `ID(0)` as a possible sentinel
only in application code; the package does not declare it invalid.

## Common Mistakes

`ID.Time()` does not validate whether an ID was issued by Discord. Decoding
accepts numeric JSON for compatibility with local fixtures, but marshaling
always emits strings and Discord's canonical wire form is strings; do not
strip the quotes from marshaled output or re-encode IDs as numbers. Do not use
`fmt.Sprint` on arbitrary JSON numbers as a replacement for parsing.

## API Walkthrough

The complete exported API is `DiscordEpoch`, `ID`, `ID.String`, `ID.Time`,
`ID.MarshalJSON`, `ID.IsZero`, `Parse`, `MustParse`, `IDs`, `IDs.MarshalJSON`,
and `IDs.UnmarshalJSON`.

## Examples

The Quick Start program is complete and runnable. Models throughout
[`../models/`](../models/README.md) use these types for IDs.

## Related APIs

- [`../json/`](../json/README.md)
- [`../models/`](../models/README.md)
- [`../rest/`](../rest/README.md)
