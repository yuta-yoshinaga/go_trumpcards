#!/usr/bin/env bash
set -uo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT" || exit 1
python3 .claude/skills/game-improve/scripts/test_game_improve.py
exit "$?"
