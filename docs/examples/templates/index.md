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

## Voice template setup

The music and recorder templates use cgo for Opus. On Windows x64, put a
MinGW-w64-compatible C compiler on `PATH`, such as LLVM-MinGW or MSYS2 GCC.
In PowerShell, enable cgo and select the compiler before running `go test` or
`go run` from a voice template directory.

```powershell
$env:CGO_ENABLED = '1'
$env:CC = 'clang' # Use 'gcc' with MSYS2 GCC.
go test ./...
```

On macOS, use the Xcode Command Line Tools. On Apple Silicon, also install
Opus and `pkgconf` so `pkg-config` can locate the system Opus library.

```sh
xcode-select --install
brew install opus pkgconf
CGO_ENABLED=1 go test ./...
```

Music playback also requires `ffmpeg` and `yt-dlp` on `PATH`. On Windows,
run each bot in its own PowerShell window with its own `DISCORD_TOKEN`
environment variable. The automated `run-voice-bots.sh` scenario runner
requires Bash.
