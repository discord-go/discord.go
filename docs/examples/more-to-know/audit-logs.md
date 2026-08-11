# Audit Logs

## Overview

Audit logs are a guild REST resource. `discord.go` exposes them through
`b.Rest.GetAuditLog` and models entries with `auditlog.AuditLogEntry`. The
library also has a typed `OnGuildAuditLogEntryCreate` handler for new audit-log
events. Reading the REST history requires the `ViewAuditLog` permission.

## Tutorial: Read Recent Entries

1. Restrict the command to a guild with `bot.GuildOnly`.
2. Require `permissions.ViewAuditLog` for the invoking member.
3. Defer the interaction before the REST request.
4. Bound the request with `context.WithTimeout`.
5. Render only safe summary fields and do not expose sensitive metadata.

The REST endpoint accepts `AuditLogParams` for `UserID`, `ActionType`,
`Before`, `After`, and `Limit`. Discord returns associated users, threads,
webhooks, and entries in one `auditlog.AuditLog` object.

## Complete Runnable Example

Copy to `examples/audit-logs/main.go`, set `DISCORD_TOKEN`, and run it. Invoke
`/audit-log` in a guild where the bot and your member can view the audit log.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/discord-go/discord.go/auditlog"
	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/permissions"
	"github.com/discord-go/discord.go/rest"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}

	router := bot.NewRouter()
	router.Command("audit-log", "Show recent message deletions", func(ctx *bot.InteractionContext) {
		if ctx.GuildID == nil {
			_ = ctx.ReplyEphemeral("This command must run in a guild.")
			return
		}
		if err := ctx.DeferEphemeral(); err != nil {
			log.Printf("defer audit log: %v", err)
			return
		}
		actionType := int(auditlog.MESSAGE_DELETE)
		requestCtx, cancel := context.WithTimeout(ctx.Context(), 8*time.Second)
		defer cancel()
		result, err := ctx.Bot.Rest.GetAuditLog(requestCtx, ctx.GuildID(), rest.AuditLogParams{
			ActionType: &actionType,
			Limit:      10,
		})
		if err != nil {
			_, _ = ctx.FollowupEphemeral(fmt.Sprintf("Could not read the audit log: %v", err))
			return
		}

		embed := messages.NewEmbedBuilder().
			SetTitle("Recent message deletions").
			SetColor(0x5865F2).
			AddField("Entries", fmt.Sprintf("%d", len(result.AuditLogEntries)), true).
			AddField("Associated users", fmt.Sprintf("%d", len(result.Users)), true).
			SetFooter("Audit log data is permission-protected", "").
			Build()
		if _, err := ctx.FollowupComplex(rest.ExecuteWebhookParams{
			Embeds: []messages.Embed{embed},
			Flags:  messages.FlagEphemeral,
		}); err != nil {
			log.Printf("audit log follow-up: %v", err)
		}
	})

	auditCommand, _ := router.Lookup("audit-log")
	auditCommand.Use(bot.GuildOnly()).Use(bot.RequirePermissions(permissions.ViewAuditLog))

	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	b.OnGuildAuditLogEntryCreate(func(ctx *bot.GuildAuditLogEntryContext) {
		log.Printf("audit event guild=%s action=%d entry=%s", ctx.GuildID.String(), ctx.ActionType, ctx.ID.String())
	})
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Note About The Example

`router.Lookup` returns the registered `*bot.Command`, so middleware can be
attached after registration. For new code, chaining the result is usually
clearer:

```go
router.Command("audit-log", "Show recent entries", handler).
	Use(bot.GuildOnly()).
	Use(bot.RequirePermissions(permissions.ViewAuditLog))
```

The typed Gateway event is useful for live notifications, but it is not a
replacement for the historical REST query. Audit-log events may arrive with
only the fields Discord supplies for that action; use IDs and fetch related
resources when a complete object is required.

## Common Mistakes

- Requesting the audit log in a DM.
- Assuming `Guilds` grants the member `ViewAuditLog` permission.
- Fetching an unbounded history in one request.
- Logging `Changes.NewValue` or `OldValue` without considering privacy.
- Performing a moderation action without an audit-log reason.

## Expected Result

`/audit-log` returns a private embed with the number of recent message-delete
entries and associated users. New audit events are logged while the bot is
connected.

## Related Pages

- [Permissions](permissions.md)
- [Common Errors](common-errors.md)
- [Gateway Intents](gateway-intents.md)
