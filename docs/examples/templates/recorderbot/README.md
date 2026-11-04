# Recorder bot template

This Go 1.22+ template exercises Discord gateway voice events, the voice
handshake, UDP transport decryption, DAVE decryption when enabled, and Opus
decoding through discord.go. It writes one 48 kHz stereo PCM WAV file per
speaker under `recordings/`.

The bot records only after a user runs `!record` while in a voice channel. It
posts a visible notice in the channel and stops on `!stoprecord`. Use it only in
a server and voice channel where recording is expected and permitted.

## Run

Enable the Guild Voice States and Message Content intents needed
by the bot in the Discord developer settings, invite it with View Channels,
Connect, Speak, and Send Messages permissions, then run:

```sh
export DISCORD_TOKEN='your-recorder-bot-token'
go run .
```

Set `RECORDINGS_DIR` to change the output directory. Files are named by Discord
user ID. Keep the output directory private; it contains recorded conversations.

For a two-bot integration check, start the `musicbot` template with its own token
and use `./docs/examples/templates/run-voice-bots.sh` from the repository root. To avoid putting tokens
in shell history, read them silently into environment variables first:

```sh
read -rsp 'Music bot token: ' MUSIC_BOT_TOKEN; echo
export MUSIC_BOT_TOKEN
read -rsp 'Recorder bot token: ' RECORDER_BOT_TOKEN; echo
export RECORDER_BOT_TOKEN
export VOICE_TEST_GUILD_ID='your-test-server-id'
export VOICE_TEST_CHANNEL_ID='your-test-voice-channel-id'
export VOICE_TEST_TEXT_CHANNEL_ID='your-test-text-channel-id'
./docs/examples/templates/run-voice-bots.sh
```

When the three voice-test IDs are set, both bots join automatically; the recorder
announces the test in the chosen text channel and the music bot sends a short
synthetic tone. Stop both bots with Ctrl-C and inspect the recorder's WAV output.

The template currently supports one active voice recording session per bot
process. It records decoded PCM WAV rather than original Opus packets.
