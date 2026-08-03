#!/usr/bin/env bash
# count-audit.sh — read-only audit of the game-count assertions that a new game must bump.
#
# Prints the source-of-truth count (RegisterKVGame calls per category) alongside every
# hardcoded assertion, and flags any mismatch. Run from the repo root. Safe to run anytime;
# it edits nothing. Exit code 0 = all consistent, 1 = at least one mismatch.
#
# The #1 recurring new-game miss is forgetting one of these assertions: tsc only runs in CI
# (`bun run check` is biome-only), so a stale frontend count slips past local checks and
# fails the expensive, OOM-prone CI round-trip. Run this before committing instead.
set -u

cd "$(git rev-parse --show-toplevel 2>/dev/null || echo .)" || exit 2

REG_TEST=internal/infrastructure/games/registry_test.go
TUT_HOOK=frontend/src/hooks/useTutorialProgress.test.ts
TUT_PANEL=frontend/src/components/tutorial/TutorialProgressPanel.test.tsx

fail=0

# --- source of truth: RegisterKVGame calls per category sub-package ---
# Buckets are discovered, not listed. This block used to name casino/classic/solo
# literally and was never updated when ADR-0032 added `extra`, so TOTAL was short
# by that worker's games and every frontend assertion looked wrong (219 vs 144).
# A size bucket is added roughly once a year -- long enough to forget this file.
bucket_count() { grep -rc 'RegisterKVGame' "internal/infrastructure/games/$1/" 2>/dev/null | awk -F: '{s+=$2} END{print s+0}'; }

buckets=$(find internal/infrastructure/games -mindepth 1 -maxdepth 1 -type d -printf '%f\n' | sort)
total=0
sot_line=""
for b in $buckets; do
  n=$(bucket_count "$b")
  [ "$n" = "0" ] && continue
  total=$((total + n))
  sot_line="$sot_line  $b=$n"
  eval "count_$b=$n"
done
cas=${count_casino:-0}; cla=${count_classic:-0}; sol=${count_solo:-0}

echo "── Source of truth (RegisterKVGame calls) ──────────────────────"
printf ' %s  TOTAL=%s\n\n' "$sot_line" "$total"

check() { # label  actual  expected
  if [ "$2" = "$3" ]; then
    printf '  ✅ %-52s %s\n' "$1" "$2"
  else
    printf '  ❌ %-52s %s (expected %s)\n' "$1" "$2" "$3"
    fail=1
  fi
}

# Extract the numeric value from the first line matching ERE $2 in file $1.
# Strip // comments first so an inline trailing number (e.g. "= 47 // up from 45")
# can't be mistaken for the value; take the last remaining digit run on the line.
grepval() { grep -E "$2" "$1" 2>/dev/null | head -1 | sed 's|//.*||' | grep -oE '[0-9]+' | tail -1; }

echo "── Go: $REG_TEST ───────────"
check "expectedCasino"  "$(grepval "$REG_TEST" 'expectedCasino[[:space:]]*=')"  "$cas"
check "expectedClassic" "$(grepval "$REG_TEST" 'expectedClassic[[:space:]]*=')" "$cla"
check "expectedSolo"    "$(grepval "$REG_TEST" 'expectedSolo[[:space:]]*=')"    "$sol"
for b in $buckets; do
  case "$b" in casino|classic|solo) continue ;; esac
  n=$(eval "echo \${count_$b:-}")
  [ -z "$n" ] && continue
  # expectedExtra2 -> the constant is the bucket name with a capitalised first letter
  konst="expected$(printf '%s' "$b" | cut -c1 | tr '[:lower:]' '[:upper:]')$(printf '%s' "$b" | cut -c2-)"
  check "$konst" "$(grepval "$REG_TEST" "$konst[[:space:]]*=")" "$n"
done
echo

echo "── Frontend (tsc-only — NOT caught by 'bun run check') ─────────"

# A file that derives the count from gameRoutes cannot drift, so there is no
# literal to audit and a blank "actual" is the *correct* state, not a stale one
# (#4652 removed these literals on purpose). Reporting ❌ there would push the
# next person to write the number back in — the exact drift this script exists
# to catch. Verify the derivation instead: the identifier must be defined from
# gameRoutes.length in the same file.
derives_count() { # file  identifier
  grep -qE "const $2 = gameRoutes\.length;" "$1" 2>/dev/null
}

# Report a count assertion, accepting a gameRoutes-derived value in place of a literal.
check_count() { # label  file  actual  expected
  if [ -z "$3" ] && derives_count "$2" TOTAL_GAMES; then
    printf '  ✅ %-52s derived from gameRoutes.length\n' "$1"
    return
  fi
  check "$1" "$3" "$4"
}

check_count "useTutorialProgress.test.ts totalCount" "$TUT_HOOK" \
  "$(grepval "$TUT_HOOK" 'totalCount\).toBe\([0-9]')" "$total"
# TutorialProgressPanel.test.tsx has three total assertions on separate lines.
# getByText(/N/) has no unique anchor token, so we take the largest of the
# getByText(/…/) numbers (the file also asserts /0/ and /3/); this assumes the
# total is the largest such literal in the file — true today (124 > 3 > 0).
# The total assertion may be the derived form `getByText(new RegExp(String(TOTAL_GAMES)))`.
# Leave the actual blank in that case so check_count reports it as derived — otherwise the
# largest *remaining* literal (the unrelated /0/ and /3/ assertions) is read as the total.
if grep -qE 'getByText\(new RegExp\(String\(TOTAL_GAMES\)\)\)' "$TUT_PANEL" 2>/dev/null; then
  panel_text=""
else
  panel_text=$(grep -E 'getByText\(/[0-9]+/\)' "$TUT_PANEL" 2>/dev/null | sed 's|//.*||' | grep -oE '/[0-9]+/' | tr -d '/' | sort -rn | head -1)
fi
panel_links=$(grepval "$TUT_PANEL" 'links\.length\)\.toBe\([0-9]')
panel_incomplete=$(grepval "$TUT_PANEL" 'incompleteMarkers\.length\)\.toBe\([0-9]')
check_count "TutorialProgressPanel getByText(/N/)"    "$TUT_PANEL" "$panel_text"       "$total"
check_count "TutorialProgressPanel links.length"      "$TUT_PANEL" "$panel_links"      "$total"
check_count "TutorialProgressPanel incompleteMarkers" "$TUT_PANEL" "$panel_incomplete" "$total"
echo

if [ "$fail" -eq 0 ]; then
  echo "VERDICT: ✅ all count assertions consistent ($total games)"
else
  echo "VERDICT: ❌ mismatch — bump the ❌ lines to match the source-of-truth column above"
fi
exit "$fail"
