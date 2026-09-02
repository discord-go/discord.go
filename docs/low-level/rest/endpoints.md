# Endpoint Groups

## Overview

The REST client exposes endpoint methods by resource family. Each method takes
a context, typed IDs, and a typed parameter where needed; it builds the HTTP
method and route internally. Callers should not concatenate Discord API paths
or hand-write string IDs when a method exists.

## Architecture

### Applications And Commands

`GetCurrentApplication`, `ModifyCurrentApplication`, application emoji methods,
activity instances, role-connection metadata, SKUs, entitlements, and
`GetGateway`/`GetGatewayBot` cover application resources. Global and guild
command methods include create, edit, delete, list, bulk overwrite, and
command-permission methods. `CreateCommandParams` and
`ApplicationCommandPermission` define their payloads.

### Channels And Messages

`GetChannel`, `ModifyChannel`, `DeleteChannel`, `CreateGuildChannel`, DM
creation, invites, webhooks, typing, and permission-overwrite methods cover
channels. `GetChannelInvites` lists a channel's invites and
`GetGuildInvites` lists every guild invite with full metadata (`Uses`,
`MaxUses`, `MaxAge`, `Temporary`, `CreatedAt`, inviter), which is what
invite trackers snapshot. Message methods include `CreateMessage`,
`CreateMessageComplex`, multipart send, `GetChannelMessage`, list, edit,
delete, bulk delete, crosspost, pin/unpin, reactions, polls, and search.
Use `messages.MessageSend`, `GetMessagesParams`, and `EditMessageParams`.
`FetchAllMessages` paginates backward through channel history;
`FormatTranscript` formats messages into a readable text transcript;
`GenerateTranscript` combines both and returns a `File` for upload.
`StringPtr` is a convenience helper for `*string` parameter fields.

### Guilds, Members, And Roles

Guild methods create, fetch, modify, leave, delete, prune, preview, widget,
welcome screen, onboarding, vanity URL, integrations, bans, and role
management. Member methods add, fetch, list, search, modify, remove, add or
remove role, and voice-state operations. `AddGuildMemberRole` and
`RemoveGuildMemberRole` are the dedicated endpoints for single-role changes
(prefer these over `ModifyGuildMember` when adding or removing one role).
`BulkBanGuildMembers` bans up to 200 users in a single request and returns
`BulkBanResult` with `BannedUsers` and `FailedUsers` ID lists. Role position
and channel position methods accept `RolePosition` and `GuildChannelPosition`.

### Threads And Archives

`StartThread`, `StartThreadWithMessage`, `ListActiveThreads`, public/private/
joined archived-thread lists, `ListThreadMembers`, `GetThreadMember`,
`AddThreadMember`, `RemoveThreadMember`, `JoinThread`, and `LeaveThread` cover
thread lifecycle. Query values use `ArchivedThreadsParams` and
`ThreadMembersParams`.

### Webhooks And Interactions

Webhook management includes create, get, modify, delete, token get/modify,
execute, execute with files/options, and webhook message get/edit/delete.
Interaction methods create the initial response, optionally multipart or with
a result, and get/edit/delete the original response or follow-up messages.
Interaction webhook routes use no-auth helpers because the token is in the
URL.

### Expressions, Events, And Moderation

Emoji, sticker, sticker pack, soundboard, AutoMod, scheduled-event, and stage
instance methods cover the expression and event resources. Their parameter
types are `CreateEmojiParams`, `ModifyEmojiParams`, `CreateStickerParams`,
`ModifyStickerParams`, AutoMod create/modify params, scheduled-event params,
soundboard params, and stage params.

### Monetization And Voice

Entitlement, test-entitlement, subscription, and SKU methods cover
application monetization. `ListVoiceRegions`, `GetCurrentGuildVoiceState`,
`GetGuildVoiceState`, and voice-state modification methods provide REST data
for [`../voice/`](../voice/README.md).

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/rest"
)

func main() {
	command := interactions.NewSlashCommandBuilder("ping", "Check health").Build()
	params := rest.CreateCommandParams{Name: command.Name, Description: command.Description, Type: command.Type}
	fmt.Println(params.Name, params.Description, *params.Type == interactions.ApplicationCommandTypeChatInput)
}
```

The example constructs a command request without sending it. A real call uses
`CreateGlobalApplicationCommand` or `CreateGuildApplicationCommand` with an
application ID and context.

## Common Patterns

Use `Get*WithOptions` variants when Discord exposes query flags such as counts,
localizations, or thread IDs. Use list parameter structs for pagination and
preserve snowflake cursors. Prefer typed methods over `Request` so routes,
serialization, and response types stay consistent.

## Best Practices

Check required permissions and endpoint-specific limits before writes. Use
`WithReason` on moderation and guild operations. Treat destructive methods such
as delete guild, ban, bulk delete, and attachment removal as explicit user
actions with audit logging.

## Common Mistakes

Global and guild application commands have different scopes. Message bulk
delete accepts 2-100 IDs. An edit's attachments list is a replacement list,
not an additive patch. Interaction follow-ups use an interaction token rather
than the bot Authorization header.

## API Walkthrough

This page is the endpoint API split: application/commands, channels/messages,
guilds/members/roles, threads, webhooks/interactions, expressions/events/
moderation, monetization, and Gateway/voice information. Parameter and
response structs are exported alongside each family; request transport details
are in [`requests.md`](requests.md) and upload details in [`uploads.md`](uploads.md).

## Examples

The Quick Start program is complete and runnable. It shows how a model builder
feeds a typed REST parameter.

## Related APIs

- [`README.md`](README.md)
- [`requests.md`](requests.md)
- [`uploads.md`](uploads.md)
