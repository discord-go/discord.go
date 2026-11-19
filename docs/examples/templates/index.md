# Templates

The template directories are larger application starting points with their own
`go.mod` files and dependency sets. Run tests from each template directory.

- [`auditbot`](./auditbot/README.md) demonstrates event-driven audit logging.
- [`musicbot`](./musicbot/README.md) demonstrates voice state, queue, and audio transport integration.
- [`recorderbot`](./recorderbot/README.md) demonstrates an opt-in voice recording flow that saves per-user WAV files.

For a two-bot voice integration check, start `musicbot` and `recorderbot` with
separate bot tokens using `run-voice-bots.sh`, then use `!join` / `!play` and
`!record` in a test server. The recorder announces when recording starts; use
`!stoprecord` to close the files. See each template README for setup details.
