#!/usr/bin/env bash
# Live voice interoperability runner. Requires two bot tokens and a dedicated
# test guild/channel. Every scenario writes WAV files and verifies that the
# recorder captured non-silent audio unless the scenario is a failure check.
set -euo pipefail

: "${MUSIC_BOT_TOKEN:?Set MUSIC_BOT_TOKEN in the environment}"
: "${RECORDER_BOT_TOKEN:?Set RECORDER_BOT_TOKEN in the environment}"
: "${VOICE_TEST_GUILD_ID:?Set VOICE_TEST_GUILD_ID in the environment}"
: "${VOICE_TEST_CHANNEL_ID:?Set VOICE_TEST_CHANNEL_ID in the environment}"
: "${VOICE_TEST_TEXT_CHANNEL_ID:?Set VOICE_TEST_TEXT_CHANNEL_ID in the environment}"

scenario=${VOICE_TEST_SCENARIO:-baseline}
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
build_dir="$(mktemp -d)"
record_root="$(mktemp -d)"
music_pid=''
recorder_pid=''
music_log="$record_root/music.log"
recorder_log="$record_root/recorder.log"

stop_pid() {
	local name=$1
	local pid=${!name:-}
	[[ -z "$pid" ]] || kill -TERM "$pid" 2>/dev/null || true
	[[ -z "$pid" ]] || wait "$pid" 2>/dev/null || true
	printf -v "$name" '%s' ''
}

cleanup() {
	stop_pid music_pid
	stop_pid recorder_pid
	rm -rf "$build_dir"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

wait_for_file() {
	local path=$1
	local deadline=$((SECONDS + 30))
	while (( SECONDS < deadline )); do
		[[ -f "$path" ]] && return 0
		sleep 1
	done
	return 1
}

verify_wav() {
	local path=$1
	local label=$2
	if [[ ! -s "$path" ]]; then
		echo "FAIL: $label produced no recording at $path" >&2
		return 1
	fi
	if ! command -v ffmpeg >/dev/null 2>&1; then
		echo "WARN: ffmpeg not found; skipping non-silent check for $label" >&2
		return 0
	fi
	local stats
	stats=$(ffmpeg -hide_banner -nostats -i "$path" -af volumedetect -f null - 2>&1 || true)
	local mean
	mean=$(sed -n 's/.*mean_volume: \([-0-9.]*\) dB.*/\1/p' <<<"$stats" | tail -1)
	if [[ -z "$mean" ]]; then
		echo "FAIL: could not read volume from $label ($path)" >&2
		return 1
	fi
	if awk -v m="$mean" 'BEGIN { exit !(m < -50) }'; then
		echo "FAIL: $label is silent (mean_volume ${mean} dB)" >&2
		return 1
	fi
	echo "PASS: $label mean_volume ${mean} dB"
}

start_music() {
	local log=${1:-$music_log}
	(
		cd "$repo_root/docs/examples/templates/musicbot"
		exec env DISCORD_TOKEN="$MUSIC_BOT_TOKEN" "$build_dir/musicbot"
	) >"$log" 2>&1 &
	music_pid=$!
}

start_recorder() {
	local dir=$1
	local log=${2:-$recorder_log}
	mkdir -p "$dir"
	(
		cd "$repo_root/docs/examples/templates/recorderbot"
		exec env DISCORD_TOKEN="$RECORDER_BOT_TOKEN" RECORDINGS_DIR="$dir" "$build_dir/recorderbot"
	) >"$log" 2>&1 &
	recorder_pid=$!
}

build_bots() {
	(cd "$repo_root/docs/examples/templates/musicbot" && go build -o "$build_dir/musicbot" ./cmd/bot)
	(cd "$repo_root/docs/examples/templates/recorderbot" && go build -o "$build_dir/recorderbot" .)
}

run_baseline() {
	local dir="$record_root/baseline"
	start_music
	start_recorder "$dir"
	wait_for_file "$dir/1520751988764049418.wav" || true
	sleep 8
	stop_pid recorder_pid
	stop_pid music_pid
	verify_wav "$dir/1520751988764049418.wav" baseline
}

run_membership() {
	local first="$record_root/membership-1"
	local second="$record_root/membership-2"
	start_music
	start_recorder "$first"
	wait_for_file "$first/1520751988764049418.wav" || true
	sleep 6
	stop_pid recorder_pid
	echo "membership: recorder left; waiting 10s"
	sleep 10
	start_recorder "$second" "$record_root/recorder-2.log"
	wait_for_file "$second/1520751988764049418.wav" || true
	sleep 10
	stop_pid recorder_pid
	stop_pid music_pid
	verify_wav "$first/1520751988764049418.wav" "membership before leave"
	verify_wav "$second/1520751988764049418.wav" "membership after rejoin"
}

run_reconnect() {
	local dir="$record_root/reconnect"
	start_music
	start_recorder "$dir"
	wait_for_file "$dir/1520751988764049418.wav" || true
	sleep 6
	stop_pid music_pid
	echo "reconnect: music bot stopped; waiting 5s"
	sleep 5
	start_music "$record_root/music-2.log"
	sleep 12
	stop_pid recorder_pid
	stop_pid music_pid
	verify_wav "$dir/1520751988764049418.wav" reconnect
}

run_long() {
	local dir="$record_root/long"
	local duration=${VOICE_TEST_LONG_DURATION:-2m}
	export VOICE_TEST_TONE_DURATION="$duration"
	start_music
	start_recorder "$dir"
	wait_for_file "$dir/1520751988764049418.wav" || true
	sleep 60
	stop_pid recorder_pid
	stop_pid music_pid
	verify_wav "$dir/1520751988764049418.wav" "long playback ($duration)"
}

run_failure() {
	local log="$record_root/failure.log"
	# An invalid channel ID must fail the voice join without hanging the runner.
	(
		cd "$repo_root/docs/examples/templates/musicbot"
		exec env DISCORD_TOKEN="$MUSIC_BOT_TOKEN" VOICE_TEST_GUILD_ID=1 VOICE_TEST_CHANNEL_ID=1 \
			"$build_dir/musicbot"
	) >"$log" 2>&1 &
	music_pid=$!
	sleep 10
	stop_pid music_pid
	if grep -qE 'voice test join|Voice handshake failed|voice handshake' "$log"; then
		echo "PASS: failure scenario reported a voice join/handshake error"
	else
		echo "FAIL: failure scenario did not report a voice error" >&2
		tail -20 "$log" >&2
		return 1
	fi
}

build_bots
case "$scenario" in
baseline) run_baseline ;;
membership) run_membership ;;
reconnect) run_reconnect ;;
long) run_long ;;
failure) run_failure ;;
*)
	echo "Unknown VOICE_TEST_SCENARIO=$scenario (baseline, membership, reconnect, long, failure)" >&2
	exit 2
	;;
esac
