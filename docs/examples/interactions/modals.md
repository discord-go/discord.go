# Modals

## Overview

A modal is a form shown in response to a command or component interaction. A
text input is wrapped in an action row, and the modal submission is a separate
interaction. `ShowModalBuilder` acknowledges the opening interaction;
`router.Modal` handles the later submission.

## Tutorial: Open And Submit A Form

1. Build one or more `TextInput` values.
2. Put each input in an action row.
3. Build a `ModalData` with its own custom ID and title.
4. Call `ctx.ShowModalBuilder` from the command or button route.
5. Read submitted fields with `ModalValue` or `ModalValues` in the modal route.

Use different IDs for the modal (`feedback:submit`) and its fields (`body`).
The route ID is not the same thing as a text-input ID.

## Complete Runnable Example

Copy to `examples/modals/main.go`, set `DISCORD_TOKEN`, and run it. Invoke
`/feedback`, submit a message, and inspect the private confirmation.

```go
package main

import (
	"log"
	"os"
	"strings"

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
	router.Command("feedback", "Open a feedback form", func(ctx *bot.InteractionContext) {
		body := components.NewTextInputBuilder().
			SetCustomID("body").
			SetLabel("Feedback").
			SetStyle(components.TextInputStyleParagraph).
			SetPlaceholder("Tell us what worked or did not work").
			SetRequired(true).
			SetMaxLength(1000).
			Build()
		row := components.NewActionRowBuilder().AddComponents(body).Build()
		modal := components.NewModalBuilder().
			SetCustomID("feedback:submit").
			SetTitle("Send feedback").
			AddComponents(row).
			Build()
		if err := ctx.ShowModalBuilder(modal); err != nil {
			log.Printf("show modal: %v", err)
		}
	})

	router.Modal("feedback:submit", func(ctx *bot.InteractionContext) {
		body := strings.TrimSpace(ctx.ModalValue("body"))
		if body == "" {
			_ = ctx.ReplyEphemeral("Feedback cannot be empty.")
			return
		}
		if err := ctx.ReplyEphemeral("Thanks for the feedback."); err != nil {
			log.Printf("modal response: %v", err)
			return
		}
		log.Printf("received feedback with %d characters", len(body))
	})

	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Explanation

`ShowModalBuilder` is the initial response to `/feedback`. The later submit
event has its own response deadline and must be acknowledged independently.
`ModalValue` walks the nested action-row structure and returns the value by text
input custom ID. `ModalValues` is useful when a form contains several known
fields.

Text inputs support `TextInputStyleShort` and `TextInputStyleParagraph`, plus
required, length, placeholder, and default value settings. Discord validates
the declared limits, but the server must still validate meaning, format, and
authorization.

## Common Mistakes

- Passing a bare text input to a modal instead of wrapping it in an action row.
- Calling `Reply` after `ShowModalBuilder` succeeded.
- Logging full form values when they may contain personal or secret data.
- Treating labels and placeholders as validation.
- Putting secrets or authorization state in a modal custom ID.

For a long operation after submission, call `ctx.Defer()` and then use
`Followup` or `EditReply`. Keep the modal itself immediate.

## Expected Result

`/feedback` opens a paragraph input. A non-empty submission receives an
ephemeral confirmation and only its length is logged. Empty input receives an
ephemeral validation response.

## Related Pages

- [Action Rows](action-rows.md)
- [Buttons](buttons.md)
- [Interactions](interactions.md)
- [Collectors](../more-to-know/collectors.md)
