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

- **Intent sets: merge by default, replace explicitly.** `bot.WithIntents(...)`
  adds to the default set, so `intents.MessageContent` is never dropped by
  accident. `bot.WithIntentsExclusive(...)` replaces the defaults and logs
  every dropped privileged intent.
- **Guild/channel IDs on event models are values, not pointers.**
  `Channel.GuildID`, `Message.GuildID`, `Interaction.GuildID`/`ChannelID`,
  and `VoiceState.GuildID`/`ChannelID` are plain `snowflake.ID` values;
  check `IsZero()` instead of comparing against `nil`. (REST request/params
  structs keep pointers where "omit the field" and "send zero" are
  different.)
- **Members are cached from gateway events** (`GUILD_CREATE` with the
  GuildMembers intent, plus `GUILD_MEMBER_ADD`/`UPDATE`/`REMOVE`). REST member
  fetches also populate the cache. Permission helpers read this cache.
- **Interaction responses are one-shot.** After `Reply` or `Defer`, use the
  followup/edit methods (`EditReply`, `Followup`), never a second `Reply`.
- **Option values are exact.** Snowflake options decode as strings, so IDs
  above 2^53 keep full precision; `ctx.OptionSnowflake("target")` returns a
  `snowflake.ID` directly.

## Upgrading to v0.14.0

### WithIntents merges into the defaults

`bot.WithIntents` used to replace the default intent set
(`Guilds | GuildMessages | MessageContent`), so a call that omitted
`MessageContent` silently emptied message content. It now unions the given
intents with `bot.DefaultIntents()`, and multiple calls union with each other.
To replace the defaults entirely, for example to opt out of `MessageContent`,
use `bot.WithIntentsExclusive`; it logs a warning for every privileged intent
(`GuildMembers`, `GuildPresences`, `MessageContent`) the replacement drops. A
non-zero `bot.Config.Intents` keeps its exact-set meaning and gains the same
warning.

Bots that already passed the full default set explicitly see no change. Bots
that relied on `WithIntents` to narrow the defaults must switch to
`WithIntentsExclusive`.

### Option accessors renamed to Option*

`InteractionContext` option readers are named after what they read:
`OptionString`, `OptionInt`, `OptionFloat`, `OptionBool`, `OptionSnowflake`,
`OptionUser`, `OptionRole`, `OptionChannel`, and `Option` for the raw option.
The previous names remain as deprecated aliases, so existing code compiles
unchanged:

| Deprecated alias | Replacement |
|---|---|
| `GetStringOption`, `GetString` | `OptionString` |
| `GetIntOption`, `GetInt` | `OptionInt` |
| `GetFloatOption`, `GetFloat` | `OptionFloat` |
| `GetBoolOption`, `GetBool` | `OptionBool` |
| `GetSnowflake` | `OptionSnowflake` |
| `GetUserID` | `OptionUser` |
| `GetRoleID` | `OptionRole` |
| `GetChannelID` | `OptionChannel` |
| `GetOption` | `Option` |
