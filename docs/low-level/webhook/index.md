# Webhooks

## Overview

The `webhook` package contains the response model for a Discord webhook. It is
separate from [`channels.Webhook`](../channels/README.md), which is the small
webhook reference embedded in channel-oriented responses. Create, execute,
edit, and delete operations are methods on [`../rest/`](../rest/README.md).

## Architecture

`Webhook` carries ID, type, guild and channel IDs, optional creator user, name,
avatar hash, token, application ID, source guild and source channel, and URL.
`TypeIncoming` is 1, `TypeChannelFollower` is 2, and
`TypeApplication` is 3. Most identifiers use `snowflake.ID` with
string JSON encoding. The token is optional in response models but is a
credential when present.

## Quick Start

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/discord-go/discord.go/webhook"
)

func main() {
	var hook webhook.Webhook
	if err := json.Unmarshal([]byte(`{"id":"100","type":1,"channel_id":"200","name":"alerts","token":"secret"}`), &hook); err != nil {
		panic(err)
	}
	fmt.Println(hook.ID, hook.Type == webhook.TypeIncoming, hook.Name, hook.Token != "")
}
```

## Creating Webhook Models

There is no constructor because this package only represents data. Struct
literals are appropriate for tests and decoded REST responses. Use
`Webhook.Type` to distinguish incoming webhooks from follower and application
webhooks before assuming a token or editable channel. `SourceGuild` and
`SourceChannel` are populated for relevant follower metadata and remain nil
otherwise.

## Using Webhooks

For token-authenticated execution, use REST methods such as
`ExecuteWebhook`, `ExecuteWebhookWithOptions`, and their multipart variants.
Interaction webhook URLs use the interaction token and the REST no-auth
helpers. `Webhook.Token` should never be logged or included in a client-facing
model. The webhook URL field is informational; build routes through the REST
API rather than concatenating untrusted values.

## Common Patterns

Store only the webhook ID and an encrypted token when persistence is needed.
Use `messages` and `components` payload types for content. Set the REST
`Wait` option when the caller needs the created message; otherwise execution
can return without a message body.

## Best Practices

Redact tokens in logs and error reports. Treat a 404 or 401 as a credential or
resource problem rather than retrying indefinitely. Use `context.Context` for
every REST operation and use multipart helpers for attachments.

## Common Mistakes

`webhook.Webhook` and `channels.Webhook` are different types. A webhook token
is not a bot token. The model does not validate URL, name, or type values and
does not fetch the webhook. Discord does not sign incoming webhook payloads
with Ed25519 like it does for interactions — secure webhook endpoints with
HTTPS, validate the webhook token in the URL path, and consider rate-limiting
or IP-restricting the endpoint.

## API Walkthrough

The complete public API is `Webhook`, `Type`, and
`TypeIncoming`, `TypeChannelFollower`, and
`TypeApplication`.

## Examples

The Quick Start program is complete and runnable. See [`../rest/uploads.md`](../rest/uploads.md)
for execution with files and [`../messages/`](../messages/README.md) for
payload construction.

## Related APIs

- [`../rest/`](../rest/README.md)
- [`../channels/`](../channels/README.md)
- [`../messages/`](../messages/README.md)
