# Select Menus

## Overview

Select menus send selected values in a message-component interaction. The
`components` package supports string, user, role, mentionable, and channel
select builders. String selects carry application-defined option values; the
other types return Discord IDs as strings through `ctx.Values()`.

## Tutorial: Build And Read A Select

1. Build options with `NewSelectOptionBuilder` for a string select.
2. Set a stable custom ID and optional min/max values.
3. Put the menu in an action row.
4. Route the ID with `router.Select`.
5. Read and validate every value returned by `ctx.Values()` before using it.

## Complete Runnable Example

Copy to `examples/select-menus/main.go`, set `DISCORD_TOKEN`, and run it.
Invoke `/select`, choose an option, and inspect the updated message.

```go
package main

import (
	"fmt"
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
	router.Command("select", "Choose a deployment environment", func(ctx *bot.InteractionContext) {
		options := []components.SelectOption{
			components.NewSelectOptionBuilder().SetLabel("Development").SetValue("dev").SetDescription("Local testing").Build(),
			components.NewSelectOptionBuilder().SetLabel("Staging").SetValue("stage").SetDescription("Pre-release testing").Build(),
			components.NewSelectOptionBuilder().SetLabel("Production").SetValue("prod").SetDescription("Released service").Build(),
		}
		menu := components.NewStringSelectMenuBuilder().
			SetCustomID("select:environment").
			SetPlaceholder("Choose one environment").
			AddOptions(options[0], options[1], options[2]).
			SetMinValues(1).
			SetMaxValues(1).
			Build()
		row := components.NewActionRowBuilder().AddComponents(menu).Build()
		if err := ctx.ReplyComplex(&interactions.InteractionCallbackData{
			Content:    "Select an environment.",
			Components: []components.Component{row},
		}); err != nil {
			log.Printf("select response: %v", err)
		}
	})

	router.Select("select:environment", func(ctx *bot.InteractionContext) {
		values := ctx.Values()
		if len(values) != 1 || (values[0] != "dev" && values[0] != "stage" && values[0] != "prod") {
			_ = ctx.ReplyEphemeral("That selection is not valid.")
			return
		}
		if err := ctx.UpdateContent(fmt.Sprintf("Selected environment: %s", values[0])); err != nil {
			log.Printf("select update: %v", err)
		}
	})

	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Select Types

- `NewStringSelectMenuBuilder` uses application-defined `SelectOption` values.
- `NewUserSelectMenuBuilder` returns user IDs.
- `NewRoleSelectMenuBuilder` returns role IDs.
- `NewChannelSelectMenuBuilder` returns channel IDs and can restrict channel
  types with `SetChannelTypes`.
- `RoleSelect`, `UserSelect`, and `ChannelSelect` do not have string options;
  Discord populates their candidates.

All selected values are client input. Check count, format, guild ownership, and
the actor's authority before a REST operation. `router.SelectPrefix` is useful
for a family of menus, but the suffix still needs validation.

## Common Mistakes

Wrong:

```go
router.Select("roles", func(ctx *bot.InteractionContext) {
	// A selected role ID is not permission to modify that role.
	_ = ctx.UpdateContent(ctx.Values()[0])
})
```

Correct handlers check the value and authorization before changing state, and
they handle an empty selection safely. Never index `ctx.Values()` without
checking its length.

## Expected Result

`/select` renders a single-choice string menu. Valid values update the source
message; malformed or unexpected values receive a private error response.

## Related Pages

- [Action Rows](action-rows.md)
- [Interactions](interactions.md)
- [Modals](modals.md)
- [Permissions](../more-to-know/permissions.md)
