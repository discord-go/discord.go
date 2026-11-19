# Templates

The template directories are larger application starting points with their own
`go.mod` files and dependency sets. Run tests from each template directory.

- [`auditbot`](https://github.com/discord-go/discord.go/tree/main/docs/examples/templates/auditbot) demonstrates event-driven audit logging.
- [`musicbot`](https://github.com/discord-go/discord.go/tree/main/docs/examples/templates/musicbot) demonstrates voice state, queue, and audio transport integration.
- [`recorderbot`](https://github.com/discord-go/discord.go/tree/main/docs/examples/templates/recorderbot) demonstrates an opt-in voice recording flow that saves per-user WAV files.

For a two-bot voice integration check, start `musicbot` and `recorderbot` with
separate bot tokens using `run-voice-bots.sh`, then use `!join` / `!play` and
`!record` in a test server. The recorder announces when recording starts; use
`!stoprecord` to close the files. See each template README for setup details.
