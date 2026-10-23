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
mapfile -t deleted < <({ git -C "$wt" diff --name-only --diff-filter=D origin/develop...HEAD; git -C "$wt" diff --name-only --diff-filter=D; git -C "$wt" diff --cached --name-only --diff-filter=D; } | sort -u)
declare -A deleted_paths=()
for path in "${deleted[@]}"; do deleted_paths["$path"]=1; done
frontend=0; full_sweep=0; page_change=0
declare -a explicit_specs=() pages=() helper_paths=() skipped_paths=() game_pages=() selected_page_specs=()
declare -a unmapped_presenters=()
find_page() {
  local wanted="${1,,}" page_file base
  while IFS= read -r page_file; do
    base="${page_file##*/}"
    [[ "${base,,}" == "$wanted" ]] && { printf '%s\n' "$base"; return 0; }
  done < <(find "$wt/frontend/src/pages" -maxdepth 1 -type f -iname '*Page.tsx' | sort)
  return 1
}
for path in "${changed[@]}"; do
  if [[ "$path" =~ ^internal/adapter/presenter/([A-Za-z0-9]+)WebPresenter\.go$ ]]; then
    candidate="${BASH_REMATCH[1]}Page.tsx"
    matched_page="$(find_page "$candidate" || true)"
    if [[ -n "$matched_page" ]]; then pages+=("$matched_page"); page_change=1
    else unmapped_presenters+=("$path"); fi
    continue
  fi
  [[ "$path" == frontend/* ]] || continue
  frontend=1
  case "$path" in
    frontend/e2e/*.spec.ts) explicit_specs+=("${path#frontend/e2e/}") ;;
    frontend/e2e/*.ts) helper_paths+=("${path#frontend/e2e/}") ;;
    frontend/src/pages/*Page.tsx) page_change=1; pages+=("${path#frontend/src/pages/}") ;;
    frontend/src/i18n/locales/*|frontend/src/**/*.test.ts|frontend/src/**/*.test.tsx|frontend/src/*.test.ts|frontend/src/*.test.tsx) ;;
    frontend/playwright.config.ts) full_sweep=1 ;;
    frontend/src/components/*/*)
      component_dir="${path#frontend/src/components/}"; component_dir="${component_dir%%/*}"
      game="$(python3 - "$component_dir" <<'PY'
import re, sys
print(''.join(part[:1].upper() + part[1:] for part in re.split(r'[-_ ]+', sys.argv[1])))
PY
)"
      candidate="${game}Page.tsx"
      matched_page="$(find_page "$candidate" || true)"
      if [[ -n "$matched_page" ]]; then
        references=0
        while IFS= read -r page_file; do
          [[ -f "$page_file" ]] || continue
          if grep -Eq "from[[:space:]]+['\"][^'\"]*/${component_dir}/|import[[:space:]]+['\"][^'\"]*/${component_dir}/" "$page_file"; then
            page_basename="${page_file##*/}"
            [[ "${page_basename,,}" == "${matched_page,,}" ]] || references=1
          fi
        done < <(find "$wt/frontend/src/pages" -maxdepth 1 -name '*Page.tsx' -type f)
        if (( references )); then full_sweep=1
        else pages+=("$matched_page"); page_change=1; game_pages+=("$matched_page")
        fi
      else
        full_sweep=1
      fi
      ;;
    frontend/src/*)
      if [[ -n "${deleted_paths[$path]:-}" && ! -e "$wt/$path" ]]; then
        skipped_paths+=("$path (deleted source)")
      elif [[ "$path" == frontend/src/types/* ]]; then
        diff="$( { git -C "$wt" diff origin/develop...HEAD -- "$path"; git -C "$wt" diff HEAD -- "$path"; } )"
        if ! grep -Eq '^[+-][[:space:]]*export[[:space:]]+(const|enum|function|class|let)\b' <<<"$diff"; then
          skipped_paths+=("$path (type-only diff)")
        else
          full_sweep=1
        fi
      else
        full_sweep=1
      fi
      ;;
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
sweep=(desktop-board-visibility desktop-hand-visibility game-message-visibility mobile-viewport text-contrast route-sweep)
repeat="${E2E_SWEEP_REPEAT:-10}"
[[ "$repeat" =~ ^[0-9]+$ ]] || { echo "E2E_SWEEP_REPEAT must be numeric" >&2; exit 2; }
page_repeat="${E2E_PAGE_REPEAT:-3}"
[[ "$page_repeat" =~ ^[0-9]+$ ]] || { echo "E2E_PAGE_REPEAT must be numeric" >&2; exit 2; }
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
# A removed route repeats unless it is re-added in the same file.
deleted_by_file, added_by_file = {}, {}
for diff in diffs:
    relevant = False
    path = None
    for line in diff.splitlines():
        if line.startswith("diff --git "):
            parts = line.split()
            path = parts[2][2:] if len(parts) >= 4 else None
            relevant = path is not None and path.startswith("frontend/e2e/") and path.endswith(".ts")
            continue
        if not relevant or line.startswith("---") or not line or line[0] not in "+-":
            continue
        routes = re.findall(r"""['"](/[a-z0-9-]+)['"]""", line[1:])
        target = added_by_file if line[0] == "+" else deleted_by_file
        target.setdefault(path, set()).update(routes)
same_file_added = {route for path, routes in deleted_by_file.items() for route in routes & added_by_file.get(path, set())}
removed = set().union(*deleted_by_file.values()) if deleted_by_file else set()
for route in sorted(removed - same_file_added):
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
if (( page_change )); then frontend=1; fi
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
      if [[ -f "$wt/frontend/e2e/$route.spec.ts" ]]; then run_specs+=("$route.spec.ts"); selected_page_specs+=("$route.spec.ts"); fi
      for spec in "$wt/frontend/e2e/$route-"*.spec.ts; do
        if [[ -f "$spec" ]]; then run_specs+=("${spec##*/}"); selected_page_specs+=("${spec##*/}"); fi
      done
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
if (( full_sweep || page_change || ${#removed_route_args[@]} )); then
  for name in "${sweep[@]}"; do
    spec="$name.spec.ts"
    [[ -f "$wt/frontend/e2e/$spec" ]] || { echo "e2egate: sweep spec $spec does not exist" >&2; exit 2; }
  done
fi
mapfile -t run_specs < <(printf '%s\n' "${run_specs[@]}" | sort -u | while IFS= read -r spec; do [[ -f "$wt/frontend/e2e/$spec" ]] && printf '%s\n' "$spec"; done)
page_specs=(); sweep_specs=(); other_specs=()
for spec in "${run_specs[@]}"; do
  name="${spec%.spec.ts}"
  if [[ " ${sweep[*]} " == *" $name "* ]]; then sweep_specs+=("e2e/$spec")
  elif [[ " ${selected_page_specs[*]} " == *" $spec "* ]] || { (( page_change )) && [[ -f "$wt/frontend/src/pages/${name^}Page.tsx" ]]; }; then page_specs+=("e2e/$spec")
  else other_specs+=("e2e/$spec")
  fi
done
declare -a call_vars=()
add_call() {
  local var="call_args_${#call_vars[@]}"
  declare -g -a "$var"
  local -n args="$var"
  args=("$@")
  call_vars+=("$var")
}
if (( frontend )); then
  if (( ${#page_specs[@]} )) && (( page_repeat > 1 )); then add_call "${page_specs[@]}" "--repeat-each=$page_repeat"
  elif (( ${#page_specs[@]} )); then add_call "${page_specs[@]}"
  fi
  if (( ${#sweep_specs[@]} )); then
    base_args=("${other_specs[@]}")
    if (( full_sweep )); then
      if (( ${#base_args[@]} )); then base_args+=(---); fi
      base_args+=("${sweep_specs[@]}")
      add_call "${base_args[@]}"
      if (( ${#removed_route_args[@]} )); then
        add_call "${sweep_specs[@]}" --grep "$(IFS='|'; echo "${removed_route_args[*]}")" --pass-with-no-tests "${removed_repeat_arg[@]}"
      fi
    else
      if (( ${#base_args[@]} )); then base_args+=(---); fi
      base_args+=("${sweep_specs[@]}")
      all_route_args=("${run_args[@]}" "${removed_route_args[@]}")
      all_repeat_arg=("${repeat_arg[@]}")
      (( ${#removed_route_args[@]} && repeat > 1 )) && all_repeat_arg=("--repeat-each=$repeat")
      if (( ${#all_route_args[@]} )); then base_args+=(--grep "$(IFS='|'; echo "${all_route_args[*]}")" --pass-with-no-tests "${all_repeat_arg[@]}"); fi
      add_call "${base_args[@]}"
    fi
  elif (( ${#page_specs[@]} || ${#other_specs[@]} )); then
    ((${#other_specs[@]})) && add_call "${other_specs[@]}"
  fi
fi
if (( dry_run )); then
  for path in "${unmapped_presenters[@]}"; do printf 'NOTE e2egate: no page for %s\n' "$path"; done
  for skipped in "${skipped_paths[@]}"; do printf 'NOTE e2egate skipped full sweep for %s\n' "$skipped"; done
  run=0
  display_call() {
    local spec_name="$1" filters="$2"
    local -n display_specs="$spec_name"
    (( ${#display_specs[@]} )) || return 0
    run=$((run+1)); printf 'run %d: %s' "$run" "${display_specs[*]}"
    [[ -z "$filters" ]] || printf ' %s' "$filters"
    printf '\n'
  }
  for var in "${call_vars[@]}"; do
    declare -n local_args="$var"
    segment=(); page_display=(); other_display=(); sweep_display=(); filters=()
    for arg in "${local_args[@]}" ---; do
      if [[ "$arg" == --- ]]; then
        for spec in "${segment[@]}"; do
          if [[ " ${page_specs[*]} " == *" $spec "* ]]; then page_display+=("$spec")
          elif [[ " ${sweep_specs[*]} " == *" $spec "* ]]; then sweep_display+=("$spec")
          else other_display+=("$spec"); fi
        done
        segment=()
      elif [[ "$arg" == e2e/*.spec.ts ]]; then segment+=("$arg")
      else filters+=("$arg"); fi
    done
    if (( ${#sweep_display[@]} )); then
      display_call page_display ""
      display_call other_display ""
      display_call sweep_display "${filters[*]}"
    else
      display_call page_display "${filters[*]}"
      display_call other_display "${filters[*]}"
    fi
  done
  exit 0
fi
for path in "${unmapped_presenters[@]}"; do printf 'NOTE e2egate: no page for %s\n' "$path" >&2; done
for skipped in "${skipped_paths[@]}"; do printf 'NOTE e2egate skipped full sweep for %s\n' "$skipped" >&2; done
for var in "${call_vars[@]}"; do
  declare -n local_args="$var"
  "$B/e2e.sh" "$slot" "${local_args[@]}"
done
