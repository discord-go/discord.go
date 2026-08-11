# Modals And Text Inputs

## Overview

A modal is an interaction response containing a title, custom ID, and text
inputs. It is shown in response to an interaction and later produces a modal
submit interaction. `components.ModalBuilder` creates the callback data while
`bot.InteractionContext` provides `ShowModal` and submission accessors.

## Architecture

Discord requires each text input to be inside an action row. The high-level
sequence is:

1. A command or button handler builds text inputs and action rows.
2. `ShowModal` or `ShowModalBuilder` sends the modal as the initial response.
3. Discord sends a new modal-submit interaction with the modal custom ID.
4. `Router.Modal` dispatches it, and `ModalValue` or `ModalValues` reads the
   submitted values by text-input custom ID.

The modal custom ID identifies the form; each text-input custom ID identifies a
field. They are independent namespaces and should be stable.

## Quick Start

This complete program shows a modal from `/profile` and replies with the
submitted name.

```go
package main

import (
	"log"
	"os"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/intents"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}

	router := bot.NewRouter()
	router.Command("profile", "Open a profile form", func(ctx *bot.InteractionContext) {
		input := components.NewTextInputBuilder().
			SetCustomID("name").SetLabel("Display name").
			SetStyle(components.TextInputStyleShort).SetRequired(true).Build()
		row := components.NewActionRowBuilder().AddComponents(input).Build()
		if err := ctx.ShowModal("profile-form", "Profile", row); err != nil {
			log.Printf("show modal: %v", err)
		}
	})
	router.Modal("profile-form", func(ctx *bot.InteractionContext) {
		name := ctx.ModalValue("name")
		if err := ctx.Reply("Saved name: " + name); err != nil {
			log.Printf("reply: %v", err)
		}
	})

	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

Install the application with `applications.commands` and invoke `/profile` in a
test guild.

## Creating/Configuration

Create a text input with `NewTextInputBuilder`. Set `CustomID`, `Style`, and
`Label`; optional fields include placeholder, default value, required flag, and
minimum or maximum length. Wrap each input in
`NewActionRowBuilder().AddComponents(input).Build()`.

You can construct a modal directly with
`NewModalBuilder().SetCustomID(...).SetTitle(...).AddComponents(...)` and pass
its `ModalData` to `ShowModalBuilder`, or call `ShowModal` with the custom ID,
title, and action-row components.

## Using

### Basic: one field

Register one modal route and read `ModalValue("field-id")`. Reply to the submit
interaction exactly once.

### Intermediate: multiple fields

Use distinct IDs such as `display-name`, `timezone`, and `bio`; call
`ModalValues()` to get a `map[string]string`, then validate each value before
writing application state.

### Advanced: scoped workflows

Include a workflow or resource key in the modal custom ID and use
`AwaitInteraction` with a filter for the expected submitter. This is useful when
several forms can be open at once, but the route must still validate ownership.

## Common Patterns

- Use short IDs that are stable across deployments.
- Set `TextInputStyleParagraph` for long text and `Short` for names or codes.
- Set `Required`, `MinLength`, and `MaxLength` in the UI, then validate again in
  the submit handler.
- Reply ephemerally when form validation errors should remain private.
- Keep form parsing separate from database writes so malformed input cannot
  partially update state.

## Best Practices

### Validate on submission

Why: client-side input constraints are not a security boundary.

Pros: the application remains correct when payloads are replayed or changed.

Cons: validation code is duplicated between UI constraints and server logic.

### Use meaningful custom IDs

Why: `ModalValue` looks up fields by exact custom ID.

Pros: handlers are self-documenting and easy to test.

Cons: renaming an ID is a protocol change that requires updating both builder
and handler.

### Keep modal work fast or defer later operations

Why: the submit interaction also needs an acknowledgement.

Pros: users see confirmation quickly.

Cons: a defer followed by an edit adds a request and requires response-state
management.

## Common Mistakes

Incorrect: passing a text input directly to a modal.

```go
_ = ctx.ShowModal("form", "Form", input)
```

Correct: wrap it in an action row.

```go
row := components.NewActionRowBuilder().AddComponents(input).Build()
_ = ctx.ShowModal("form", "Form", row)
```

Incorrect: using the modal ID to read a field.

```go
value := ctx.ModalValue("profile-form")
```

Correct: use the text input custom ID.

```go
value := ctx.ModalValue("name")
```

## API Walkthrough

- `TextInputStyle` has `TextInputStyleShort` and `TextInputStyleParagraph`.
- `TextInput` exposes `CustomID`, `Style`, `Label`, `MinLength`, `MaxLength`,
  `Required`, `Value`, and `Placeholder`, plus `Type` and `MarshalJSON`.
- `NewTextInputBuilder` returns a builder with `SetCustomID`, `SetCustomId`,
  `SetStyle`, `SetLabel`, `SetPlaceholder`, `SetValue`, `SetRequired`,
  `SetMinLength`, `SetMaxLength`, and `Build`.
- `ActionRow` contains `[]Component`; `NewActionRowBuilder`,
  `AddComponents`, and `Build` create the required wrapper.
- `ModalData` contains `CustomID`, `Title`, and `Components`.
- `NewModalBuilder`, `SetCustomID`, `SetCustomId`, `SetTitle`, `AddComponents`,
  and `Build` construct `ModalData` without importing `interactions`.
- `InteractionContext.ShowModal(customID, title string, ...components.Component) error`,
  `ShowModalBuilder(modal components.ModalData) error`, and
  `ShowModalComplex(*interactions.InteractionCallbackData) error` send modal
  callbacks.
- `InteractionContext.IsModalSubmit`, `CustomID`, `ModalValue`, `ModalValues`,
  `Reply`, `ReplyEphemeral`, `Defer`, `EditReply`, and `Followup` are the core
  submit-handler methods.
- `Router.Modal(customID string, handler InteractionHandler) *InteractionRoute`
  registers an exact modal route. `Router.ModalPrefix(prefix string, handler
  InteractionHandler) *InteractionRoute` registers a prefix-matched modal
  route for dynamic modal IDs like `supreq_stop_modal_<requestID>`. Prefix
  matching uses longest-match-first, so overlapping prefixes resolve to the
  most specific handler.

## Examples

- [Modals example](../examples/interactions/modals.md)
- [Buttons opening modals](../examples/interactions/buttons.md)

## Related APIs

- [`components.md`](components.md) for component serialization.
- [`interactions.md`](interactions.md) for response deadlines and follow-ups.
- [`collectors.md`](collectors.md) for one-shot form workflows.
