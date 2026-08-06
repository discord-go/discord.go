# Interaction Responses And Data

## Overview

`bot.InteractionContext` wraps an `interactions.Interaction` and provides the
complete high-level response lifecycle. It can identify the interaction type,
read command options, acknowledge a command or component, edit the original
response, send follow-ups, display a modal, or answer autocomplete.

Every interaction gets one initial acknowledgement. A normal reply, defer,
component update, modal, autocomplete result, or pong consumes that initial
response. After a reply or defer, use the original-response and follow-up
methods rather than replying again.

## Architecture

The gateway produces an `interactions.Interaction` containing IDs, token, type,
member/user, channel, guild, and raw `Data`. `bot` decodes command data and
component data into `InteractionContext` helpers. Initial responses go through
`rest.Client.CreateInteractionResponse`; follow-ups use the interaction webhook
endpoints. The context applies short deadlines to initial acknowledgements and
longer event-context deadlines to later REST operations.

## Quick Start

This complete program registers a command, reads an option, and replies. It
needs only the guilds intent for the command interaction path.

```go
package main

import (
	"log"
	"os"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/interactions"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}

	router := bot.NewRouter()
	router.Command("echo", "Echo text", func(ctx *bot.InteractionContext) {
		if err := ctx.Reply(ctx.GetStringOption("text")); err != nil {
			log.Printf("reply: %v", err)
		}
	}, interactions.ApplicationCommandOption{
		Type:        interactions.ApplicationCommandOptionTypeString,
		Name:        "text",
		Description: "Text to echo",
		Required:    true,
	})

	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

For a command that may take longer than the initial response window, replace
the immediate reply with `Defer` and finish with `EditReply` or `Followup`.

## Creating/Configuration

Handlers receive `*bot.InteractionContext`. The embedded
`*interactions.Interaction` exposes `ID`, `ApplicationID`, `Type`, `GuildID`,
`ChannelID`, `Member`, `User`, `Token`, `Message`, `AppPermissions`, `Locale`,
and `GuildLocale`. These fields can be nil or empty for different interaction
types, so use the predicate methods before reading them.

Use `interactions.InteractionCallbackData` for embeds, components, flags,
attachments, choices, modal fields, and polls. `messages.FlagEphemeral` makes an
initial or follow-up response visible only to the invoking user.

## Using

### Basic: reply once

Use `Reply`, `ReplyEphemeral`, `ReplyEmbed`, or `ReplyComplex`. The complex form
accepts `*interactions.InteractionCallbackData`, which is the most flexible way
to include components.

### Intermediate: defer and edit

Call `Defer` or `DeferEphemeral` immediately, perform work, then call
`EditReply`, `EditReplyComplex`, or `DeleteReply`. `GetReply` retrieves the
original message if the application needs its ID or final content.

### Advanced: component and modal interactions

For a button or select, use `Update` or `UpdateContent` to edit the triggering
message, or `DeferUpdate` when the edit will happen later. For modal submissions,
use `ModalValue` or `ModalValues`. For autocomplete, inspect `FocusedOption` and
return `Autocomplete` choices.

## Common Patterns

- Check `IsChatInputCommand`, `IsContextMenuCommand`, `IsButton`,
  `IsSelectMenu`, `IsModalSubmit`, and `IsAutocomplete` before type-specific
  access.
- Use `HasOption` before treating zero or an empty string as a supplied value.
- Use `Subcommand` and `SubcommandGroup` for nested command definitions.
- Use `FollowupEphemeral` for private progress or validation messages after a
  defer.
- Use `HasResponded`, `Replied`, `Deferred`, and `Ephemeral` in shared helper
  code that may be called by multiple paths.

## Best Practices

### Acknowledge before doing network work

Why: Discord imposes a short initial interaction response window.

Pros: long-running commands can complete reliably and show a thinking state.

Cons: a defer creates an extra edit or follow-up request and may be less
responsive for trivial handlers.

### Prefer typed accessors

Why: option values are decoded through `json.Number` and may be nested under
subcommands.

Pros: `GetIntOption`, `GetFloatOption`, and ID helpers avoid repeated decoding.

Cons: invalid or absent values return zero values, so use `HasOption` when that
distinction matters.

### Treat interaction tokens as short-lived credentials

Why: follow-up methods use `ApplicationID` and `Token` to call webhook routes.

Pros: the context keeps token handling out of most application code.

Cons: storing contexts or tokens for later use is fragile; persist a Discord
message ID and use an authenticated REST path when appropriate.

## Common Mistakes

Incorrect: reply twice as if the second call were another initial response.

```go
_ = ctx.Reply("first")
_ = ctx.Reply("second")
```

Correct: reply once, then send a follow-up.

```go
if err := ctx.Reply("first"); err != nil {
	return
}
_, _ = ctx.Followup("second")
```

Incorrect: update a command interaction that has no triggering message.

```go
_ = ctx.UpdateContent("changed")
```

Correct: use `Update` only for component interactions; use `Reply` or `EditReply`
for commands.

```go
if ctx.IsMessageComponent() {
	_ = ctx.UpdateContent("changed")
} else {
	_ = ctx.Reply("This is a command response")
}
```

## API Walkthrough

- `InteractionContext.CommandName`, `CommandType`, `CustomID`, `ComponentType`,
  `Values`, `ModalValue`, `ModalValues`, `FocusedOption`, `Options`,
  `Subcommand`, `SubcommandGroup`, `HasOption`, and `GetOption` inspect decoded
  interaction data.
- `GetStringOption`, `GetIntOption`, `GetFloatOption`, `GetBoolOption`,
  `GetUserID`, `GetRoleID`, and `GetChannelID` return typed option values.
- `IsChatInputCommand`, `IsCommand`, `IsContextMenuCommand`,
  `IsUserContextMenuCommand`, `IsMessageContextMenuCommand`, `IsRepliable`,
  `InGuild`, `IsAutocomplete`, `IsMessageComponent`, `IsButton`, `IsSelectMenu`,
  and `IsModalSubmit` identify valid response paths.
- `MemberPermissions`, `BotPermissions`, and `TargetID` expose permission and
  context-command data.
- `Reply`, `ReplyEphemeral`, `ReplyEmbed`, `ReplyComplex`, and
  `ReplyComplexWithFiles` send initial responses.
- `Defer`, `DeferEphemeral`, `Update`, `UpdateContent`, `DeferUpdate`, `Pong`,
  `LaunchActivity`, `Autocomplete`, `ShowModal`, `ShowModalBuilder`, and
  `ShowModalComplex` send specialized initial callbacks.
- `GetReply`, `EditReply`, `EditReplyComplex`, `EditReplyWithFiles`, and
  `DeleteReply` manage the original response.
- `Followup`, `FollowupEmbed`, `FollowupEphemeral`, `FollowupComplex`,
  `FollowupComplexWithFiles`, `GetFollowup`, `EditFollowup`, and
  `DeleteFollowup` manage follow-up messages.
- `HasResponded`, `Replied`, `Deferred`, and `Ephemeral` report response state.
- `interactions.InteractionResponse` contains `Type` and callback `Data`;
  `InteractionCallbackData` contains content, embeds, flags, components,
  attachments, choices, modal fields, thread fields, and polls.
- Response type constants include `Pong`, `ChannelMessageWithSource`, both
  deferred forms, `UpdateMessage`, autocomplete, `Modal`, and
  `LaunchActivity`. `interactions.VerifySignature` verifies an Ed25519 HTTP
  interaction request from a public key, timestamp, signature, and body.

## Examples

- [Slash commands](../examples/commands/slash-commands.md)
- [Autocomplete](../examples/commands/autocomplete.md)
- [Buttons](buttons.md)
- [Modals](modals.md)

## Related APIs

- [`commands.md`](commands.md) for routing and command options.
- [`components.md`](components.md) for callback components.
- [`embeds.md`](embeds.md) for rich responses.
- [`../low-level/interactions/README.md`](../low-level/interactions/README.md) for raw interaction models.
