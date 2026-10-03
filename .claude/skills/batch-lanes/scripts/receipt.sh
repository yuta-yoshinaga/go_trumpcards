#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
# usage: receipt.sh <slot>  -> status, diff stat, vitest on changed tests, check, typecheck
page_test_for() {
  case "$1" in
    frontend/src/pages/*Page.tsx) printf '%s\n' "${1%.tsx}.test.tsx" ;;
    frontend/src/utils/*.ts|frontend/src/utils/**/*.ts) printf '%s\n' "${1%.ts}.test.ts" ;;
    *) return 1 ;;
  esac
}
if [[ "${BATCH_TESTING:-}" = 1 ]]; then return 0; fi
wt="$WT_ROOT/wt-ib$1"; cd $wt || exit 1
echo "== $(git branch --show-current)"; git status --porcelain; git diff --stat
changed=$(git status --porcelain | awk '{print $2}')
tests=$(printf '%s\n' "$changed" | grep -E '\.test\.tsx?$' | sed 's#^frontend/##' || true)
while IFS= read -r file; do
  case "$file" in
    frontend/src/pages/*Page.tsx|frontend/src/utils/*.ts|frontend/src/utils/**/*.ts)
      candidate=$(page_test_for "$file")
      [ -f "$candidate" ] && tests="${tests}${tests:+$'\n'}${candidate#frontend/}"
      ;;
  esac
done <<< "$changed"
tests=$(printf '%s\n' "$tests" | sort -u | sed '/^$/d')
go=$(git status --porcelain | awk '{print $2}' | grep -E '\.go$')
cd frontend
[ -n "$tests" ] && { bunx vitest run $tests 2>&1 | grep -E 'Test Files|Tests |FAIL' | head; }
game=$(git -C .. branch --show-current | sed -E 's#^feat/[0-9]+-##'); [ -n "$game" ] && { echo "== vitest *$game*"; bunx vitest run "$game" 2>&1 | grep -E 'Test Files|Tests |FAIL' | head -5; }
sib=$(git -C .. status --porcelain | awk '{print $2}' | grep -E '^frontend/src/(components|hooks|utils)/.*\.tsx?$' | grep -v '\.test\.' | sed -E 's#^frontend/##; s#\.(tsx?)$#.test.\1#' | while read f; do [ -f "$f" ] && echo "$f"; done)
[ -n "$sib" ] && { echo "== vitest shared siblings"; bunx vitest run $sib 2>&1 | grep -E 'Test Files|Tests |FAIL' | head -20; }
git -C .. status --porcelain | grep -qE 'pages/(BlackJackPage|PineapplePage|SevenCardStudPage|HorsePage|FiveCardStudPage)\.tsx|components/VideoPokerGameContent\.tsx' && echo "NOTE SHARED-VARIANT PAGE: new t() keys must exist in every variant namespace"
for comp in $(git -C .. status --porcelain | awk '{print $2}' | grep -E '^frontend/src/components/[A-Za-z]+\.tsx$' | sed -E 's#.*/##; s#\.tsx$##'); do
  users=$(grep -lE "from '\.\./components/${comp}'" src/pages/*.tsx 2>/dev/null | sed 's/\.tsx$/.test.tsx/' | while read f; do [ -f "$f" ] && echo "$f"; done)
  n=$(echo "$users" | grep -c .); if [ "$n" -gt 0 ] && [ "$n" -le 40 ]; then echo "== vitest importers of $comp ($n)"; bunx vitest run $users 2>&1 | grep -E 'Test Files|Tests |FAIL' | head -10; elif [ "$n" -gt 40 ]; then echo "NOTE $comp has $n importers (not run locally; CI covers)"; fi
done
for u in $(git -C .. status --porcelain | awk '$1=="??"{print $2}' | grep -E '^frontend/src/(utils|hooks)/[^/]+\.tsx?$' | grep -v '\.test\.'); do t=$(echo "$u" | sed -E 's#^frontend/##; s#\.(tsx?)$#.test.\1#'); [ -f "$t" ] || echo "MISSING_TEST for $u"; done
fmt=$(git -C .. status --porcelain | awk "{print \$2}" | grep -E "^frontend/.*\.(tsx?|json)$" | sed "s#^frontend/##"); [ -n "$fmt" ] && bunx biome check --write $fmt >/dev/null 2>&1
bun run check >/tmp/ck$1.log 2>&1 && echo CHECK=ok || { echo CHECK=FAIL; tail -20 /tmp/ck$1.log; }
flock /tmp/ib-tsc.lock bun run typecheck >/tmp/tc$1.log 2>&1 && echo TYPE=ok || { echo TYPE=FAIL; tail -20 /tmp/tc$1.log; }
cd ..
if [ -n "$go" ] || git status --porcelain | grep -q internal/i18n/locales/; then go build ./... && echo BUILD=ok; flock /tmp/ib-gotest.lock go test -tags test ./... 2>&1 | grep -vE "^(ok|\?)" | tail -15; echo GOTEST_DONE
elif git status --porcelain | grep -q frontend/src/i18n/locales/; then echo "GOTEST(frontend-locale guards)"; flock /tmp/ib-gotest.lock go test -tags test ./internal/infrastructure/games/ ./internal/adapter/presenter/ ./internal/i18n/... 2>&1 | grep -vE "^(ok|\?)" | tail -15; echo GOTEST_DONE; fi
# An empty file list must not reach grep: with no operands it reads stdin and hangs forever.
mut_files=$(git status --porcelain | awk '{print $2}' | grep -E '\.(tsx?|go)$')
[ -n "$mut_files" ] && grep -nE '\{false &&|if \(true\)' $mut_files </dev/null 2>/dev/null && echo MUTANT_LEFT
if [ -n "$go" ]; then export PATH=$PATH:$(go env GOPATH)/bin; gi=$(goimports -l $go); [ -n "$gi" ] && echo "GOIMPORTS: $gi"; flock /tmp/ib-lint.lock golangci-lint run --build-tags test ./internal/... 2>&1 | tail -3 | sed 's/^/LINT: /'; fi
"$B/fieldcheck.sh" "$WT_ROOT/wt-ib$1"
python3 "$B/e2echeck.py" "$WT_ROOT/wt-ib$1"
