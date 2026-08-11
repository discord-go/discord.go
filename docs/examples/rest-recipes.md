# REST Recipes

## Overview

This page provides common REST operation examples using the `rest.Client`.

## Create a Channel

```go
import (
    "context"
    "github.com/discord-go/discord.go/rest"
)

func createTextChannel(ctx context.Context, restClient *rest.Client, guildID string) error {
    params := rest.CreateChannelParams{
        Name: "new-channel",
        Type: 0, // GuildText
    }
    return restClient.CreateGuildChannel(ctx, guildID, params)
}
```

## Bulk Delete Messages

```go
func bulkDelete(ctx context.Context, restClient *rest.Client, channelID string, messageIDs []string) error {
    return restClient.BulkDeleteMessages(ctx, channelID, messageIDs)
}
```

## Kick a Member

```go
func kickMember(ctx context.Context, restClient *rest.Client, guildID, userID string) error {
    ctx = rest.WithReason(ctx, "Rule violation")
    return restClient.RemoveGuildMember(ctx, guildID, userID)
}
```

## Create a Ban

```go
func banMember(ctx context.Context, restClient *rest.Client, guildID, userID string, deleteMessageDays int) error {
    ctx = rest.WithReason(ctx, "Rule violation")
    return restClient.CreateGuildBan(ctx, guildID, userID, rest.CreateBanParams{
        DeleteMessageDays: deleteMessageDays,
    })
}
```

## Create a Role

```go
func createRole(ctx context.Context, restClient *rest.Client, guildID string) error {
    params := rest.CreateGuildRoleParams{
        Name:        "New Role",
        Color:       0x5865F2,
        Permissions: 0,
        Mentionable: true,
    }
    return restClient.CreateGuildRole(ctx, guildID, params)
}
```

## Send a Message

```go
func sendMessage(ctx context.Context, restClient *rest.Client, channelID string) error {
    params := rest.CreateMessageParams{
        Content: "Hello!",
    }
    return restClient.CreateMessage(ctx, channelID, params)
}
```

## Send a Message with Embed

```go
func sendEmbed(ctx context.Context, restClient *rest.Client, channelID string) error {
    params := rest.CreateMessageParams{
        Embeds: []messages.Embed{
            {
                Title:       "Notification",
                Description: "Something happened",
                Color:       0x5865F2,
            },
        },
    }
    return restClient.CreateMessage(ctx, channelID, params)
}
```

## Get Audit Logs

```go
func getAuditLogs(ctx context.Context, restClient *rest.Client, guildID string) error {
    logs, err := restClient.GetGuildAuditLogs(ctx, guildID, rest.AuditLogParams{})
    if err != nil {
        return err
    }
    for _, entry := range logs.AuditLogEntries {
        log.Printf("action %d by %s", entry.ActionType, entry.UserID)
    }
    return nil
}
```

## Common Patterns

- Use `rest.WithReason` to set audit log reasons.
- Use `context.WithTimeout` for all REST calls.
- Use bulk endpoints instead of individual calls for batch operations.
- Check for `*rest.APIError` to handle API-specific errors.
- Use `rest.StringPtr` for `*string` parameter fields like
  `EditMessageParams.Content`.
- Use `rest.GenerateTranscript` to produce a transcript file from channel
  message history, suitable for ticket-bot close flows.

## Best Practices

- Set deadlines on contexts to prevent unbounded waits.
- Use the rate limiter (default) to avoid 429s.
- Log API error codes for debugging.
- Use multipart for file uploads.
