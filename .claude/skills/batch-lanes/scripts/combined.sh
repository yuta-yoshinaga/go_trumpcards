#!/usr/bin/env bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
set -euo pipefail

usage() { echo "usage: combined.sh [-r repeat] [-k] <spec>..." >&2; exit 2; }
repeat=3 keep=0
while getopts ':r:k' opt; do
  case "$opt" in
    r) repeat="$OPTARG"; [[ "$repeat" =~ ^[1-9][0-9]*$ ]] || usage ;;
    k) keep=1 ;;
    *) usage ;;
  esac
done
shift $((OPTIND - 1))
(( $# > 0 )) || usage
require_branch_re
# playwright.config.ts reuses any server already on :8080 outside CI, which would test that
# server's build instead of the combined one; refuse rather than report someone else's result.
if ss -ltnH 'sport = :8080' 2>/dev/null | grep -q .; then
  echo "PORT 8080 BUSY: stop that server so the combined build is the one under test" >&2; exit 1
fi
wt="$WT_ROOT/wt-combined"
cd "$REPO"
git fetch -q origin
if [[ -e "$wt" ]]; then
  [[ -d "$wt/.git" || -f "$wt/.git" ]] || { echo "NOT A WORKTREE $wt" >&2; exit 1; }
  [[ -z "$(git -C "$wt" status --porcelain)" ]] || { echo "DIRTY $wt" >&2; exit 1; }
  git -C "$wt" checkout -q --detach origin/develop
else
  git worktree add -q --detach "$wt" origin/develop
fi
cleanup() {
  if (( keep == 0 )) && { [[ -d "$wt" ]] || [[ -f "$wt/.git" ]]; }; then
    git -C "$REPO" worktree remove --force "$wt" >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT
merged=() conflicted=()
branches="$(gh pr list --state open --json headRefName -q '.[].headRefName')"
while IFS= read -r branch; do
  [[ -n "$branch" ]] || continue
  if [[ "$branch" =~ $BATCH_BRANCH_RE ]]; then
    if git -C "$wt" merge --no-edit "origin/$branch" >/dev/null 2>&1; then
      merged+=("$branch")
    else
      files="$(git -C "$wt" diff --name-only --diff-filter=U | paste -sd, -)"
      git -C "$wt" merge --abort >/dev/null 2>&1 || true
      conflicted+=("$branch")
      echo "CONFLICT $branch ${files:-unknown}" >&2
    fi
  fi
done <<< "$branches"
if [[ ! -d "$wt/frontend/node_modules" ]]; then (cd "$wt/frontend" && bun install --frozen-lockfile); fi
(cd "$wt/frontend" && bun run build)
set +e
(cd "$wt/frontend" && bunx playwright test "$@" "--repeat-each=$repeat" --reporter=line)
test_rc=$?
set -e
printf 'MERGED: %s\n' "${merged[*]:-none}"
printf 'CONFLICTED: %s\n' "${conflicted[*]:-none}"
if (( test_rc == 0 )); then echo 'TESTS: PASS'; else echo "TESTS: FAIL (exit $test_rc)"; fi
(( test_rc == 0 && ${#conflicted[@]} == 0 ))
