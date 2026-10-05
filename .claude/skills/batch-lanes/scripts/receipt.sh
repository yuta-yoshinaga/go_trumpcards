#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
# go_trumpcards 固有の規約: feat/<issue>- branch names and the shared-variant page list below.
page_test_for() {
  case "$1" in
    frontend/src/pages/*Page.tsx) printf '%s\n' "${1%.tsx}.test.tsx" ;;
    frontend/src/utils/*.ts|frontend/src/utils/**/*.ts) printf '%s\n' "${1%.ts}.test.ts" ;;
    *) return 1 ;;
  esac
}
if [[ "${BATCH_TESTING:-}" = 1 ]]; then return 0 2>/dev/null || exit 0; fi
slot="$1"; wt="$WT_ROOT/wt-ib$slot"; cd "$wt" || exit 1
echo "== $(git branch --show-current)"; git status --porcelain; git diff --stat
mapfile -d '' -t changed < <({ git diff --name-only -z HEAD; git ls-files -o --exclude-standard -z; } | sort -zu)
tests=() go_files=() siblings=() components=() new_sources=() fmt=() mut_files=()
for file in "${changed[@]}"; do
  [[ "$file" =~ \.test\.tsx?$ ]] && tests+=("${file#frontend/}")
  [[ "$file" =~ \.go$ ]] && go_files+=("$file")
  case "$file" in
    frontend/src/pages/*Page.tsx)
      candidate=$(page_test_for "$file"); [[ -f "$candidate" ]] && tests+=("${candidate#frontend/}") ;;
  esac
  if [[ "$file" =~ ^frontend/src/(components|hooks|utils)/.*\.tsx?$ && ! "$file" =~ \.test\. ]]; then
    candidate="${file#frontend/}"; candidate="${candidate%.tsx}"; candidate="${candidate%.ts}"
    [[ "$file" == *.tsx ]] && candidate+=".test.tsx" || candidate+=".test.ts"
    [[ -f "frontend/$candidate" ]] && siblings+=("$candidate")
  fi
  [[ "$file" =~ ^frontend/src/components/[A-Za-z]+\.tsx$ ]] && components+=("${file##*/}")
  [[ "$file" =~ ^frontend/src/(utils|hooks)/[^/]+\.tsx?$ && ! "$file" =~ \.test\. ]] && new_sources+=("$file")
  [[ "$file" =~ ^frontend/.*\.(tsx?|json)$ ]] && fmt+=("${file#frontend/}")
  [[ "$file" =~ \.(tsx?|go)$ ]] && mut_files+=("$file")
done
go="${go_files[*]}"
cd frontend || exit 1
((${#tests[@]})) && { bunx vitest run "${tests[@]}" 2>&1 | grep -E 'Test Files|Tests |FAIL' | head; }
game=$(git -C .. branch --show-current | sed -E 's#^feat/[0-9]+-##'); [ -n "$game" ] && { echo "== vitest *$game*"; bunx vitest run "$game" 2>&1 | grep -E 'Test Files|Tests |FAIL' | head -5; }
((${#siblings[@]})) && { echo "== vitest shared siblings"; bunx vitest run "${siblings[@]}" 2>&1 | grep -E 'Test Files|Tests |FAIL' | head -20; }
for file in "${changed[@]}"; do [[ "$file" =~ pages/(BlackJackPage|PineapplePage|SevenCardStudPage|HorsePage|FiveCardStudPage)\.tsx|components/VideoPokerGameContent\.tsx ]] && { echo "NOTE SHARED-VARIANT PAGE: new t() keys must exist in every variant namespace"; break; }; done
for compfile in "${components[@]}"; do comp="${compfile%.tsx}"; users=(); while IFS= read -r f; do [[ -f "$f" ]] && users+=("$f"); done < <(grep -lE "from '\.\./components/${comp}'" src/pages/*.tsx 2>/dev/null | sed 's/\.tsx$/.test.tsx/')
  n=${#users[@]}; if ((n > 0 && n <= 40)); then echo "== vitest importers of $comp ($n)"; bunx vitest run "${users[@]}" 2>&1 | grep -E 'Test Files|Tests |FAIL' | head -10; elif ((n > 40)); then echo "NOTE $comp has $n importers (not run locally; CI covers)"; fi
done
# a .ts hook is often tested from a .test.tsx file (it renders providers), so accept either extension
for u in "${new_sources[@]}"; do t="${u#frontend/}"; t="${t%.tsx}"; t="${t%.ts}"; [ -f "$t.test.ts" ] || [ -f "$t.test.tsx" ] || echo "MISSING_TEST for $u"; done
((${#fmt[@]})) && bunx biome check --write "${fmt[@]}" >/dev/null 2>&1
mkdir -p "$BATCH_STATE"
bun run check >"$BATCH_STATE/ck$slot.log" 2>&1 && echo CHECK=ok || { echo CHECK=FAIL; tail -20 "$BATCH_STATE/ck$slot.log"; }
flock "$BATCH_STATE/tsc.lock" bun run typecheck >"$BATCH_STATE/tc$slot.log" 2>&1 && echo TYPE=ok || { echo TYPE=FAIL; tail -20 "$BATCH_STATE/tc$slot.log"; }
cd ..
if [ -n "$go" ] || printf '%s\n' "${changed[@]}" | grep -q internal/i18n/locales/; then go build ./... && echo BUILD=ok; flock "$BATCH_STATE/gotest.lock" go test -tags test ./... 2>&1 | grep -vE '^(ok|\?)' | tail -15; echo GOTEST_DONE
elif printf '%s\n' "${changed[@]}" | grep -q frontend/src/i18n/locales/; then echo "GOTEST(frontend-locale guards)"; flock "$BATCH_STATE/gotest.lock" go test -tags test ./internal/infrastructure/games/ ./internal/adapter/presenter/ ./internal/i18n/... 2>&1 | grep -vE '^(ok|\?)' | tail -15; echo GOTEST_DONE; fi
if ((${#mut_files[@]})); then grep -nE '\{false &&|if \(true\)' -- "${mut_files[@]}" </dev/null 2>/dev/null && echo MUTANT_LEFT; fi
if [ -n "$go" ]; then export PATH=$PATH:"$(go env GOPATH)/bin"; gi=$(goimports -l "${go_files[@]}"); [ -n "$gi" ] && echo "GOIMPORTS: $gi"; flock "$BATCH_STATE/lint.lock" golangci-lint run --build-tags test ./internal/... 2>&1 | tail -3 | sed 's/^/LINT: /'; fi
"$B/fieldcheck.sh" "$wt"
python3 "$B/e2echeck.py" "$wt"
