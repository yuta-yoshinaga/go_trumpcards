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
repeat="${E2E_SWEEP_REPEAT:-10}"
[[ "$repeat" =~ ^[0-9]+$ ]] || { echo "E2E_SWEEP_REPEAT must be numeric" >&2; exit 2; }
repeat_arg=()
if (( page_change && ! full_sweep && repeat > 1 )); then repeat_arg+=("--repeat-each=$repeat"); fi
mapfile -t deleted_routes < <(python3 - "$wt" <<'PYROUTES'
import re, subprocess, sys
root = sys.argv[1]
diffs = []
for args in (["diff", "origin/develop...HEAD"], ["diff", "HEAD"]):
    result = subprocess.run(["git", "-C", root, *args], text=True, stdout=subprocess.PIPE, stderr=subprocess.DEVNULL)
    if result.returncode == 0:
        diffs.append(result.stdout)
removed, added = set(), set()
for diff in diffs:
    relevant = False
    for line in diff.splitlines():
        if line.startswith("diff --git "):
            parts = line.split()
            relevant = len(parts) >= 4 and parts[2].startswith("a/frontend/e2e/") and parts[2].endswith(".ts")
            continue
        if not relevant or line.startswith("---") or not line or line[0] not in "+-":
            continue
        routes = re.findall(r"""['"](/[a-z0-9-]+)['"]""", line[1:])
        (added if line[0] == "+" else removed).update(routes)
untracked = subprocess.run(["git", "-C", root, "ls-files", "--others", "--exclude-standard", "--", "frontend/e2e"], text=True, stdout=subprocess.PIPE, stderr=subprocess.DEVNULL)
if untracked.returncode == 0:
    for path in untracked.stdout.splitlines():
        if not path.endswith(".ts"):
            continue
        try:
            with open(root + "/" + path, encoding="utf-8") as source:
                for line in source:
                    added.update(re.findall(r"""['"](/[a-z0-9-]+)['"]""", line))
        except OSError:
            pass
for route in sorted(removed - added):
    print(route)
PYROUTES
)
removed_route_args=()
for route in "${deleted_routes[@]}"; do
  removed_route_args+=("/$(printf '%s' "${route#/}" | sed 's/[][(){}.^$*+?|\\]/\\&/g')\\b")
done
removed_repeat_arg=()
if (( ${#removed_route_args[@]} && repeat > 1 )); then removed_repeat_arg+=("--repeat-each=$repeat"); fi
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
for path, name in re.findall(r"\{[^}]*?path:\s*['\"]([^'\"]+)['\"][^}]*?page:\s*['\"]([^'\"]+)['\"]", text, re.S):
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
  if (( full_sweep || page_change || ${#removed_route_args[@]} )); then
    run_specs+=("${sweep[@]/%/.spec.ts}")
  fi
fi
mapfile -t run_specs < <(printf '%s\n' "${run_specs[@]}" | sort -u | while IFS= read -r spec; do [[ -f "$wt/frontend/e2e/$spec" ]] && printf '%s\n' "$spec"; done)
page_specs=(); sweep_specs=(); other_specs=()
for spec in "${run_specs[@]}"; do
  name="${spec%.spec.ts}"
  if [[ " ${sweep[*]} " == *" $name "* ]]; then sweep_specs+=("e2e/$spec")
  elif (( page_change )) && [[ -f "$wt/frontend/src/pages/${name^}Page.tsx" ]]; then page_specs+=("e2e/$spec")
  else other_specs+=("e2e/$spec")
  fi
done
if (( dry_run )); then
  run=0
  if (( ${#page_specs[@]} )); then run=$((run+1)); printf 'run %d: %s\n' "$run" "${page_specs[*]}"; fi
  if (( ${#other_specs[@]} )); then run=$((run+1)); printf 'run %d: %s\n' "$run" "${other_specs[*]}"; fi
  if (( ${#sweep_specs[@]} )); then
    if (( full_sweep )); then
      run=$((run+1)); printf 'run %d: %s\n' "$run" "${sweep_specs[*]}"
      if (( ${#removed_route_args[@]} )); then run=$((run+1)); printf 'run %d: %s --grep %s --pass-with-no-tests%s\n' "$run" "${sweep_specs[*]}" "$(IFS='|'; echo "${removed_route_args[*]}")" "${removed_repeat_arg[*]:+ ${removed_repeat_arg[*]}}"; fi
    else
      all_route_args=("${run_args[@]}" "${removed_route_args[@]}")
      all_repeat_arg=("${repeat_arg[@]}")
      (( ${#removed_route_args[@]} && repeat > 1 )) && all_repeat_arg=("--repeat-each=$repeat")
      run=$((run+1))
      if (( ${#all_route_args[@]} )); then printf 'run %d: %s --grep %s --pass-with-no-tests%s\n' "$run" "${sweep_specs[*]}" "$(IFS='|'; echo "${all_route_args[*]}")" "${all_repeat_arg[*]:+ ${all_repeat_arg[*]}}"
      else printf 'run %d: %s\n' "$run" "${sweep_specs[*]}"; fi
    fi
  fi
  exit 0
fi
(( frontend )) || exit 0
if (( ${#page_specs[@]} || ${#sweep_specs[@]} || ${#other_specs[@]} )); then
  base_args=("${page_specs[@]}" "${other_specs[@]}")
  if (( ${#sweep_specs[@]} )); then
    if (( full_sweep )); then
      if (( ${#base_args[@]} )); then base_args+=(---); fi
      base_args+=("${sweep_specs[@]}")
      "$B/e2e.sh" "$slot" "${base_args[@]}"
      if (( ${#removed_route_args[@]} )); then "$B/e2e.sh" "$slot" "${sweep_specs[@]}" --grep "$(IFS='|'; echo "${removed_route_args[*]}")" --pass-with-no-tests "${removed_repeat_arg[@]}"; fi
      exit 0
    fi
    if (( ${#base_args[@]} )); then base_args+=(---); fi
    base_args+=("${sweep_specs[@]}")
    all_route_args=("${run_args[@]}" "${removed_route_args[@]}")
    all_repeat_arg=("${repeat_arg[@]}")
    (( ${#removed_route_args[@]} && repeat > 1 )) && all_repeat_arg=("--repeat-each=$repeat")
    if (( ${#all_route_args[@]} )); then base_args+=(--grep "$(IFS='|'; echo "${all_route_args[*]}")" --pass-with-no-tests "${all_repeat_arg[@]}"); fi
  fi
  "$B/e2e.sh" "$slot" "${base_args[@]}"
elif (( ${#removed_route_args[@]} )); then
  "$B/e2e.sh" "$slot" "${sweep[@]/%/.spec.ts}" --grep "$(IFS='|'; echo "${removed_route_args[*]}")" --pass-with-no-tests "${removed_repeat_arg[@]}"
fi
