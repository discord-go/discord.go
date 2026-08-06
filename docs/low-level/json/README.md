# JSON

## Overview

The `json` package is a narrow compatibility seam around the standard
`encoding/json` package. `Marshal` and `Unmarshal` currently call the standard
implementation, while `RawMessage` is an alias of `encoding/json.RawMessage`.
Protocol packages import this seam so a future parser change does not require
changing every Discord model.

## Architecture

`Marshal(v any) ([]byte, error)` follows standard JSON rules, including
struct tags, `omitempty`, custom `MarshalJSON` methods, and errors for values
that cannot be represented. `Unmarshal(data []byte, v any) error` requires a
non-nil pointer target and honors custom unmarshalers such as model snowflake
parsers. `RawMessage` delays decoding and is used for Gateway `d` payloads and
unknown interaction data.

## Quick Start

```go
package main

import (
	"fmt"

	discordjson "github.com/discord-go/discord.go/json"
)

func main() {
	data, err := discordjson.Marshal(struct {
		Name string `json:"name"`
	}{Name: "gateway"})
	if err != nil {
		panic(err)
	}

	var decoded struct{ Name string `json:"name"` }
	if err := discordjson.Unmarshal(data, &decoded); err != nil {
		panic(err)
	}
	fmt.Println(string(data), decoded.Name)
}
```

## Using RawMessage

`RawMessage` is a byte slice that implements JSON marshaling and unmarshaling.
Decode an envelope first, inspect an event name or opcode, then decode `Data`
into the matching type. Copy the bytes if they must outlive the enclosing
buffer or be handed to another goroutine.

## Common Patterns

Use the package alias `discordjson` when the standard library is also imported
as `json`. Keep Discord IDs in `snowflake.ID` fields so their struct tags and
custom parsing preserve string-encoded IDs. For arbitrary JSON numbers decoded
into `interface{}`, remember that the default `encoding/json` behavior uses
`float64`; use a typed field or a decoder configured with `UseNumber` when
precision matters.

## Best Practices

Always check both marshal and unmarshal errors. Decode into a pointer, not a
value. Prefer the repository's model types over `map[string]any` when a model
exists, because custom unmarshalers handle Discord's nullable and string-ID
conventions.

## Common Mistakes

`RawMessage` does not validate or interpret the nested object by itself.
`Marshal` is not a streaming encoder and does not add a newline. The wrapper
does not change standard JSON number behavior or make a malformed payload
valid.

## API Walkthrough

The entire public API is `Marshal`, `Unmarshal`, and the `RawMessage` alias.
There are no options, constructors, or package-level state.

## Examples

The Quick Start program is complete and runnable. Gateway-specific envelope
handling is shown in [`../gateway/events.md`](../gateway/events.md).

## Related APIs

- [`../snowflake/`](../snowflake/README.md)
- [`../gateway/events.md`](../gateway/events.md)
- [`../interactions/`](../interactions/README.md)
