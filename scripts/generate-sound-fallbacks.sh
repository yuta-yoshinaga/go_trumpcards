#!/usr/bin/env bash
# Generates .mp3 fallbacks for every .ogg sound asset in frontend/public/sounds/.
#
# Why: SoundProvider's SOUND_FILES lists an .mp3 fallback per sound, and
# Safari/iOS cannot decode Ogg Vorbis — without real .mp3 files those users
# get a silent app. Re-run this script whenever an .ogg clip is added or
# re-recorded so the two formats never drift.
#
# Usage: ./scripts/generate-sound-fallbacks.sh
# Requires: ffmpeg (any recent build; a static build in ~/.local/bin works).
set -euo pipefail

SOUNDS_DIR="$(cd "$(dirname "$0")/.." && pwd)/frontend/public/sounds"

if ! command -v ffmpeg >/dev/null 2>&1; then
  echo "ERROR: ffmpeg not found on PATH. Install it first (e.g. apt install ffmpeg)." >&2
  exit 1
fi

shopt -s nullglob
oggs=("$SOUNDS_DIR"/*.ogg)
if [ ${#oggs[@]} -eq 0 ]; then
  echo "ERROR: no .ogg files found in $SOUNDS_DIR" >&2
  exit 1
fi

for ogg in "${oggs[@]}"; do
  mp3="${ogg%.ogg}.mp3"
  # -q:a 4 (~130 kbps VBR) keeps short UI clips small with no audible loss.
  ffmpeg -y -loglevel error -i "$ogg" -codec:a libmp3lame -q:a 4 "$mp3"
  echo "generated: $(basename "$mp3")"
done

echo "done: ${#oggs[@]} mp3 fallbacks in $SOUNDS_DIR"
