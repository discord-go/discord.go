# Embeds And Rich Messages

## Overview

An embed is structured rich content attached to a Discord message or interaction
response. `messages.EmbedBuilder` provides a fluent way to construct titles,
descriptions, colors, links, timestamps, images, thumbnails, authors, footers,
and fields. Validate the result before sending when content is user-generated.

## Architecture

The builder returns a `messages.Embed` value. A message context sends it with
`ReplyEmbed` or `ReplyComplex`; an interaction context sends it with the matching
response helpers. The REST layer serializes embeds inside `MessageSend`,
`InteractionCallbackData`, or `ExecuteWebhookParams`.

Embeds are separate from Components V2. A Components V2 payload must use
`messages.FlagIsComponentsV2` and follows Discord's component payload rules;
regular content and embeds should not be mixed into that representation unless
the target Discord endpoint explicitly allows it.

## Quick Start

This complete program replies to `/status` with a validated embed.

```go
package main

import (
	"log"
	"os"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/messages"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}
	router := bot.NewRouter()
	router.Command("status", "Show bot status", func(ctx *bot.InteractionContext) {
		embed := messages.NewEmbedBuilder().
			SetTitle("Status").
			SetDescription("The bot is online").
			SetColor(0x5865F2).
			AddField("Gateway", "connected", true).
			Build()
		if err := embed.Validate(); err != nil {
				log.Printf("embed: %v", err)
				return
			}
		if err := ctx.ReplyEmbed(embed); err != nil {
			log.Printf("reply: %v", err)
		}
	})
	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

Run it in a test guild with `DISCORD_TOKEN=... go run .`.

## Creating/Configuration

Start with `messages.NewEmbedBuilder()`. Set only the fields needed for the
message, then call `Build()`. `Embed.Validate()` checks title, description,
footer, author, field, and total-character limits. Color is an integer RGB value
such as `0x5865F2`; `SetTimestamp` formats a `time.Time` as RFC3339.

For complete control, create `messages.Embed` and its nested `EmbedFooter`,
`EmbedImage`, `EmbedAuthor`, and `EmbedField` values directly. Use
`MessageSend.Embeds` or `InteractionCallbackData.Embeds` when combining embeds
with other supported fields.

## Using

### Basic: title and description

Use `SetTitle`, `SetDescription`, and `SetColor`, then `ReplyEmbed`.

### Intermediate: fields and media

Use `AddField(name, value, inline)`, `SetAuthor`, `SetFooter`, `SetThumbnail`,
and `SetImage`. Keep field values concise and validate before sending.

### Advanced: messages and files

Use `ReplyComplex` with `messages.MessageSend` or
`ReplyComplexWithFiles` when the embed references uploaded content. For an
interaction, use `ReplyComplexWithFiles` and `rest.File` values.

## Common Patterns

- Use a consistent color per status or feature.
- Keep fields short enough for mobile clients and screen readers.
- Set `AllowedMentions` explicitly when embed text includes user-controlled
  strings.
- Use `SetURL` for a title that should link to a canonical resource.
- Validate every generated embed, especially those built from user input.

## Best Practices

### Validate before sending

Why: Discord enforces field and total embed limits.

Pros: failures happen before a REST request and error messages are actionable.

Cons: validation currently covers documented size constraints, not every remote
policy or URL rule.

### Prefer structured fields over formatting tricks

Why: fields have consistent layout semantics.

Pros: clearer desktop and mobile rendering.

Cons: too many fields make an embed dense; use a normal message for long prose.

### Keep user input untrusted

Why: descriptions and field values can contain mentions or misleading content.

Pros: explicit allowed mentions and length checks reduce abuse.

Cons: sanitizing or truncating input can change what users intended to see.

## Common Mistakes

Incorrect: exceeding the field limit without checking.

```go
for i := 0; i < 30; i++ {
	embed.Fields = append(embed.Fields, messages.EmbedField{Name: "x", Value: "y"})
}
```

Correct: keep at most 25 fields and validate.

```go
if len(embed.Fields) > 25 {
	return errors.New("too many embed fields")
}
if err := embed.Validate(); err != nil {
	return err
}
```

Incorrect: treating Components V2 as another embed container.

```go
_ = ctx.ReplyComplex(&interactions.InteractionCallbackData{
	Flags: messages.FlagIsComponentsV2,
	Embeds: []messages.Embed{embed},
})
```

Correct: use regular embed response data, or construct a valid Components V2
payload separately.

```go
_ = ctx.ReplyEmbed(embed)
```

## API Walkthrough

- `Embed` has `Title`, `Type`, `Description`, `URL`, `Timestamp`, `Color`,
  `Footer`, `Image`, `Thumbnail`, `Video`, `Provider`, `Author`, and `Fields`.
- `EmbedFooter`, `EmbedImage`, `EmbedVideo`, `EmbedProvider`, `EmbedAuthor`,
  and `EmbedField` model nested embed values.
- `(*Embed).Validate() error` checks the Discord character and field limits.
- `NewEmbedBuilder() *EmbedBuilder` creates a builder.
- `SetTitle`, `SetDescription`, `SetURL`, `SetTimestamp`, `SetColor`,
  `SetFooter`, `SetImage`, `SetThumbnail`, `SetAuthor`, `AddField`, and `Build`
  configure and return an embed builder or value.
- `MessageSend` contains `Content`, `Embeds`, `AllowedMentions`, `Components`,
  `Attachments`, `Flags`, `Poll`, and related message fields.
- `NewMessageSendBuilder() *MessageSendBuilder` creates a fluent builder for
  `MessageSend` with `SetContent`, `SetEmbeds`/`AddEmbed`,
  `SetComponents`/`AddComponent`, `SetFlags`/`AddFlag`, `SetAllowedMentions`,
  `SetMessageReference`, `SetNonce`, `SetEnforceNonce`, `SetPoll`,
  `AddAttachment`, `SetStickerIDs`, and `Build`.
- `InteractionCallbackData` contains `Content`, `Embeds`, `Flags`, `Components`,
  `Attachments`, and interaction-specific fields.
- `MessageContext.ReplyEmbed`, `ReplyComplex`, and `ReplyComplexWithFiles`,
  plus `InteractionContext.ReplyEmbed`, `ReplyComplex`, and
  `ReplyComplexWithFiles`, send embed payloads.

## Examples

- [Slash commands](../examples/commands/slash-commands.md)
- [Basic client](../examples/setup/basic-client.md)
- [Messages low-level guide](../low-level/messages/README.md)

## Related APIs

- [`interactions.md`](interactions.md) for interaction response data.
- [`components.md`](components.md) for Components V2, a separate message model.
- [`resources.md`](resources.md) for fetching and editing messages.
