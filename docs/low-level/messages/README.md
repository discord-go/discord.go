# Messages

## Overview

The `messages` package models received messages and outbound message payloads.
It includes embeds, attachments, mentions, reactions, polls, message
references, calls, snapshots, sticker items, components, and message flags.
`MessageSend` is a channel-message payload; interaction callback data belongs
to [`../interactions/`](../interactions/README.md).

## Architecture

`Message` is a broad response model with IDs, author, content, timestamps,
mentions, attachments, embeds, reactions, references, threads, components,
stickers, polls, calls, and interaction metadata. Its custom unmarshaller
decodes component interface values. `MessageSend` omits zero fields and can
carry content, embeds, allowed mentions, a reference, components, stickers,
attachments, flags, a nonce, a poll, and the `EnforceNonce` switch.

`EmbedBuilder` creates `Embed` values with title, description, URL, RFC3339
timestamp, color, footer, image, thumbnail, author, and fields. `Embed.Validate`
checks title 256, description 4096, footer 2048, author 256, field count 25,
field name 256, field value 1024, and total counted characters 6000. It does
not validate every URL or Discord-specific content rule.

## Quick Start

```go
package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/discord-go/discord.go/messages"
)

func main() {
	embed := messages.NewEmbedBuilder().
		SetTitle("Build report").
		SetDescription("All checks passed").
		SetColor(0x2ecc71).
		SetTimestamp(time.Unix(0, 0)).
		AddField("Status", "ok", true).
		Build()
	if err := embed.Validate(); err != nil {
		panic(err)
	}
	payload := messages.MessageSend{
		Content: "Deployment complete",
		Embeds: []messages.Embed{embed},
		AllowedMentions: &messages.AllowedMentions{},
	}
	data, _ := json.Marshal(payload)
	fmt.Println(string(data))
}
```

## Creating Messages

Use `MessageSend` for normal channel creation and `EditMessageParams` from
REST for edits. Pointer slices in edit parameters distinguish "leave as is"
from "replace with an empty list." `MessageReference` supports replies and
cross-channel references. `AllowedMentions` can parse everyone, roles, or
users, or allow explicit role/user IDs; set `RepliedUser` deliberately.

## Using Embeds And Attachments

`Embed` contains `EmbedFooter`, `EmbedImage`, `EmbedVideo`, `EmbedProvider`,
`EmbedAuthor`, and `EmbedField`. `Attachment` is a received attachment;
`AttachmentSend` is the descriptor for an upload and uses a string file ID.
The REST multipart helpers create the actual `files[n]` parts. For an edit,
include every attachment that should remain.

## Using Polls And Reactions

`Poll` contains question media, answer media, expiry, multiselect, layout, and
optional results. `PollAnswer`, `PollMedia`, `PollResults`, and `AnswerCount`
model the nested data. `Reaction` contains total and burst/normal counts,
whether the current user reacted, the emoji, and burst colors.

## Common Patterns

Use `FlagEphemeral` and `FlagIsComponentsV2` only where the receiving endpoint
supports them. `FlagIsComponentsV2` is `1 << 15`; setting it does not build a
V2 component tree. Use `components` values in `MessageSend.Components` and
validate embeds before calling REST.

## Best Practices

Use explicit allowed mentions for user-controlled content. Keep nonces stable
when retrying idempotent sends and use `EnforceNonce` when supported. Preserve
pointer fields in partial edits. Treat received message arrays as snapshots.

## Common Mistakes

`Embed.Validate` checks the documented character and field limits but does not
validate every URL or Discord-specific content rule. `MessageSend.Attachments`
alone does not upload bytes. An empty `AllowedMentions` object is different
from omitting it when application defaults are involved.

## API Walkthrough

The exported API includes `Message`, `MessageSend`, `MessageType` and its
constants, `MessageReference`, `MessageActivity`, `MessageCall` and its
unmarshaller, `MessageSnapshot`, `InteractionMetadata`, `Attachment`,
`AttachmentSend`, `Embed` and all nested embed types, `EmbedBuilder` and its
setters, `AllowedMentions` and its unmarshaller, `AllowedMentionType` and
constants, `Reaction`, `ReactionCountDetails`, `Poll`, `PollAnswer`,
`PollMedia`, `PollResults`, `AnswerCount`, all message flag constants, and
`Message.UnmarshalJSON`.

## Examples

The Quick Start program is complete and runnable. Multipart examples are in
[`../rest/uploads.md`](../rest/uploads.md), and component examples are in
[`../components/`](../components/README.md).

## Related APIs

- [`../components/`](../components/README.md)
- [`../interactions/`](../interactions/README.md)
- [`../rest/`](../rest/README.md)
