# Music bot template

This Go 1.22+ example joins the voice channel of a user who invokes a text
command, queues audio from `yt-dlp`, and streams Opus through discord.go.

Install `ffmpeg` and `yt-dlp`, enable the Guild Voice States and Message Content
intents, and grant the bot View Channels, Connect, Speak, and Send Messages
permissions. Then run:

```sh
export DISCORD_TOKEN='your-music-bot-token'
go run ./cmd/bot
```

Commands are `!join`, `!play <search or URL>`, `!queue`, `!stop`, `!leave`, and
`!info`. Keep the token in the environment; do not commit it or paste it into
source files.
