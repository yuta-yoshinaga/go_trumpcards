#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPTS="$ROOT/.claude/skills/rebucket-game/scripts"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

python3 - "$SCRIPTS/move-game.py" <<'PY'
import importlib.util
import pathlib
import sys

path = pathlib.Path(sys.argv[1])
spec = importlib.util.spec_from_file_location("move_game", path)
mg = importlib.util.module_from_spec(spec)
spec.loader.exec_module(mg)

for bucket in mg.BUCKETS:
    line = "//go:build !js || !wasm || " + bucket
    assert mg.TAG_RE.findall(line) == [bucket], (bucket, mg.TAG_RE.findall(line))

casino_extra10 = "//go:build !js || !wasm || casino || extra10"
extra_extra11 = "//go:build !js || !wasm || extra || extra11"
assert set(mg.TAG_RE.findall(casino_extra10)) == {"casino", "extra10"}
assert set(mg.TAG_RE.findall(extra_extra11)) == {"extra", "extra11"}
PY

# Euchre.go defines sameColorSuit and FiveHundred.go (also extra11) calls it, so euchre
# alone strands a symbol; only the euchre+fivehundred unit is CLEAN. With the extra8-11
# tags unread, the standalone unit came out `[CLEAN] extra11 / euchre ... (retag 0)`.
python3 "$SCRIPTS/movability.py" --all > "$TMP/movability.txt"
if ! grep -qE '^extra11 +/ euchre / .*\(lose: +[1-9]' "$TMP/movability.txt"; then
  echo 'FAIL: movability.py --all does not report the standalone euchre unit as losing a symbol' >&2
  grep -E '/ euchre' "$TMP/movability.txt" >&2 || true
  exit 1
fi
if ! grep -qE '^\[CLEAN\] extra11 +/ euchre, fivehundred /' "$TMP/movability.txt"; then
  echo 'FAIL: movability.py --all no longer reports euchre+fivehundred as one CLEAN unit' >&2
  exit 1
fi

python3 "$SCRIPTS/crossrefs.py" extra5 euchre > "$TMP/crossrefs.txt"
if ! grep -q 'sameColorSuit' "$TMP/crossrefs.txt"; then
  echo 'FAIL: crossrefs.py extra5 euchre did not report sameColorSuit' >&2
  exit 1
fi

if grep -n 'extra\[' "$SCRIPTS/movability.py" "$SCRIPTS/crossrefs.py"; then
  echo 'FAIL: hardcoded extra bucket regex returned' >&2
  exit 1
fi

printf 'rebucket-game script tests: PASS\n'
