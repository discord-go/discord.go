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

The queue is per guild. The FFmpeg source in `internal/player` accepts a local
file path or a media URL, converts it to 48 kHz stereo PCM, encodes 20 ms Opus
frames, and supplies them to `voiceplayer` for pacing. It requires FFmpeg
and the example's C-backed Opus dependency. `!play` uses yt-dlp to resolve a
YouTube search or URL. YouTube extraction may require an up-to-date yt-dlp
version. See the [voice template setup](../index.md#voice-template-setup) for
the native dependencies on Windows and macOS.

For a controlled test, the two-bot runner also accepts
`VOICE_TEST_MEDIA_INPUT` with an operator-selected local file path, URL, or
search term. The recorder announces that music and other channel participants
may be recorded.
With that variable unset, the runner sends a synthetic tone instead. For a
longer tone, set `VOICE_TEST_TONE_DURATION` (for example `10m`).
Set `VOICE_TEST_MEDIA_LOOP=1` to replay the selected media until the bot stops.

`run-voice-bots.sh` selects a scenario with `VOICE_TEST_SCENARIO`: `baseline`,
`membership` (recorder leaves and rejoins), `reconnect` (music bot restarts),
`long`, or `failure` (invalid channel ID). Each scenario verifies the recorded
WAV is non-silent. See `../voice-test-plan.md` for the full procedure.
