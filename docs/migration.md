# Migration Guide

Coming from discordgo or discord.js? This page maps the concepts you already
know to discord.go's names and structure.

## Concept Map

| discordgo | discord.js | discord.go |
|---|---|---|
| `discordgo.Session` | `Client` | `bot.Bot` (via `bot.New`) |
| `session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate))` | `client.on(Events.MessageCreate, ...)` | `bot.OnMessageCreate(func(ctx *bot.MessageContext))` |
| `s.InteractionCreate` handler | `client.on(Events.InteractionCreate, ...)` | `bot.Router` + `router.Command("ping", ...)` |
| `s.ChannelMessageSend(channelID, content)` | `channel.send(...)` | `ctx.Reply(content)` or `bot.Rest.SendMessage(channelID, ...)` |
| `s.GuildMember(guildID, userID)` | `guild.members.fetch(id)` | `bot.FetchMember(ctx, guildID, userID)` |
| `discordgo.MessageEmbed{...}` | `EmbedBuilder` | `messages.NewEmbedBuilder()` |
| `i.Member.Permissions` | `interaction.memberPermissions` | `ctx.MemberPermissions()` on `*bot.InteractionContext` |
| `discordgo.Intents` constants | `GatewayIntentBits` | `intents.Guilds`, `intents.MessageContent`, ... |
| `discordgo.New(...)` + `OpenGateway` | `client.login(...)` | `bot.New(token, ...opts)` then `bot.Run()` |

## The Biggest Differences

### One bot object, typed contexts

discordgo hands you the raw session and event struct in every handler. Here,
`bot.Bot` composes the gateway, REST client, cache, and router, and each event
arrives as a typed context (`*bot.MessageContext`, `*bot.InteractionContext`,
`*bot.ReadyContext`, ...) with helper methods. The raw event struct is still
there: `MessageContext` embeds `*events.MessageCreate`, and every context has
`Decode(v any)` for fields the library does not model yet.

### Router-first command handling

Instead of one giant `InteractionCreate` switch, register commands on a router:

```go
router := bot.NewRouter()
router.Command("ping", "Check bot status", func(ctx *bot.InteractionContext) {
    _ = ctx.Reply("Pong!")
})
client := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
```

Subcommands are declared with builder helpers (`router.Subcommand(...)`,
`router.SubcommandGroup(...)`); inside the handler, `ctx.Subcommand()`,
`ctx.SubcommandGroup()`, and `ctx.SubcommandOptions()` resolve the nesting for
you, including subcommands inside groups.

### Permissions are computed, not just read

On interactions, `ctx.MemberPermissions()` reads the field Discord sends with
the payload. On gateway messages, Discord does not send permissions, so
`MessageContext.MemberPermissions()` computes effective channel permissions
from the cached guild roles and channel overwrites. Enable the `Guilds`
intent (guild/channel/role data) and `GuildMembers` (member roles) for it to
resolve; without them it returns zero ("unknown", not "no permissions").

### Rate limits are built in

The REST client waits on rate-limit headers and retries 429s automatically.
Set `rest.Client.MaxRetries` to bound the retry loop (it returns
`*rest.RateLimitError` when exhausted); the default is unbounded. Inspect
budget with `rest.Client.BucketState(route)`.

## Common Gotchas

- **Intents replace, they do not merge.** `bot.WithIntents(...)` overrides the
  default set. An explicit list without `intents.MessageContent` silently
  empties message content for prefix commands.
- **`Channel.GuildID` is a value, not a pointer.** DM channels carry the zero
  value; check `GuildID.IsZero()` instead of comparing against `nil`.
- **Members are cached from gateway events** (`GUILD_CREATE` with the
  GuildMembers intent, plus `GUILD_MEMBER_ADD`/`UPDATE`/`REMOVE`). REST member
  fetches also populate the cache. Permission helpers read this cache.
- **Interaction responses are one-shot.** After `Reply` or `Defer`, use the
  followup/edit methods (`EditReply`, `Followup`), never a second `Reply`.
- **Option values are exact.** Snowflake options decode as strings, so IDs
  above 2^53 keep full precision; `ctx.GetSnowflake("target")` returns a
  `snowflake.ID` directly.
