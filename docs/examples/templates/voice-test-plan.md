# Live voice interoperability scenarios

Use a dedicated test guild and two freshly generated bot tokens supplied via
environment variables. Do not store credentials or record participants without
notice. See the recorderbot README for the runner and its visible notice.

The runner selects a scenario with `VOICE_TEST_SCENARIO`:

```sh
export MUSIC_BOT_TOKEN='...'
export RECORDER_BOT_TOKEN='...'
export VOICE_TEST_GUILD_ID='...'
export VOICE_TEST_CHANNEL_ID='...'
export VOICE_TEST_TEXT_CHANNEL_ID='...'
export VOICE_TEST_SCENARIO=membership
./run-voice-bots.sh
```

Each scenario writes WAV files to a temporary directory and checks the mean
volume with `ffmpeg -af volumedetect`; a recording below -50 dB is treated as
silence. The recorder writes a standard 44-byte PCM WAV header, so the files
open in normal audio tools.

1. **Baseline (`baseline`):** both bots join, the music bot plays, the recorder
   captures audio, and the WAV is non-silent.
2. **Membership change (`membership`):** the recorder leaves for ten seconds
   and rejoins while the music bot keeps playing. Both the pre-leave and
   post-rejoin WAVs must be non-silent. This exercises the DAVE re-add
   transition and the add-only commit path.
3. **Reconnect (`reconnect`):** the music bot stops and restarts while the
   recorder stays connected. The recorder's WAV must contain audio after the
   new music-bot session joins. Record whether a new voice session is required
   and whether resources exit cleanly.
4. **Long playback (`long`):** the music bot sends a synthetic tone for
   `VOICE_TEST_LONG_DURATION` (default `2m`; the plan's manual check uses
   `10m`). Inspect recording continuity, packet loss, CPU and memory use. The
   music bot accepts `VOICE_TEST_TONE_DURATION` up to one hour.
5. **Failures (`failure`):** the music bot is started with an invalid channel
   ID. The runner requires a clear voice join or handshake error within ten
   seconds and no hung process. Also exercise invalid channel permissions and
   an unreachable endpoint manually; confirm bounded timeouts and cleanup.

Automated unit tests and the baseline scenario do not by themselves prove the
membership, reconnect, long-playback, or failure scenarios. Network disruption
and participant changes should only be performed in the dedicated test guild.
Record the date, version, duration, and actual result for each run.
