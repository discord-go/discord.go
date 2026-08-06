# Emojis And Stickers

## Overview

The `emojis` package contains wire models for custom emojis and Discord
stickers. It exports `Emoji`, `Sticker`, `StickerItem`, `StickerPack`,
`StickerType`, and `StickerFormatType`; it does not perform CRUD. Use the
resource methods in [`../rest/`](../rest/README.md) for retrieval and writes.

## Architecture

`Emoji.ID`, role IDs, and sticker IDs use `snowflake.ID` with
`json:",string"` tags. Discord sometimes sends an emoji's `roles` as an array
of strings; `Emoji.UnmarshalJSON` converts those strings into `roles.Role`
values containing only the ID. It also parses the emoji ID and returns a parse
error for malformed IDs. The remaining fields preserve Discord's optional
state: `Name` is a pointer, and `User` is optional.

`StickerTypeStandard` is 1 and `StickerTypeGuild` is 2. Sticker format values
are PNG 1, APNG 2, Lottie 3, and GIF 4. `StickerPack` contains full stickers,
SKU information, and optional banner and cover IDs.

## Quick Start

Models can be created directly and round-tripped through `encoding/json`.

```go
package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	input := []byte(`{"id":"100","name":"party","roles":["200"],"animated":true}`)
	var emoji struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Roles    []string `json:"roles"`
		Animated bool `json:"animated"`
	}
	if err := json.Unmarshal(input, &emoji); err != nil {
		panic(err)
	}
	output, _ := json.Marshal(emoji)
	fmt.Println(string(output))
}
```

The small anonymous type above demonstrates the Discord wire shape. A real
program should use `emojis.Emoji` so the custom role parsing is applied:

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/discord-go/discord.go/emojis"
)

func main() {
	var emoji emojis.Emoji
	if err := json.Unmarshal([]byte(`{"id":"100","name":"party","roles":["200"]}`), &emoji); err != nil {
		panic(err)
	}
	fmt.Println(emoji.ID, emoji.Roles[0].ID)
}
```

## Creating And Using Models

`Emoji` includes `RequireColons`, `Managed`, `Animated`, and `Available`, plus
optional `User` and role metadata. `Sticker` includes `PackID`, `Tags`, type,
format, availability, guild ownership, optional uploader, and sort value.
`StickerItem` is the reduced message form with ID, name, and format. `StickerPack`
groups a pack's name, description, stickers, SKU, banner asset, and cover
sticker.

## Common Patterns

Use `StickerItem` when processing a message and `Sticker` when you need full
metadata. Build CDN URLs from a sticker ID with [`../cdn/`](../cdn/README.md).
Use pointers for nullable API fields and preserve `nil` when re-serializing
partial data. Pass model values to REST methods rather than hand-writing JSON.

## Best Practices

Check `json.Unmarshal` errors because malformed snowflakes are rejected by the
custom unmarshaller. Do not treat a missing emoji name as an empty name; the
pointer distinguishes null from a present string. Keep the format enum value
with the sticker because Lottie and GIF handling differs from PNG downloads.

## Common Mistakes

The package does not provide `CreateEmoji` or `DeleteSticker`; those are REST
operations. Do not assume `Emoji.Roles` contains complete role objects after
unmarshal: the API supplies IDs, so each element has only `Role.ID` populated.

## API Walkthrough

The exported model API is `Emoji` and `Emoji.UnmarshalJSON`, `Sticker`,
`StickerItem`, `StickerPack`, `StickerType` with its two constants, and
`StickerFormatType` with its four constants. There are no constructors or
validation methods.

## Examples

The second Quick Start program is complete and runnable with the module's
imports. For REST creation and edits, continue with [`../rest/uploads.md`](../rest/uploads.md)
and [`../rest/endpoints.md`](../rest/endpoints.md).

## Related APIs

- [`../cdn/`](../cdn/README.md)
- [`../rest/`](../rest/README.md)
- [`../roles/`](../roles/README.md)
