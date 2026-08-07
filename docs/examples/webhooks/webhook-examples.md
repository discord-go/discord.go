# Webhook Examples

## Overview

This page covers executing webhooks, editing and deleting webhook messages, and
securing webhook endpoints. Webhook execution is done through the REST client.

## Executing a Webhook

Use `rest.ExecuteWebhook` to send a message via a webhook URL:

```go
import (
    "context"
    "github.com/discord-go/discord.go/rest"
)

func sendWebhook(ctx context.Context, restClient *rest.Client, webhookID, webhookToken string) error {
    params := rest.ExecuteWebhookParams{
        Content: "Hello from discord.go!",
    }
    return restClient.ExecuteWebhook(ctx, webhookID, webhookToken, true, params)
}
```

The `wait` parameter (4th argument) controls whether Discord returns the created
message. Set to `true` when you need the message ID.

## Executing with Embeds

```go
params := rest.ExecuteWebhookParams{
    Embeds: []messages.Embed{
        {
            Title:       "Alert",
            Description: "Service is down",
            Color:       0xFF0000,
        },
    },
    Username: "Alert Bot",
    AvatarURL: "https://example.com/avatar.png",
}
err := restClient.ExecuteWebhook(ctx, webhookID, webhookToken, true, params)
```

## Executing with Attachments

Use multipart upload for file attachments:

```go
params := rest.ExecuteWebhookParams{
    Content: "Here is a file",
    Attachments: []rest.Attachment{
        {Filename: "report.pdf", Reader: fileReader},
    },
}
err := restClient.ExecuteWebhookWithFiles(ctx, webhookID, webhookToken, true, params)
```

## Editing a Webhook Message

```go
func editWebhookMessage(ctx context.Context, restClient *rest.Client, webhookID, webhookToken, messageID string) error {
    params := rest.EditWebhookMessageParams{
        Content: "Updated content",
    }
    return restClient.EditWebhookMessage(ctx, webhookID, webhookToken, messageID, params)
}
```

## Deleting a Webhook Message

```go
func deleteWebhookMessage(ctx context.Context, restClient *rest.Client, webhookID, webhookToken, messageID string) error {
    return restClient.DeleteWebhookMessage(ctx, webhookID, webhookToken, messageID)
}
```

## Getting a Webhook

```go
webhook, err := restClient.GetWebhook(ctx, webhookID)
if err != nil {
    return err
}
fmt.Println(webhook.Name, webhook.ChannelID)
```

## Securing Webhook Endpoints

Discord does not sign incoming webhook payloads with Ed25519. To secure webhook
receiver endpoints:

- Use HTTPS for the webhook URL.
- Validate the webhook token in the URL path.
- Consider IP-restricting the endpoint to Discord's IP ranges.
- Rate-limit incoming requests to prevent flooding.

Interaction webhooks (for slash commands) are signed by Discord. Use
`interactions.VerifyRequest` to verify those.

## Common Patterns

- Store webhook ID and token separately, not as a full URL.
- Use `wait=true` when you need the message ID for later edits.
- Set `context.Context` deadlines on all webhook REST calls.
- Use multipart for file uploads.

## Best Practices

- Redact webhook tokens from logs.
- Do not expose webhook tokens in client-side code.
- Use ephemeral messages for user-specific responses via interaction followups.
- Clean up unused webhooks.

## Common Mistakes

- Confusing webhook tokens with bot tokens.
- Not using HTTPS for webhook endpoints.
- Forgetting that `wait=true` is needed to get the message ID.
- Not handling 429 rate limits on webhook execution.
