# FAQ

## How long do global slash commands take to propagate?

Up to one hour. Use `bot.WithGuildCommandSync(guildID)` during development for
near-instant propagation.

## What intents do I need?

`intents.Guilds` for slash commands. Add `intents.GuildMessages` and
`intents.MessageContent` for prefix commands. Voice requires
`intents.GuildVoiceStates`.

## How do I shard?

Use `bot.WithShards(count)` or set `numShards` to 0 for automatic detection
via the `gateway/bot` endpoint. Shards start with 5-second delays between
concurrency buckets.

## How do I handle interaction security?

Use `interactions.VerifyRequest` (not `VerifySignature`) to validate incoming
interaction webhooks. `VerifyRequest` enforces both the Ed25519 signature and
timestamp freshness to prevent replay attacks. For a complete HTTP server,
use `interactions.NewServer(publicKey, handler)` which verifies signatures
and timestamps automatically, handles pings, and dispatches to your handler.

## How do I configure voice?

Voice requires `intents.GuildVoiceStates`. Use `bot.JoinVoiceChannel` to
join, then create a `voice.Client` from the voice server update event. The
library handles AES-256-GCM transport encryption (with a cached cipher.AEAD
for performance) and DAVE MLS end-to-end encryption.

## How are bot tokens stored?

Bot tokens are stored in unexported fields on `rest.Client` and
`gateway.Client`. They are set via `rest.New`, `SetToken`, `SetBearerToken`,
or `SetBotToken`. No code with a reference to these clients can read the
token directly. `bot.Start` validates token format and returns
`ErrInvalidToken` for malformed tokens.

## What Discord API version does the library target?

Discord API v10. The gateway URL includes `v=10` and the REST base URL is
`https://discord.com/api/v10`.

## What Go version is required?

Go 1.26 or newer, as declared in `go.mod`.
