#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
dry_run=0
if [[ "${1:-}" == --dry-run ]]; then dry_run=1; shift; fi
slot="${1:?usage: e2egate.sh [--dry-run] <slot>}"; shift
[[ "$slot" =~ ^[0-9]+$ ]] || { echo "invalid slot: $slot" >&2; exit 2; }
wt="$WT_ROOT/wt-ib$slot"
[[ -d "$wt" ]] || { echo "missing lane worktree: $wt" >&2; exit 2; }
mapfile -t changed < <({ git -C "$wt" diff --name-only origin/develop...HEAD; git -C "$wt" diff --name-only; git -C "$wt" diff --cached --name-only; git -C "$wt" ls-files --others --exclude-standard; } | sort -u)
frontend=0; full_sweep=0; page_change=0
declare -a explicit_specs=() pages=() helper_paths=()
for path in "${changed[@]}"; do
  [[ "$path" == frontend/* ]] || continue
  frontend=1
  case "$path" in
    frontend/e2e/*.spec.ts) explicit_specs+=("${path#frontend/e2e/}") ;;
    frontend/e2e/*.ts) helper_paths+=("${path#frontend/e2e/}") ;;
    frontend/src/pages/*Page.tsx) page_change=1; pages+=("${path#frontend/src/pages/}") ;;
    frontend/src/i18n/locales/*|frontend/src/**/*.test.ts|frontend/src/**/*.test.tsx|frontend/src/*.test.ts|frontend/src/*.test.tsx) ;;
    frontend/playwright.config.ts) full_sweep=1 ;;
    frontend/src/*) full_sweep=1 ;;
  esac
done
# Helpers are exercised by every E2E spec that imports them.
for helper in "${helper_paths[@]}"; do
  import_path="${helper%.*}"
  while IFS= read -r spec; do
    [[ -n "$spec" ]] || continue
    grep -Eq "from[[:space:]]+['\"][^'\"]*${import_path//./\\.}(['\"]|\.)|import[[:space:]]+['\"][^'\"]*${import_path//./\\.}(['\"]|\.)" "$wt/frontend/e2e/$spec" && explicit_specs+=("$spec")
  done < <(cd "$wt/frontend/e2e" && printf '%s\n' *.spec.ts)
done
sweep=(desktop-board-visibility desktop-hand-visibility game-message-visibility hand-visibility mobile-viewport text-contrast duplicate-text)
declare -a run_specs=() run_args=()
(( ${#explicit_specs[@]} )) && mapfile -t explicit_specs < <(printf '%s\n' "${explicit_specs[@]}" | sort -u)
if (( frontend )); then
  # For page changes, the game spec runs without a grep; only the sweep is route-filtered.
  if (( page_change )); then
    for page in "${pages[@]}"; do
      page="${page%Page.tsx}"
      route="$(python3 - "$wt/frontend/src/constants/gameRoutes.ts" "$page" <<'PY'
import re, sys
text = open(sys.argv[1]).read()
for path, name in re.findall(r"\{\s*path:\s*['\"]([^'\"]+)['\"],.*?page:\s*['\"]([^'\"]+)['\"]", text, re.S):
    if name == sys.argv[2]: print(path); break
PY
)"
      [[ -n "$route" ]] || continue
      route="${route#/}"; [[ -n "$route" ]] || route=blackjack
      [[ -f "$wt/frontend/e2e/$route.spec.ts" ]] && run_specs+=("$route.spec.ts")
      if (( ! full_sweep )); then
        run_args+=("/$(printf '%s' "$route" | sed 's/[][(){}.^$*+?|\\]/\\&/g')\\b")
      fi
    done
  fi
  run_specs+=("${explicit_specs[@]}")
  if (( full_sweep )); then
    run_specs+=("${sweep[@]/%/.spec.ts}")
  elif (( page_change )); then
    run_specs+=("${sweep[@]/%/.spec.ts}")
  fi
fi
mapfile -t run_specs < <(printf '%s\n' "${run_specs[@]}" | sort -u | while IFS= read -r spec; do [[ -f "$wt/frontend/e2e/$spec" ]] && printf '%s\n' "$spec"; done)
if (( dry_run )); then
  run=0
  if (( page_change )); then
    page_specs=(); sweep_specs=()
    for spec in "${run_specs[@]}"; do
      case "$spec" in *.spec.ts) name="${spec%.spec.ts}"; if [[ " ${sweep[*]} " == *" $name "* ]]; then sweep_specs+=("e2e/$spec"); else page_specs+=("e2e/$spec"); fi ;;
      esac
    done
    if (( ${#page_specs[@]} )); then run=1; printf 'run %d: %s\n' "$run" "${page_specs[*]}"; fi
    if (( ${#sweep_specs[@]} )); then
      run=$((run+1))
      if (( ${#run_args[@]} )); then printf 'run %d: %s --grep %s\n' "$run" "${sweep_specs[*]}" "$(IFS='|'; echo "${run_args[*]}")"
      else printf 'run %d: %s\n' "$run" "${sweep_specs[*]}"; fi
    fi
  elif (( ${#run_specs[@]} )); then printf 'run 1: %s\n' "${run_specs[*]/#/e2e/}"; fi
  exit 0
fi
(( frontend )) || exit 0
if (( page_change )); then
  page_specs=(); sweep_specs=()
  for spec in "${run_specs[@]}"; do
    name="${spec%.spec.ts}"
    if [[ " ${sweep[*]} " == *" $name "* ]]; then sweep_specs+=("e2e/$spec"); else page_specs+=("e2e/$spec"); fi
  done
  args=(); for spec in "${page_specs[@]}"; do args+=("e2e/$spec"); done
  if (( ${#sweep_specs[@]} )); then
    (( ${#args[@]} )) && args+=(---)
    for spec in "${sweep_specs[@]}"; do args+=("e2e/$spec"); done
    (( ${#run_args[@]} )) && args+=(--grep "$(IFS='|'; echo "${run_args[*]}")")
  fi
  (( ${#args[@]} )) && "$B/e2e.sh" "$slot" "${args[@]}"
elif (( ${#run_specs[@]} )); then
  args=(); for spec in "${run_specs[@]}"; do args+=("e2e/$spec"); done
  "$B/e2e.sh" "$slot" "${args[@]}"
fi
