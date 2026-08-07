# Glossary

## Overview

This page defines terms used throughout the discord.go documentation.

## Terms

**Application**
A Discord application registered in the Developer Portal. Has a client ID,
client secret, and one or more bot tokens.

**Audit Log Reason**
A header (`X-Audit-Log-Reason`) attached to REST requests that modifies guild
state. Appears in the guild's audit log. Limited to 512 characters.

**Bucket**
A rate limit group identified by Discord's `X-RateLimit-Bucket` header.
Requests in the same bucket share a rate limit.

**Cache**
An in-memory or external store for Discord entities (guilds, channels, users,
etc.). Used to avoid redundant REST calls. Eventually consistent.

**Collector**
A mechanism that waits for a specific gateway event (e.g., a button click)
with a filter and timeout. See `bot.AwaitInteraction` and `bot.AwaitMessage`.

**Command Sync**
The process of registering slash commands with Discord. Global sync takes up
to one hour; guild sync is near-instant.

**Components V2**
Discord's updated component system with containers, sections, separators, and
media galleries. Activated by setting `FlagIsComponentsV2` on message flags.

**DAVE**
Discord Audio Visual End-to-end encryption. An MLS-based protocol for
end-to-end encrypted voice. Supported by discord.go.

**Dispatcher**
The component that receives raw gateway payloads and fans them out to
registered handlers.

**Gateway**
Discord's WebSocket connection for real-time events. Uses API v10.

**Heartbeat**
A periodic ping sent over the gateway to keep the connection alive. The
round-trip time is available via `Heartbeater.Ping`.

**Identify**
The payload sent after connecting to the gateway to authenticate. Includes the
token, intents, and shard information. Rate-limited to 1 per 5 seconds.

**Intent**
A bit flag that declares which event categories the bot wants to receive.
Privileged intents (MessageContent, GuildMembers, GuildPresences) require
approval in the Developer Portal.

**Interaction**
A user-initiated event from a slash command, button, select menu, or modal.
Must be acknowledged within 3 seconds.

**Middleware**
A function that wraps a command handler to add cross-cutting concerns like
permission checks, validation, and cooldowns.

**Nonce**
A unique value used in voice encryption to prevent replay. A 4-byte counter
zero-padded to 12 bytes for AES-256-GCM.

**Opus**
The audio codec used by Discord for voice. Frames are typically 20ms.

**Presence**
A user's online status and activity. Requires the `GuildPresences` privileged
intent to receive for other users.

**Rate Limit**
Discord's per-bucket and global request limits. The library handles these
automatically. Global limit is 50 requests/second.

**Resume**
Reconnecting to the gateway and replaying missed events using a session ID
and sequence number.

**Session**
The gateway session state, including session ID, resume URL, and sequence
number. Used to resume after disconnections.

**Shard**
A single gateway connection responsible for a subset of guilds. Guilds are
assigned to shards by `(guildID >> 22) % numShards`.

**Shard Manager**
The component that manages multiple shard connections, including startup
ordering, concurrency buckets, and shutdown.

**Snowflake**
Discord's ID format. A 64-bit integer encoding timestamp, worker ID, process
ID, and increment. Represented as `snowflake.ID` in discord.go.

**Token**
A bot's authentication credential. Prefixed with `Bot ` in REST Authorization
headers. Never commit to version control. Stored in unexported fields on
`rest.Client` and `gateway.Client`, accessible only via `SetToken`,
`SetBearerToken`, or `SetBotToken`. `bot.Start` validates the token format
(three dot-separated segments) and returns `ErrInvalidToken` if malformed.

**Webhook**
A mechanism for sending messages to a channel without a bot user. Has an ID
and token. Discord does not sign incoming webhook payloads.

**Interaction Server**
An `http.Handler` (`interactions.Server`) that receives Discord interaction
webhooks, verifies their Ed25519 signature and timestamp freshness, and
dispatches them to a user-provided handler. Auto-responds to pings and
defaults to a deferred response when the handler returns nil.

**Paginator**
A helper (`bot.Paginator`) that creates a paginated message with prev/next/stop
button navigation. Pages are provided as `PaginatorPage` structs. Supports
per-user restriction and timeout.

**Reaction Collector**
A one-shot wait (`bot.AwaitReaction`) for a matching `MESSAGE_REACTION_ADD`
event. Uses a filter function and context for cancellation/timeout.

**Nonce Counter**
A 4-byte counter used in voice AES-256-GCM encryption. Incremented per
packet and zero-padded to 12 bytes for GCM. The library checks for overflow
before each send to prevent nonce reuse.
