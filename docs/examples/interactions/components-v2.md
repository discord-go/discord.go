# Components V2

## Overview

Components V2 provides typed builders for text displays, separators, sections, containers, media galleries, files, and action rows. The source-backed example sends those values in an interaction callback with `messages.FlagIsComponentsV2`, uploads a generated JSON attachment, and routes a channel select interaction.

## Prerequisites

- Go `1.26.4` or newer.
- `DISCORD_TOKEN` set to a bot token.
- A test guild and a channel where the bot can send messages.
- `Guilds` enabled in the Portal and selected in the bot.
- Valid public image URLs for media gallery or thumbnail content when replacing the example values.
- The bot must be allowed to attach files when using the file portion of the example.

## Architecture

Builders return concrete component values that implement `components.Component`. An interaction response carries them through `interactions.InteractionCallbackData.Components`. A V2 payload must include `messages.FlagIsComponentsV2`; otherwise Discord interprets the payload as a legacy message. Files are uploaded separately as `rest.File` values and referenced by `attachment://filename` in a file component.

## Quick Start

Run the repository source from its root:

```bash
export DISCORD_TOKEN='replace-with-a-bot-token'
go run ./docs/examples/code/v2_components
```

Invoke `/v2-components`, choose a channel in the rendered select, and inspect the uploaded `embed-export.json` attachment. Set `V2_ATTACHMENT` to replace the generated content with a local file.

## Complete Runnable Example

[`examples/v2_components/main.go`](../code/v2_components/main.go) is the complete runnable program. It includes `package main`, imports, token validation, all builders, the V2 flag, a multipart file, the `channel_select` route, and `b.Run()`.

Run that exact source with the command in Quick Start. Do not remove the `FlagIsComponentsV2` field or the attachment when copying the complete example.

## Explanation

The example creates components in memory, composes them into a container, and sends the complete callback with `ctx.ReplyComplexWithFiles`. The file component points to `attachment://embed-export.json`, and the attached `rest.File` has the matching name. `router.Select` matches the channel select's custom ID and uses `ctx.Values()` to read selected IDs.

V2 layout components and interactive action rows can coexist in the callback. A link button in the example is navigation only; it does not produce a router interaction.

## Basic Usage

- Build a `TextDisplay` with `NewTextDisplayBuilder`.
- Compose separators and sections with their corresponding builders.
- Put interactive menus in an action row.
- Set `Flags: messages.FlagIsComponentsV2` on the callback data.
- Use `ReplyComplex` for payloads without files and `ReplyComplexWithFiles` for uploads.

## Intermediate Usage

- Use containers and accent colors to create a consistent message hierarchy.
- Use `NewMediaGalleryBuilder` with validated, stable media URLs.
- Use `NewFileBuilder().SetURL("attachment://name")` with an upload whose `Name` is `name`.
- Route select menus with `router.Select` or `router.SelectPrefix`.
- Reply to a select interaction with `UpdateContent` or a new complex update.

## Advanced Usage

- Validate component composition and payload size before sending.
- Generate attachment bytes in memory with `rest.NewAttachmentBuilderFromBytes` and validate local file sizes with the REST helpers.
- Use `bot.WithCommandSync` to keep a large V2 command set scoped during development.
- Version custom IDs and revalidate selected channel, role, or user IDs against current permissions.
- Use a dedicated media proxy or allowlist if external image URLs cannot be trusted.

## Common Patterns

- Define reusable functions that return `components.Component` or `components.Container`.
- Keep the callback data and files in the same operation so attachment references cannot drift.
- Use ephemeral responses for invalid selections and public updates for shared workflow state.
- Limit select values and validate every value before a REST action.
- Log attachment names and sizes, not full user-provided files.

## Best Practices

- Always include `messages.FlagIsComponentsV2` for V2 payloads.
- Keep attachment names stable and safe; do not trust a user-provided path.
- Enforce upload size limits before constructing multipart requests.
- Set interaction deadlines by replying immediately or deferring before slow generation.
- Clean up temporary files and stop any background media generation when the bot shuts down.

## Common Mistakes with wrong/correct examples

### Wrong

```go
_ = ctx.ReplyComplex(&interactions.InteractionCallbackData{
	Components: []components.Component{container},
})
```

### Correct

```go
_ = ctx.ReplyComplex(&interactions.InteractionCallbackData{
	Flags:      messages.FlagIsComponentsV2,
	Components: []components.Component{container},
})
```

### Wrong

```go
file := components.NewFileBuilder().SetURL("attachment://report.json").Build()
attachment := rest.NewAttachmentBuilderFromBytes("other.json", data).Build()
```

### Correct

```go
file := components.NewFileBuilder().SetURL("attachment://report.json").Build()
attachment := rest.NewAttachmentBuilderFromBytes("report.json", data).Build()
```

The examples are excerpts; the linked source is the complete runnable program.

## Expected Result

`/v2-components` returns a V2 layout containing text, separators, sections, galleries, links, a file, and a channel select. A selection invokes `channel_select` and updates the message with the selected channel ID. A custom `V2_ATTACHMENT` path is uploaded under the expected filename.

## Related Pages

- [Examples Overview](README.md)
- [Buttons](buttons.md)
- [Collectors](../more-to-know/collectors.md)
- [Full Template](../advanced/full-template.md)
- [Complete source: `examples/v2_components/main.go`](../code/v2_components/main.go)
