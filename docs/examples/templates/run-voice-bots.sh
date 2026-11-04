#!/usr/bin/env bash
set -euo pipefail

: "${MUSIC_BOT_TOKEN:?Set MUSIC_BOT_TOKEN in the environment}"
: "${RECORDER_BOT_TOKEN:?Set RECORDER_BOT_TOKEN in the environment}"
: "${VOICE_TEST_GUILD_ID:?Set VOICE_TEST_GUILD_ID in the environment}"
: "${VOICE_TEST_CHANNEL_ID:?Set VOICE_TEST_CHANNEL_ID in the environment}"
: "${VOICE_TEST_TEXT_CHANNEL_ID:?Set VOICE_TEST_TEXT_CHANNEL_ID in the environment}"

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"

(
	cd "$repo_root/docs/examples/templates/musicbot"
	exec env DISCORD_TOKEN="$MUSIC_BOT_TOKEN" VOICE_TEST_GUILD_ID="$VOICE_TEST_GUILD_ID" VOICE_TEST_CHANNEL_ID="$VOICE_TEST_CHANNEL_ID" go run ./cmd/bot
) &
music_pid=$!

(
	cd "$repo_root/docs/examples/templates/recorderbot"
	exec env DISCORD_TOKEN="$RECORDER_BOT_TOKEN" VOICE_TEST_GUILD_ID="$VOICE_TEST_GUILD_ID" VOICE_TEST_CHANNEL_ID="$VOICE_TEST_CHANNEL_ID" VOICE_TEST_TEXT_CHANNEL_ID="$VOICE_TEST_TEXT_CHANNEL_ID" go run .
) &
recorder_pid=$!

cleanup() {
	kill "$music_pid" "$recorder_pid" 2>/dev/null || true
	wait "$music_pid" "$recorder_pid" 2>/dev/null || true
}
trap cleanup EXIT INT TERM

wait
