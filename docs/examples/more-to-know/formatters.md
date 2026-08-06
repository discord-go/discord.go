# Formatters

## Overview

Discord formatting is protocol text, not a separate formatter package in this
repository. Build mentions, timestamps, inline code, and code blocks with Go's
`fmt` and `strings` packages. Use `snowflake.ID.String()` for IDs and
`messages.AllowedMentions` to control which generated mentions can notify
users.

## Tutorial: Format Safe Protocol Text

1. Parse or obtain a typed `snowflake.ID`.
2. Use the correct Discord token, such as `<@id>` or `<t:unix:F>`.
3. Escape user-controlled Markdown before placing it in formatted text.
4. Restrict allowed mentions when the message includes generated mention text.
5. Keep raw IDs and formatted text separate in application state.

Useful forms include `<@USER_ID>`, `<@&ROLE_ID>`, `<#CHANNEL_ID>`, `<t:UNIX:R>`,
`` `inline` ``, and triple-backtick code blocks. Discord's renderer is client
behavior, so use plain text when formatting is not essential.

## Complete Runnable Example

Copy to `examples/formatters/main.go`, set `DISCORD_TOKEN`, and run it. Invoke
`/format` with a user option.

```go
package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/snowflake"
)

func escapeInline(value string) string {
	return strings.NewReplacer("\\", "\\\\", "`", "\\`", "*", "\\*", "_", "\\_", "~", "\\~").Replace(value)
}

func userMention(id snowflake.ID) string {
	return fmt.Sprintf("<@%s>", id.String())
}

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}

	router := bot.NewRouter()
	router.Command("format", "Show Discord formatting examples", func(ctx *bot.InteractionContext) {
		userID := ctx.GetUserID("user")
		if userID == 0 {
			_ = ctx.ReplyEphemeral("A valid user is required.")
			return
		}
		now := time.Now()
		content := fmt.Sprintf("Mention: %s\nTimestamp: <t:%d:F> (<t:%d:R>)\nInline: `%s`", userMention(userID), now.Unix(), now.Unix(), escapeInline("safe-value"))
		data := &interactions.InteractionCallbackData{
			Content: content,
			AllowedMentions: &messages.AllowedMentions{
				Users: []snowflake.ID{userID},
			},
		}
		if err := ctx.ReplyComplex(data); err != nil {
			log.Printf("format response: %v", err)
		}
	})

	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Allowed Mentions

`messages.AllowedMentions` can allow specific users or roles, or use the
`Parse` field for broad categories. Prefer explicit `Users` and `Roles` lists
for generated messages. If a message echoes user input, do not allow arbitrary
mentions from that input.

## Common Mistakes

- Treating a user name as an ID and producing an invalid mention.
- Formatting a millisecond timestamp as seconds; Discord timestamps use Unix
  seconds.
- Passing unescaped user text into a Markdown code block.
- Using `Parse: []messages.AllowedMentionType{messages.AllowedMentionTypeEveryone}`
  for ordinary generated content.
- Logging rendered content that contains private IDs or user input.

## Expected Result

`/format` sends a controlled user mention, absolute and relative timestamps,
and escaped inline code. Only the selected user is allowed to be mentioned.

## Related Pages

- [Embeds](embeds.md)
- [Permissions](permissions.md)
- [Common Errors](common-errors.md)
