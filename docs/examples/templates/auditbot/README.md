# Audit bot template

This template logs selected guild events, voice state changes, and thread
activity. It uses its own Go module and local replacements for the checked-out
library.

1. Create a Discord bot and invite it to a test server with access to the log
   channel. Grant only the permissions needed for events and messages you use.
2. Enable Message Content in the Developer Portal if you want message content;
   the example also requests guild, moderation, voice-state, and message intents.
3. From this directory, set `DISCORD_TOKEN` and optionally `LOG_CHANNEL_ID`,
   then run `go run .`. Without the channel ID, events are printed locally.
4. Trigger an event in the test server and check the output. Stop with Ctrl-C.

```sh
export DISCORD_TOKEN='your-bot-token'
export LOG_CHANNEL_ID='your-log-channel-id'
go run .
```

Keep credentials outside source control. Audit output can contain server or
member data; control access to the destination and retention period.
