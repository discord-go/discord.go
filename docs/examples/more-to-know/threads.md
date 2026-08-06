# Threads

## Overview

Threads are channel resources. `discord.go` exposes creation, membership, and
archived-thread methods on `rest.Client`. The most useful workflow is to send
an initial message, then call `StartThreadWithMessage` with that message ID.

## Tutorial: Start A Thread From A Follow-up

1. Restrict the command to a guild.
2. Check `CreatePublicThreads` and `SendMessagesInThreads` as appropriate.
3. Defer the command if creating and editing resources will take time.
4. Send a follow-up message and use its returned ID as the thread starter.
5. Bound the REST request and report the resulting thread ID.

Use `StartThread` for a thread not attached to an existing message. Set
`AutoArchiveDuration`, `RateLimitPerUser`, `Invitable`, and `AppliedTags` in the
typed parameter structs as the channel type requires.

## Complete Runnable Example

Copy to `examples/threads/main.go`, set `DISCORD_TOKEN`, and run it. Invoke
`/thread` in a guild text channel.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/permissions"
	"github.com/discord-go/discord.go/rest"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}

	router := bot.NewRouter()
	router.Command("thread", "Start a discussion thread", func(ctx *bot.InteractionContext) {
		if ctx.ChannelID == nil || ctx.GuildID == nil {
			_ = ctx.ReplyEphemeral("This command must run in a guild channel.")
			return
		}
		if err := ctx.Defer(); err != nil {
			log.Printf("defer thread: %v", err)
			return
		}
		prompt, err := ctx.Followup("Starting a discussion thread...")
		if err != nil {
			log.Printf("thread prompt: %v", err)
			return
		}
		requestCtx, cancel := context.WithTimeout(ctx.Context(), 8*time.Second)
		defer cancel()
		thread, err := ctx.Bot.Rest.StartThreadWithMessage(requestCtx, *ctx.ChannelID, prompt.ID, rest.StartThreadWithMessageParams{
			Name:                "discussion",
			AutoArchiveDuration: 60,
		})
		if err != nil {
			message := "Could not start the thread."
			_, _ = ctx.EditFollowup(prompt.ID, rest.EditMessageParams{Content: &message})
			log.Printf("start thread: %v", err)
			return
		}
		message := fmt.Sprintf("Thread created: %s", thread.ID.String())
		if _, err := ctx.EditFollowup(prompt.ID, rest.EditMessageParams{Content: &message}); err != nil {
			log.Printf("edit thread prompt: %v", err)
		}
	})

	required := permissions.CreatePublicThreads | permissions.SendMessagesInThreads
	threadCommand, _ := router.Lookup("thread")
	threadCommand.Use(bot.GuildOnly()).Use(bot.RequirePermissions(required))

	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Membership And Listing

Use `JoinThread`, `LeaveThread`, `AddThreadMember`, `RemoveThreadMember`,
`GetThreadMember`, and `ListThreadMembers` for membership. Use
`ListActiveThreads` for a guild and the three archived-thread methods for
public, private, or joined-private history. Set `ThreadMembersParams.WithMember`
only when the extra member objects are needed.

## Common Mistakes

- Starting a public thread without the channel's create-thread permission.
- Using the interaction ID instead of the starter message ID.
- Forgetting `SendMessagesInThreads` for the bot's follow-up work.
- Treating `AutoArchiveDuration` as a permanent retention setting.
- Listing archived threads without pagination or a bounded limit.

## Expected Result

`/thread` creates a discussion thread attached to the follow-up message and edits
that message with the new thread ID.

## Related Pages

- [Permissions](permissions.md)
- [Webhooks](webhooks.md)
- [Common Errors](common-errors.md)
