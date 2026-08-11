# Buttons

## Overview

Buttons are interactive message components with a label, style, optional emoji,
and either a custom ID or URL. A custom-ID button sends a
`MESSAGE_COMPONENT` interaction; a link button opens its URL and never invokes
a bot route.

## Architecture

`components.ButtonBuilder` creates a `components.Button`. The button travels in
an action row inside `messages.MessageSend.Components` or
`interactions.InteractionCallbackData.Components`. `bot.Router.Button` matches
one exact custom ID and `ButtonPrefix` matches the longest registered prefix,
so overlapping prefixes (e.g. `supreq_cost_` and `supreq_cost_done_`) resolve
to the most specific handler.
Handlers receive `*bot.InteractionContext`; `UpdateContent` edits the message
that contained the button, while `Reply` sends a separate interaction response.

## Quick Start

This complete program sends a button and replaces its message after a click.

```go
package main

import (
	"log"
	"os"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/interactions"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}

	router := bot.NewRouter()
	router.Command("confirm", "Show a confirmation button", func(ctx *bot.InteractionContext) {
		button := components.NewButtonBuilder().
			SetCustomID("confirm").SetLabel("Confirm").
			SetStyle(components.ButtonStyleSuccess).Build()
		row := components.NewActionRowBuilder().AddComponents(button).Build()
		if err := ctx.ReplyComplex(&interactions.InteractionCallbackData{
			Components: []components.Component{row},
		}); err != nil {
			log.Printf("reply: %v", err)
		}
	})
	router.Button("confirm", func(ctx *bot.InteractionContext) {
		if err := ctx.UpdateContent("Confirmed"); err != nil {
			log.Printf("update: %v", err)
		}
	})

	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

Install the application with the command scope and test the command in a guild.

## Creating/Configuration

Create an interactive button with `components.NewButtonBuilder`, then set
`CustomID`, `Label`, `Style`, and optionally `Disabled` or `Emoji`. Use one of
`ButtonStylePrimary`, `Secondary`, `Success`, or `Danger` for custom-ID buttons.
Use `ButtonStyleLink` with `SetURL` for navigation-only buttons. Premium buttons
use the `SKUID` field directly on `components.Button` when that Discord feature
is required.

Put buttons in an action row with `NewActionRowBuilder().AddComponents(...)`.
Register the route before `Run`; use `ButtonPrefix` when the ID contains data,
for example `ticket:close:123`.

## Using

### Basic: exact route

Use a constant custom ID and `router.Button`. Call `UpdateContent` for a quick
message replacement or `Update` for a complete component/embeds payload.

### Intermediate: prefix route

Use `router.ButtonPrefix("ticket:close:", handler)` for a family of buttons.
The current `CustomID` is available through `ctx.CustomID()`. The portion
after the matched prefix is available through `ctx.Suffix()` — for example,
a button with custom ID `ticket:close:123` dispatched under the prefix
`ticket:close:` yields `ctx.Suffix()` == `"123"`. Note that `Suffix()`
returns an empty string when the custom ID equals the prefix exactly (e.g.
`"ticket:close:"` with no suffix); handlers should guard against empty
suffixes. Validate the suffix before changing state.

### Advanced: defer and collect

Call `DeferUpdate` when a click starts work that cannot finish immediately, then
edit the original response with `EditReplyComplex` or send a follow-up. For one
user-specific button flow, use `AwaitInteraction` with a filter that checks user,
message, guild, and custom ID.

## Common Patterns

- Use `confirm:<resource-id>` only when the ID is validated and bounded.
- Disable a button in the updated message after a successful action.
- Use `InteractionRoute.Use` for route-specific middleware.
- Make link buttons visually distinct from state-changing custom-ID buttons.
- Reply ephemerally when a failed action should not modify a public message.

## Best Practices

### Use exact IDs for fixed actions

Why: exact matching is unambiguous and easy to audit.

Pros: fewer accidental matches and simple tests.

Cons: many resources can require many registrations unless a prefix route is
used.

### Scope prefix routes

Why: a prefix handler can receive clicks from any message that contains that
prefix.

Pros: one route supports many resources.

Cons: the handler must parse the ID and enforce authorization itself.

### Acknowledge every click

Why: Discord expects a component interaction acknowledgement.

Pros: users get immediate feedback and the component does not remain stuck.

Cons: a failed update still needs an error path, commonly an ephemeral reply or
follow-up.

## Common Mistakes

Incorrect: creating a button without a custom ID and expecting a route.

```go
button := components.NewButtonBuilder().SetLabel("Confirm").Build()
router.Button("confirm", handler)
```

Correct: use the same custom ID in both places.

```go
button := components.NewButtonBuilder().
	SetCustomID("confirm").SetLabel("Confirm").
	SetStyle(components.ButtonStyleSuccess).Build()
router.Button("confirm", handler)
```

Incorrect: routing a link button.

```go
button := components.NewButtonBuilder().SetURL("https://example.com").Build()
router.Button("docs", handler)
```

Correct: choose either a URL button or an interactive custom-ID button.

```go
button := components.NewButtonBuilder().
	SetCustomID("docs").SetLabel("Docs").
	SetStyle(components.ButtonStylePrimary).Build()
```

## API Walkthrough

- `components.ButtonStyle` has `Primary`, `Secondary`, `Success`, `Danger`,
  `Link`, and `Premium` constants.
- `components.Button` has `Style`, `Label`, `Emoji`, `CustomID`, `URL`,
  `Disabled`, and `SKUID` fields, plus `Type() ComponentType` and
  `MarshalJSON()`.
- `NewButtonBuilder() *ButtonBuilder` creates a builder. Its
  `SetCustomID`, `SetLabel`, `SetURL`, `SetStyle`, `SetDisabled`, and `Build`
  methods return or produce the button.
- `NewActionRowBuilder`, `ActionRowBuilder.AddComponents`, and `Build` wrap a
  button in a legacy action row.
- `Router.Button(customID string, handler InteractionHandler)` registers an
  exact route; `ButtonPrefix(prefix string, handler InteractionHandler)`
  registers a prefix route. Both return `*InteractionRoute`.
- `InteractionRoute.Use(Middleware) *InteractionRoute` adds route middleware.
- `InteractionContext.CustomID() string` returns the full custom ID;
  `Suffix() string` returns the part after the matched prefix for prefix
  routes (empty for exact-match routes). `IsButton() bool`, `Update`,
  `UpdateContent`, `DeferUpdate`, `Reply`, `ReplyEphemeral`, `EditReplyComplex`,
  and `Followup` are the normal button response methods.

## Examples

- [Buttons example](../examples/interactions/buttons.md)
- [Components V2 example](../examples/interactions/components-v2.md)

## Related APIs

- [`components.md`](components.md) for all component types and builders.
- [`interactions.md`](interactions.md) for response lifecycle rules.
- [`collectors.md`](collectors.md) for scoped one-shot button flows.
- [`permissions.md`](permissions.md) for route and command authorization.
