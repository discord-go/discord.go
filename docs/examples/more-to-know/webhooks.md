# Webhooks

## Overview

Webhooks let an application publish messages with a webhook identity. The REST
API separates bot-authenticated management (`CreateWebhook`, `GetWebhook`,
`DeleteWebhook`) from token-authenticated execution (`ExecuteWebhook`). A
webhook token is a credential; never log it or expose it to users.

## Tutorial: Create And Execute

1. Require `ManageWebhooks` for the command that creates a webhook.
2. Defer before the management and execution REST calls.
3. Create the webhook in the current channel.
4. Check that Discord returned a token before executing it.
5. Use `ExecuteWebhook` with `Wait: true` behavior from the convenience method.
6. Store the webhook ID and token in a secret store when the webhook is meant
   to survive the command; do not create one per request in production.

## Complete Runnable Example

Copy to `examples/webhooks/main.go`, set `DISCORD_TOKEN`, and run it. Invoke
`/webhook` in a channel where the bot can manage webhooks.

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
	router.Command("webhook", "Create and execute a tutorial webhook", func(ctx *bot.InteractionContext) {
		if ctx.ChannelID == nil || ctx.GuildID == nil {
			_ = ctx.ReplyEphemeral("This command must run in a guild channel.")
			return
		}
		if err := ctx.Defer(); err != nil {
			log.Printf("defer webhook: %v", err)
			return
		}
		requestCtx, cancel := context.WithTimeout(ctx.Context(), 10*time.Second)
		defer cancel()
		created, err := ctx.Bot.Rest.CreateWebhook(requestCtx, *ctx.ChannelID, rest.CreateWebhookParams{
			Name: "discord.go tutorial",
		})
		if err != nil {
			_, _ = ctx.FollowupEphemeral("Could not create the webhook.")
			log.Printf("create webhook: %v", err)
			return
		}
		if created.Token == "" {
			_, _ = ctx.FollowupEphemeral("Discord did not return an executable webhook token.")
			return
		}
		message, err := ctx.Bot.Rest.ExecuteWebhook(requestCtx, created.ID, created.Token, rest.ExecuteWebhookParams{
			Content:  "Published through a discord.go webhook.",
			Username: "Tutorial webhook",
		})
		if err != nil {
			_, _ = ctx.FollowupEphemeral("Could not execute the webhook.")
			log.Printf("execute webhook: %v", err)
			return
		}
		if message == nil {
			_, _ = ctx.Followup("Webhook executed without a returned message.")
			return
		}
		_, _ = ctx.Followup(fmt.Sprintf("Webhook message created: %s", message.ID.String()))
	})

	webhookCommand, _ := router.Lookup("webhook")
	webhookCommand.Use(bot.GuildOnly()).Use(bot.RequirePermissions(permissions.ManageWebhooks))

	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Management And Token Execution

Use bot-authenticated methods to list, modify, or delete webhooks by ID. Use
`GetWebhookWithToken`, `ModifyWebhookWithToken`, and `DeleteWebhookWithToken`
only when an application intentionally holds the token. `ExecuteWebhookWithOptions`
adds thread and component query options; `ExecuteWebhookWithFiles` handles
multipart uploads. The convenience `ExecuteWebhook` waits for and returns the
created message.

The example creates a webhook for clarity. A production service should create
one during setup, store its token in a secret manager, rotate it when exposure
is suspected, and reuse it.

## Common Mistakes

- Logging `Webhook.Token` or including it in a response.
- Accepting a webhook ID and token directly from a user without authorization.
- Creating a new webhook for every event and exhausting the channel's resources.
- Calling an execution endpoint with bot-token assumptions instead of the
  webhook token.
- Omitting a timeout around create, execute, edit, or delete requests.

## Expected Result

`/webhook` creates a managed webhook, publishes one message through its token,
and returns only the resulting message ID to the invoker.

## Related Pages

- [Permissions](permissions.md)
- [Threads](threads.md)
- [Common Errors](common-errors.md)
- [Embeds](embeds.md)
