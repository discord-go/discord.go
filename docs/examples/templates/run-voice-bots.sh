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

# Select the music bot as the only recorded speaker. The bot ID can be supplied
# directly; otherwise resolve it without putting the token in process arguments.
if [[ -z "${VOICE_TEST_SPEAKER_ID:-}" ]]; then
	VOICE_TEST_SPEAKER_ID=$(MUSIC_BOT_TOKEN="$MUSIC_BOT_TOKEN" python3 - <<'PY'
import json
import os
import sys
import urllib.error
import urllib.request

request = urllib.request.Request(
    "https://discord.com/api/v10/users/@me",
    headers={"Authorization": "Bot " + os.environ["MUSIC_BOT_TOKEN"],
             "User-Agent": "DiscordBot (https://github.com/discord-go/discord.go, voice-test)"},
)
try:
    with urllib.request.urlopen(request, timeout=15) as response:
        identity = json.load(response)
    if not identity.get("bot") or not str(identity.get("id", "")).isdigit():
        raise ValueError("identity is not a bot")
    print(identity["id"])
except (OSError, urllib.error.URLError, ValueError, KeyError) as exc:
    print("Unable to resolve the music bot ID: " + type(exc).__name__, file=sys.stderr)
    sys.exit(1)
PY
	)
fi
if [[ ! "$VOICE_TEST_SPEAKER_ID" =~ ^[0-9]+$ || "$VOICE_TEST_SPEAKER_ID" == "0" ]]; then
	echo 'VOICE_TEST_SPEAKER_ID must be a Discord user ID' >&2
	exit 2
fi
export VOICE_TEST_SPEAKER_ID
export VOICE_TEST_TONE_DURATION="${VOICE_TEST_TONE_DURATION:-2m}"

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
	wait_for_file "$dir/$VOICE_TEST_SPEAKER_ID.wav" || true
	sleep 8
	stop_pid recorder_pid
	stop_pid music_pid
	verify_wav "$dir/$VOICE_TEST_SPEAKER_ID.wav" baseline
}

run_membership() {
	local first="$record_root/membership-1"
	local second="$record_root/membership-2"
	start_music
	start_recorder "$first"
	wait_for_file "$first/$VOICE_TEST_SPEAKER_ID.wav" || true
	sleep 6
	stop_pid recorder_pid
	echo "membership: recorder left; waiting 10s"
	sleep 10
	start_recorder "$second" "$record_root/recorder-2.log"
	wait_for_file "$second/$VOICE_TEST_SPEAKER_ID.wav" || true
	sleep 10
	stop_pid recorder_pid
	stop_pid music_pid
	verify_wav "$first/$VOICE_TEST_SPEAKER_ID.wav" "membership before leave"
	verify_wav "$second/$VOICE_TEST_SPEAKER_ID.wav" "membership after rejoin"
}

run_reconnect() {
	local dir="$record_root/reconnect"
	local wav="$dir/$VOICE_TEST_SPEAKER_ID.wav"
	local before_size
	start_music
	start_recorder "$dir"
	wait_for_file "$wav" || true
	sleep 6
	stop_pid music_pid
	echo "reconnect: music bot stopped; waiting 5s"
	sleep 5
	if [[ ! -f "$wav" ]]; then
		echo 'FAIL: reconnect produced no recording before the restart' >&2
		return 1
	fi
	before_size=$(wc -c < "$wav")
	start_music "$record_root/music-2.log"
	sleep 12
	stop_pid recorder_pid
	stop_pid music_pid
	verify_wav "$wav" "reconnect full recording"
	python3 - "$wav" "$before_size" "$dir/after-reconnect.wav" <<'PY'
import sys
import wave

source, offset, target = sys.argv[1], int(sys.argv[2]), sys.argv[3]
with wave.open(source, "rb") as src:
    frame_size = src.getnchannels() * src.getsampwidth()
    src.setpos(min(src.getnframes(), max(0, (offset - 44) // frame_size)))
    frames = src.readframes(src.getnframes())
    with wave.open(target, "wb") as out:
        out.setparams(src.getparams())
        out.writeframes(frames)
PY
	verify_wav "$dir/after-reconnect.wav" "reconnect after restart"
}

run_long() {
	local dir="$record_root/long"
	local duration=${VOICE_TEST_LONG_DURATION:-2m}
	local seconds
	seconds=$(python3 - "$duration" <<'PY'
import re
import sys

text = sys.argv[1]
parts = list(re.finditer(r"(?:\d+(?:\.\d*)?|\.\d+)(?:ms|us|µs|ns|h|m|s)", text))
if not parts or "".join(p.group() for p in parts) != text:
    sys.exit("Invalid VOICE_TEST_LONG_DURATION")
units = {"h": 3600, "m": 60, "s": 1, "ms": 1e-3, "us": 1e-6, "µs": 1e-6, "ns": 1e-9}
total = 0.0
for part in parts:
    match = re.fullmatch(r"(\d+(?:\.\d*)?|\.\d+)(ms|us|µs|ns|h|m|s)", part.group())
    total += float(match.group(1)) * units[match.group(2)]
if not 0.02 <= total <= 3600:
    sys.exit("VOICE_TEST_LONG_DURATION must be between 20ms and 1h")
print(total)
PY
	)
	export VOICE_TEST_TONE_DURATION="$duration"
	start_music
	start_recorder "$dir"
	wait_for_file "$dir/$VOICE_TEST_SPEAKER_ID.wav" || true
	sleep "$seconds"
	stop_pid recorder_pid
	stop_pid music_pid
	verify_wav "$dir/$VOICE_TEST_SPEAKER_ID.wav" "long playback ($duration)"
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
